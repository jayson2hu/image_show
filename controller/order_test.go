package controller_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func TestCreateOrderReturnsPaymentLink(t *testing.T) {
	engine := setupAuthTest(t)
	configurePaymentForTest()
	token := createTokenForRole(t, 1)
	pkg := firstPackage(t)

	rec := adminJSON(engine, http.MethodPost, "/api/orders", map[string]interface{}{
		"package_id": pkg.ID,
		"pay_method": "alipay",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("create order status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"pay_url":"https://pay.example.com/submit.php"`) {
		t.Fatalf("missing pay url: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"out_trade_no"`) || !strings.Contains(rec.Body.String(), `"sign"`) {
		t.Fatalf("missing epay params: %s", rec.Body.String())
	}
}

func TestPaymentNotifyCreditsUserAndIsIdempotent(t *testing.T) {
	engine := setupAuthTest(t)
	configurePaymentForTest()
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	pkg := firstPackage(t)

	create := adminJSON(engine, http.MethodPost, "/api/orders", map[string]interface{}{
		"package_id": pkg.ID,
		"pay_method": "wechat",
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create order status=%d body=%s", create.Code, create.Body.String())
	}
	var order model.Order
	if err := model.DB.Where("user_id = ?", userID).First(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}

	body := signedNotifyBody(order, "wxpay", "EPAY123", "TRADE_SUCCESS")
	first := postForm(engine, "/api/payment/notify", body)
	if first.Code != http.StatusOK || strings.TrimSpace(first.Body.String()) != "success" {
		t.Fatalf("notify status=%d body=%s", first.Code, first.Body.String())
	}
	second := postForm(engine, "/api/payment/notify", body)
	if second.Code != http.StatusOK || strings.TrimSpace(second.Body.String()) != "success" {
		t.Fatalf("duplicate notify status=%d body=%s", second.Code, second.Body.String())
	}

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Credits != pkg.Credits {
		t.Fatalf("expected credits %.2f, got %.2f", pkg.Credits, user.Credits)
	}
	if user.CreditsExpiry == nil {
		t.Fatal("expected credits expiry after payment")
	}
	var logCount int64
	if err := model.DB.Model(&model.CreditLog{}).Where("user_id = ? AND type = ?", userID, service.CreditLogTypePaymentTopup).Count(&logCount).Error; err != nil {
		t.Fatalf("count credit logs: %v", err)
	}
	if logCount != 1 {
		t.Fatalf("expected one payment credit log, got %d", logCount)
	}
	var paid model.Order
	if err := model.DB.First(&paid, order.ID).Error; err != nil {
		t.Fatalf("load paid order: %v", err)
	}
	if paid.Status != service.OrderStatusPaid || paid.PayTradeNo != "EPAY123" || paid.PaidAt == nil {
		t.Fatalf("unexpected paid order: %+v", paid)
	}
}

func TestPaymentNotifyExtendsExistingCreditsExpiry(t *testing.T) {
	engine := setupAuthTest(t)
	configurePaymentForTest()
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	pkg := firstPackage(t)
	existingExpiry := time.Now().Add(10 * 24 * time.Hour)
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"credits":        3,
		"credits_expiry": existingExpiry,
	}).Error; err != nil {
		t.Fatalf("set existing credits expiry: %v", err)
	}

	create := adminJSON(engine, http.MethodPost, "/api/orders", map[string]interface{}{
		"package_id": pkg.ID,
		"pay_method": "wechat",
	}, token)
	if create.Code != http.StatusOK {
		t.Fatalf("create order status=%d body=%s", create.Code, create.Body.String())
	}
	var order model.Order
	if err := model.DB.Where("user_id = ?", userID).First(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}

	body := signedNotifyBody(order, "wxpay", "EPAY456", "TRADE_SUCCESS")
	rec := postForm(engine, "/api/payment/notify", body)
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "success" {
		t.Fatalf("notify status=%d body=%s", rec.Code, rec.Body.String())
	}

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Credits != 3+pkg.Credits {
		t.Fatalf("expected credits %.2f, got %.2f", 3+pkg.Credits, user.Credits)
	}
	if user.CreditsExpiry == nil {
		t.Fatal("expected credits expiry")
	}
	expectedMin := existingExpiry.Add(time.Duration(pkg.ValidDays)*24*time.Hour - time.Minute)
	if user.CreditsExpiry.Before(expectedMin) {
		t.Fatalf("expected expiry to extend from existing expiry, existing=%s valid_days=%d got=%s", existingExpiry, pkg.ValidDays, user.CreditsExpiry)
	}
}

func TestExpiredOrderIsClosed(t *testing.T) {
	setupAuthTest(t)
	configurePaymentForTest()
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	pkg := firstPackage(t)
	old := time.Now().Add(-31 * time.Minute)
	order := model.Order{
		OrderNo:   "OLDORDER",
		UserID:    userID,
		PackageID: pkg.ID,
		Amount:    pkg.Price,
		Status:    service.OrderStatusPending,
		PayMethod: "alipay",
		CreatedAt: old,
		UpdatedAt: old,
	}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatalf("create old order: %v", err)
	}
	if err := service.ExpireStaleOrders(); err != nil {
		t.Fatalf("expire orders: %v", err)
	}
	var expired model.Order
	if err := model.DB.First(&expired, order.ID).Error; err != nil {
		t.Fatalf("load expired order: %v", err)
	}
	if expired.Status != service.OrderStatusExpired {
		t.Fatalf("expected expired status, got %d", expired.Status)
	}
}

func configurePaymentForTest() {
	config.AppConfig.ServerAddress = "https://app.example.com"
	config.AppConfig.PayAddress = "https://pay.example.com"
	config.AppConfig.EpayID = "pid"
	config.AppConfig.EpayKey = "secret"
	config.AppConfig.EpayPayMethods = "alipay,wxpay"
}

func firstPackage(t *testing.T) model.Package {
	t.Helper()
	var pkg model.Package
	if err := model.DB.Where("status = ?", 1).Order("id ASC").First(&pkg).Error; err != nil {
		t.Fatalf("load package: %v", err)
	}
	return pkg
}

func signedNotifyBody(order model.Order, payType, tradeNo, status string) string {
	params := map[string]string{
		"pid":          "pid",
		"type":         payType,
		"trade_no":     tradeNo,
		"out_trade_no": order.OrderNo,
		"name":         "test",
		"money":        "9.90",
		"trade_status": status,
		"sign_type":    "MD5",
	}
	signed := epay.GenerateParams(params, "secret")
	values := url.Values{}
	for key, value := range signed {
		values.Set(key, value)
	}
	return values.Encode()
}

func postForm(engine http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
