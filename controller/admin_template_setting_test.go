package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jayson2hu/image-show/model"
)

func TestAdminPromptTemplateCRUDAndSettings(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)

	create := adminJSON(engine, http.MethodPost, "/api/admin/prompt-templates", map[string]interface{}{
		"category":   "default",
		"label":      "Product",
		"prompt":     "clean product photo",
		"sort_order": 5,
		"status":     1,
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create template=%d body=%s", create.Code, create.Body.String())
	}
	var template model.PromptTemplate
	if err := json.Unmarshal(create.Body.Bytes(), &template); err != nil {
		t.Fatalf("decode template: %v", err)
	}

	update := adminJSON(engine, http.MethodPut, "/api/admin/prompt-templates/"+jsonNumber(template.ID), map[string]interface{}{
		"category":   "repair",
		"label":      "Repair",
		"prompt":     "repair details",
		"sort_order": 2,
		"status":     2,
	}, token)
	if update.Code != http.StatusOK {
		t.Fatalf("update template=%d body=%s", update.Code, update.Body.String())
	}
	list := adminRequest(engine, http.MethodGet, "/api/admin/prompt-templates", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list template=%d body=%s", list.Code, list.Body.String())
	}
	del := adminRequest(engine, http.MethodDelete, "/api/admin/prompt-templates/"+jsonNumber(template.ID), token)
	if del.Code != http.StatusOK {
		t.Fatalf("delete template=%d body=%s", del.Code, del.Body.String())
	}

	settings := adminJSON(engine, http.MethodPut, "/api/admin/settings", map[string]interface{}{
		"items": map[string]string{
			"register_enabled":                   "false",
			"site_name":                          "Image Show",
			"credit_exhausted_message":           "请联系管理员开通额度",
			"credit_exhausted_wechat_qrcode_url": "https://cdn.example.com/wechat.png",
			"credit_exhausted_qq":                "123456",
		},
	}, token)
	if settings.Code != http.StatusOK {
		t.Fatalf("update settings=%d body=%s", settings.Code, settings.Body.String())
	}
	if model.RegisterEnabled() {
		t.Fatal("expected register_enabled=false")
	}
	getSettings := adminRequest(engine, http.MethodGet, "/api/admin/settings", token)
	if getSettings.Code != http.StatusOK {
		t.Fatalf("get settings=%d body=%s", getSettings.Code, getSettings.Body.String())
	}
	var settingsResp struct {
		Items map[string]string `json:"items"`
	}
	if err := json.Unmarshal(getSettings.Body.Bytes(), &settingsResp); err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	for _, key := range []string{"r2_endpoint", "r2_access_key", "r2_secret_key", "r2_bucket", "r2_public_url"} {
		if _, ok := settingsResp.Items[key]; !ok {
			t.Fatalf("missing r2 setting %s in %#v", key, settingsResp.Items)
		}
	}
	for _, key := range []string{"register_gift_credits", "credit_exhausted_message", "credit_exhausted_wechat_qrcode_url", "credit_exhausted_qq"} {
		if _, ok := settingsResp.Items[key]; !ok {
			t.Fatalf("missing support setting %s in %#v", key, settingsResp.Items)
		}
	}
	contact := adminRequest(engine, http.MethodGet, "/api/support/contact", "")
	if contact.Code != http.StatusOK {
		t.Fatalf("support contact=%d body=%s", contact.Code, contact.Body.String())
	}
	var contactResp map[string]string
	if err := json.Unmarshal(contact.Body.Bytes(), &contactResp); err != nil {
		t.Fatalf("decode support contact: %v", err)
	}
	if contactResp["credit_exhausted_message"] != "请联系管理员开通额度" || contactResp["credit_exhausted_wechat_qrcode_url"] == "" || contactResp["credit_exhausted_qq"] != "123456" {
		t.Fatalf("unexpected support contact: %#v", contactResp)
	}
}
