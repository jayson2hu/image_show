package service

import (
	"errors"
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

func TestUpdateGenerationChannel(t *testing.T) {
	setupServiceDB(t)
	channel := model.Channel{Name: "primary", BaseURL: "http://primary", Status: 1, Weight: 1}
	if err := model.DB.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}
	generation := model.Generation{Prompt: "channel", Quality: "medium", Size: "1024x1024", Status: 1}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	channelID := channel.ID

	updateGenerationChannel(generation.ID, ChannelUse{ID: &channelID, Name: channel.Name})

	var updated model.Generation
	if err := model.DB.First(&updated, generation.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if updated.ChannelID == nil || *updated.ChannelID != channel.ID || updated.ChannelName != "primary" {
		t.Fatalf("unexpected channel attribution: %+v", updated)
	}
}

func TestChannelFromError(t *testing.T) {
	channelID := int64(12)
	err := ChannelError{Channel: ChannelUse{ID: &channelID, Name: "fallback"}, Err: errors.New("failed")}

	channel := channelFromError(err)

	if channel.ID == nil || *channel.ID != channelID || channel.Name != "fallback" {
		t.Fatalf("unexpected channel from error: %+v", channel)
	}
}
