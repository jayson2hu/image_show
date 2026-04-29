package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func TestCreditBalanceAndLogs(t *testing.T) {
	engine := setupAuthTest(t)
	user := model.User{Email: "credits@example.com", Role: 1, Status: 1, Credits: 3}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := model.DB.Create(&model.CreditLog{UserID: user.ID, Type: 1, Amount: 3, Balance: 3}).Error; err != nil {
		t.Fatalf("create credit log: %v", err)
	}
	token, err := service.GenerateToken(user.ID, user.Role)
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	balanceReq := httptest.NewRequest(http.MethodGet, "/api/credits/balance", nil)
	balanceReq.Header.Set("Authorization", "Bearer "+token)
	balanceRec := httptest.NewRecorder()
	engine.ServeHTTP(balanceRec, balanceReq)
	if balanceRec.Code != http.StatusOK {
		t.Fatalf("balance status=%d body=%s", balanceRec.Code, balanceRec.Body.String())
	}
	var balanceResp struct {
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal(balanceRec.Body.Bytes(), &balanceResp); err != nil {
		t.Fatalf("decode balance: %v", err)
	}
	if balanceResp.Balance != 3 {
		t.Fatalf("unexpected balance: %v", balanceResp.Balance)
	}

	logsReq := httptest.NewRequest(http.MethodGet, "/api/credits/logs?page=1&pageSize=10", nil)
	logsReq.Header.Set("Authorization", "Bearer "+token)
	logsRec := httptest.NewRecorder()
	engine.ServeHTTP(logsRec, logsReq)
	if logsRec.Code != http.StatusOK || !json.Valid(logsRec.Body.Bytes()) {
		t.Fatalf("logs response invalid: status=%d body=%s", logsRec.Code, logsRec.Body.String())
	}
}

func TestExpiredCreditsUnavailable(t *testing.T) {
	setupAuthTest(t)
	expired := time.Now().Add(-time.Hour)
	user := model.User{Email: "expired@example.com", Role: 1, Status: 1, Credits: 3, CreditsExpiry: &expired}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	balance, err := service.GetBalance(user.ID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance != 0 {
		t.Fatalf("expected expired balance 0, got %v", balance)
	}
	if err := service.Deduct(user.ID, 0.2, 1); err != service.ErrCreditsExpired {
		t.Fatalf("expected ErrCreditsExpired, got %v", err)
	}
}
