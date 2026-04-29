package service

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"sync"
	"time"

	"github.com/jayson2hu/image-show/config"
)

type verificationEntry struct {
	Code      string
	ExpiresAt time.Time
	SentAt    time.Time
}

var verificationStore = struct {
	sync.Mutex
	items map[string]verificationEntry
}{items: make(map[string]verificationEntry)}

func SendVerificationCode(email string) error {
	verificationStore.Lock()
	if entry, ok := verificationStore.items[email]; ok && time.Since(entry.SentAt) < time.Minute {
		verificationStore.Unlock()
		return ErrVerificationTooFrequent
	}
	code, err := randomCode()
	if err != nil {
		verificationStore.Unlock()
		return err
	}
	verificationStore.items[email] = verificationEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		SentAt:    time.Now(),
	}
	verificationStore.Unlock()

	return deliverVerificationCode(email, code)
}

func VerifyCode(email, code string) bool {
	verificationStore.Lock()
	defer verificationStore.Unlock()

	entry, ok := verificationStore.items[email]
	if !ok || time.Now().After(entry.ExpiresAt) || entry.Code != code {
		return false
	}
	delete(verificationStore.items, email)
	return true
}

func PeekVerificationCode(email string) string {
	verificationStore.Lock()
	defer verificationStore.Unlock()
	return verificationStore.items[email].Code
}

func randomCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func deliverVerificationCode(email, code string) error {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		log.Printf("verification code for %s: %s", email, code)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUser
	}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Image Show verification code\r\n" +
		"\r\n" +
		"Your verification code is: " + code + "\r\n")
	return smtp.SendMail(addr, auth, from, []string{email}, msg)
}
