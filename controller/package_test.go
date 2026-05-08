package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jayson2hu/image-show/model"
)

func TestPackageListAndAdminCRUD(t *testing.T) {
	engine := setupAuthTest(t)
	public := adminRequest(engine, http.MethodGet, "/api/packages", "")
	if public.Code != http.StatusOK {
		t.Fatalf("public packages status=%d body=%s", public.Code, public.Body.String())
	}
	var publicResp struct {
		Items []model.Package `json:"items"`
	}
	_ = json.Unmarshal(public.Body.Bytes(), &publicResp)
	if len(publicResp.Items) != 3 || publicResp.Items[0].Name != "Starter Pack" || publicResp.Items[1].Name != "Standard Pack" || publicResp.Items[2].Name != "Pro Pack" {
		t.Fatalf("unexpected default package names: %+v", publicResp.Items)
	}
	if publicResp.Items[0].Credits != 10 || publicResp.Items[1].Price != 39.9 || publicResp.Items[2].Credits != 100 {
		t.Fatalf("unexpected default packages: %+v", publicResp.Items)
	}

	token := createTokenForRole(t, 10)
	create := adminJSON(engine, http.MethodPost, "/api/admin/packages", map[string]interface{}{
		"name":       "测试包",
		"credits":    5,
		"price":      4.9,
		"valid_days": 15,
		"sort_order": 9,
		"status":     1,
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create package=%d body=%s", create.Code, create.Body.String())
	}
	var pkg model.Package
	if err := json.Unmarshal(create.Body.Bytes(), &pkg); err != nil {
		t.Fatalf("decode package: %v", err)
	}
	update := adminJSON(engine, http.MethodPut, "/api/admin/packages/"+jsonNumber(pkg.ID), map[string]interface{}{
		"name":       "下架包",
		"credits":    6,
		"price":      5.9,
		"valid_days": 20,
		"sort_order": 10,
		"status":     2,
	}, token)
	if update.Code != http.StatusOK {
		t.Fatalf("update package=%d body=%s", update.Code, update.Body.String())
	}
	del := adminRequest(engine, http.MethodDelete, "/api/admin/packages/"+jsonNumber(pkg.ID), token)
	if del.Code != http.StatusOK {
		t.Fatalf("delete package=%d body=%s", del.Code, del.Body.String())
	}
}
