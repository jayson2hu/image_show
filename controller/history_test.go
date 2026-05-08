package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/jayson2hu/image-show/model"
)

func TestUserGenerationHistoryDetailAndSoftDelete(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	otherID := userID + 100
	generation := model.Generation{
		UserID:   &userID,
		Prompt:   "history",
		Quality:  "low",
		Size:     "1024x1024",
		Status:   3,
		ImageURL: "data:image/png;base64,aGVsbG8=",
	}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if err := model.DB.Create(&model.Generation{UserID: &otherID, Prompt: "other", Status: 3}).Error; err != nil {
		t.Fatalf("create other generation: %v", err)
	}

	list := adminRequest(engine, http.MethodGet, "/api/generations?page=1&pageSize=10", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var listResp struct {
		Total int64 `json:"total"`
	}
	_ = json.Unmarshal(list.Body.Bytes(), &listResp)
	if listResp.Total != 1 {
		t.Fatalf("expected one owned generation, got %d", listResp.Total)
	}
	if strings.Contains(list.Body.String(), "data:image/png;base64,aGVsbG8=") {
		t.Fatalf("list response should not include inline image data: %s", list.Body.String())
	}

	image := adminRequest(engine, http.MethodGet, "/api/generations/"+jsonNumber(generation.ID)+"/image", token)
	if image.Code != http.StatusOK {
		t.Fatalf("image status=%d body=%s", image.Code, image.Body.String())
	}
	if got := image.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected image/png content type, got %s", got)
	}
	if image.Body.String() != "hello" {
		t.Fatalf("unexpected image body: %q", image.Body.String())
	}

	detail := adminRequest(engine, http.MethodGet, "/api/generations/"+jsonNumber(generation.ID), token)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}

	del := adminRequest(engine, http.MethodDelete, "/api/generations/"+jsonNumber(generation.ID), token)
	if del.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", del.Code, del.Body.String())
	}
	list = adminRequest(engine, http.MethodGet, "/api/generations", token)
	_ = json.Unmarshal(list.Body.Bytes(), &listResp)
	if listResp.Total != 0 {
		t.Fatalf("expected deleted generation hidden, got %d", listResp.Total)
	}
}

func TestUserGenerationHistoryFilters(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	otherID := userID + 100
	items := []model.Generation{
		{UserID: &userID, Prompt: "red book cover", Size: "1024x1024", Status: 3},
		{UserID: &userID, Prompt: "product photo", Size: "1536x1024", Status: 4},
		{UserID: &userID, Prompt: "red poster", Size: "1536x1024", Status: 3},
		{UserID: &otherID, Prompt: "red private", Size: "1536x1024", Status: 3},
	}
	if err := model.DB.Create(&items).Error; err != nil {
		t.Fatalf("create generations: %v", err)
	}

	list := adminRequest(engine, http.MethodGet, "/api/generations?keyword=red&status=3&size=1536x1024", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var listResp struct {
		Total int64              `json:"total"`
		Items []model.Generation `json:"items"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Items) != 1 || listResp.Items[0].Prompt != "red poster" {
		t.Fatalf("unexpected filtered response: total=%d items=%+v", listResp.Total, listResp.Items)
	}

	invalid := adminRequest(engine, http.MethodGet, "/api/generations?status=bad", token)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid status 400, got %d body=%s", invalid.Code, invalid.Body.String())
	}
}

func TestAdminGenerationBatchDelete(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)
	userID := int64(42)
	one := model.Generation{UserID: &userID, Prompt: "one", Status: 3}
	two := model.Generation{UserID: &userID, Prompt: "two", Status: 3}
	if err := model.DB.Create(&one).Error; err != nil {
		t.Fatalf("create one: %v", err)
	}
	if err := model.DB.Create(&two).Error; err != nil {
		t.Fatalf("create two: %v", err)
	}

	list := adminRequest(engine, http.MethodGet, "/api/admin/generations?user_id=42", token)
	if list.Code != http.StatusOK {
		t.Fatalf("admin list status=%d body=%s", list.Code, list.Body.String())
	}
	batch := adminJSON(engine, http.MethodDelete, "/api/admin/generations/batch", map[string]interface{}{
		"ids":       []int64{one.ID, two.ID},
		"delete_r2": false,
	}, token)
	if batch.Code != http.StatusOK {
		t.Fatalf("batch delete status=%d body=%s", batch.Code, batch.Body.String())
	}
	var count int64
	if err := model.DB.Model(&model.Generation{}).Where("is_deleted = ?", true).Count(&count).Error; err != nil {
		t.Fatalf("count deleted: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected two soft deleted generations, got %d", count)
	}
}
