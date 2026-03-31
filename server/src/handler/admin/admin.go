package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/service"
)

// 管理员Handler
type AdminHandler struct {
	authSvc     *service.AuthService
	orderSvc    *service.OrderService
	settleSvc   *service.SettleService
	transferSvc *service.TransferService
	userSvc     *service.AuthService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		authSvc:     service.NewAuthService(),
		orderSvc:    service.NewOrderService(),
		settleSvc:   service.NewSettleService(),
		transferSvc: service.NewTransferService(),
		userSvc:     service.NewAuthService(),
	}
}

// 登录页面
func (h *AdminHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/login.html", nil)
}

// 登录处理
func (h *AdminHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	token, err := h.authSvc.AdminLogin(username, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.SetCookie("admin_token", token, 86400*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "登录成功"})
}

// 登出
func (h *AdminHandler) Logout(c *gin.Context) {
	c.SetCookie("admin_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已退出"})
}

// 首页
func (h *AdminHandler) Index(c *gin.Context) {
	// 统计信息
	var orderCount, userCount int64
	var todayMoney float64

	config.DB.Model(&model.Order{}).Count(&orderCount)
	config.DB.Model(&model.User{}).Count(&userCount)

	today := time.Now().Format("2006-01-02")
	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", today).
		Select("COALESCE(SUM(money), 0)").Scan(&todayMoney)

	c.HTML(http.StatusOK, "admin/index.html", gin.H{
		"order_count":  orderCount,
		"user_count":   userCount,
		"today_money": todayMoney,
	})
}

// 商户列表
func (h *AdminHandler) UserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	config.DB.Model(&model.User{}).Count(&total)
	config.DB.Offset(offset).Limit(pageSize).Order("uid DESC").Find(&users)

	c.HTML(http.StatusOK, "admin/ulist.html", gin.H{
		"users":  users,
		"total":  total,
		"page":   page,
		"pages":  (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// 订单列表
func (h *AdminHandler) OrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	offset := (page - 1) * pageSize

	status := c.DefaultQuery("status", "-1")
	tradeNo := c.Query("trade_no")

	var orders []model.Order
	var total int64

	query := config.DB.Model(&model.Order{})
	if status != "-1" {
		query = query.Where("status = ?", status)
	}
	if tradeNo != "" {
		query = query.Where("trade_no LIKE ?", "%"+tradeNo+"%")
	}

	query.Count(&total)
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders)

	c.HTML(http.StatusOK, "admin/order.html", gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
	})
}

// 结算列表
func (h *AdminHandler) SettleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var settles []model.Settle
	var total int64

	config.DB.Model(&model.Settle{}).Count(&total)
	config.DB.Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&settles)

	c.HTML(http.StatusOK, "admin/settle.html", gin.H{
		"settles": settles,
		"total":   total,
		"page":    page,
	})
}

// 转账列表
func (h *AdminHandler) TransferList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var transfers []model.Transfer
	var total int64

	config.DB.Model(&model.Transfer{}).Count(&total)
	config.DB.Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&transfers)

	c.HTML(http.StatusOK, "admin/transfer.html", gin.H{
		"transfers": transfers,
		"total":     total,
		"page":      page,
	})
}

// 系统设置页面
func (h *AdminHandler) Settings(c *gin.Context) {
	mod := c.DefaultQuery("mod", "site")

	// 加载所有配置
	var configs []model.Config
	config.DB.Find(&configs)

	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.K] = cfg.V
	}

	c.HTML(http.StatusOK, "admin/set.html", gin.H{
		"mod":    mod,
		"paygo/config": configMap,
	})
}

