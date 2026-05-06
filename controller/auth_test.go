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
	"golang.org/x/crypto/bcrypt"
)

func TestAuthFlow(t *testing.T) {
	engine := setupAuthTest(t)
	email := "user@example.com"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := model.User{Email: email, Username: "user", PasswordHash: string(passwordHash), Role: 1, Status: 1}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	rec := postJSON(engine, "/api/auth/login", map[string]string{
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
	if err := model.DB.Where("user_id = ? AND success = ?", user.ID, true).First(&loginLog).Error; err != nil {
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

func TestEmailRegistrationDisabled(t *testing.T) {
	engine := setupAuthTest(t)

	rec := postJSON(engine, "/api/auth/send-code", map[string]string{"email": "user@example.com"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("send-code status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(engine, "/api/auth/register", map[string]string{
		"email":    "user@example.com",
		"password": "password123",
		"code":     "123456",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("register status = %d body = %s", rec.Code, rec.Body.String())
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
