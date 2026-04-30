package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jayson2hu/image-show/common"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_DRIVER", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("MOCK_SUB2API", "")
	AppConfig = nil

	cfg := LoadConfig()

	if cfg.Port != 3000 {
		t.Fatalf("expected default port 3000, got %d", cfg.Port)
	}
	if cfg.DBDriver != "sqlite" {
		t.Fatalf("expected sqlite DB driver, got %q", cfg.DBDriver)
	}
	if cfg.RedisDB != 1 {
		t.Fatalf("expected redis db 1, got %d", cfg.RedisDB)
	}
	if cfg.Sub2APIBaseURL != "http://sub2api:8080" {
		t.Fatalf("unexpected sub2api url: %q", cfg.Sub2APIBaseURL)
	}
	if cfg.ImageModel != "gpt-image-2" {
		t.Fatalf("unexpected image model: %q", cfg.ImageModel)
	}
	if cfg.WeChatEnabled {
		t.Fatal("expected wechat auth disabled by default")
	}
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	t.Setenv("PORT", "3100")
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/image_show?sslmode=disable")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("MOCK_SUB2API", "true")
	t.Setenv("WECHAT_AUTH_ENABLED", "true")
	t.Setenv("WECHAT_SERVER_ADDRESS", "https://wechat.example.com")
	t.Setenv("WECHAT_SERVER_TOKEN", "token")
	t.Setenv("WECHAT_QRCODE_URL", "https://wechat.example.com/qrcode.png")
	AppConfig = nil

	cfg := LoadConfig()

	if cfg.Port != 3100 {
		t.Fatalf("expected overridden port 3100, got %d", cfg.Port)
	}
	if cfg.DBDriver != "postgres" {
		t.Fatalf("expected postgres DB driver, got %q", cfg.DBDriver)
	}
	if cfg.RedisDB != 2 {
		t.Fatalf("expected redis db 2, got %d", cfg.RedisDB)
	}
	if !cfg.MockSub2API {
		t.Fatal("expected mock sub2api enabled")
	}
	if !cfg.WeChatEnabled || cfg.WeChatServer != "https://wechat.example.com" || cfg.WeChatToken != "token" || cfg.WeChatQRCode == "" {
		t.Fatalf("unexpected wechat config: %+v", cfg)
	}
}

func TestLoadEnvReadsDotEnv(t *testing.T) {
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
		_ = os.Unsetenv("PORT")
		AppConfig = nil
	})

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=3200\n"), 0600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	_ = os.Unsetenv("PORT")
	common.LoadEnv()

	cfg := LoadConfig()
	if cfg.Port != 3200 {
		t.Fatalf("expected .env port 3200, got %d", cfg.Port)
	}
}
