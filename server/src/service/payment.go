package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
)

type PaymentService struct {
	orderSvc *OrderService
	authSvc  *AuthService
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		orderSvc: NewOrderService(),
		authSvc:  NewAuthService(),
	}
}

// 支付参数
type SubmitParams struct {
	UID       uint
	OutTradeNo string
	Type      int    // 支付类型ID
	ChannelID int    // 通道ID（submit2指定通道）
	Name      string
	Money     float64
	NotifyURL string
	ReturnURL string
	Param     string
	IP        string
	Device    string // pc/mobile
	Method    string // web/jump/jsapi/scan
}

// 获取可用支付方式
func (s *PaymentService) GetAvailableTypes(uid uint) ([]model.PayType, error) {
	// 获取商户的用户组
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		return nil, err
	}

	// 获取用户组配置
	var group model.Group
	config.DB.First(&group, user.GID)

	var availableTypes []model.PayType

	// 如果有自定义通道配置
	if group.Settings != "" {
		var settings map[string]interface{}
		json.Unmarshal([]byte(group.Settings), &settings)
		if types, ok := settings["types"]; ok {
			typeIDs := types.([]interface{})
			for _, t := range typeIDs {
				var payType model.PayType
				if config.DB.First(&payType, int(t.(float64))).Error == nil {
					availableTypes = append(availableTypes, payType)
				}
			}
			return availableTypes, nil
		}
	}

	// 默认返回所有开启的支付类型
	config.DB.Where("status = 1").Find(&availableTypes)
	return availableTypes, nil
}

// 获取可用通道
func (s *PaymentService) GetAvailableChannels(uid uint, typeID int) ([]model.Channel, error) {
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		return nil, err
	}

	// 获取用户组
	var group model.Group
	config.DB.First(&group, user.GID)

	// 获取通道列表
	var channels []model.Channel
	query := config.DB.Where("type = ? AND status = 1", typeID)

	// 如果有自定义配置
	if group.Config != "" {
		var groupConfig map[string]interface{}
		json.Unmarshal([]byte(group.Config), &groupConfig)
		if channelIDs, ok := groupConfig["channels"]; ok {
			ids := channelIDs.([]interface{})
			idStrs := make([]string, len(ids))
			for i, v := range ids {
				idStrs[i] = strconv.Itoa(int(v.(float64)))
			}
			query = query.Where("id IN (?)", idStrs)
		}
	}

	query.Find(&channels)
	return channels, nil
}

// 选择通道（轮询或指定）
func (s *PaymentService) SelectChannel(uid uint, typeID int, channelID int) (*model.Channel, error) {
	if channelID > 0 {
		// 指定通道
		var channel model.Channel
		result := config.DB.Where("id = ? AND type = ? AND status = 1", channelID, typeID).First(&channel)
		if result.Error != nil {
			return nil, errors.New("指定通道不存在或已关闭")
		}
		return &channel, nil
	}

	// 轮询选择
	var roll model.Roll
	result := config.DB.Where("type = ? AND status = 1", typeID).First(&roll)
	if result.Error != nil {
		// 没有轮询配置，直接查询可用通道
		var channel model.Channel
		r := config.DB.Where("type = ? AND status = 1", typeID).First(&channel)
		if r.Error != nil {
			return nil, errors.New("没有可用的支付通道")
		}
		return &channel, nil
	}

	// 解析轮询配置
	var rollInfo struct {
		Channels []int `json:"channels"`
		Weights  []int `json:"weights"`
	}
	json.Unmarshal([]byte(roll.Info), &rollInfo)

	if len(rollInfo.Channels) == 0 {
		return nil, errors.New("轮询配置错误")
	}

	// 根据权重随机选择
	if roll.Kind == 1 && len(rollInfo.Weights) > 0 {
		totalWeight := 0
		for _, w := range rollInfo.Weights {
			totalWeight += w
		}
		rand := time.Now().UnixNano() % int64(totalWeight)
		curWeight := 0
		for i, w := range rollInfo.Weights {
			curWeight += w
			if int64(curWeight) >= rand {
				channelID = rollInfo.Channels[i]
				break
			}
		}
	} else {
		// 简单轮询
		currentIndex := roll.Index
		channelID = rollInfo.Channels[currentIndex%len(rollInfo.Channels)]

		// 更新索引
		config.DB.Model(&roll).Update("index", (currentIndex+1)%len(rollInfo.Channels))
	}

	var channel model.Channel
	if config.DB.First(&channel, channelID).Error != nil {
		return nil, errors.New("通道不存在")
	}

	return &channel, nil
}

