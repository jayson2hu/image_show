package service

import (
	"strings"
	"time"

	"github.com/jayson2hu/image-show/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthResult struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func Register(email, password, code, ip, anonymousID string) (*AuthResult, error) {
	if !model.RegisterEnabled() {
		return nil, ErrRegisterDisabled
	}
	if !registerEmailDomainAllowed(email) {
		return nil, ErrEmailDomainNotAllowed
	}
	if !VerifyCode(email, code) {
		return nil, ErrInvalidVerificationCode
	}

	var count int64
	if err := model.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         1,
		Status:       1,
		LastLoginIP:  ip,
	}
	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := RegisterGift(tx, &user); err != nil {
			return err
		}
		if anonymousID != "" {
			if err := tx.Model(&model.Generation{}).
				Where("anonymous_id = ? AND user_id IS NULL", anonymousID).
				Update("user_id", user.ID).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.AnonymousIdentity{}).
				Where("anonymous_id = ?", anonymousID).
				Update("claimed_by_user_id", user.ID).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	token, err := GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func registerEmailDomainAllowed(email string) bool {
	raw := strings.TrimSpace(model.GetSettingValue("register_email_domain_allowlist", ""))
	if raw == "" {
		return true
	}
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return false
	}
	domain := strings.ToLower(strings.TrimSpace(email[at+1:]))
	for _, item := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == ';' || r == ' '
	}) {
		allowed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(item)), "@")
		if allowed != "" && domain == allowed {
			return true
		}
	}
	return false
}

func Login(email, password, ip, userAgent string) (*AuthResult, error) {
	var user model.User
	if err := model.DB.Where("email = ?", email).First(&user).Error; err != nil {
		recordLoginLog(0, ip, userAgent, "email", false)
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		recordLoginLog(user.ID, ip, userAgent, "email", false)
		return nil, ErrInvalidCredentials
	}
	if user.Status != 1 {
		recordLoginLog(user.ID, ip, userAgent, "email", false)
		return nil, ErrUserDisabled
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ip
	if err := model.DB.Save(&user).Error; err != nil {
		return nil, err
	}
	recordLoginLog(user.ID, ip, userAgent, "email", true)

	token, err := GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func recordLoginLog(userID int64, ip, userAgent, method string, success bool) {
	_ = model.DB.Create(&model.LoginLog{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		Method:    method,
		Success:   success,
	}).Error
}
