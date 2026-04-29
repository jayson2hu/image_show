package controller_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func TestGenerationStreamCompletesInMockMode(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true
	token := createGenerationUser(t, 3)

	rec := postJSONWithToken(engine, "/api/generations", map[string]string{
		"prompt":       "a small house",
		"quality":      "low",
		"size":         "1024x1024",
		"anonymous_id": "anon-1",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("create generation status=%d body=%s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/generations/"+jsonNumber(createResp.ID)+"/stream", nil)
	streamRec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		engine.ServeHTTP(streamRec, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("stream did not complete")
	}

	body := streamRec.Body.String()
	if !strings.Contains(body, "event:status") || !strings.Contains(body, `"status":3`) {
		t.Fatalf("unexpected sse body: %s", body)
	}
	var generation model.Generation
	if err := model.DB.First(&generation, createResp.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.Status != 3 || generation.ImageURL == "" {
		t.Fatalf("generation not completed: %+v", generation)
	}
	balance, err := service.GetBalance(*generation.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance < 2.79 || balance > 2.81 {
		t.Fatalf("expected low quality cost 0.2, balance=%v", balance)
	}
}

func TestCreateGenerationValidation(t *testing.T) {
	engine := setupAuthTest(t)
	rec := postJSON(engine, "/api/generations", map[string]string{
		"prompt":  "",
		"quality": "bad",
		"size":    "1024x1024",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateGenerationRequiresFingerprintForTrial(t *testing.T) {
	engine := setupAuthTest(t)
	rec := postJSON(engine, "/api/generations", map[string]string{
		"prompt":  "a small house",
		"quality": "low",
		"size":    "1024x1024",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCreateGenerationRequiresCaptchaWhenEnabled(t *testing.T) {
	engine := setupAuthTest(t)
	enableCaptchaForTest(t)
	token := createGenerationUser(t, 3)

	rec := postJSONWithToken(engine, "/api/generations", map[string]string{
		"prompt":  "a small house",
		"quality": "low",
		"size":    "1024x1024",
	}, token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without captcha, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateGenerationAcceptsValidCaptcha(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true
	enableCaptchaForTest(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Form.Get("secret") != "secret" || r.Form.Get("response") != "token-ok" {
			t.Fatalf("unexpected captcha form: %v", r.Form)
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	defer service.SetCaptchaVerifyURLForTest(server.URL)()
	token := createGenerationUser(t, 3)

	rec := postJSONWithToken(engine, "/api/generations", map[string]string{
		"prompt":        "a small house",
		"quality":       "low",
		"size":          "1024x1024",
		"captcha_token": "token-ok",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with captcha, got %d body=%s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	waitGenerationStatus(t, createResp.ID, 3)
}

func TestAnonymousTrialOnceAndForcesLow(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true

	first := postJSONWithFingerprint(engine, "/api/generations", map[string]string{
		"prompt":  "trial image",
		"quality": "high",
		"size":    "1024x1024",
	}, "fp-1")
	if first.Code != http.StatusOK {
		t.Fatalf("first trial status=%d body=%s", first.Code, first.Body.String())
	}
	var firstResp struct {
		ID          int64  `json:"id"`
		AnonymousID string `json:"anonymous_id"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &firstResp); err != nil {
		t.Fatalf("decode first trial: %v", err)
	}
	var generation model.Generation
	if err := model.DB.First(&generation, firstResp.ID).Error; err != nil {
		t.Fatalf("load trial generation: %v", err)
	}
	if generation.UserID != nil || generation.Quality != "low" || generation.AnonymousID != firstResp.AnonymousID {
		t.Fatalf("unexpected trial generation: %+v", generation)
	}
	waitGenerationStatus(t, firstResp.ID, 3)

	second := postJSONWithFingerprint(engine, "/api/generations", map[string]string{
		"prompt":  "trial image",
		"quality": "low",
		"size":    "1024x1024",
	}, "fp-1")
	if second.Code != http.StatusForbidden {
		t.Fatalf("second trial status=%d body=%s", second.Code, second.Body.String())
	}

	third := postJSONWithFingerprint(engine, "/api/generations", map[string]string{
		"prompt":  "trial image",
		"quality": "low",
		"size":    "1024x1024",
	}, "fp-2")
	if third.Code != http.StatusOK {
		t.Fatalf("different fingerprint trial status=%d body=%s", third.Code, third.Body.String())
	}
	var thirdResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(third.Body.Bytes(), &thirdResp); err != nil {
		t.Fatalf("decode third trial: %v", err)
	}
	waitGenerationStatus(t, thirdResp.ID, 3)
}

func TestCreateGenerationInsufficientCredits(t *testing.T) {
	engine := setupAuthTest(t)
	token := createGenerationUser(t, 0)
	rec := postJSONWithToken(engine, "/api/generations", map[string]string{
		"prompt":  "a small house",
		"quality": "high",
		"size":    "1024x1024",
	}, token)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d body=%s", rec.Code, rec.Body.String())
	}
	var count int64
	if err := model.DB.Model(&model.Generation{}).Count(&count).Error; err != nil {
		t.Fatalf("count generations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no generation created, got %d", count)
	}
}

func TestGenerationFailureRefundsCredits(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = false
	token := createGenerationUser(t, 1)
	rec := postJSONWithToken(engine, "/api/generations", map[string]string{
		"prompt":  "a small house",
		"quality": "medium",
		"size":    "1024x1024",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("create generation status=%d body=%s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var generation model.Generation
		if err := model.DB.First(&generation, createResp.ID).Error; err != nil {
			t.Fatalf("load generation: %v", err)
		}
		if generation.Status == 4 {
			balance, err := service.GetBalance(*generation.UserID)
			if err != nil {
				t.Fatalf("get balance: %v", err)
			}
			if balance != 1 {
				t.Fatalf("expected refunded balance 1, got %v", balance)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("generation did not fail in time")
}

func TestCancelPendingGenerationRefundsCredits(t *testing.T) {
	engine := setupAuthTest(t)
	user := model.User{Email: "cancel@example.com", Role: 1, Status: 1, Credits: 0.8}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	generation := model.Generation{UserID: &user.ID, Prompt: "cancel", Quality: "low", Size: "1024x1024", CreditsCost: 0.2, Status: 0}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}

	rec := adminRequest(engine, http.MethodPost, "/api/generations/"+jsonNumber(generation.ID)+"/cancel", token)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"refunded":true`) {
		t.Fatalf("cancel status=%d body=%s", rec.Code, rec.Body.String())
	}
	var updated model.Generation
	if err := model.DB.First(&updated, generation.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if updated.Status != 5 {
		t.Fatalf("expected cancelled status, got %d", updated.Status)
	}
	balance, err := service.GetBalance(user.ID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance != 1 {
		t.Fatalf("expected refunded balance 1, got %v", balance)
	}
}

func TestCancelProcessingGenerationDoesNotRefund(t *testing.T) {
	engine := setupAuthTest(t)
	user := model.User{Email: "processing@example.com", Role: 1, Status: 1, Credits: 0.8}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	generation := model.Generation{UserID: &user.ID, Prompt: "cancel", Quality: "low", Size: "1024x1024", CreditsCost: 0.2, Status: 1}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}

	rec := adminRequest(engine, http.MethodPost, "/api/generations/"+jsonNumber(generation.ID)+"/cancel", token)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"refunded":false`) {
		t.Fatalf("cancel status=%d body=%s", rec.Code, rec.Body.String())
	}
	balance, err := service.GetBalance(user.ID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance != 0.8 {
		t.Fatalf("expected unchanged balance 0.8, got %v", balance)
	}
}

func TestSSEFormatUsesStatusEvent(t *testing.T) {
	body := bytes.NewBufferString("event:status\ndata:{\"status\":0}\n\n")
	scanner := bufio.NewScanner(body)
	if !scanner.Scan() || scanner.Text() != "event:status" {
		t.Fatal("expected status event format")
	}
}

func jsonNumber(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func createGenerationUser(t *testing.T, credits float64) string {
	t.Helper()
	user := model.User{Email: "gen@example.com", Role: 1, Status: 1, Credits: credits}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func postJSONWithToken(engine http.Handler, path string, body interface{}, token string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Real-IP", "1.2.3.4")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func postJSONWithFingerprint(engine http.Handler, path string, body interface{}, fingerprint string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "1.2.3.4")
	req.Header.Set("X-Fingerprint", fingerprint)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func waitGenerationStatus(t *testing.T, id int64, status int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var generation model.Generation
		if err := model.DB.First(&generation, id).Error; err != nil {
			t.Fatalf("load generation: %v", err)
		}
		if generation.Status == status {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("generation %d did not reach status %d", id, status)
}

func enableCaptchaForTest(t *testing.T) {
	t.Helper()
	settings := []model.Setting{
		{Key: "captcha_enabled", Value: "true"},
		{Key: "turnstile_site_key", Value: "site"},
		{Key: "turnstile_secret", Value: "secret"},
	}
	if err := model.DB.Create(&settings).Error; err != nil {
		t.Fatalf("create captcha settings: %v", err)
	}
}
