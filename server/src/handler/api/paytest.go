package api

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gopay/src/config"

	"github.com/gin-gonic/gin"
)

const testNotifySessionTTL = 30 * time.Minute

type testNotifyEvent struct {
	ReceivedAt   int64             `json:"received_at"`
	TradeNo      string            `json:"trade_no"`
	OutTradeNo   string            `json:"out_trade_no"`
	Status       string            `json:"status"`
	Money        string            `json:"money"`
	SignType     string            `json:"sign_type"`
	SignValid    bool              `json:"sign_valid"`
	VerifyReason string            `json:"verify_reason"`
	Form         map[string]string `json:"form"`
}

type testNotifySession struct {
	Token     string           `json:"token"`
	NotifyURL string           `json:"notify_url"`
	ReturnURL string           `json:"return_url"`
	CreatedAt int64            `json:"created_at"`
	ExpiresAt int64            `json:"expires_at"`
	HitCount  int              `json:"hit_count"`
	LastEvent *testNotifyEvent `json:"last_event"`
}

type paytestStore struct {
	mu       sync.Mutex
	sessions map[string]*testNotifySession
}

func newPaytestStore() *paytestStore {
	return &paytestStore{
		sessions: make(map[string]*testNotifySession),
	}
}

func normalizeBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return ""
	}
	if strings.TrimSpace(u.Scheme) == "" || strings.TrimSpace(u.Host) == "" {
		return ""
	}
	return strings.TrimRight(u.String(), "/")
}

func resolveTestNotifyBaseURL(c *gin.Context) string {
	if v := normalizeBaseURL(resolveRequestBaseURL(c)); v != "" {
		return v
	}
	if v := normalizeBaseURL(config.Get("localurl")); v != "" {
		return v
	}
	if v := normalizeBaseURL(config.Get("apiurl")); v != "" {
		return v
	}
	return "http://127.0.0.1:8080"
}

func generateTestNotifyToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(buf)), nil
}

func (h *PayHandler) cleanupExpiredTestNotifySessions(now time.Time) {
	if h.paytest == nil {
		return
	}
	h.paytest.mu.Lock()
	defer h.paytest.mu.Unlock()

	nowUnix := now.Unix()
	for token, s := range h.paytest.sessions {
		if s == nil || s.ExpiresAt <= nowUnix {
			delete(h.paytest.sessions, token)
		}
	}
}

// 创建测试回调会话（用于 paytest 页面自动生成 notify_url）
func (h *PayHandler) CreateTestNotifySession(c *gin.Context) {
	now := time.Now()
	h.cleanupExpiredTestNotifySessions(now)

	token, err := generateTestNotifyToken()
	if err != nil {
		log.Printf("[pay_test_notify_session_failed] reason=generate token failed, error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "生成测试会话失败"})
		return
	}

	baseURL := resolveTestNotifyBaseURL(c)
	session := &testNotifySession{
		Token:     token,
		NotifyURL: baseURL + "/api/pay/test_notify/" + token,
		ReturnURL: baseURL + "/",
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(testNotifySessionTTL).Unix(),
		HitCount:  0,
		LastEvent: nil,
	}

	h.paytest.mu.Lock()
	h.paytest.sessions[token] = session
	h.paytest.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": session})
}

// 查询测试回调会话状态
func (h *PayHandler) GetTestNotifySession(c *gin.Context) {
	token := strings.ToUpper(strings.TrimSpace(c.Param("token")))
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "token不能为空"})
		return
	}

	h.cleanupExpiredTestNotifySessions(time.Now())

	h.paytest.mu.Lock()
	session, ok := h.paytest.sessions[token]
	h.paytest.mu.Unlock()
	if !ok || session == nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "测试会话不存在或已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": session})
}

// 测试回调接收器（供 paytest 页面使用）
func (h *PayHandler) TestNotifyCallback(c *gin.Context) {
	token := strings.ToUpper(strings.TrimSpace(c.Param("token")))
	if token == "" {
		log.Printf("[pay_test_notify_failed] reason=empty token")
		c.String(http.StatusOK, "fail")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		log.Printf("[pay_test_notify_failed] token=%s, reason=parse form failed, error=%s", token, err.Error())
		c.String(http.StatusOK, "fail")
		return
	}

	form := make(map[string]string, len(c.Request.PostForm))
	for k, vv := range c.Request.PostForm {
		if len(vv) == 0 {
			continue
		}
		form[k] = strings.TrimSpace(vv[0])
	}

	signValid := false
	verifyReason := ""
	tradeNo := strings.TrimSpace(form["trade_no"])
	sign := strings.TrimSpace(form["sign"])
	signType := strings.ToUpper(strings.TrimSpace(form["sign_type"]))
	if signType == "" {
		signType = "HMAC-SHA256"
	}

	if sign == "" {
		verifyReason = "missing sign"
	} else if tradeNo == "" {
		verifyReason = "missing trade_no"
	} else {
		order, err := h.orderSvc.GetOrder(tradeNo)
		if err != nil {
			verifyReason = "order not found"
		} else {
			user, err := h.authSvc.GetUser(order.UID)
			if err != nil {
				verifyReason = "merchant not found"
			} else {
				signParams := map[string]string{
					"trade_no":     form["trade_no"],
					"out_trade_no": form["out_trade_no"],
					"type":         form["type"],
					"status":       form["status"],
					"money":        form["money"],
					"realmoney":    form["realmoney"],
				}
				expected := h.authSvc.MakeSign(signParams, user.Key)
				signValid = strings.EqualFold(sign, expected)
				if !signValid {
					verifyReason = "sign mismatch"
				}
			}
		}
	}

	event := &testNotifyEvent{
		ReceivedAt:   time.Now().Unix(),
		TradeNo:      tradeNo,
		OutTradeNo:   strings.TrimSpace(form["out_trade_no"]),
		Status:       strings.TrimSpace(form["status"]),
		Money:        strings.TrimSpace(form["money"]),
		SignType:     signType,
		SignValid:    signValid,
		VerifyReason: verifyReason,
		Form:         form,
	}

	h.cleanupExpiredTestNotifySessions(time.Now())

	h.paytest.mu.Lock()
	session, ok := h.paytest.sessions[token]
	if ok && session != nil {
		session.HitCount++
		session.LastEvent = event
		session.ExpiresAt = time.Now().Add(testNotifySessionTTL).Unix()
	}
	h.paytest.mu.Unlock()

	if !ok {
		log.Printf("[pay_test_notify_failed] token=%s, trade_no=%s, reason=session not found", token, tradeNo)
		c.String(http.StatusOK, "fail")
		return
	}

	log.Printf("[pay_test_notify_received] token=%s, trade_no=%s, sign_valid=%t, verify_reason=%s", token, tradeNo, signValid, verifyReason)
	c.String(http.StatusOK, "success")
}
