package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func TestAdminMonitorSummaryAndAlert(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 10)
	now := time.Now()
	userID := int64(10)
	if err := model.DB.Create(&model.Generation{UserID: &userID, Prompt: "p", Quality: "low", Size: "1024x1024", Status: 3, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if err := model.DB.Create(&model.Generation{UserID: &userID, Prompt: "bad", Quality: "low", Size: "1024x1024", Status: 4, ErrorMsg: "sub2api status 503: upstream unavailable", CreatedAt: now}).Error; err != nil {
		t.Fatalf("create failed generation: %v", err)
	}
	if err := model.DB.Create(&model.CreditLog{UserID: userID, Type: 2, Amount: -2.5, Balance: 7.5, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create credit log: %v", err)
	}
	if err := model.DB.Create(&model.Order{OrderNo: "MONITOR", UserID: userID, Amount: 9.9, Status: service.OrderStatusPaid, PaidAt: &now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := model.DB.Create(&model.Setting{Key: "monitor_daily_credit_threshold", Value: "1"}).Error; err != nil {
		t.Fatalf("create threshold: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/admin/monitor/summary", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary status=%d body=%s", rec.Code, rec.Body.String())
	}
	var summary service.MonitorSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	if summary.GenerationCount != 2 || summary.FailedCount != 1 || summary.FailureRate != 0.5 || summary.CreditsConsumed != 2.5 || !summary.AlertTriggered {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(summary.FailureReasons) != 1 || summary.FailureReasons[0].Category != "upstream_unavailable" || len(summary.RecentFailures) != 1 {
		t.Fatalf("unexpected failure details: %+v", summary)
	}

	check := adminRequest(engine, http.MethodPost, "/api/admin/monitor/check", token)
	if check.Code != http.StatusOK {
		t.Fatalf("check status=%d body=%s", check.Code, check.Body.String())
	}
	var result service.MonitorAlertResult
	if err := json.Unmarshal(check.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode alert result: %v", err)
	}
	if !result.Triggered || !result.Sent {
		t.Fatalf("expected alert sent, got %+v", result)
	}

	duplicate := adminRequest(engine, http.MethodPost, "/api/admin/monitor/check", token)
	if duplicate.Code != http.StatusOK {
		t.Fatalf("duplicate status=%d body=%s", duplicate.Code, duplicate.Body.String())
	}
	var duplicateResult service.MonitorAlertResult
	_ = json.Unmarshal(duplicate.Body.Bytes(), &duplicateResult)
	if !duplicateResult.Triggered || !duplicateResult.Skipped {
		t.Fatalf("expected duplicate skipped, got %+v", duplicateResult)
	}
}
