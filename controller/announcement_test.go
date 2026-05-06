package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jayson2hu/image-show/model"
)

func TestAnnouncementAdminCRUDAndPublicActive(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)

	empty := adminRequest(engine, http.MethodGet, "/api/announcement", "")
	if empty.Code != http.StatusOK {
		t.Fatalf("empty active announcement=%d body=%s", empty.Code, empty.Body.String())
	}

	create := adminJSON(engine, http.MethodPost, "/api/admin/announcements", map[string]interface{}{
		"title":      "维护通知",
		"content":    "今晚 23:00 进行短暂维护",
		"status":     1,
		"sort_order": 1,
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create announcement=%d body=%s", create.Code, create.Body.String())
	}
	var item model.Announcement
	if err := json.Unmarshal(create.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode announcement: %v", err)
	}
	if item.Title != "维护通知" || item.Content == "" {
		t.Fatalf("unexpected created announcement: %#v", item)
	}

	active := adminRequest(engine, http.MethodGet, "/api/announcement", "")
	if active.Code != http.StatusOK {
		t.Fatalf("active announcement=%d body=%s", active.Code, active.Body.String())
	}
	var activeResp struct {
		Item *model.Announcement `json:"item"`
	}
	if err := json.Unmarshal(active.Body.Bytes(), &activeResp); err != nil {
		t.Fatalf("decode active announcement: %v", err)
	}
	if activeResp.Item == nil || activeResp.Item.ID != item.ID {
		t.Fatalf("unexpected active announcement: %#v", activeResp.Item)
	}

	update := adminJSON(engine, http.MethodPut, "/api/admin/announcements/"+jsonNumber(item.ID), map[string]interface{}{
		"title":      "维护完成",
		"content":    "服务已恢复",
		"status":     2,
		"sort_order": 2,
	}, token)
	if update.Code != http.StatusOK {
		t.Fatalf("update announcement=%d body=%s", update.Code, update.Body.String())
	}

	list := adminRequest(engine, http.MethodGet, "/api/admin/announcements", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list announcement=%d body=%s", list.Code, list.Body.String())
	}

	inactive := adminRequest(engine, http.MethodGet, "/api/announcement", "")
	if inactive.Code != http.StatusOK {
		t.Fatalf("inactive active announcement=%d body=%s", inactive.Code, inactive.Body.String())
	}
	var inactiveResp struct {
		Item *model.Announcement `json:"item"`
	}
	if err := json.Unmarshal(inactive.Body.Bytes(), &inactiveResp); err != nil {
		t.Fatalf("decode inactive announcement: %v", err)
	}
	if inactiveResp.Item != nil {
		t.Fatalf("expected no active announcement, got %#v", inactiveResp.Item)
	}

	del := adminRequest(engine, http.MethodDelete, "/api/admin/announcements/"+jsonNumber(item.ID), token)
	if del.Code != http.StatusOK {
		t.Fatalf("delete announcement=%d body=%s", del.Code, del.Body.String())
	}
}
