package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/jayson2hu/image-show/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestSeedDefaultAdminCreatesAndResetsAdmin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	cfg := &config.Config{AdminEmail: "admin@image-show.local", AdminPassword: "Admin123456"}

	if err := seedDefaultAdmin(db, cfg); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	var admin User
	if err := db.Where("email = ?", cfg.AdminEmail).First(&admin).Error; err != nil {
		t.Fatalf("load admin: %v", err)
	}
	if admin.Role != 10 || admin.Status != 1 {
		t.Fatalf("unexpected admin flags: role=%d status=%d", admin.Role, admin.Status)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(cfg.AdminPassword)); err != nil {
		t.Fatalf("admin password mismatch: %v", err)
	}

	cfg.AdminPassword = "Admin654321"
	if err := seedDefaultAdmin(db, cfg); err != nil {
		t.Fatalf("reseed admin: %v", err)
	}
	if err := db.Where("email = ?", cfg.AdminEmail).First(&admin).Error; err != nil {
		t.Fatalf("reload admin: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(cfg.AdminPassword)); err != nil {
		t.Fatalf("admin password was not reset: %v", err)
	}
}
