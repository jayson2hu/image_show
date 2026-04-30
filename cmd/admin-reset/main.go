package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
)

func main() {
	email := flag.String("email", "", "admin email")
	password := flag.String("password", "", "admin password")
	flag.Parse()

	common.LoadEnv()
	cfg := config.LoadConfig()
	if *email != "" {
		cfg.AdminEmail = *email
	}
	if *password != "" {
		cfg.AdminPassword = *password
	}
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		fmt.Fprintln(os.Stderr, "admin email and password are required")
		os.Exit(1)
	}
	if err := model.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "init db: %v\n", err)
		os.Exit(1)
	}
	defer model.CloseDB()

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}
	admin := model.User{
		Email:        cfg.AdminEmail,
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         10,
		Status:       1,
	}
	if err := model.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "email"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"username":      admin.Username,
			"password_hash": admin.PasswordHash,
			"role":          admin.Role,
			"status":        admin.Status,
		}),
	}).Create(&admin).Error; err != nil {
		fmt.Fprintf(os.Stderr, "reset admin: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("admin reset: %s\n", cfg.AdminEmail)
}
