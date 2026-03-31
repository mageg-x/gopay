package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"paygo/src/config"
	"paygo/src/model"
)

type OrderService struct {
	authSvc *AuthService
}

func NewOrderService() *OrderService {
	return &OrderService{
		authSvc: NewAuthService(),
	}
}

// 生成订单号
func (s *OrderService) GenTradeNo() string {
	now := time.Now()
	// 格式: YYYYMMDDHHMMSS + 6位随机数
	randNum := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	return fmt.Sprintf("%s%s", now.Format("20060102150405"), randNum)
}

// 创建订单
func (s *OrderService) CreateOrder(uid uint, outTradeNo, name, notifyURL, returnURL, param string,
	money float64, payType int, channelID int, ip string) (*model.Order, error) {

	// 获取商户信息
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		return nil, errors.New("商户不存在")
	}

	if user.Status != 0 {
		return nil, errors.New("商户已被禁用")
	}

	if user.Pay != 1 {
		return nil, errors.New("商户没有支付权限")
	}

	// 获取通道信息
	var channel model.Channel
	result := config.DB.First(&channel, channelID)
	if result.Error != nil {
		return nil, errors.New("通道不存在")
	}

	if channel.Status != 1 {
		return nil, errors.New("通道已关闭")
	}

	// 检查金额限制
	if channel.Paymin != "" {
		minMoney, _ := strconv.ParseFloat(channel.Paymin, 10)
		if money < minMoney {
			return nil, fmt.Errorf("最低支付金额%.2f", minMoney)
		}
	}
	if channel.Paymax != "" {
		maxMoney, _ := strconv.ParseFloat(channel.Paymax, 10)
		if money > maxMoney {
			return nil, fmt.Errorf("最高支付金额%.2f", maxMoney)
		}
	}

	// 检查用户组费率
	var group model.Group
	config.DB.First(&group, user.GID)

	// 计算实际金额和商户可得
	// 费率 = 通道费率 + 用户组加成
	rate := channel.Rate
	if user.Mode == 1 {
		// 加费模式
		rate = rate + (100 - rate) * 0.5
	} else if user.Mode == 2 {
		// 减费模式
		rate = rate * 0.5
	}

	// 计算商户可得
	getmoney := money * rate / 100
	profitmoney := money - getmoney

	// 平台实际收入
	realmoney := money * channel.Costrate / 100

	order := &model.Order{
		TradeNo:     s.GenTradeNo(),
		OutTradeNo:  outTradeNo,
		UID:         uid,
		Tid:         0,
		Type:        payType,
		Channel:     channelID,
		Name:        name,
		Money:       money,
		Realmoney:   realmoney,
		Getmoney:    getmoney,
		Profitmoney: profitmoney,
		NotifyURL:   notifyURL,
		ReturnURL:   returnURL,
		Param:       param,
		Addtime:     time.Now(),
		Date:        time.Now().Format("2006-01-02"),
		IP:          ip,
		Status:      model.OrderStatusPending,
		Notify:      0,
		Invite:      user.Upid,
		Subchannel:  0,
		Version:     0,
	}

	// 如果有扩展信息
	if param != "" {
		order.Param = param
	}

	result = config.DB.Create(order)
	if result.Error != nil {
		return nil, errors.New("创建订单失败")
	}

	return order, nil
}

// 订单支付成功回调
func (s *OrderService) OrderPaid(tradeNo, apiTradeNo, buyer string) error {
	var order model.Order
	result := config.DB.Where("trade_no = ? AND status = ?", tradeNo, model.OrderStatusPending).First(&order)
	if result.Error != nil {
		return errors.New("订单不存在或已处理")
	}

	// 更新订单状态
	now := time.Now()
	config.DB.Model(&order).Updates(map[string]interface{}{
		"status":      model.OrderStatusPaid,
		"api_trade_no": apiTradeNo,
		"buyer":        buyer,
		"endtime":      now,
		"notifytime":   now,
	})

	// 给商户加款
	var user model.User
	config.DB.First(&user, order.UID)

	oldMoney := user.Money
	newMoney := oldMoney + order.Getmoney

	config.DB.Model(&user).Update("money", newMoney)

	// 记录资金变动
	record := &model.Record{
		UID:      order.UID,
		Action:   1, // 订单收入
		Money:    order.Getmoney,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "order",
		TradeNo:  tradeNo,
		Date:     now,
	}
	config.DB.Create(record)

	// 邀请人奖励
	if order.Invite > 0 {
		s.addInviteMoney(order.Invite, order.UID, order.Money, tradeNo)
	}

	// 通知商户
	go s.notifyMerchant(order)

	return nil
}

