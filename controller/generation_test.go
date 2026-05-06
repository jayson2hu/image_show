package controller_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

var testPNGBytes = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

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
	if balance != 2 {
		t.Fatalf("expected 1024x1024 size cost 1, balance=%v", balance)
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

func TestGenerationOptionsDefaultSizesIncludeStableRatios(t *testing.T) {
	engine := setupAuthTest(t)
	rec := adminRequest(engine, http.MethodGet, "/api/generation/options", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("options status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Sizes []string `json:"sizes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode options: %v", err)
	}
	want := "square,portrait_3_4,story,landscape_4_3,widescreen"
	if got := strings.Join(resp.Sizes, ","); got != want {
		t.Fatalf("unexpected default sizes: got %s want %s", got, want)
	}
}

func TestGenerationOptionsReturnsSameSizesForAnonymousAndLoggedIn(t *testing.T) {
	engine := setupAuthTest(t)
	if err := model.DB.Create(&model.Setting{
		Key:   "enabled_image_sizes",
		Value: "square,portrait_3_4,story,landscape_4_3,widescreen",
	}).Error; err != nil {
		t.Fatalf("create size setting: %v", err)
	}

	anonymous := adminRequest(engine, http.MethodGet, "/api/generation/options", "")
	if anonymous.Code != http.StatusOK {
		t.Fatalf("anonymous options status=%d body=%s", anonymous.Code, anonymous.Body.String())
	}
	var anonymousResp struct {
		Sizes       []string `json:"sizes"`
		SizeOptions []struct {
			Value      string  `json:"value"`
			Label      string  `json:"label"`
			Ratio      string  `json:"ratio"`
			CreditCost float64 `json:"credit_cost"`
		} `json:"size_options"`
	}
	if err := json.Unmarshal(anonymous.Body.Bytes(), &anonymousResp); err != nil {
		t.Fatalf("decode anonymous options: %v", err)
	}
	anonymousSizes := strings.Join(anonymousResp.Sizes, ",")
	for _, expected := range []string{"square", "portrait_3_4", "story", "landscape_4_3", "widescreen"} {
		if !strings.Contains(anonymousSizes, expected) {
			t.Fatalf("expected size %s, got %#v", expected, anonymousResp.Sizes)
		}
	}
	if len(anonymousResp.SizeOptions) != 5 {
		t.Fatalf("unexpected anonymous sizes: %#v", anonymousResp.Sizes)
	}
	if len(anonymousResp.SizeOptions) == 0 || anonymousResp.SizeOptions[0].Label == "" || anonymousResp.SizeOptions[0].Ratio == "" {
		t.Fatalf("expected ratio size options, got %#v", anonymousResp.SizeOptions)
	}
	expectedOptions := map[string]struct {
		label string
		ratio string
		cost  float64
	}{
		"square":        {label: "方形", ratio: "1:1", cost: 1},
		"portrait_3_4":  {label: "竖版", ratio: "3:4", cost: 2},
		"story":         {label: "故事版", ratio: "9:16", cost: 2},
		"landscape_4_3": {label: "横版", ratio: "4:3", cost: 2},
		"widescreen":    {label: "宽屏", ratio: "16:9", cost: 2},
	}
	for _, item := range anonymousResp.SizeOptions {
		expected, ok := expectedOptions[item.Value]
		if !ok {
			t.Fatalf("unexpected size option: %#v", item)
		}
		if item.Ratio != expected.ratio || item.Label != expected.label || item.CreditCost != expected.cost {
			t.Fatalf("unexpected option for %s: got ratio=%s label=%s cost=%v", item.Value, item.Ratio, item.Label, item.CreditCost)
		}
	}

	token := createGenerationUser(t, 1)
	loggedIn := adminRequest(engine, http.MethodGet, "/api/generation/options", token)
	if loggedIn.Code != http.StatusOK {
		t.Fatalf("logged in options status=%d body=%s", loggedIn.Code, loggedIn.Body.String())
	}
	var loggedInResp struct {
		Sizes []string `json:"sizes"`
	}
	if err := json.Unmarshal(loggedIn.Body.Bytes(), &loggedInResp); err != nil {
		t.Fatalf("decode logged in options: %v", err)
	}
	if strings.Join(loggedInResp.Sizes, ",") != anonymousSizes {
		t.Fatalf("expected same sizes for anonymous and logged in, anonymous=%#v loggedIn=%#v", anonymousResp.Sizes, loggedInResp.Sizes)
	}
}

func TestGenerationOptionsUpgradesLegacyDefaultSizeSetting(t *testing.T) {
	engine := setupAuthTest(t)
	if err := model.DB.Create(&model.Setting{
		Key:   "enabled_image_sizes",
		Value: "1280x720,720x1280,1024x1024,1536x1024,1024x1536",
	}).Error; err != nil {
		t.Fatalf("create legacy size setting: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/generation/options", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("options status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Sizes []string `json:"sizes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode options: %v", err)
	}
	want := "square,portrait_3_4,story,landscape_4_3,widescreen"
	if got := strings.Join(resp.Sizes, ","); got != want {
		t.Fatalf("unexpected upgraded sizes: got %s want %s", got, want)
	}
}

func TestCreateGenerationAcceptsAspectRatioKey(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true

	rec := postJSONWithFingerprint(engine, "/api/generations", map[string]string{
		"prompt":  "wide image",
		"quality": "medium",
		"size":    "widescreen",
	}, "aspect-ratio-fp")
	if rec.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var generation model.Generation
	if err := model.DB.First(&generation, resp.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.Size != "1792x1008" || generation.CreditsCost != 2 {
		t.Fatalf("unexpected generation size/cost: %+v", generation)
	}
	waitGenerationStatus(t, resp.ID, 3)
}

func TestCreateGenerationStillAcceptsMappedPixelSize(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true

	rec := postJSONWithFingerprint(engine, "/api/generations", map[string]string{
		"prompt":  "square image",
		"quality": "medium",
		"size":    "1024x1024",
	}, "pixel-size-fp")
	if rec.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var generation model.Generation
	if err := model.DB.First(&generation, resp.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.Size != "1024x1024" || generation.CreditsCost != 1 {
		t.Fatalf("unexpected generation size/cost: %+v", generation)
	}
	waitGenerationStatus(t, resp.ID, 3)
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

func TestCreateGenerationValidatesOutputOptions(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true
	token := createGenerationUser(t, 3)
	rec := postJSONWithToken(engine, "/api/generations", map[string]interface{}{
		"prompt":        "a small house",
		"quality":       "medium",
		"size":          "1024x1024",
		"output_format": "gif",
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid format 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = postJSONWithToken(engine, "/api/generations", map[string]interface{}{
		"prompt":             "a small house",
		"quality":            "medium",
		"size":               "1024x1024",
		"output_format":      "jpeg",
		"background":         "transparent",
		"output_compression": 101,
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid compression 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	compression := 80
	rec = postJSONWithToken(engine, "/api/generations", map[string]interface{}{
		"prompt":             "a small house",
		"quality":            "medium",
		"size":               "1024x1024",
		"output_format":      "webp",
		"background":         "transparent",
		"output_compression": compression,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected valid output options 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var generation model.Generation
	if err := model.DB.First(&generation, resp.ID).Error; err != nil {
		t.Fatalf("load generation: %v", err)
	}
	if generation.OutputFormat != "webp" || generation.Background != "transparent" || generation.OutputCompression == nil || *generation.OutputCompression != compression {
		t.Fatalf("unexpected output options saved: %+v", generation)
	}
	waitGenerationStatus(t, resp.ID, 3)
}

func TestCreateImageEditCompletesInMockMode(t *testing.T) {
	engine := setupAuthTest(t)
	config.AppConfig.MockSub2API = true
	token := createGenerationUser(t, 3)

	rec := postMultipartEditWithToken(engine, token, map[string]string{
		"prompt":  "make it brighter",
		"quality": "low",
		"size":    "1024x1024",
	}, "image", "source.png", "image/png", testPNGBytes)
	if rec.Code != http.StatusOK {
		t.Fatalf("create edit status=%d body=%s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create edit: %v", err)
	}
	waitGenerationStatus(t, createResp.ID, 3)
	var generation model.Generation
	if err := model.DB.First(&generation, createResp.ID).Error; err != nil {
		t.Fatalf("load edit generation: %v", err)
	}
	if generation.Mode != service.GenerationModeEdit || generation.ImageURL == "" {
		t.Fatalf("unexpected edit generation: %+v", generation)
	}
}

func TestCreateImageEditRejectsUnsupportedFileType(t *testing.T) {
	engine := setupAuthTest(t)
	token := createGenerationUser(t, 3)

	rec := postMultipartEditWithToken(engine, token, map[string]string{
		"prompt":  "make it brighter",
		"quality": "low",
		"size":    "1024x1024",
	}, "image", "source.txt", "text/plain", []byte("not image"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateImageEditRequiresLogin(t *testing.T) {
	engine := setupAuthTest(t)

	rec := postMultipartEditWithToken(engine, "", map[string]string{
		"prompt":  "make it brighter",
		"quality": "low",
		"size":    "1024x1024",
	}, "image", "source.png", "image/png", testPNGBytes)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAnonymousTrialOnceUsesStandardQuality(t *testing.T) {
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
	if generation.UserID != nil || generation.Quality != "medium" || generation.AnonymousID != firstResp.AnonymousID {
		t.Fatalf("unexpected trial generation: %+v", generation)
	}
	waitGenerationStatus(t, firstResp.ID, 3)

	second := postJSONWithFingerprint(engine, "/api/generations", map[string]string{
		"prompt":  "trial image",
		"quality": "low",
		"size":    "1024x1024",
	}, "fp-1")
	if second.Code != http.StatusPaymentRequired {
		t.Fatalf("second trial status=%d body=%s", second.Code, second.Body.String())
	}
	assertJSONError(t, second, "free_trial_exhausted")
	if strings.Contains(second.Body.String(), "free trial used") || strings.Contains(second.Body.String(), "please register") {
		t.Fatalf("trial exhausted response should not expose legacy message: %s", second.Body.String())
	}
	var secondBody map[string]string
	if err := json.Unmarshal(second.Body.Bytes(), &secondBody); err != nil {
		t.Fatalf("decode second trial body: %v", err)
	}
	if secondBody["message"] == "" {
		t.Fatalf("expected friendly message in second trial body: %s", second.Body.String())
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
	assertJSONError(t, rec, "insufficient_credits")
	var count int64
	if err := model.DB.Model(&model.Generation{}).Count(&count).Error; err != nil {
		t.Fatalf("count generations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no generation created, got %d", count)
	}
}

func TestCreateGenerationCreditsExpired(t *testing.T) {
	engine := setupAuthTest(t)
	expiry := time.Now().Add(-time.Hour)
	user := model.User{Email: "expired@example.com", Role: 1, Status: 1, Credits: 3, CreditsExpiry: &expiry}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	rec := postJSONWithToken(engine, "/api/generations", map[string]string{
		"prompt":  "a small house",
		"quality": "medium",
		"size":    "1024x1024",
	}, token)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertJSONError(t, rec, "credits_expired")
}

func TestCreateImageEditInsufficientCredits(t *testing.T) {
	engine := setupAuthTest(t)
	token := createGenerationUser(t, 0)
	rec := postMultipartEditWithToken(engine, token, map[string]string{
		"prompt":  "make it brighter",
		"quality": "low",
		"size":    "1024x1024",
	}, "image", "source.png", "image/png", testPNGBytes)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertJSONError(t, rec, "insufficient_credits")
}

func TestCreateImageEditCreditsExpired(t *testing.T) {
	engine := setupAuthTest(t)
	expiry := time.Now().Add(-time.Hour)
	user := model.User{Email: "edit-expired@example.com", Role: 1, Status: 1, Credits: 3, CreditsExpiry: &expiry}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	rec := postMultipartEditWithToken(engine, token, map[string]string{
		"prompt":  "make it brighter",
		"quality": "low",
		"size":    "1024x1024",
	}, "image", "source.png", "image/png", testPNGBytes)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertJSONError(t, rec, "credits_expired")
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

func TestStreamGenerationUsesUTF8EventStream(t *testing.T) {
	engine := setupAuthTest(t)
	generation := model.Generation{Prompt: "cat", Quality: "low", Size: "1024x1024", Status: 3, ImageURL: "https://example.com/image.png"}
	if err := model.DB.Create(&generation).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/generations/"+jsonNumber(generation.ID)+"/stream", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") || !strings.Contains(strings.ToLower(contentType), "charset=utf-8") {
		t.Fatalf("unexpected content type: %q", contentType)
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

func postMultipartEditWithToken(engine http.Handler, token string, fields map[string]string, fileField, filename, contentType string, data []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		_ = writer.WriteField(key, value)
	}
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="`+fileField+`"; filename="`+filename+`"`)
	partHeader.Set("Content-Type", contentType)
	part, _ := writer.CreatePart(partHeader)
	_, _ = part.Write(data)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/generations/edit", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-Real-IP", "1.2.3.4")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func assertJSONError(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	t.Helper()
	var resp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v body=%s", err, rec.Body.String())
	}
	if resp.Error != expected || resp.Message == "" {
		t.Fatalf("unexpected error response: got error=%q message=%q want error=%q", resp.Error, resp.Message, expected)
	}
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
