package epay

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"paygo/src/plugin"
)

// 聚合支付插件(彩虹易支付)
type EpayPlugin struct {
	plugin.BasePlugin
}

func New() plugin.Plugin {
	return &EpayPlugin{}
}

func init() {
	plugin.Register("epay", New)
}

func (p *EpayPlugin) GetInfo() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:       "epay",
		Showname:   "彩虹易支付",
		Author:     "彩虹聚合",
		Link:       "https://codepay.pro/",
		Types:      []string{"alipay", "wxpay", "qqpay", "bank"},
		Transtypes: []string{"alipay", "wxpay", "bank"},
		Inputs: map[string]plugin.InputConfig{
			"appid":  {Name: "商户ID", Type: "input"},
			"appkey": {Name: "商户密钥", Type: "input"},
			"appurl": {Name: "接口地址", Type: "input"},
		},
		Select: map[string]string{
			"1": "支付宝扫码",
			"2": "微信扫码",
			"3": "QQ扫码",
			"4": "支付宝跳转",
			"5": "微信跳转",
		},
		Note: "<p>聚合支付接口，支持多种支付方式</p>",
	}
}

func (p *EpayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	returnURL := params["return_url"].(string)
	payType := params["paytype"].(string) // alipay/wxpay/qqpay/bank

	// 构造支付链接
	payURL := p.buildPayURL(tradeNo, money, name, notifyURL, returnURL, payType)

	return plugin.SubmitResult{
		Type: "jump",
		URL:  payURL,
	}, nil
}

func (p *EpayPlugin) buildPayURL(tradeNo string, money float64, name, notifyURL, returnURL, payType string) string {
	// TODO: 从配置获取真实参数
	appid := "10001"
	appkey := "xxx"
	appurl := "https://pay.example.com"

	params := url.Values{}
	params.Set("pid", appid)
	params.Set("out_trade_no", tradeNo)
	params.Set("money", strconv.FormatFloat(money, 'f', 2, 64))
	params.Set("name", name)
	params.Set("notify_url", notifyURL)
	params.Set("return_url", returnURL)
	params.Set("type", payType)

	// 签名
	signStr := fmt.Sprintf("out_trade_no=%s&pid=%s&money=%s&key=%s",
		tradeNo, appid, strconv.FormatFloat(money, 'f', 2, 64), appkey)

	return appurl + "/submit?" + params.Encode() + "&sign=" + signStr
}

func (p *EpayPlugin) Mapi(params map[string]interface{}) (plugin.SubmitResult, error) {
	return p.Submit(params)
}

func (p *EpayPlugin) Notify(tradeNo string) (plugin.NotifyResult, error) {
	return plugin.NotifyResult{
		Success:    true,
		TradeNo:    tradeNo,
		APITradeNo: "EP" + time.Now().Format("20060102150405"),
		Amount:     100.00,
		Buyer:      "",
		Message:    "成功",
	}, nil
}

func (p *EpayPlugin) Return(tradeNo string) (plugin.ReturnResult, error) {
	return plugin.ReturnResult{
		Success: true,
		TradeNo: tradeNo,
		Message: "支付成功",
	}, nil
}

func (p *EpayPlugin) OK(tradeNo string) (string, error) {
	return "支付成功", nil
}

func (p *EpayPlugin) Refund(params map[string]interface{}) (plugin.RefundResult, error) {
	return plugin.RefundResult{Code: 0}, nil
}

func (p *EpayPlugin) Transfer(params map[string]interface{}) (plugin.TransferResult, error) {
	return plugin.TransferResult{Code: 0}, nil
}

func (p *EpayPlugin) TransferQuery(params map[string]interface{}) (plugin.TransferQueryResult, error) {
	return plugin.TransferQueryResult{Code: 0}, nil
}
