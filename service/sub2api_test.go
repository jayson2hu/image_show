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
	var gotRequest imageGenerationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIP = r.Header.Get("X-Real-IP")
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Fatalf("missing authorization header")
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "abc"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "key", map[string]string{"X-Test": "ok"})
	result, err := client.GenerateImage("prompt", "medium", "1024x1024", "1.2.3.4", ImageOptions{})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if result.Base64Data != "abc" || gotIP != "1.2.3.4" {
		t.Fatalf("unexpected result=%+v ip=%s", result, gotIP)
	}
	if gotRequest.Model != "gpt-image-2" || gotRequest.Size != "1024x1024" {
		t.Fatalf("unexpected request: %+v", gotRequest)
	}
	if gotRequest.OutputFormat != "" || gotRequest.OutputCompression != nil || gotRequest.Background != "" {
		t.Fatalf("expected empty output options, got %+v", gotRequest)
	}
}

func TestSub2APIClientGenerateImagePassesOutputOptions(t *testing.T) {
	compression := 82
	var gotRequest imageGenerationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "abc"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "", nil)
	_, err := client.GenerateImage("prompt", "medium", "1024x1024", "", ImageOptions{
		OutputFormat:      "webp",
		OutputCompression: &compression,
		Background:        "transparent",
	})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if gotRequest.OutputFormat != "webp" || gotRequest.OutputCompression == nil || *gotRequest.OutputCompression != 82 || gotRequest.Background != "transparent" {
		t.Fatalf("unexpected output options: %+v", gotRequest)
	}
}

func TestSub2APIClientGenerateImageUsesConfiguredModel(t *testing.T) {
	config.AppConfig = &config.Config{ImageModel: "gpt-image-1.5"}
	t.Cleanup(func() { config.AppConfig = nil })

	var gotRequest imageGenerationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "configured-model"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "", nil)
	if _, err := client.GenerateImage("prompt", "medium", "1024x1024", "", ImageOptions{}); err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if gotRequest.Model != "gpt-image-1.5" {
		t.Fatalf("expected configured model, got %+v", gotRequest)
	}
}

func TestSub2APIClientEditImageUsesMultipartEditsEndpoint(t *testing.T) {
	var gotPath string
	var gotModel string
	var gotPrompt string
	var gotSize string
	var gotOutputFormat string
	var gotBackground string
	var gotOutputCompression string
	var gotFileContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		gotModel = r.FormValue("model")
		gotPrompt = r.FormValue("prompt")
		gotSize = r.FormValue("size")
		gotOutputFormat = r.FormValue("output_format")
		gotBackground = r.FormValue("background")
		gotOutputCompression = r.FormValue("output_compression")
		file, header, err := r.FormFile("image")
		if err != nil {
			t.Fatalf("read image: %v", err)
		}
		defer file.Close()
		gotFileContentType = header.Header.Get("Content-Type")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "edit-ok"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "", nil)
	compression := 76
	result, err := client.EditImage("change it", "medium", "1536x1024", "1.2.3.4", []byte("image-bytes"), "source.png", "image/png", ImageOptions{
		OutputFormat:      "webp",
		OutputCompression: &compression,
		Background:        "transparent",
	})
	if err != nil {
		t.Fatalf("EditImage: %v", err)
	}
	if result.Base64Data != "edit-ok" || gotPath != "/v1/images/edits" || gotModel != "gpt-image-2" || gotPrompt != "change it" || gotSize != "1536x1024" || gotFileContentType != "image/png" {
		t.Fatalf("unexpected edit request result=%+v path=%s model=%s prompt=%s size=%s contentType=%s", result, gotPath, gotModel, gotPrompt, gotSize, gotFileContentType)
	}
	if gotOutputFormat != "webp" || gotBackground != "transparent" || gotOutputCompression != "76" {
		t.Fatalf("unexpected edit output options format=%s background=%s compression=%s", gotOutputFormat, gotBackground, gotOutputCompression)
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
	result, err := client.GenerateImage("prompt", "medium", "1024x1024", "", ImageOptions{})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if result.Base64Data != "retry-ok" || attempts != 2 {
		t.Fatalf("unexpected result=%+v attempts=%d", result, attempts)
	}
}

func TestSub2APIClientGenerateImageRetriesCloudflareTimeout(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(cloudflareTimeoutStatus)
			_, _ = w.Write([]byte(`error code: 524`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{{"b64_json": "retry-after-524"}},
		})
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "", nil)
	result, err := client.GenerateImage("prompt", "medium", "1024x1024", "", ImageOptions{})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if result.Base64Data != "retry-after-524" || attempts != 2 {
		t.Fatalf("unexpected result=%+v attempts=%d", result, attempts)
	}
}

func TestSub2APIClientGenerateImageMapsCloudflareTimeoutError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(cloudflareTimeoutStatus)
		_, _ = w.Write([]byte(`error code: 524`))
	}))
	defer server.Close()

	client := NewSub2APIClient(server.URL, "", nil)
	_, err := client.GenerateImage("prompt", "medium", "1024x1024", "", ImageOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sub2api upstream timeout: image generation exceeded upstream proxy timeout, please retry or switch channel" {
		t.Fatalf("unexpected error: %v", err)
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

	result, err := GenerateImageViaChannels("prompt", "medium", "1024x1024", "1.2.3.4", ImageOptions{})
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

	result, err := NewSub2APIClient("http://unused", "", nil).GenerateImage("prompt", "low", "1024x1024", "", ImageOptions{})
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
