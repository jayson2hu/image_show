package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jayson2hu/image-show/config"
)

func TestInitDBSQLiteMigratesTables(t *testing.T) {
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = CloseDB()
		_ = os.Chdir(originalDir)
		config.AppConfig = nil
		DB = nil
	})

	config.AppConfig = &config.Config{DBDriver: "sqlite"}
	if err := InitDB(); err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "data", "image_show.db")); err != nil {
		t.Fatalf("sqlite db file not created: %v", err)
	}

	models := []interface{}{
		&User{},
		&Generation{},
		&CreditLog{},
		&LoginLog{},
		&Channel{},
		&Setting{},
		&PromptTemplate{},
		&AnonymousIdentity{},
	}
	for _, model := range models {
		if !DB.Migrator().HasTable(model) {
			t.Fatalf("missing table for %T", model)
		}
	}
}