// 保存设置
func (h *AdminHandler) SaveSettings(c *gin.Context) {
	mod := c.PostForm("mod")

	if mod == "account" {
		// 管理员密码修改
		oldPwd := c.PostForm("old_pwd")
		newPwd := c.PostForm("new_pwd")
		confirmPwd := c.PostForm("confirm_pwd")

		if newPwd != confirmPwd {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "两次密码不一致"})
			return
		}

		cfg := config.AppConfig
		if oldPwd != cfg.AdminPwd {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "原密码错误"})
			return
		}

		h.authSvc.SaveConfig("admin_pwd", newPwd)
		cfg.AdminPwd = newPwd
	} else {
		// 其他配置保存
		cfgKeys := []string{
			"sitename", "localurl", "apiurl", "kfqq",
			"reg_open", "settle_money", "settle_alipay", "settle_wxpay",
			"transfer_alipay", "transfer_wxpay", "login_alipay", "login_qq", "login_wx",
		}

		for _, k := range cfgKeys {
			v := c.PostForm(k)
			if v != "" {
				h.authSvc.SaveConfig(k, v)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功"})
}

// 通道管理
func (h *AdminHandler) ChannelList(c *gin.Context) {
	var channels []model.Channel
	config.DB.Find(&channels)

	c.HTML(http.StatusOK, "admin/channel.html", gin.H{
		"channels": channels,
	})
}

// 插件管理
func (h *AdminHandler) PluginList(c *gin.Context) {
	var plugins []model.Plugin
	config.DB.Find(&plugins)

	c.HTML(http.StatusOK, "admin/plugin.html", gin.H{
		"plugins": plugins,
	})
}

// AJAX: 获取订单列表
func (h *AdminHandler) AjaxOrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.PostForm("page"))
	pageSize, _ := strconv.Atoi(c.PostForm("limit"))
	status := c.PostForm("status")

	query := config.DB.Model(&model.Order{})
	if status != "" && status != "-1" {
		query = query.Where("status = ?", status)
	}

	var orders []model.Order
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  orders,
	})
}

// AJAX: 订单操作
func (h *AdminHandler) AjaxOrderOp(c *gin.Context) {
	action := c.PostForm("action")
	tradeNo := c.PostForm("trade_no")

	var err error
	switch action {
	case "refund":
		moneyStr := c.PostForm("money")
		money, _ := strconv.ParseFloat(moneyStr, 10)
		err = h.orderSvc.Refund(tradeNo, money)
	case "freeze":
		err = h.orderSvc.Freeze(tradeNo)
	case "unfreeze":
		err = h.orderSvc.Unfreeze(tradeNo)
	case "notify":
		// 重新通知
		err = nil
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 商户操作
func (h *AdminHandler) AjaxUserOp(c *gin.Context) {
	action := c.PostForm("action")
	uid, _ := strconv.Atoi(c.PostForm("uid"))

	switch action {
	case "reset_key":
		// 重置密钥
		// TODO: 实现重置密钥
	case "set_status":
		status, _ := strconv.Atoi(c.PostForm("status"))
		config.DB.Model(&model.User{}).Where("uid = ?", uid).Update("status", status)
	case "recharge":
		money, _ := strconv.ParseFloat(c.PostForm("money"), 10)
		typ := c.PostForm("type")
		h.transferSvc.AdminChangeMoney(uint(uid), money, typ, "管理员操作")
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 结算操作
func (h *AdminHandler) AjaxSettleOp(c *gin.Context) {
	action := c.PostForm("action")
	id, _ := strconv.Atoi(c.PostForm("id"))

	var err error
	switch action {
	case "approve":
		err = h.settleSvc.ApproveSettle(uint(id))
	case "reject":
		reason := c.PostForm("reason")
		err = h.settleSvc.RejectSettle(uint(id), reason)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// 统计API
func (h *AdminHandler) AjaxStats(c *gin.Context) {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	var todayOrderMoney, yesterdayOrderMoney float64
	var todayOrderCount, yesterdayOrderCount int64
	var userCount int64

	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", today).
		Select("COALESCE(SUM(money), 0)").Scan(&todayOrderMoney)
	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", today).
		Count(&todayOrderCount)

	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", yesterday).
		Select("COALESCE(SUM(money), 0)").Scan(&yesterdayOrderMoney)
	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", yesterday).
		Count(&yesterdayOrderCount)

	config.DB.Model(&model.User{}).Count(&userCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"today_order_money":    todayOrderMoney,
			"today_order_count":    todayOrderCount,
			"yesterday_order_money": yesterdayOrderMoney,
			"yesterday_order_count": yesterdayOrderCount,
			"user_count": userCount,
		},
	})
}
