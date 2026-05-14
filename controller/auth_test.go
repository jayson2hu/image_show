package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/middleware"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/router"
	"github.com/jayson2hu/image-show/service"
)

func TestAuthFlow(t *testing.T) {
	engine := setupAuthTest(t)
	email := "user@example.com"

	rec := postJSON(engine, "/api/auth/send-code", map[string]string{"email": email})
	if rec.Code != http.StatusOK {
		t.Fatalf("send-code status = %d body = %s", rec.Code, rec.Body.String())
	}

	code := service.PeekVerificationCode(email)
	rec = postJSON(engine, "/api/auth/register", map[string]string{
		"email":    email,
		"password": "password123",
		"code":     code,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("register status = %d body = %s", rec.Code, rec.Body.String())
	}
	var registerResp struct {
		Token string `json:"token"`
		User  struct {
			ID      int64   `json:"id"`
			Email   string  `json:"email"`
			Credits float64 `json:"credits"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registerResp.Token == "" || registerResp.User.Credits != 10 {
		t.Fatalf("unexpected register response: %+v", registerResp)
	}
	var createdUser model.User
	if err := model.DB.First(&createdUser, registerResp.User.ID).Error; err != nil {
		t.Fatalf("load created user: %v", err)
	}
	if createdUser.CreditsExpiry == nil {
		t.Fatal("expected credits expiry")
	}

	rec = postJSON(engine, "/api/auth/login", map[string]string{
		"email":    email,
		"password": "password123",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d body = %s", rec.Code, rec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("expected login token")
	}
	var loginLog model.LoginLog
	if err := model.DB.Where("user_id = ? AND success = ?", registerResp.User.ID, true).First(&loginLog).Error; err != nil {
		t.Fatalf("expected successful login log: %v", err)
	}
	if loginLog.IP != "1.2.3.4" || loginLog.UserAgent != "auth-test" || loginLog.Method != "email" {
		t.Fatalf("unexpected login log: %+v", loginLog)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "PasswordHash") || strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatalf("me response invalid: status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized without token, got %d", rec.Code)
	}
}

func TestRegisterGiftCreditsConfigurable(t *testing.T) {
	engine := setupAuthTest(t)
	email := "gift@example.com"
	if err := model.DB.Create(&model.Setting{Key: "register_gift_credits", Value: "12"}).Error; err != nil {
		t.Fatalf("create gift setting: %v", err)
	}
	if err := service.SendVerificationCode(email); err != nil {
		t.Fatalf("send code: %v", err)
	}

	rec := postJSON(engine, "/api/auth/register", map[string]string{
		"email":    email,
		"password": "password123",
		"code":     service.PeekVerificationCode(email),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("register status = %d body = %s", rec.Code, rec.Body.String())
	}
	var registerResp struct {
		User struct {
			Credits float64 `json:"credits"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registerResp.User.Credits != 12 {
		t.Fatalf("expected configurable gift credits 12, got %v", registerResp.User.Credits)
	}
}

func TestSendCodeRateLimit(t *testing.T) {
	engine := setupAuthTest(t)
	email := "limit@example.com"

	first := postJSON(engine, "/api/auth/send-code", map[string]string{"email": email})
	second := postJSON(engine, "/api/auth/send-code", map[string]string{"email": email})

	if first.Code != http.StatusOK {
		t.Fatalf("first send-code status = %d", first.Code)
	}
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second send-code status = %d", second.Code)
	}
}

func TestRegisterDisabled(t *testing.T) {
	engine := setupAuthTest(t)
	email := "disabled@example.com"
	if err := model.DB.Create(&model.Setting{Key: "register_enabled", Value: "false"}).Error; err != nil {
		t.Fatalf("create setting: %v", err)
	}

	sendCode := postJSON(engine, "/api/auth/send-code", map[string]string{"email": email})
	if sendCode.Code != http.StatusForbidden {
		t.Fatalf("expected send-code 403 when registration disabled, got %d body=%s", sendCode.Code, sendCode.Body.String())
	}

	rec := postJSON(engine, "/api/auth/register", map[string]string{
		"email":    email,
		"password": "password123",
		"code":     "123456",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "registration is disabled") {
		t.Fatalf("expected friendly disabled message, got %s", rec.Body.String())
	}
}

func TestRegisterEmailDomainAllowlist(t *testing.T) {
	engine := setupAuthTest(t)
	if err := model.DB.Create(&model.Setting{Key: "register_email_domain_allowlist", Value: "example.com\ncompany.com"}).Error; err != nil {
		t.Fatalf("create allowlist: %v", err)
	}

	blockedEmail := "blocked@other.com"
	if err := service.SendVerificationCode(blockedEmail); err != nil {
		t.Fatalf("send blocked code: %v", err)
	}
	blocked := postJSON(engine, "/api/auth/register", map[string]string{
		"email":    blockedEmail,
		"password": "password123",
		"code":     service.PeekVerificationCode(blockedEmail),
	})
	if blocked.Code != http.StatusForbidden || !strings.Contains(blocked.Body.String(), "邮箱后缀") {
		t.Fatalf("expected domain forbidden, got %d body=%s", blocked.Code, blocked.Body.String())
	}

	allowedEmail := "allowed@example.com"
	if err := service.SendVerificationCode(allowedEmail); err != nil {
		t.Fatalf("send allowed code: %v", err)
	}
	allowed := postJSON(engine, "/api/auth/register", map[string]string{
		"email":    allowedEmail,
		"password": "password123",
		"code":     service.PeekVerificationCode(allowedEmail),
	})
	if allowed.Code != http.StatusOK {
		t.Fatalf("expected allowed register, got %d body=%s", allowed.Code, allowed.Body.String())
	}
}

func setupAuthTest(t *testing.T) *gin.Engine {
	t.Helper()
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = model.CloseDB()
		model.DB = nil
		config.AppConfig = nil
		service.ResetTrialStoreForTest()
		middleware.ResetRateLimitForTest()
		_ = service.CloseRedis()
		_ = os.Chdir(originalDir)
	})

	config.AppConfig = &config.Config{
		AppEnv:    "test",
		Port:      3000,
		DBDriver:  "sqlite",
		JWTSecret: "test-secret",
	}
	if err := model.InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router.Register(engine)
	return engine
}

func postJSON(engine http.Handler, path string, body interface{}) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "1.2.3.4")
	req.Header.Set("User-Agent", "auth-test")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
