package controller_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/model"
)

func TestConversationCRUD(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)

	create := adminJSON(engine, http.MethodPost, "/api/conversations", map[string]string{"title": "测试会话"}, token)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var created model.Conversation
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created conversation: %v", err)
	}
	if created.ID == 0 || created.Title != "测试会话" || created.UserID != tokenUserID(t, token) || created.IsDeleted {
		t.Fatalf("unexpected created conversation: %+v", created)
	}

	detail := adminRequest(engine, http.MethodGet, "/api/conversations/"+strconv.FormatInt(created.ID, 10), token)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}

	rename := adminJSON(engine, http.MethodPatch, "/api/conversations/"+strconv.FormatInt(created.ID, 10), map[string]string{"title": "重命名会话"}, token)
	if rename.Code != http.StatusOK {
		t.Fatalf("rename status=%d body=%s", rename.Code, rename.Body.String())
	}
	var renamed model.Conversation
	if err := json.Unmarshal(rename.Body.Bytes(), &renamed); err != nil {
		t.Fatalf("decode renamed conversation: %v", err)
	}
	if renamed.Title != "重命名会话" {
		t.Fatalf("unexpected renamed conversation: %+v", renamed)
	}

	deleted := adminRequest(engine, http.MethodDelete, "/api/conversations/"+strconv.FormatInt(created.ID, 10), token)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}

	list := adminRequest(engine, http.MethodGet, "/api/conversations", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var listResp struct {
		Items []model.Conversation `json:"items"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Items) != 0 {
		t.Fatalf("expected deleted conversation hidden from list, got %+v", listResp.Items)
	}

	var stored model.Conversation
	if err := model.DB.First(&stored, created.ID).Error; err != nil {
		t.Fatalf("soft deleted row should remain: %v", err)
	}
	if !stored.IsDeleted {
		t.Fatalf("expected soft deleted row, got %+v", stored)
	}
}

func TestConversationListSearchAndCursor(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	otherToken := createTokenForRole(t, 1)
	otherUserID := tokenUserID(t, otherToken)
	now := time.Now()

	items := []model.Conversation{
		{UserID: userID, Title: "alpha prompt", LastMsgAt: now.Add(-3 * time.Minute)},
		{UserID: userID, Title: "beta prompt", LastMsgAt: now.Add(-2 * time.Minute)},
		{UserID: userID, Title: "gamma idea", LastMsgAt: now.Add(-time.Minute)},
		{UserID: otherUserID, Title: "alpha hidden", LastMsgAt: now},
		{UserID: userID, Title: "deleted alpha", LastMsgAt: now, IsDeleted: true},
	}
	if err := model.DB.Create(&items).Error; err != nil {
		t.Fatalf("create conversations: %v", err)
	}

	search := adminRequest(engine, http.MethodGet, "/api/conversations?q=prompt", token)
	if search.Code != http.StatusOK {
		t.Fatalf("search status=%d body=%s", search.Code, search.Body.String())
	}
	var searchResp struct {
		Items []model.Conversation `json:"items"`
	}
	if err := json.Unmarshal(search.Body.Bytes(), &searchResp); err != nil {
		t.Fatalf("decode search: %v", err)
	}
	if len(searchResp.Items) != 2 {
		t.Fatalf("expected two prompt conversations, got %+v", searchResp.Items)
	}
	for _, item := range searchResp.Items {
		if item.UserID != userID || item.IsDeleted || item.Title == "alpha hidden" {
			t.Fatalf("unexpected search item: %+v", item)
		}
	}

	page1 := adminRequest(engine, http.MethodGet, "/api/conversations?limit=2", token)
	if page1.Code != http.StatusOK {
		t.Fatalf("page1 status=%d body=%s", page1.Code, page1.Body.String())
	}
	var page1Resp struct {
		Items      []model.Conversation `json:"items"`
		NextCursor string               `json:"next_cursor"`
	}
	if err := json.Unmarshal(page1.Body.Bytes(), &page1Resp); err != nil {
		t.Fatalf("decode page1: %v", err)
	}
	if len(page1Resp.Items) != 2 || page1Resp.NextCursor == "" {
		t.Fatalf("unexpected page1: %+v", page1Resp)
	}

	page2 := adminRequest(engine, http.MethodGet, "/api/conversations?limit=2&cursor="+page1Resp.NextCursor, token)
	if page2.Code != http.StatusOK {
		t.Fatalf("page2 status=%d body=%s", page2.Code, page2.Body.String())
	}
	var page2Resp struct {
		Items      []model.Conversation `json:"items"`
		NextCursor string               `json:"next_cursor"`
	}
	if err := json.Unmarshal(page2.Body.Bytes(), &page2Resp); err != nil {
		t.Fatalf("decode page2: %v", err)
	}
	if len(page2Resp.Items) != 1 || page2Resp.NextCursor != "" {
		t.Fatalf("unexpected page2: %+v", page2Resp)
	}
}

func TestConversationAuthAndOwnership(t *testing.T) {
	engine := setupAuthTest(t)
	ownerToken := createTokenForRole(t, 1)
	ownerID := tokenUserID(t, ownerToken)
	otherToken := createTokenForRole(t, 1)
	conversation := model.Conversation{UserID: ownerID, Title: "private", LastMsgAt: time.Now()}
	if err := model.DB.Create(&conversation).Error; err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	path := "/api/conversations/" + strconv.FormatInt(conversation.ID, 10)

	unauthorized := adminRequest(engine, http.MethodGet, path, "")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	for _, tc := range []struct {
		name   string
		method string
		rec    func() int
	}{
		{name: "get", method: http.MethodGet, rec: func() int { return adminRequest(engine, http.MethodGet, path, otherToken).Code }},
		{name: "patch", method: http.MethodPatch, rec: func() int {
			return adminJSON(engine, http.MethodPatch, path, map[string]string{"title": "x"}, otherToken).Code
		}},
		{name: "delete", method: http.MethodDelete, rec: func() int { return adminRequest(engine, http.MethodDelete, path, otherToken).Code }},
	} {
		if code := tc.rec(); code != http.StatusNotFound {
			t.Fatalf("%s expected ownership 404, got %d", tc.name, code)
		}
	}

	missing := adminRequest(engine, http.MethodGet, "/api/conversations/999999", ownerToken)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected missing 404, got %d body=%s", missing.Code, missing.Body.String())
	}
}
