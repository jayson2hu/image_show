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

	rec := adminRequest(engine, http.MethodGet, "/api/site/config", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("site config=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode site config: %v", err)
	}
	if resp["site_title"] != "来看看巴" || resp["seo_description"] != "输入提示词生成图片" {
		t.Fatalf("unexpected site config: %#v", resp)
	}
	if _, ok := resp["wechat_server_token"]; ok {
		t.Fatalf("site config leaked secret: %#v", resp)
	}
}
