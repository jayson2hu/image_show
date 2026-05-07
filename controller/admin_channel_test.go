package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/model"
)

func TestAdminChannelCRUD(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)

	create := adminJSON(engine, http.MethodPost, "/api/admin/channels", map[string]interface{}{
		"name":     "sub2api",
		"base_url": "http://sub2api:8080/",
		"api_key":  "key",
		"status":   1,
		"weight":   2,
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var channel model.Channel
	if err := json.Unmarshal(create.Body.Bytes(), &channel); err != nil {
		t.Fatalf("decode channel: %v", err)
	}
	if channel.BaseURL != "http://sub2api:8080" || channel.Weight != 2 {
		t.Fatalf("unexpected channel: %+v", channel)
	}

	update := adminJSON(engine, http.MethodPut, "/api/admin/channels/"+jsonNumber(channel.ID), map[string]interface{}{
		"name":     "disabled",
		"base_url": "http://sub2api:8080",
		"status":   2,
		"weight":   1,
	}, token)
	if update.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", update.Code, update.Body.String())
	}

	list := adminRequest(engine, http.MethodGet, "/api/admin/channels", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if !bytes.Contains(list.Body.Bytes(), []byte(`"recent_success_count":0`)) || !bytes.Contains(list.Body.Bytes(), []byte(`"recent_failed_count":0`)) {
		t.Fatalf("list missing channel stats: %s", list.Body.String())
	}

	del := adminRequest(engine, http.MethodDelete, "/api/admin/channels/"+jsonNumber(channel.ID), token)
	if del.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", del.Code, del.Body.String())
	}
}

func TestAdminChannelRecentGenerationStats(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)
	channel := model.Channel{Name: "stats", BaseURL: "http://sub2api:8080", Status: 1, Weight: 1}
	if err := model.DB.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}
	old := time.Now().Add(-25 * time.Hour)
	if err := model.DB.Create(&[]model.Generation{
		{Prompt: "ok", Size: "1024x1024", Status: 3, ChannelID: &channel.ID, ChannelName: channel.Name, CreatedAt: time.Now()},
		{Prompt: "bad", Size: "1024x1024", Status: 4, ChannelID: &channel.ID, ChannelName: channel.Name, CreatedAt: time.Now()},
		{Prompt: "old", Size: "1024x1024", Status: 4, ChannelID: &channel.ID, ChannelName: channel.Name, CreatedAt: old},
	}).Error; err != nil {
		t.Fatalf("create generations: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/admin/channels", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"recent_success_count":1`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"recent_failed_count":1`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"recent_failure_rate":0.5`)) {
		t.Fatalf("unexpected stats response: %s", rec.Body.String())
	}
}

func TestAdminChannelTestEndpoint(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	channel := model.Channel{Name: "ok", BaseURL: server.URL, APIKey: "key", Status: 1, Weight: 1}
	if err := model.DB.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}

	rec := adminRequest(engine, http.MethodPost, "/api/admin/channels/"+jsonNumber(channel.ID)+"/test", token)
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"ok":true`)) {
		t.Fatalf("test status=%d body=%s", rec.Code, rec.Body.String())
	}
	var tested model.Channel
	if err := model.DB.First(&tested, channel.ID).Error; err != nil {
		t.Fatalf("load tested channel: %v", err)
	}
	if tested.LastTestAt == nil || !tested.LastTestSuccess || tested.LastTestStatus != http.StatusOK || tested.LastTestError != "" {
		t.Fatalf("unexpected channel test result: %+v", tested)
	}
}

func adminJSON(engine http.Handler, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
