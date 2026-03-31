package service

import (
	"errors"
	"fmt"
	"time"

	"paygo/src/config"
	"paygo/src/model"
)

type TransferService struct {
	authSvc *AuthService
}

func NewTransferService() *TransferService {
	return &TransferService{
		authSvc: NewAuthService(),
	}
}

// 创建转账
func (s *TransferService) CreateTransfer(uid uint, transferType, account, username string, money float64, desc string) (*model.Transfer, error) {
	// 获取商户
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		return nil, errors.New("商户不存在")
	}

	if user.Transfer != 1 {
		return nil, errors.New("商户没有转账权限")
	}

	// 检查转账方式是否开启
	cfgKey := "transfer_alipay"
	switch transferType {
	case "alipay":
		cfgKey = "transfer_alipay"
	case "wxpay":
		cfgKey = "transfer_wxpay"
	case "qqpay":
		cfgKey = "transfer_qqpay"
	case "bank":
		cfgKey = "transfer_bank"
	}

	enabled := s.authSvc.GetConfig(cfgKey)
	if enabled != "1" {
		return nil, errors.New("该转账方式未开启")
	}

	// 检查余额
	if user.Money < money {
		return nil, errors.New("余额不足")
	}

	// 生成转账单号
	bizNo := fmt.Sprintf("T%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	// 计算手续费（假设0.1%）
	costRate := 0.001
	costMoney := money * costRate

	tx := config.DB.Begin()

	// 扣除余额
	oldMoney := user.Money
	newMoney := oldMoney - money
	tx.Model(&user).Update("money", newMoney)

	// 创建转账记录
	transfer := &model.Transfer{
		BizNo:     bizNo,
		UID:       uid,
		Type:      transferType,
		Channel:   0, // TODO: 选择通道
		Account:   account,
		Username:  username,
		Money:     money,
		Costmoney: costMoney,
		Status:    0, // 处理中
		API:       1, // API发起
		Desc:      desc,
	}

	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建转账记录失败")
	}

	// 记录资金变动
	record := &model.Record{
		UID:      uid,
		Action:   3, // 转账
		Money:    -money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "transfer",
		TradeNo:  bizNo,
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	// TODO: 异步执行转账

	return transfer, nil
}

// 转账查询
func (s *TransferService) QueryTransfer(bizNo string) (*model.Transfer, error) {
	var transfer model.Transfer
	result := config.DB.Where("biz_no = ?", bizNo).First(&transfer)
	if result.Error != nil {
		return nil, errors.New("转账记录不存在")
	}
	return &transfer, nil
}

// 获取商户转账记录
func (s *TransferService) GetUserTransfers(uid uint, page, pageSize int) ([]model.Transfer, int64, error) {
	var transfers []model.Transfer
	var total int64

	query := config.DB.Model(&model.Transfer{}).Where("uid = ?", uid)
	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&transfers)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transfers, total, nil
}

// 更新转账状态
func (s *TransferService) UpdateTransferStatus(bizNo string, status int, result string) error {
	return config.DB.Model(&model.Transfer{}).Where("biz_no = ?", bizNo).
		Updates(map[string]interface{}{
			"status": status,
			"result": result,
		}).Error
}

// 余额查询
func (s *TransferService) QueryBalance(uid uint) (float64, error) {
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		return 0, errors.New("商户不存在")
	}
	return user.Money, nil
}

// 转账退款（人工退款）
func (s *TransferService) RefundTransfer(bizNo string) error {
	var transfer model.Transfer
	if config.DB.First(&transfer, "biz_no = ?", bizNo).Error != nil {
		return errors.New("转账记录不存在")
	}

	if transfer.Status != 1 { // 只有成功的才能退款
		return errors.New("状态不允许退款")
	}

	// 退还余额给商户
	tx := config.DB.Begin()

	var user model.User
	tx.First(&user, transfer.UID)

	oldMoney := user.Money
	newMoney := oldMoney + transfer.Money
	tx.Model(&user).Update("money", newMoney)

	// 更新转账状态
	tx.Model(&transfer).Updates(map[string]interface{}{
		"status": 3, // 已退款
	})

	// 记录资金变动
	record := &model.Record{
		UID:      transfer.UID,
		Action:   9, // 转账退款
		Money:    transfer.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "transfer_refund",
		TradeNo:  bizNo,
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return nil
}

// 获取转账记录详情
func (s *TransferService) GetTransferDetail(bizNo string) (map[string]interface{}, error) {
	var transfer model.Transfer
	if config.DB.First(&transfer, "biz_no = ?", bizNo).Error != nil {
		return nil, errors.New("转账记录不存在")
	}

	var user model.User
	config.DB.First(&user, transfer.UID)

	return map[string]interface{}{
		"transfer": transfer,
		"user":     user.Username,
	}, nil
}

// 获取商户资金记录
func (s *TransferService) GetUserRecords(uid uint, action int, page, pageSize int) ([]model.Record, int64, error) {
	var records []model.Record
	var total int64

	query := config.DB.Model(&model.Record{}).Where("uid = ?", uid)
	if action >= 0 {
		query = query.Where("action = ?", action)
	}

	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return records, total, nil
}

// 管理员加款/扣款
func (s *TransferService) AdminChangeMoney(uid uint, money float64, typ string, remark string) error {
	var user model.User
	if config.DB.First(&user, uid).Error != nil {
		return errors.New("商户不存在")
	}

	oldMoney := user.Money
	newMoney := oldMoney + money

	if newMoney < 0 {
		return errors.New("余额不能为负数")
	}

	tx := config.DB.Begin()

	tx.Model(&user).Update("money", newMoney)

	record := &model.Record{
		UID:      uid,
		Money:    money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     typ,
		Date:     time.Now(),
	}

	if typ == "admin_add" {
		record.Action = 5 // 后台加款
	} else if typ == "admin_sub" {
		record.Action = 6 // 后台扣款
	}

	tx.Create(record)

	tx.Commit()

	return nil
}
