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

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing nosniff header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("missing frame header")
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("missing csp header")
	}
}

func TestIPBlacklistBlocksRequest(t *testing.T) {
	setupMiddlewareDB(t)
	if err := model.DB.Create(&model.Setting{Key: "ip_blacklist", Value: "1.2.3.4,5.6.7.8"}).Error; err != nil {
		t.Fatalf("create setting: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RealIP(), middleware.IPBlacklist())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Real-IP", "1.2.3.4")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func setupMiddlewareDB(t *testing.T) {
	t.Helper()
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
	t.Cleanup(func() {
		_ = model.CloseDB()
		config.AppConfig = nil
		model.DB = nil
		_ = os.Chdir(originalDir)
	})
}