// 添加邀请奖励
func (s *OrderService) addInviteMoney(inviteUID, uid uint, money float64, tradeNo string) {
	// 获取邀请人信息
	var inviteUser model.User
	if config.DB.First(&inviteUser, inviteUID).Error != nil {
		return
	}

	// 获取用户组配置
	var group model.Group
	if config.DB.First(&group, inviteUser.GID).Error != nil {
		return
	}

	// 计算奖励比例（默认1%）
	rate := 0.01
	if group.Settings != "" {
		var settings map[string]interface{}
		json.Unmarshal([]byte(group.Settings), &settings)
		if v, ok := settings["invite_rate"]; ok {
			rate, _ = strconv.ParseFloat(fmt.Sprintf("%v", v), 10)
		}
	}

	reward := money * rate

	oldMoney := inviteUser.Money
	newMoney := oldMoney + reward

	config.DB.Model(&inviteUser).Update("money", newMoney)

	// 记录
	record := &model.Record{
		UID:      inviteUID,
		Action:   7, // 邀请返现
		Money:    reward,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "invite",
		TradeNo:  tradeNo,
		Date:     time.Now(),
	}
	config.DB.Create(record)
}

// 通知商户
func (s *OrderService) notifyMerchant(order model.Order) {
	if order.NotifyURL == "" {
		return
	}

	// TODO: 发送HTTP通知
	// 构造通知数据
	params := map[string]string{
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"type":         strconv.Itoa(order.Type),
		"status":       "1",
	}

	// 获取商户密钥进行签名
	var user model.User
	config.DB.First(&user, order.UID)

	params["sign"] = s.authSvc.MakeSign(params, user.Key)

	// 发送通知
	// http.PostForm(order.NotifyURL, params)
}