// 提交支付
func (s *PaymentService) SubmitPayment(params SubmitParams) (map[string]interface{}, error) {
	// 获取商户
	user, err := s.authSvc.GetUser(params.UID)
	if err != nil {
		return nil, errors.New("商户不存在")
	}

	if user.Status != 0 {
		return nil, errors.New("商户已被禁用")
	}

	if user.Pay != 1 {
		return nil, errors.New("商户没有支付权限")
	}

	// 选择通道
	channel, err := s.SelectChannel(params.UID, params.Type, params.ChannelID)
	if err != nil {
		return nil, err
	}

	// 检查金额限制
	if channel.Paymin != "" {
		minMoney, _ := strconv.ParseFloat(channel.Paymin, 10)
		if params.Money < minMoney {
			return nil, errors.New("最低支付金额" + channel.Paymin)
		}
	}
	if channel.Paymax != "" {
		maxMoney, _ := strconv.ParseFloat(channel.Paymax, 10)
		if params.Money > maxMoney {
			return nil, errors.New("最高支付金额" + channel.Paymax)
		}
	}

	// 检查域名授权
	domain := strings.Split(params.Param, "|")[0]
	if domain != "" && !s.orderSvc.CheckDomainAuth(params.UID, domain) {
		// 检查是否在禁止名单中
		if s.orderSvc.IsBlacklisted(params.IP) {
			return nil, errors.New("IP禁止访问")
		}
	}

	// 计算费率
	rate := channel.Rate
	if user.Mode == 1 {
		rate = rate + (100-rate)*0.5
	} else if user.Mode == 2 {
		rate = rate * 0.5
	}

	// 计算金额
	getmoney := params.Money * rate / 100
	profitmoney := params.Money - getmoney
	costrate := channel.Costrate
	if costrate == 0 {
		costrate = rate
	}
	realmoney := params.Money * costrate / 100

	// 创建订单
	tradeNo := s.orderSvc.GenTradeNo()

	order := &model.Order{
		TradeNo:     tradeNo,
		OutTradeNo:  params.OutTradeNo,
		UID:         params.UID,
		Type:        params.Type,
		Channel:     int(channel.ID),
		Name:        params.Name,
		Money:       params.Money,
		Realmoney:   realmoney,
		Getmoney:    getmoney,
		Profitmoney: profitmoney,
		NotifyURL:   params.NotifyURL,
		ReturnURL:   params.ReturnURL,
		Param:       params.Param,
		Addtime:     time.Now(),
		Date:        time.Now().Format("2006-01-02"),
		IP:          params.IP,
		Status:      model.OrderStatusPending,
		Notify:      0,
	}

	if err := config.DB.Create(order).Error; err != nil {
		return nil, errors.New("创建订单失败")
	}

	// 加载插件并提交
	pluginName := channel.Plugin
	pluginHandler := plugin.GetHandler(pluginName)
	if pluginHandler == nil {
		return nil, errors.New("支付通道插件不存在")
	}

	// 构造插件参数
	pluginParams := map[string]interface{}{
		"trade_no":   tradeNo,
		"out_trade_no": params.OutTradeNo,
		"money":       params.Money,
		"name":        params.Name,
		"notify_url":  params.NotifyURL,
		"return_url":  params.ReturnURL,
		"param":       params.Param,
		"ip":          params.IP,
		"device":      params.Device,
		"method":      params.Method,
		"channel":     channel,
	}

	// 调用插件提交
	result, err := pluginHandler.Submit(pluginParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"trade_no": tradeNo,
		"result":   result,
		"order":    order,
	}, nil
}

// 扫码支付提交
func (s *PaymentService) SubmitScanPayment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "scan"
	return s.SubmitPayment(params)
}

// JSAPI支付
func (s *PaymentService) SubmitJSAPIPayment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "jsapi"
	return s.SubmitPayment(params)
}

// APP支付
func (s *PaymentService) SubmitAppPayment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "app"
	return s.SubmitPayment(params)
}

// H5支付
func (s *PaymentService) SubmitH5Payment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "wap"
	return s.SubmitPayment(params)
}

// 支付回调处理
func (s *PaymentService) HandleNotify(tradeNo string, pluginName string) (map[string]interface{}, error) {
	pluginHandler := plugin.GetHandler(pluginName)
	if pluginHandler == nil {
		return nil, errors.New("插件不存在")
	}

	// 调用插件处理回调
	result, err := pluginHandler.Notify(tradeNo)
	if err != nil {
		return nil, err
	}

	if result.Success {
		// 更新订单状态
		err = s.orderSvc.OrderPaid(result.TradeNo, result.APITradeNo, result.Buyer)
		if err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"success": result.Success,
		"message": result.Message,
	}, nil
}

// 同步回调处理
func (s *PaymentService) HandleReturn(tradeNo string, pluginName string) (plugin.ReturnResult, error) {
	pluginHandler := plugin.GetHandler(pluginName)
	if pluginHandler == nil {
		return plugin.ReturnResult{}, errors.New("插件不存在")
	}

	return pluginHandler.Return(tradeNo)
}

// 退款
func (s *PaymentService) Refund(tradeNo string, money float64) error {
	order, err := s.orderSvc.GetOrder(tradeNo)
	if err != nil {
		return err
	}

	// 获取通道的插件名称
	var channel model.Channel
	if config.DB.First(&channel, order.Channel).Error != nil {
		return errors.New("通道不存在")
	}

	pluginHandler := plugin.GetHandler(channel.Plugin)
	if pluginHandler == nil {
		return errors.New("插件不存在")
	}

	_, err = pluginHandler.Refund(map[string]interface{}{
		"trade_no": tradeNo,
		"money":    money,
	})
	return err
}

// 获取通道配置
func (s *PaymentService) GetChannelConfig(channelID int) (*model.Channel, error) {
	var channel model.Channel
	result := config.DB.First(&channel, channelID)
	if result.Error != nil {
		return nil, errors.New("通道不存在")
	}
	return &channel, nil
}

// 获取插件列表
func (s *PaymentService) GetPluginList() ([]model.Plugin, error) {
	var plugins []model.Plugin
	config.DB.Find(&plugins)
	return plugins, nil
}

// 获取通道列表
func (s *PaymentService) GetChannelList(typeID int) ([]model.Channel, error) {
	var channels []model.Channel
	query := config.DB.Where("status = 1")
	if typeID > 0 {
		query = query.Where("type = ?", typeID)
	}
	query.Find(&channels)
	return channels, nil
}

// 手动补单
func (s *PaymentService) ManualFillOrder(tradeNo, apiTradeNo, buyer string) error {
	return s.orderSvc.OrderPaid(tradeNo, apiTradeNo, buyer)
}

// config.Now() 需要定义
