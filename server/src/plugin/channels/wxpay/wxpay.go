package wxpay

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"paygo/src/plugin"
)

// 微信支付插件
type WxpayPlugin struct {
	plugin.BasePlugin
}

func New() plugin.Plugin {
	return &WxpayPlugin{}
}

func init() {
	plugin.Register("wxpay", New)
}

func (p *WxpayPlugin) GetInfo() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:       "wxpay",
		Showname:   "微信支付官方",
		Author:     "微信支付",
		Link:       "https://pay.weixin.qq.com/",
		Types:      []string{"wxpay"},
		Transtypes: []string{"wxpay", "bank"},
		Inputs: map[string]plugin.InputConfig{
			"appid":     {Name: "公众账号ID", Type: "input"},
			"appkey":    {Name: "API密钥", Type: "input"},
			"appsecret": {Name: "AppSecret", Type: "input"},
			"appmchid":  {Name: "商户号", Type: "input"},
		},
		Select: map[string]string{
			"1": "公众号支付",
			"2": "原生扫码支付",
			"3": "APP支付",
			"4": "H5支付",
			"5": "小程序支付",
		},
		Note: "<p>微信支付官方接口</p>",
	}
}

// 提交支付
func (p *WxpayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	method := params["method"].(string)

	switch method {
	case "scan":
		return p.submitScan(params)
	case "jsapi":
		return p.submitJSAPI(params)
	case "app":
		return p.submitApp(params)
	case "wap":
		return p.submitH5(params)
	default:
		return p.submitScan(params)
	}
}

func (p *WxpayPlugin) submitScan(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)

	// 构造统一下单参数
	xmlData := p.buildUnifiedOrderXML(tradeNo, money, name, notifyURL, ip, "NATIVE")

	// 调用微信统一下单接口
	codeURL := p.callUnifiedOrder(xmlData)

	return plugin.SubmitResult{
		Type: "qrcode",
		URL:  codeURL,
	}, nil
}

func (p *WxpayPlugin) submitJSAPI(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)
	openid := params["openid"].(string)

	// 构造统一下单参数
	xmlData := p.buildUnifiedOrderXML(tradeNo, money, name, notifyURL, ip, "JSAPI")
	xmlData += fmt.Sprintf("<openid>%s</openid>", openid)

	// 调用微信统一下单接口
	prepayID := p.callUnifiedOrder(xmlData)

	// 构造JSAPI调起参数
	jsApiParams := p.buildJSAPIPayParams(prepayID)

	return plugin.SubmitResult{
		Type: "jsapi",
		Data: jsApiParams,
	}, nil
}

func (p *WxpayPlugin) submitApp(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)

	// 构造统一下单参数
	xmlData := p.buildUnifiedOrderXML(tradeNo, money, name, notifyURL, ip, "APP")

	// 调用微信统一下单接口
	prepayID := p.callUnifiedOrder(xmlData)

	// 构造APP调起参数
	appParams := p.buildAPPPayParams(prepayID)

	return plugin.SubmitResult{
		Type: "app",
		Data: appParams,
	}, nil
}

func (p *WxpayPlugin) submitH5(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)

	// 构造统一下单参数
	xmlData := p.buildUnifiedOrderXML(tradeNo, money, name, notifyURL, ip, "MWEB")

	// 调用微信统一下单接口
	mwebURL := p.callUnifiedOrder(xmlData)

	return plugin.SubmitResult{
		Type: "jump",
		URL:  mwebURL,
	}, nil
}

func (p *WxpayPlugin) buildUnifiedOrderXML(tradeNo string, money float64, name, notifyURL, ip, tradeType string) string {
	// TODO: 获取真实配置
	appid := ""
	mchid := ""
	apiKey := ""

	nonceStr := p.generateNonceStr(32)
	totalFee := int(money * 100) // 分

	xml := fmt.Sprintf(`<xml>
<appid>%s</appid>
<mch_id>%s</mch_id>
<nonce_str>%s</nonce_str>
<body>%s</body>
<out_trade_no>%s</out_trade_no>
<total_fee>%d</total_fee>
<spbill_create_ip>%s</spbill_create_ip>
<notify_url>%s</notify_url>
<trade_type>%s</trade_type>
</xml>`, appid, mchid, nonceStr, name, tradeNo, totalFee, ip, notifyURL, tradeType)

	// 签名
	sign := p.signXML(xml, apiKey)
	xml = strings.Replace(xml, "</xml>", fmt.Sprintf("<sign>%s</sign></xml>", sign), 1)

	return xml
}

