package controller_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

func TestCreateMessageCreatesGenerationAndUpdatesConversation(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true
	token := createGenerationUser(t, 3)
	userID := tokenUserID(t, token)

	conversation := model.Conversation{UserID: userID, Title: "New chat", LastMsgAt: time.Now()}
	if err := model.DB.Create(&conversation).Error; err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	rec := adminJSON(engine, http.MethodPost, "/api/conversations/"+strconv.FormatInt(conversation.ID, 10)+"/messages", map[string]interface{}{
		"prompt":      "a glass cabin near the lake",
		"size":        "square",
		"style_id":    "realistic",
		"scene_id":    "1",
		"layered":     true,
		"layer_count": 4,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		Message      model.Message `json:"message"`
		GenerationID int64         `json:"generation_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Message.ID == 0 || createResp.GenerationID == 0 {
		t.Fatalf("expected message and generation ids, got %+v", createResp)
	}
	if createResp.Message.GenerationID == nil || *createResp.Message.GenerationID != createResp.GenerationID {
		t.Fatalf("message not linked to generation: %+v", createResp.Message)
	}
	if createResp.Message.TaskKind != "text2img" || createResp.Message.Size != "1024x1024" || !createResp.Message.Layered || createResp.Message.LayerCount != 4 {
		t.Fatalf("unexpected message payload: %+v", createResp.Message)
	}

	var generation model.Generation
	if err := model.DB.First(&generation, createResp.GenerationID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.MessageID == nil || *generation.MessageID != createResp.Message.ID {
		t.Fatalf("generation not linked to message: %+v", generation)
	}
	waitGenerationStatus(t, createResp.GenerationID, 3)

	var updated model.Conversation
	if err := model.DB.First(&updated, conversation.ID).Error; err != nil {
		t.Fatalf("load conversation: %v", err)
	}
	if updated.MsgCount != 1 || updated.Title != "a glass cabin near the lake" || !updated.IsLayered {
		t.Fatalf("conversation not updated: %+v", updated)
	}
}

func TestListMessagesUsesConversationOwnership(t *testing.T) {
	engine := setupAuthTest(t)
	ownerToken := createTokenForRole(t, 1)
	ownerID := tokenUserID(t, ownerToken)
	otherToken := createTokenForRole(t, 1)
	otherID := tokenUserID(t, otherToken)

	owned := model.Conversation{UserID: ownerID, Title: "owned", LastMsgAt: time.Now()}
	other := model.Conversation{UserID: otherID, Title: "other", LastMsgAt: time.Now()}
	if err := model.DB.Create(&owned).Error; err != nil {
		t.Fatalf("create owned conversation: %v", err)
	}
	if err := model.DB.Create(&other).Error; err != nil {
		t.Fatalf("create other conversation: %v", err)
	}

	messages := []model.Message{
		{ConversationID: owned.ID, UserID: ownerID, Prompt: "first", TaskKind: "text2img", CreatedAt: time.Now().Add(-time.Minute)},
		{ConversationID: owned.ID, UserID: ownerID, Prompt: "second", TaskKind: "text2img", CreatedAt: time.Now()},
		{ConversationID: other.ID, UserID: otherID, Prompt: "hidden", TaskKind: "text2img", CreatedAt: time.Now()},
	}
	if err := model.DB.Create(&messages).Error; err != nil {
		t.Fatalf("create messages: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/conversations/"+strconv.FormatInt(owned.ID, 10)+"/messages", ownerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Items []model.Message `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Items) != 2 || listResp.Items[0].Prompt != "first" || listResp.Items[1].Prompt != "second" {
		t.Fatalf("unexpected list response: %+v", listResp.Items)
	}

	forbidden := adminRequest(engine, http.MethodGet, "/api/conversations/"+strconv.FormatInt(owned.ID, 10)+"/messages", otherToken)
	if forbidden.Code != http.StatusNotFound {
		t.Fatalf("expected ownership 404, got %d body=%s", forbidden.Code, forbidden.Body.String())
	}
}

func TestCreateMultipartMessageCreatesImageEditGeneration(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true
	token := createGenerationUser(t, 3)
	userID := tokenUserID(t, token)

	conversation := model.Conversation{UserID: userID, Title: "edit chat", LastMsgAt: time.Now()}
	if err := model.DB.Create(&conversation).Error; err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("prompt", "restore this image")
	_ = writer.WriteField("size", "square")
	_ = writer.WriteField("style_id", "restore")
	part, err := writer.CreateFormFile("image", "source.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(testPNGBytes); err != nil {
		t.Fatalf("write image: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/conversations/"+strconv.FormatInt(conversation.ID, 10)+"/messages", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create multipart message status=%d body=%s", rec.Code, rec.Body.String())
	}

	var createResp struct {
		Message      model.Message `json:"message"`
		GenerationID int64         `json:"generation_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Message.TaskKind != "img2img_generic" || createResp.Message.AttachmentName != "source.png" || createResp.Message.AttachmentSize == 0 {
		t.Fatalf("unexpected multipart message: %+v", createResp.Message)
	}

	var generation model.Generation
	if err := model.DB.First(&generation, createResp.GenerationID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.Mode != "edit" || generation.MessageID == nil || *generation.MessageID != createResp.Message.ID {
		t.Fatalf("unexpected edit generation: %+v", generation)
	}
	waitGenerationStatus(t, createResp.GenerationID, 3)
}
