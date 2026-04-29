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
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	t.Setenv("PORT", "3100")
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/image_show?sslmode=disable")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("MOCK_SUB2API", "true")
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