// 订单退款
func (s *OrderService) Refund(tradeNo string, money float64) error {
	var order model.Order
	result := config.DB.Where("trade_no = ?", tradeNo).First(&order)
	if result.Error != nil {
		return errors.New("订单不存在")
	}

	if order.Status != model.OrderStatusPaid {
		return errors.New("订单状态不允许退款")
	}

	if order.Refundmoney+money > order.Getmoney {
		return errors.New("退款金额超过可退金额")
	}

	// 扣除商户余额
	var user model.User
	config.DB.First(&user, order.UID)

	if user.Money < money {
		return errors.New("商户余额不足")
	}

	oldMoney := user.Money
	newMoney := oldMoney - money

	tx := config.DB.Begin()

	// 扣除余额
	tx.Model(&user).Update("money", newMoney)

	// 更新退款金额
	refundmoney := order.Refundmoney + money
	tx.Model(&order).Update("refundmoney", refundmoney)

	// 如果完全退款，更新状态
	if refundmoney >= order.Getmoney {
		tx.Model(&order).Update("status", model.OrderStatusRefunded)
	}

	// 记录资金变动
	record := &model.Record{
		UID:      order.UID,
		Action:   4, // 退款
		Money:    -money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "refund",
		TradeNo:  tradeNo,
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return nil
}

// 冻结订单
func (s *OrderService) Freeze(tradeNo string) error {
	return config.DB.Model(&model.Order{}).Where("trade_no = ?", tradeNo).
		Update("status", model.OrderStatusFrozen).Error
}

// 解冻订单
func (s *OrderService) Unfreeze(tradeNo string) error {
	return config.DB.Model(&model.Order{}).Where("trade_no = ? AND status = ?", tradeNo, model.OrderStatusFrozen).
		Update("status", model.OrderStatusPaid).Error
}

// 查询订单
func (s *OrderService) GetOrder(tradeNo string) (*model.Order, error) {
	var order model.Order
	result := config.DB.Where("trade_no = ?", tradeNo).First(&order)
	if result.Error != nil {
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}

// 按商户订单号查询
func (s *OrderService) GetOrderByOutTradeNo(outTradeNo string, uid uint) (*model.Order, error) {
	var order model.Order
	result := config.DB.Where("out_trade_no = ? AND uid = ?", outTradeNo, uid).First(&order)
	if result.Error != nil {
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}

// 获取商户订单列表
func (s *OrderService) GetUserOrders(uid uint, status int, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := config.DB.Model(&model.Order{}).Where("uid = ?", uid)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return orders, total, nil
}

// 获取订单类型名称
func (s *OrderService) GetTypeName(typeID int) string {
	var payType model.PayType
	if config.DB.First(&payType, typeID).Error == nil {
		return payType.Showname
	}
	return "未知"
}

// 检查订单是否超时（未支付订单超过30分钟）
func (s *OrderService) IsOrderTimeout(order model.Order) bool {
	if order.Status != model.OrderStatusPending {
		return false
	}
	return time.Since(order.Addtime) > 30*time.Minute
}

// 删除超时未支付订单
func (s *OrderService) CleanTimeoutOrders() (int64, error) {
	timeout := time.Now().Add(-30 * time.Minute)
	result := config.DB.Where("status = ? AND addtime < ?", model.OrderStatusPending, timeout).Delete(&model.Order{})
	return result.RowsAffected, result.Error
}

// 获取订单统计
func (s *OrderService) GetOrderStats(uid uint, startDate, endDate string) (map[string]interface{}, error) {
	type Stats struct {
		Total   float64
		Count   int64
		Today   float64
		TodayCount int64
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 总计
	var totalMoney float64
	var totalCount int64

	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ?", uid, model.OrderStatusPaid).Select("COALESCE(SUM(money), 0)").Scan(&totalMoney)
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ?", uid, model.OrderStatusPaid).Count(&totalCount)

	var todayMoney float64
	var todayCount int64
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ? AND date = ?", uid, model.OrderStatusPaid, today).Select("COALESCE(SUM(money), 0)").Scan(&todayMoney)
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ? AND date = ?", uid, model.OrderStatusPaid, today).Count(&todayCount)

	return map[string]interface{}{
		"total_money":    totalMoney,
		"total_count":    totalCount,
		"today_money":    todayMoney,
		"today_count":    todayCount,
	}, nil
}

// 检查黑名单
func (s *OrderService) IsBlacklisted(ip string) bool {
	var count int64
	config.DB.Model(&model.Blacklist{}).Where("type = 0 AND content = ?", ip).Count(&count)
	return count > 0
}

// 检查域名授权
func (s *OrderService) CheckDomainAuth(uid uint, domain string) bool {
	if domain == "" {
		return true
	}

	var count int64
	config.DB.Model(&model.Domain{}).Where("uid = ? AND domain = ? AND status = 1", uid, domain).Count(&count)
	return count > 0
}

// 获取订单详情（包含关联信息）
func (s *OrderService) GetOrderDetail(tradeNo string) (map[string]interface{}, error) {
	var order model.Order
	if config.DB.First(&order, tradeNo).Error != nil {
		return nil, errors.New("订单不存在")
	}

	// 获取商户信息
	var user model.User
	config.DB.First(&user, order.UID)

	// 获取通道信息
	var channel model.Channel
	config.DB.First(&channel, order.Channel)

	// 获取支付类型
	var payType model.PayType
	config.DB.First(&payType, order.Type)

	detail := map[string]interface{}{
		"order":       order,
		"user":        user.Username,
		"channel":     channel.Name,
		"typename":   payType.Showname,
	}

	return detail, nil
}
