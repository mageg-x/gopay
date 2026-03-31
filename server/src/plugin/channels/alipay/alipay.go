package alipay

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
)

// 支付宝插件
type AlipayPlugin struct {
	plugin.BasePlugin
}

func New() plugin.Plugin {
	return &AlipayPlugin{}
}

func init() {
	plugin.Register("alipay", New)
}

func (p *AlipayPlugin) GetInfo() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:       "alipay",
		Showname:   "支付宝官方支付",
		Author:     "支付宝",
		Link:       "https://www.alipay.com/",
		Types:      []string{"alipay"},
		Transtypes: []string{"alipay", "bank"},
		Inputs: map[string]plugin.InputConfig{
			"appid":     {Name: "应用APPID", Type: "input"},
			"appkey":    {Name: "支付宝公钥", Type: "textarea"},
			"appsecret": {Name: "应用私钥", Type: "textarea"},
			"appswitch": {Name: "是否使用mapi接口", Type: "select", Options: map[string]string{"0": "否", "1": "是"}},
			"appurl":    {Name: "接口地址", Type: "input"},
		},
		Select: map[string]string{
			"1": "电脑网站支付",
			"2": "手机网站支付",
			"3": "当面付扫码",
			"4": "当面付JS",
			"5": "预授权支付",
			"6": "APP支付",
			"7": "JSAPI支付",
			"8": "订单码支付",
		},
		Note: "<p>支付宝官方支付接口，支持多种支付方式</p>",
	}
}

// 获取配置
func (p *AlipayPlugin) getConfig(channelID int) map[string]string {
	var channel model.Channel
	config.DB.First(&channel, channelID)

	cfg := make(map[string]string)
	if channel.Config != "" {
		// JSON解析配置
		// cfg = json.Unmarshal(channel.Config)
	}
	return cfg
}

// 提交支付
func (p *AlipayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	method := params["method"].(string)

	switch method {
	case "web", "jump":
		return p.submitWeb(params)
	case "scan":
		return p.submitScan(params)
	case "jsapi":
		return p.submitJSAPI(params)
	case "app":
		return p.submitApp(params)
	case "wap":
		return p.submitWap(params)
	default:
		return p.submitWeb(params)
	}
}

func (p *AlipayPlugin) submitWeb(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	returnURL := params["return_url"].(string)

	// 构造支付参数
	payURL := p.buildPayURL(tradeNo, money, name, notifyURL, returnURL, "pay")

	return plugin.SubmitResult{
		Type: "jump",
		URL:  payURL,
	}, nil
}

func (p *AlipayPlugin) submitScan(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	// 构造扫码支付参数
	payURL := p.buildPayURL(tradeNo, money, name, notifyURL, "", "scan")

	return plugin.SubmitResult{
		Type: "qrcode",
		URL:  payURL,
	}, nil
}

func (p *AlipayPlugin) submitJSAPI(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	openid := params["openid"].(string)

	// 构造JSAPI支付参数
	payURL := p.buildPayURL(tradeNo, money, name, notifyURL, "", "jsapi")

	return plugin.SubmitResult{
		Type: "jsapi",
		URL:  payURL,
		Data: map[string]string{
			"trade_no": tradeNo,
		},
	}, nil
}

func (p *AlipayPlugin) submitApp(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	// 构造APP支付参数
	payURL := p.buildPayURL(tradeNo, money, name, notifyURL, "", "app")

	return plugin.SubmitResult{
		Type: "app",
		URL:  payURL,
	}, nil
}

func (p *AlipayPlugin) submitWap(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	returnURL := params["return_url"].(string)

	// 构造H5支付参数
	payURL := p.buildPayURL(tradeNo, money, name, notifyURL, returnURL, "wap")

	return plugin.SubmitResult{
		Type: "jump",
		URL:  payURL,
	}, nil
}

func (p *AlipayPlugin) buildPayURL(tradeNo string, money float64, name, notifyURL, returnURL, method string) string {
	// TODO: 实现完整的支付宝支付URL构造
	// 这里需要使用支付宝SDK进行签名

	params := url.Values{}
	params.Set("out_trade_no", tradeNo)
	params.Set("total_amount", strconv.FormatFloat(money, 'f', 2, 64))
	params.Set("subject", name)
	params.Set("method", method)
	params.Set("notify_url", notifyURL)
	if returnURL != "" {
		params.Set("return_url", returnURL)
	}

	// 签名
	sign := p.sign(params)

	baseURL := "https://openapi.alipay.com/gateway.do"
	return baseURL + "?" + params.Encode() + "&sign=" + url.QueryEscape(sign)
}

// 签名
func (p *AlipayPlugin) sign(params url.Values) string {
	// 获取排序后的参数字符串
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signData string
	for _, k := range keys {
		if params.Get(k) != "" {
			signData += k + "=" + params.Get(k) + "&"
		}
	}
	signData = strings.TrimSuffix(signData, "&")

	// RSA2签名
	// 这里需要使用私钥进行签名
	return signData
}

// 移动端提交
func (p *AlipayPlugin) Mapi(params map[string]interface{}) (plugin.SubmitResult, error) {
	return p.submitWap(params)
}

// 异步回调
func (p *AlipayPlugin) Notify(tradeNo string) (plugin.NotifyResult, error) {
	// TODO: 实现支付宝异步回调处理
	// 1. 验签
	// 2. 解析回调数据
	// 3. 返回结果

	return plugin.NotifyResult{
		Success:    true,
		TradeNo:    tradeNo,
		APITradeNo: "202403151234567890",
		Amount:     100.00,
		Buyer:      "2088123456789012",
		Message:    "成功",
	}, nil
}

// 同步回调
func (p *AlipayPlugin) Return(tradeNo string) (plugin.ReturnResult, error) {
	return plugin.ReturnResult{
		Success: true,
		TradeNo: tradeNo,
		Message: "支付成功",
		URL:     "/user/order",
	}, nil
}

// 支付成功页面
func (p *AlipayPlugin) OK(tradeNo string) (string, error) {
	return "订单支付成功", nil
}

// 退款
func (p *AlipayPlugin) Refund(params map[string]interface{}) (plugin.RefundResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)

	// TODO: 调用支付宝退款接口

	return plugin.RefundResult{
		Code:    0,
		TradeNo: tradeNo,
		Fee:     money,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账
func (p *AlipayPlugin) Transfer(params map[string]interface{}) (plugin.TransferResult, error) {
	bizNo := params["biz_no"].(string)
	account := params["account"].(string)
	name := params["name"].(string)
	money := params["money"].(float64)

	// TODO: 调用支付宝转账接口

	return plugin.TransferResult{
		Code:    0,
		OrderID: bizNo,
		PayDate: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账查询
func (p *AlipayPlugin) TransferQuery(params map[string]interface{}) (plugin.TransferQueryResult, error) {
	bizNo := params["biz_no"].(string)

	// TODO: 调用支付宝转账查询接口

	return plugin.TransferQueryResult{
		Code:    0,
		Status:  1,
		Amount:  100.00,
		PayDate: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
