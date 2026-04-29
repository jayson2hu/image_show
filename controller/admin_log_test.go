package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func TestAdminLogsRequireAdmin(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs/generations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminGenerationAndLoginLogs(t *testing.T) {
	engine := setupAuthTest(t)
	adminToken := createTokenForRole(t, 10)
	userID := int64(99)
	now := time.Now()
	if err := model.DB.Create(&model.Generation{UserID: &userID, Prompt: "p", Quality: "low", Size: "1024x1024", Status: 3, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if err := model.DB.Create(&model.LoginLog{UserID: userID, IP: "1.2.3.4", Method: "email", Success: true, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create login log: %v", err)
	}

	genRec := adminRequest(engine, http.MethodGet, "/api/admin/logs/generations?status=3&page=1&pageSize=10", adminToken)
	if genRec.Code != http.StatusOK {
		t.Fatalf("generation logs status=%d body=%s", genRec.Code, genRec.Body.String())
	}
	var genResp struct {
		Total int64 `json:"total"`
	}
	_ = json.Unmarshal(genRec.Body.Bytes(), &genResp)
	if genResp.Total != 1 {
		t.Fatalf("expected 1 generation log, got %d", genResp.Total)
	}

	loginRec := adminRequest(engine, http.MethodGet, "/api/admin/logs/logins?user_id=99", adminToken)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login logs status=%d body=%s", loginRec.Code, loginRec.Body.String())
	}
}

func TestAdminDeleteLogsBefore(t *testing.T) {
	engine := setupAuthTest(t)
	adminToken := createTokenForRole(t, 10)
	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now()
	if err := model.DB.Create(&model.LoginLog{UserID: 1, CreatedAt: oldTime}).Error; err != nil {
		t.Fatalf("create old log: %v", err)
	}
	if err := model.DB.Create(&model.LoginLog{UserID: 1, CreatedAt: newTime}).Error; err != nil {
		t.Fatalf("create new log: %v", err)
	}

	rec := adminRequest(engine, http.MethodDelete, "/api/admin/logs/logins?before="+url.QueryEscape(time.Now().Add(-24*time.Hour).Format(time.RFC3339)), adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", rec.Code, rec.Body.String())
	}
	var count int64
	if err := model.DB.Model(&model.LoginLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one remaining log, got %d", count)
	}
}

func createTokenForRole(t *testing.T, role int) string {
	t.Helper()
	user := model.User{Email: time.Now().String() + "@example.com", Role: role, Status: 1}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return token
}

func adminRequest(engine http.Handler, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
