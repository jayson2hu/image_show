package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func TestWeChatQRCodeAndLoginCreatesUser(t *testing.T) {
	engine := setupAuthTest(t)
	configureWeChatTestServer(t)

	qrcode := adminRequest(engine, http.MethodGet, "/api/auth/wechat/qrcode", "")
	if qrcode.Code != http.StatusOK {
		t.Fatalf("qrcode status=%d body=%s", qrcode.Code, qrcode.Body.String())
	}
	var qrResp struct {
		Enabled   bool   `json:"enabled"`
		QRCodeURL string `json:"qrcode_url"`
	}
	if err := json.Unmarshal(qrcode.Body.Bytes(), &qrResp); err != nil {
		t.Fatalf("decode qrcode: %v", err)
	}
	if !qrResp.Enabled || qrResp.QRCodeURL == "" {
		t.Fatalf("unexpected qrcode response: %+v", qrResp)
	}
	if qrcodeBody := qrcode.Body.String(); containsAny(qrcodeBody, []string{"wechat-token", "wechat_server_token", "wechat_server_address", config.AppConfig.WeChatServer}) {
		t.Fatalf("qrcode response leaked sensitive wechat server config: %s", qrcodeBody)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/auth/wechat/callback?code=new-user", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("wechat login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var loginResp struct {
		Token string     `json:"token"`
		User  model.User `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if loginResp.Token == "" || loginResp.User.WechatOpenID != "openid-new-user" || loginResp.User.Credits != 10 {
		t.Fatalf("unexpected login response: %+v", loginResp)
	}

	second := adminRequest(engine, http.MethodGet, "/api/auth/wechat/callback?code=new-user", "")
	if second.Code != http.StatusOK {
		t.Fatalf("existing wechat login status=%d body=%s", second.Code, second.Body.String())
	}
	var count int64
	if err := model.DB.Model(&model.User{}).Where("wechat_open_id = ?", "openid-new-user").Count(&count).Error; err != nil {
		t.Fatalf("count wechat users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected existing user reused, got %d", count)
	}
}

func TestWeChatBindAndUnbind(t *testing.T) {
	engine := setupAuthTest(t)
	configureWeChatTestServer(t)
	token := createTokenForRole(t, 1)

	bind := adminJSON(engine, http.MethodPost, "/api/auth/wechat/bind", map[string]string{"code": "bind-user"}, token)
	if bind.Code != http.StatusOK {
		t.Fatalf("bind status=%d body=%s", bind.Code, bind.Body.String())
	}
	var user model.User
	claimsUserID := tokenUserID(t, token)
	if err := model.DB.First(&user, claimsUserID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.WechatOpenID != "openid-bind-user" {
		t.Fatalf("unexpected openid: %s", user.WechatOpenID)
	}

	unbind := adminRequest(engine, http.MethodDelete, "/api/auth/wechat/bind", token)
	if unbind.Code != http.StatusOK {
		t.Fatalf("unbind status=%d body=%s", unbind.Code, unbind.Body.String())
	}
	if err := model.DB.First(&user, claimsUserID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if user.WechatOpenID != "" {
		t.Fatalf("expected unbound openid, got %s", user.WechatOpenID)
	}
}

func TestWeChatDisabled(t *testing.T) {
	engine := setupAuthTest(t)
	rec := adminRequest(engine, http.MethodGet, "/api/auth/wechat/callback?code=x", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected disabled 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestWeChatInvalidCodeFromServer(t *testing.T) {
	engine := setupAuthTest(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "code expired",
		})
	}))
	t.Cleanup(server.Close)
	config.AppConfig.WeChatEnabled = true
	config.AppConfig.WeChatServer = server.URL
	config.AppConfig.WeChatToken = "wechat-token"
	config.AppConfig.WeChatQRCode = "https://example.com/qrcode.png"

	rec := adminRequest(engine, http.MethodGet, "/api/auth/wechat/callback?code=expired", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid code 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func configureWeChatTestServer(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/wechat/user" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "wechat-token" {
			t.Fatalf("missing authorization: %s", r.Header.Get("Authorization"))
		}
		code := r.URL.Query().Get("code")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    "openid-" + code,
		})
	}))
	t.Cleanup(server.Close)
	config.AppConfig.WeChatEnabled = true
	config.AppConfig.WeChatServer = server.URL
	config.AppConfig.WeChatToken = "wechat-token"
	config.AppConfig.WeChatQRCode = "https://example.com/qrcode.png"
}

func tokenUserID(t *testing.T, token string) int64 {
	t.Helper()
	claims, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	return claims.UserID
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
