package service

import (
	"testing"

	"github.com/jayson2hu/image-show/model"
)

func TestRefundGenerationCreditsForFailedSave(t *testing.T) {
	setupServiceDB(t)
	user := model.User{Email: "save-refund@example.com", Role: 1, Status: 1, Credits: 0}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	generation := model.Generation{UserID: &user.ID, Prompt: "save fail", Quality: "medium", Size: "1024x1024", CreditsCost: 1, Status: 2}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}

	refundGenerationCredits(generation.ID)

	balance, err := GetBalance(user.ID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance != 1 {
		t.Fatalf("expected refunded balance 1, got %v", balance)
	}
}
