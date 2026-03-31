package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"paygo/src/config"
	"paygo/src/model"
)

type SettleService struct {
	authSvc *AuthService
}

func NewSettleService() *SettleService {
	return &SettleService{
		authSvc: NewAuthService(),
	}
}

// 申请结算
func (s *SettleService) ApplySettle(uid uint, account, username string, money float64, settleType int) (*model.Settle, error) {
	// 获取商户
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		return nil, errors.New("商户不存在")
	}

	if user.Settle != 1 {
		return nil, errors.New("商户没有结算权限")
	}

	// 检查结算方式是否开启
	cfgKey := "settle_alipay"
	switch settleType {
	case 1:
		cfgKey = "settle_alipay"
	case 2:
		cfgKey = "settle_wxpay"
	case 3:
		cfgKey = "settle_qqpay"
	case 4:
		cfgKey = "settle_bank"
	}

	enabled := s.authSvc.GetConfig(cfgKey)
	if enabled != "1" {
		return nil, errors.New("该结算方式未开启")
	}

	// 检查最低结算限额
	minMoney, _ := strconv.ParseFloat(s.authSvc.GetConfig("settle_money"), 10)
	if money < minMoney {
		return nil, fmt.Errorf("最低结算金额%.2f元", minMoney)
	}

	// 检查余额
	if user.Money < money {
		return nil, errors.New("余额不足")
	}

	// 计算实际到账金额（扣除手续费）
	rate := 0.0 // 默认0手续费
	settleRateStr := s.authSvc.GetConfig("settle_rate_" + strconv.Itoa(settleType))
	if settleRateStr != "" {
		rate, _ = strconv.ParseFloat(settleRateStr, 10)
	}
	realMoney := money * (1 - rate/100)

	tx := config.DB.Begin()

	// 扣除余额
	oldMoney := user.Money
	newMoney := oldMoney - money
	tx.Model(&user).Update("money", newMoney)

	// 创建结算记录
	settle := &model.Settle{
		UID:      uid,
		Auto:     1,
		Type:     settleType,
		Account:  account,
		Username: username,
		Money:    money,
		Realmoney: realMoney,
		Addtime:  time.Now(),
		Status:   model.SettleStatusPending,
	}

	if err := tx.Create(settle).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建结算记录失败")
	}

	// 记录资金变动
	record := &model.Record{
		UID:      uid,
		Action:   2, // 结算扣款
		Money:    -money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return settle, nil
}

// 获取商户结算记录
func (s *SettleService) GetUserSettles(uid uint, page, pageSize int) ([]model.Settle, int64, error) {
	var settles []model.Settle
	var total int64

	query := config.DB.Model(&model.Settle{}).Where("uid = ?", uid)
	query.Count(&total)

	result := query.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&settles)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return settles, total, nil
}

// 获取结算记录详情
func (s *SettleService) GetSettle(id uint) (*model.Settle, error) {
	var settle model.Settle
	result := config.DB.First(&settle, id)
	if result.Error != nil {
		return nil, errors.New("结算记录不存在")
	}
	return &settle, nil
}

// 同意结算
func (s *SettleService) ApproveSettle(id uint) error {
	var settle model.Settle
	result := config.DB.First(&settle, id)
	if result.Error != nil {
		return errors.New("结算记录不存在")
	}

	if settle.Status != model.SettleStatusPending {
		return errors.New("状态不允许操作")
	}

	tx := config.DB.Begin()

	// 更新状态
	tx.Model(&settle).Updates(map[string]interface{}{
		"status":    model.SettleStatusCompleted,
		"endtime":   time.Now(),
		"result":    "已同意",
	})

	tx.Commit()

	// TODO: 调用支付通道进行转账

	return nil
}

// 拒绝结算
func (s *SettleService) RejectSettle(id uint, reason string) error {
	var settle model.Settle
	result := config.DB.First(&settle, id)
	if result.Error != nil {
		return errors.New("结算记录不存在")
	}

	if settle.Status != model.SettleStatusPending {
		return errors.New("状态不允许操作")
	}

	tx := config.DB.Begin()

	// 退还余额给商户
	var user model.User
	tx.First(&user, settle.UID)

	oldMoney := user.Money
	newMoney := oldMoney + settle.Money
	tx.Model(&user).Update("money", newMoney)

	// 更新结算状态
	tx.Model(&settle).Updates(map[string]interface{}{
		"status":  model.SettleStatusFailed,
		"endtime": time.Now(),
		"result":  "拒绝: " + reason,
	})

	// 记录资金变动（退还）
	record := &model.Record{
		UID:      settle.UID,
		Action:   8, // 结算失败返还
		Money:    settle.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return nil
}

// 生成批量结算批次
func (s *SettleService) CreateBatch(settleIDs []uint) (*model.Batch, []model.Settle, error) {
	var settles []model.Settle
	config.DB.Where("id IN ? AND status = ?", settleIDs, model.SettleStatusPending).Find(&settles)
	if len(settles) == 0 {
		return nil, nil, errors.New("没有待处理的结算记录")
	}

	// 计算总金额
	var totalMoney float64
	for _, s := range settles {
		totalMoney += s.Money
	}

	// 生成批次号
	batchNo := fmt.Sprintf("B%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)

	batch := &model.Batch{
		Batch:   batchNo,
		Allmoney: totalMoney,
		Count:   len(settles),
		Time:    time.Now(),
		Status:  0,
	}

	tx := config.DB.Begin()

	// 创建批次
	tx.Create(batch)

	// 更新结算记录的批次号
	for _, settle := range settles {
		tx.Model(&settle).Update("batch", batchNo)
	}

	tx.Commit()

	return batch, settles, nil
}

// 执行批量转账
func (s *SettleService) ExecuteBatchTransfer(batchNo string) error {
	var batch model.Batch
	if config.DB.First(&batch, "batch = ?", batchNo).Error != nil {
		return errors.New("批次不存在")
	}

	if batch.Status == 1 {
		return errors.New("批次已处理")
	}

	var settles []model.Settle
	config.DB.Where("batch = ?", batchNo).Find(&settles)

	tx := config.DB.Begin()

	// 更新批次状态
	tx.Model(&batch).Update("status", 1)

	// 更新每条结算记录
	for _, settle := range settles {
		tx.Model(&settle).Updates(map[string]interface{}{
			"status":           model.SettleStatusCompleted,
			"endtime":          time.Now(),
			"transfer_status":  1,
			"transfer_date":    time.Now(),
		})
	}

	tx.Commit()

	// TODO: 调用支付通道进行实际转账

	return nil
}

// 获取所有待结算记录
func (s *SettleService) GetPendingSettles() ([]model.Settle, error) {
	var settles []model.Settle
	result := config.DB.Where("status = ?", model.SettleStatusPending).Find(&settles)
	if result.Error != nil {
		return nil, result.Error
	}
	return settles, nil
}
