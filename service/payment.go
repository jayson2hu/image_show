package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

const (
	OrderStatusPending  = 0
	OrderStatusPaid     = 1
	OrderStatusExpired  = 2
	OrderStatusRefunded = 3

	CreditLogTypePaymentTopup = 5
)

var (
	ErrPaymentNotConfigured = errors.New("payment not configured")
	ErrInvalidPayMethod     = errors.New("invalid payment method")
	ErrPackageUnavailable   = errors.New("package unavailable")
	ErrOrderNotPayable      = errors.New("order not payable")
	ErrPaymentVerifyFailed  = errors.New("payment verify failed")
)

type CreatedOrder struct {
	Order  model.Order       `json:"order"`
	PayURL string            `json:"pay_url"`
	Params map[string]string `json:"params"`
}

type PaymentNotify struct {
	OrderNo string
	Status  string
}

var paymentOrderLocks sync.Map
var paymentLockGuard sync.Mutex

type paymentOrderLock struct {
	mu       sync.Mutex
	refCount int
}

func CreatePaymentOrder(userID, packageID int64, payMethod string) (*CreatedOrder, error) {
	if err := ExpireStaleOrders(); err != nil {
		return nil, err
	}

	normalizedMethod, err := normalizePayMethod(payMethod)
	if err != nil {
		return nil, err
	}

	client, err := NewEpayClient()
	if err != nil {
		return nil, err
	}

	callbackAddress, err := GetCallbackAddress()
	if err != nil {
		return nil, err
	}
	notifyURL, err := url.Parse(callbackAddress + "/api/payment/notify")
	if err != nil {
		return nil, err
	}
	returnURL, err := url.Parse(callbackAddress + "/packages?pay=return")
	if err != nil {
		return nil, err
	}

	var pkg model.Package
	if err := model.DB.Where("id = ? AND status = ?", packageID, 1).First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPackageUnavailable
		}
		return nil, err
	}
	if pkg.Price < 0.01 || pkg.Credits <= 0 || pkg.ValidDays <= 0 {
		return nil, ErrPackageUnavailable
	}

	order := model.Order{
		OrderNo:   generateOrderNo(userID),
		UserID:    userID,
		PackageID: pkg.ID,
		Amount:    pkg.Price,
		Status:    OrderStatusPending,
		PayMethod: normalizedMethod,
	}
	if err := model.DB.Create(&order).Error; err != nil {
		return nil, err
	}

	uri, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           normalizedMethod,
		ServiceTradeNo: order.OrderNo,
		Name:           fmt.Sprintf("PKG:%s", pkg.Name),
		Money:          strconv.FormatFloat(pkg.Price, 'f', 2, 64),
		Device:         epay.PC,
		NotifyUrl:      notifyURL,
		ReturnUrl:      returnURL,
	})
	if err != nil {
		_ = model.DB.Model(&model.Order{}).Where("id = ? AND status = ?", order.ID, OrderStatusPending).Update("status", OrderStatusExpired).Error
		return nil, err
	}

	return &CreatedOrder{Order: order, PayURL: uri, Params: params}, nil
}

