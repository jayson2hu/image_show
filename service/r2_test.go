package service

import (
	"strings"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

func TestBuildR2KeyForAnonymousAndUser(t *testing.T) {
	userID := int64(7)
	generation := &model.Generation{
		ID:        42,
		UserID:    &userID,
		CreatedAt: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
	}
	if key := BuildR2Key(generation); key != "generations/user-7/2026-04/42.png" {
		t.Fatalf("unexpected user key: %s", key)
	}

	generation.UserID = nil
	generation.AnonymousID = "ip/fingerprint value"
	if key := BuildR2Key(generation); key != "generations/anon-ip-fingerprint-value/2026-04/42.png" {
		t.Fatalf("unexpected anonymous key: %s", key)
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
