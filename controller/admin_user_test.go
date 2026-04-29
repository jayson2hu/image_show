package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jayson2hu/image-show/model"
	"golang.org/x/crypto/bcrypt"
)

func TestAdminUserManagementAndCredits(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)
	user := model.User{Username: "managed", Email: "managed@example.com", Role: 1, Status: 1, Credits: 1}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := model.DB.Create(&model.Generation{UserID: &user.ID, Prompt: "p", Quality: "low", Size: "1024x1024", Status: 3}).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}

	list := adminRequest(engine, http.MethodGet, "/api/admin/users?keyword=managed", token)
	if list.Code != http.StatusOK {
		t.Fatalf("list users status=%d body=%s", list.Code, list.Body.String())
	}
	var listResp struct {
		Total int64 `json:"total"`
	}
	_ = json.Unmarshal(list.Body.Bytes(), &listResp)
	if listResp.Total != 1 {
		t.Fatalf("expected one user, got %d", listResp.Total)
	}

	status := adminJSON(engine, http.MethodPut, "/api/admin/users/"+jsonNumber(user.ID)+"/status", map[string]int{"status": 2}, token)
	if status.Code != http.StatusOK {
		t.Fatalf("status update=%d body=%s", status.Code, status.Body.String())
	}
	role := adminJSON(engine, http.MethodPut, "/api/admin/users/"+jsonNumber(user.ID)+"/role", map[string]int{"role": 10}, token)
	if role.Code != http.StatusOK {
		t.Fatalf("role update=%d body=%s", role.Code, role.Body.String())
	}
	topup := adminJSON(engine, http.MethodPost, "/api/admin/users/"+jsonNumber(user.ID)+"/credits", map[string]interface{}{"amount": 2.5, "remark": "manual"}, token)
	if topup.Code != http.StatusOK {
		t.Fatalf("topup=%d body=%s", topup.Code, topup.Body.String())
	}

	var updated model.User
	if err := model.DB.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updated.Status != 2 || updated.Role != 10 || updated.Credits != 3.5 {
		t.Fatalf("unexpected updated user: %+v", updated)
	}

	generations := adminRequest(engine, http.MethodGet, "/api/admin/users/"+jsonNumber(user.ID)+"/generations", token)
	if generations.Code != http.StatusOK {
		t.Fatalf("generations=%d body=%s", generations.Code, generations.Body.String())
	}
	credits := adminRequest(engine, http.MethodGet, "/api/admin/credits/logs?user_id="+jsonNumber(user.ID), token)
	if credits.Code != http.StatusOK {
		t.Fatalf("credit logs=%d body=%s", credits.Code, credits.Body.String())
	}
}

func TestBannedUserCannotLogin(t *testing.T) {
	engine := setupAuthTest(t)
	user := model.User{Username: "banned", Email: "banned@example.com", PasswordHash: bcryptHash(t, "password123"), Role: 1, Status: 2}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	rec := postJSON(engine, "/api/auth/login", map[string]string{"email": user.Email, "password": "password123"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func bcryptHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(hash)
}
