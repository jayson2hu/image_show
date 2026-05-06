package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
		"title":       "维护通知",
		"content":     "今晚 23:00 进行短暂维护",
		"status":      1,
		"notify_mode": "popup",
		"sort_order":  1,
		"starts_at":   time.Now().Add(-time.Hour).Format(time.RFC3339),
		"ends_at":     time.Now().Add(time.Hour).Format(time.RFC3339),
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create announcement=%d body=%s", create.Code, create.Body.String())
	}
	var item model.Announcement
	if err := json.Unmarshal(create.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode announcement: %v", err)
	}
	if item.Title != "维护通知" || item.Content == "" || item.NotifyMode != "popup" {
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
		"title":       "维护完成",
		"content":     "服务已恢复",
		"status":      2,
		"notify_mode": "silent",
		"sort_order":  2,
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

func TestUserAnnouncementsAndReadStatus(t *testing.T) {
	engine := setupAuthTest(t)
	adminToken := createTokenForRole(t, 10)
	userToken := createTokenForRole(t, 1)

	future := adminJSON(engine, http.MethodPost, "/api/admin/announcements", map[string]interface{}{
		"title":      "未来公告",
		"content":    "稍后展示",
		"status":     1,
		"starts_at":  time.Now().Add(time.Hour).Format(time.RFC3339),
		"sort_order": 1,
	}, adminToken)
	if future.Code != http.StatusOK {
		t.Fatalf("create future announcement=%d body=%s", future.Code, future.Body.String())
	}

	create := adminJSON(engine, http.MethodPost, "/api/admin/announcements", map[string]interface{}{
		"title":       "系统公告",
		"content":     "请查看公告中心",
		"status":      1,
		"notify_mode": "popup",
		"sort_order":  0,
	}, adminToken)
	if create.Code != http.StatusOK {
		t.Fatalf("create active announcement=%d body=%s", create.Code, create.Body.String())
	}
	var item model.Announcement
	if err := json.Unmarshal(create.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode announcement: %v", err)
	}

	list := adminRequest(engine, http.MethodGet, "/api/announcements", userToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list user announcements=%d body=%s", list.Code, list.Body.String())
	}
	var listResp struct {
		Items []struct {
			ID     int64      `json:"id"`
			ReadAt *time.Time `json:"read_at"`
		} `json:"items"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode user announcement list: %v", err)
	}
	if len(listResp.Items) != 1 || listResp.Items[0].ID != item.ID || listResp.Items[0].ReadAt != nil {
		t.Fatalf("unexpected user announcement list: %#v", listResp.Items)
	}

	read := adminRequest(engine, http.MethodPost, "/api/announcements/"+jsonNumber(item.ID)+"/read", userToken)
	if read.Code != http.StatusOK {
		t.Fatalf("mark read=%d body=%s", read.Code, read.Body.String())
	}

	listAgain := adminRequest(engine, http.MethodGet, "/api/announcements", userToken)
	if listAgain.Code != http.StatusOK {
		t.Fatalf("list after read=%d body=%s", listAgain.Code, listAgain.Body.String())
	}
	var listAgainResp struct {
		Items []struct {
			ID     int64      `json:"id"`
			ReadAt *time.Time `json:"read_at"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listAgain.Body.Bytes(), &listAgainResp); err != nil {
		t.Fatalf("decode user announcement list after read: %v", err)
	}
	if len(listAgainResp.Items) != 1 || listAgainResp.Items[0].ReadAt == nil {
		t.Fatalf("expected read_at after mark read: %#v", listAgainResp.Items)
	}
}
