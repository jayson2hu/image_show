package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

func TestSub2APIClientGenerateImagePassesHeaders(t *testing.T) {
	var gotIP string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = r.Header.Get("X-Real-IP")
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Fatalf("missing authorization header")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "abc"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "key", map[string]string{"X-Test": "ok"})
	result, err := client.GenerateImage("prompt", "medium", "1024x1024", "1.2.3.4")
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if result.Base64Data != "abc" || gotIP != "1.2.3.4" {
		t.Fatalf("unexpected result=%+v ip=%s", result, gotIP)
	}
}

func TestSub2APIClientGenerateImageRetriesTransientDecodeFailure(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			_, _ = w.Write([]byte(`{"data":[`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "retry-ok"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "", nil)
	result, err := client.GenerateImage("prompt", "medium", "1024x1024", "")
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if result.Base64Data != "retry-ok" || attempts != 2 {
		t.Fatalf("unexpected result=%+v attempts=%d", result, attempts)
	}
}

func TestGenerateImageViaChannelsFallsBack(t *testing.T) {
	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusBadGateway)
	}))
	defer fail.Close()
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"url": "https://example.com/image.png"}},
		})
	}))
	defer ok.Close()

	setupServiceDB(t)
	if err := model.DB.Create(&model.Channel{Name: "fail", BaseURL: fail.URL, Status: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create fail channel: %v", err)
	}
	if err := model.DB.Create(&model.Channel{Name: "ok", BaseURL: ok.URL, Status: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create ok channel: %v", err)
	}

	result, err := GenerateImageViaChannels("prompt", "medium", "1024x1024", "1.2.3.4")
	if err != nil {
		t.Fatalf("GenerateImageViaChannels: %v", err)
	}
	if result.URL != "https://example.com/image.png" {
		t.Fatalf("unexpected url: %s", result.URL)
	}
}

func TestGenerateImageMockMode(t *testing.T) {
	config.AppConfig = &config.Config{MockSub2API: true}
	t.Cleanup(func() { config.AppConfig = nil })

	result, err := NewSub2APIClient("http://unused", "", nil).GenerateImage("prompt", "low", "1024x1024", "")
	if err != nil {
		t.Fatalf("mock GenerateImage: %v", err)
	}
	if result.Base64Data == "" {
		t.Fatal("expected mock base64 data")
	}
}

func setupServiceDB(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = model.CloseDB()
		model.DB = nil
		config.AppConfig = nil
		_ = os.Chdir(originalDir)
	})
	config.AppConfig = &config.Config{DBDriver: "sqlite", Sub2APIBaseURL: "http://sub2api:8080"}
	if err := model.InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}
}