func (p *WxpayPlugin) signXML(xml, key string) string {
	// 解析XML获取签名内容
	// 实际应该按微信文档进行排序签名
	return md5Hash(xml + key)
}

func (p *WxpayPlugin) callUnifiedOrder(xmlData string) string {
	// TODO: 调用微信统一下单接口
	// POST https://api.mch.weixin.qq.com/pay/unifiedorder

	return "weixin://wxpay/bizpayurl?pr=xxx"
}

func (p *WxpayPlugin) buildJSAPIPayParams(prepayID string) map[string]string {
	// TODO: 获取真实配置
	appid := ""
	apiKey := ""

	nonceStr := p.generateNonceStr(32)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	signStr := fmt.Sprintf("appId=%s&nonceStr=%s&package=prepay_id=%s&signType=MD5&timeStamp=%s&key=%s",
		appid, nonceStr, prepayID, timestamp, apiKey)
	sign := md5Hash(signStr)

	return map[string]string{
		"appId":     appid,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   "prepay_id=" + prepayID,
		"signType":  "MD5",
		"paySign":   sign,
	}
}

func (p *WxpayPlugin) buildAPPPayParams(prepayID string) map[string]string {
	// TODO: 获取真实配置
	appid := ""
	mchid := ""
	apiKey := ""

	nonceStr := p.generateNonceStr(32)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	signStr := fmt.Sprintf("appid=%s&partnerid=%s&prepayid=%s&package=Sign=WXPay&timestamp=%s&nonce=%s&key=%s",
		appid, mchid, prepayID, timestamp, nonceStr, apiKey)
	sign := md5Hash(signStr)

	return map[string]string{
		"appid":     appid,
		"partnerid": mchid,
		"prepayid":  prepayID,
		"package":   "Sign=WXPay",
		"timestamp": timestamp,
		"noncestr":  nonceStr,
		"sign":      sign,
	}
}

func (p *WxpayPlugin) generateNonceStr(length int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

func md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 移动端提交
func (p *WxpayPlugin) Mapi(params map[string]interface{}) (plugin.SubmitResult, error) {
	return p.submitH5(params)
}

// 异步回调
func (p *WxpayPlugin) Notify(tradeNo string) (plugin.NotifyResult, error) {
	// TODO: 实现微信支付异步回调处理
	// 1. 验签
	// 2. 解析回调数据
	// 3. 返回结果

	return plugin.NotifyResult{
		Success:    true,
		TradeNo:    tradeNo,
		APITradeNo: "wx202403151234567890",
		Amount:     100.00,
		Buyer:      "oXXXX",
		Message:    "成功",
	}, nil
}

// 同步回调
func (p *WxpayPlugin) Return(tradeNo string) (plugin.ReturnResult, error) {
	return plugin.ReturnResult{
		Success: true,
		TradeNo: tradeNo,
		Message: "支付成功",
		URL:     "/user/order",
	}, nil
}

// 支付成功页面
func (p *WxpayPlugin) OK(tradeNo string) (string, error) {
	return "订单支付成功", nil
}

// 退款
func (p *WxpayPlugin) Refund(params map[string]interface{}) (plugin.RefundResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)

	// TODO: 调用微信退款接口

	return plugin.RefundResult{
		Code:    0,
		TradeNo: tradeNo,
		Fee:     money,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账
func (p *WxpayPlugin) Transfer(params map[string]interface{}) (plugin.TransferResult, error) {
	bizNo := params["biz_no"].(string)
	account := params["account"].(string)
	// name := params["name"].(string)
	money := params["money"].(float64)

	// TODO: 调用微信企业付款接口

	return plugin.TransferResult{
		Code:    0,
		OrderID: bizNo,
		PayDate: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账查询
func (p *WxpayPlugin) TransferQuery(params map[string]interface{}) (plugin.TransferQueryResult, error) {
	bizNo := params["biz_no"].(string)

	// TODO: 调用微信企业付款查询接口

	return plugin.TransferQueryResult{
		Code:    0,
		Status:  1,
		Amount:  100.00,
		PayDate: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