func GetUserOrder(userID, orderID int64) (*model.Order, error) {
	if err := ExpireStaleOrders(); err != nil {
		return nil, err
	}
	var order model.Order
	if err := model.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func HandlePaymentNotify(params map[string]string) (*PaymentNotify, error) {
	client, err := NewEpayClient()
	if err != nil {
		return nil, err
	}
	verifyInfo, err := client.Verify(params)
	if err != nil || !verifyInfo.VerifyStatus {
		return nil, ErrPaymentVerifyFailed
	}
	if verifyInfo.TradeStatus != epay.StatusTradeSuccess {
		return &PaymentNotify{OrderNo: verifyInfo.ServiceTradeNo, Status: verifyInfo.TradeStatus}, nil
	}

	lockPaymentOrder(verifyInfo.ServiceTradeNo)
	defer unlockPaymentOrder(verifyInfo.ServiceTradeNo)

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Where("order_no = ?", verifyInfo.ServiceTradeNo).First(&order).Error; err != nil {
			return err
		}
		if order.PayMethod != verifyInfo.Type {
			return ErrInvalidPayMethod
		}
		if order.Status == OrderStatusPaid {
			return nil
		}
		if order.Status != OrderStatusPending {
			return ErrOrderNotPayable
		}

		var pkg model.Package
		if err := tx.First(&pkg, order.PackageID).Error; err != nil {
			return err
		}
		if fmt.Sprintf("%.2f", order.Amount) != verifyInfo.Money {
			return ErrPaymentVerifyFailed
		}

		var user model.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			return err
		}
		now := time.Now()
		expiryBase := now
		if user.CreditsExpiry != nil && user.CreditsExpiry.After(now) {
			expiryBase = *user.CreditsExpiry
		}
		expiry := expiryBase.Add(time.Duration(pkg.ValidDays) * 24 * time.Hour)
		user.Credits += pkg.Credits
		user.CreditsExpiry = &expiry
		if err := tx.Model(&user).Updates(map[string]interface{}{"credits": user.Credits, "credits_expiry": user.CreditsExpiry}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Order{}).Where("id = ? AND status = ?", order.ID, OrderStatusPending).Updates(map[string]interface{}{
			"status":       OrderStatusPaid,
			"pay_trade_no": verifyInfo.TradeNo,
			"paid_at":      now,
		}).Error; err != nil {
			return err
		}

		relatedID := order.ID
		return tx.Create(&model.CreditLog{
			UserID:    user.ID,
			Type:      CreditLogTypePaymentTopup,
			Amount:    pkg.Credits,
			Balance:   user.Credits,
			RelatedID: &relatedID,
			Remark:    "payment topup: " + order.OrderNo,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &PaymentNotify{OrderNo: verifyInfo.ServiceTradeNo, Status: verifyInfo.TradeStatus}, PromoteUserGenerationsToPaidByPayment(verifyInfo.ServiceTradeNo)
}

func ExpireStaleOrders() error {
	deadline := time.Now().Add(-30 * time.Minute)
	return model.DB.Model(&model.Order{}).
		Where("status = ? AND created_at < ?", OrderStatusPending, deadline).
		Update("status", OrderStatusExpired).Error
}

func StartOrderExpiryLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = ExpireStaleOrders()
			case <-stop:
				return
			}
		}
	}()
}

func PromoteUserGenerationsToPaidByPayment(orderNo string) error {
	var order model.Order
	if err := model.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return err
	}
	return PromoteUserGenerationsToPaid(order.UserID)
}

func NewEpayClient() (*epay.Client, error) {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if strings.TrimSpace(cfg.PayAddress) == "" || strings.TrimSpace(cfg.EpayID) == "" || strings.TrimSpace(cfg.EpayKey) == "" {
		return nil, ErrPaymentNotConfigured
	}
	client, err := epay.NewClient(&epay.Config{PartnerID: cfg.EpayID, Key: cfg.EpayKey}, cfg.PayAddress)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetCallbackAddress() (string, error) {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	address := strings.TrimRight(strings.TrimSpace(cfg.ServerAddress), "/")
	if address == "" {
		return "", ErrPaymentNotConfigured
	}
	return address, nil
}

func normalizePayMethod(method string) (string, error) {
	method = strings.TrimSpace(strings.ToLower(method))
	if method == "wechat" {
		method = "wxpay"
	}
	if method == "" {
		return "", ErrInvalidPayMethod
	}

	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	for _, allowed := range strings.Split(cfg.EpayPayMethods, ",") {
		allowed = strings.TrimSpace(strings.ToLower(allowed))
		if allowed == "wechat" {
			allowed = "wxpay"
		}
		if method == allowed {
			return method, nil
		}
	}
	return "", ErrInvalidPayMethod
}

func generateOrderNo(userID int64) string {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("IMGUSR%dNO%d", userID, time.Now().UnixNano())
	}
	return fmt.Sprintf("IMGUSR%dNO%s%d", userID, strings.ToUpper(hex.EncodeToString(buf[:])), time.Now().Unix())
}

func lockPaymentOrder(orderNo string) {
	paymentLockGuard.Lock()
	var lock *paymentOrderLock
	if value, ok := paymentOrderLocks.Load(orderNo); ok {
		lock = value.(*paymentOrderLock)
	} else {
		lock = &paymentOrderLock{}
		paymentOrderLocks.Store(orderNo, lock)
	}
	lock.refCount++
	paymentLockGuard.Unlock()
	lock.mu.Lock()
}

func unlockPaymentOrder(orderNo string) {
	value, ok := paymentOrderLocks.Load(orderNo)
	if !ok {
		return
	}
	lock := value.(*paymentOrderLock)
	lock.mu.Unlock()

	paymentLockGuard.Lock()
	lock.refCount--
	if lock.refCount == 0 {
		paymentOrderLocks.Delete(orderNo)
	}
	paymentLockGuard.Unlock()
}
