package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jayson2hu/image-show/model"
)

func TestSiteConfigReturnsPublicSettings(t *testing.T) {
	engine := setupAuthTest(t)
	if err := model.DB.Create(&model.Setting{Key: "site_title", Value: "来看看巴"}).Error; err != nil {
		t.Fatalf("create site title: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "seo_description", Value: "输入提示词生成图片"}).Error; err != nil {
		t.Fatalf("create seo description: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "wechat_server_token", Value: "secret-token"}).Error; err != nil {
		t.Fatalf("create secret setting: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "credit_cost_square", Value: "3"}).Error; err != nil {
		t.Fatalf("create credit cost setting: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "register_enabled", Value: "false"}).Error; err != nil {
		t.Fatalf("create register setting: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/site/config", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("site config=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		SiteTitle       string             `json:"site_title"`
		SEODescription  string             `json:"seo_description"`
		Secret          string             `json:"wechat_server_token"`
		RegisterEnabled bool               `json:"register_enabled"`
		CreditCosts     map[string]float64 `json:"credit_costs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode site config: %v", err)
	}
	if resp.SiteTitle != "来看看巴" || resp.SEODescription != "输入提示词生成图片" {
		t.Fatalf("unexpected site config: %#v", resp)
	}
	if resp.Secret != "" {
		t.Fatalf("site config leaked secret: %#v", resp)
	}
	if resp.RegisterEnabled {
		t.Fatalf("expected register_enabled=false in public config: %#v", resp)
	}
	if resp.CreditCosts["square"] != 3 || resp.CreditCosts["portrait"] != 2 || resp.CreditCosts["widescreen"] != 2 {
		t.Fatalf("unexpected credit costs: %#v", resp.CreditCosts)
	}
}
