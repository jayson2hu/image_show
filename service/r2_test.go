package service

import (
	"strings"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

func TestBuildR2KeyForAnonymousAndUser(t *testing.T) {
	setupServiceDB(t)
	userID := int64(7)
	generation := &model.Generation{
		ID:        42,
		UserID:    &userID,
		CreatedAt: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
	}
	if key := BuildR2Key(generation); key != "generations/free/2026-04/user-7-42.png" {
		t.Fatalf("unexpected free user key: %s", key)
	}
	if err := model.DB.Create(&model.CreditLog{UserID: userID, Type: 3, Amount: 1, Balance: 1}).Error; err != nil {
		t.Fatalf("create paid credit log: %v", err)
	}
	if key := BuildR2Key(generation); key != "generations/paid/2026-04/user-7-42.png" {
		t.Fatalf("unexpected paid user key: %s", key)
	}

	generation.UserID = nil
	generation.AnonymousID = "ip/fingerprint value"
	if key := BuildR2Key(generation); key != "generations/free/2026-04/anon-ip-fingerprint-value-42.png" {
		t.Fatalf("unexpected anonymous key: %s", key)
	}
}

func TestAdminTopupPromotesFreeR2KeysToPaid(t *testing.T) {
	setupServiceDB(t)
	config.AppConfig.R2Endpoint = ""
	operatorID := int64(1)
	user := model.User{Email: "paid@example.com", Status: 1, Credits: 0}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	generation := model.Generation{
		UserID: &user.ID,
		R2Key:  "generations/free/2026-04/user-2-99.png",
		Status: 3,
	}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if err := AdminTopup(user.ID, operatorID, 10, "paid"); err != nil {
		t.Fatalf("admin topup: %v", err)
	}
	var updated model.Generation
	if err := model.DB.First(&updated, generation.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if updated.R2Key != "generations/paid/2026-04/user-2-99.png" {
		t.Fatalf("expected paid key, got %s", updated.R2Key)
	}
}

func TestStoreGeneratedImageWithoutR2ReturnsDataURL(t *testing.T) {
	setupServiceDB(t)
	config.AppConfig.R2Endpoint = ""

	generation := model.Generation{Prompt: "test", Quality: "low", Size: "1024x1024"}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	result := mockImageResult()

	imageURL, key, err := StoreGeneratedImage(generation.ID, result)
	if err != nil {
		t.Fatalf("StoreGeneratedImage: %v", err)
	}
	if key != "" || !strings.HasPrefix(imageURL, "data:image/png;base64,") {
		t.Fatalf("unexpected local store result url=%s key=%s", imageURL, key)
	}
}

func TestR2SettingsPreferAdminSettingsOverEnv(t *testing.T) {
	setupServiceDB(t)
	config.AppConfig.R2Endpoint = "https://env.r2.cloudflarestorage.com"
	config.AppConfig.R2AccessKey = "env-access"
	config.AppConfig.R2SecretKey = "env-secret"
	config.AppConfig.R2Bucket = "env-bucket"
	config.AppConfig.R2PublicURL = "https://env-cdn.example.com"
	if err := model.DB.Create(&model.Setting{Key: "r2_endpoint", Value: "https://admin.r2.cloudflarestorage.com"}).Error; err != nil {
		t.Fatalf("create endpoint setting: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "r2_access_key", Value: "admin-access"}).Error; err != nil {
		t.Fatalf("create access key setting: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "r2_secret_key", Value: "admin-secret"}).Error; err != nil {
		t.Fatalf("create secret key setting: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "r2_bucket", Value: "admin-bucket"}).Error; err != nil {
		t.Fatalf("create bucket setting: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "r2_public_url", Value: "https://admin-cdn.example.com"}).Error; err != nil {
		t.Fatalf("create public url setting: %v", err)
	}

	settings := r2SettingsFromConfig()
	if settings.endpoint != "https://admin.r2.cloudflarestorage.com" || settings.accessKey != "admin-access" || settings.secretKey != "admin-secret" || settings.bucket != "admin-bucket" || settings.publicURL != "https://admin-cdn.example.com" {
		t.Fatalf("unexpected settings: %+v", settings)
	}
}

func TestRegisterClaimsAnonymousGenerations(t *testing.T) {
	setupServiceDB(t)
	email := "claim@example.com"
	anonymousID := "anon-claim"
	if err := model.DB.Create(&model.Generation{AnonymousID: anonymousID, Prompt: "p", Quality: "low", Size: "1024x1024"}).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if err := SendVerificationCode(email); err != nil {
		t.Fatalf("send code: %v", err)
	}

	result, err := Register(email, "password123", PeekVerificationCode(email), "1.2.3.4", anonymousID)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	var generation model.Generation
	if err := model.DB.Where("anonymous_id = ?", anonymousID).First(&generation).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.UserID == nil || *generation.UserID != result.User.ID {
		t.Fatalf("generation not claimed: %+v", generation)
	}
}
