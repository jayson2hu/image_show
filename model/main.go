package model

import (
	"fmt"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/jayson2hu/image-show/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

func InitDB() error {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	db, err := openDB(cfg.DBDriver, cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(
		&User{},
		&Generation{},
		&CreditLog{},
		&LoginLog{},
		&Channel{},
		&Setting{},
		&PromptTemplate{},
		&Announcement{},
		&AnnouncementRead{},
		&Package{},
		&Order{},
		&AnonymousIdentity{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	if err := seedDefaultPackages(db); err != nil {
		return err
	}
	if err := seedDefaultAdmin(db, cfg); err != nil {
		return err
	}

	DB = db
	return nil
}

func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func seedDefaultPackages(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Package{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaults := []Package{
		{Name: "入门包", Credits: 10, Price: 9.9, ValidDays: 30, SortOrder: 1, Status: 1},
		{Name: "标准包", Credits: 50, Price: 39.9, ValidDays: 90, SortOrder: 2, Status: 1},
		{Name: "专业包", Credits: 200, Price: 99.9, ValidDays: 180, SortOrder: 3, Status: 1},
	}
	return db.Create(&defaults).Error
}

func seedDefaultAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	admin := User{
		Email:        cfg.AdminEmail,
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         10,
		Status:       1,
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "email"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"username":      admin.Username,
			"password_hash": admin.PasswordHash,
			"role":          admin.Role,
			"status":        admin.Status,
			"updated_at":    time.Now(),
		}),
	}).Create(&admin).Error
}

func openDB(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "", "sqlite":
		if err := os.MkdirAll("data", 0755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
		return gorm.Open(sqlite.Open("./data/image_show.db"), &gorm.Config{})
	case "postgres":
		if dsn == "" {
			return nil, fmt.Errorf("DATABASE_DSN is required when DB_DRIVER=postgres")
		}
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q", driver)
	}
}
