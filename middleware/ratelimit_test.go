package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/middleware"
	"github.com/jayson2hu/image-show/model"
)

func TestGenerationRateLimitUserLimit(t *testing.T) {
	engine := rateLimitRouter(t, 10, "")
	for i := 0; i < 10; i++ {
		rec := requestRateLimited(engine, "1.2.3.4", int64(10))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status=%d", i+1, rec.Code)
		}
	}
	rec := requestRateLimited(engine, "1.2.3.4", int64(10))
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("expected 429 with Retry-After, got %d", rec.Code)
	}
}

func TestGenerationRateLimitIPLimit(t *testing.T) {
	engine := rateLimitRouter(t, 0, "")
	for i := 0; i < 20; i++ {
		rec := requestRateLimited(engine, "2.2.2.2", int64(i+1))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status=%d", i+1, rec.Code)
		}
	}
	rec := requestRateLimited(engine, "2.2.2.2", int64(99))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected IP 429, got %d", rec.Code)
	}
}

func TestGenerationRateLimitDailyBudget(t *testing.T) {
	engine := rateLimitRouter(t, 2, "2")
	first := requestRateLimited(engine, "3.3.3.3", int64(1))
	second := requestRateLimited(engine, "3.3.3.4", int64(2))
	third := requestRateLimited(engine, "3.3.3.5", int64(3))
	if first.Code != http.StatusOK || second.Code != http.StatusOK || third.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected budget statuses: %d %d %d", first.Code, second.Code, third.Code)
	}
}

func rateLimitRouter(t *testing.T, userID int64, dailyBudget string) *gin.Engine {
	t.Helper()
	middleware.ResetRateLimitForTest()
	config.AppConfig = &config.Config{DBDriver: "sqlite"}
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	if err := model.InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	if dailyBudget != "" {
		if err := model.DB.Create(&model.Setting{Key: "daily_budget", Value: dailyBudget}).Error; err != nil {
			t.Fatalf("create budget setting: %v", err)
		}
	}
	t.Cleanup(func() {
		middleware.ResetRateLimitForTest()
		_ = model.CloseDB()
		config.AppConfig = nil
		model.DB = nil
		_ = os.Chdir(originalDir)
	})
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RealIP())
	if userID > 0 {
		r.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
	}
	r.POST("/limited", middleware.GenerationRateLimit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func requestRateLimited(engine http.Handler, ip string, userID int64) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/limited", nil)
	req.Header.Set("X-Real-IP", ip)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
