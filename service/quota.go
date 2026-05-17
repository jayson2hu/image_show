package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jayson2hu/image-show/model"
)

var ErrGenerationQuotaExceeded = errors.New("generation quota exceeded")
var ErrLayeredGenerationQuotaExceeded = errors.New("layered generation quota exceeded")

const (
	DefaultGuestGenerationLimit        = 5
	DefaultGuestLayeredGenerationLimit = 1
	DefaultUserGenerationLimit         = 100
	DefaultUserLayeredGenerationLimit  = 10
)

func GuestGenerationLimit() int {
	return positiveIntSetting("guest_generation_limit", DefaultGuestGenerationLimit)
}

func GuestLayeredGenerationLimit() int {
	return positiveIntSetting("guest_layered_generation_limit", DefaultGuestLayeredGenerationLimit)
}

func UserGenerationLimit() int {
	return positiveIntSetting("user_generation_limit", DefaultUserGenerationLimit)
}

func UserLayeredGenerationLimit() int {
	return positiveIntSetting("user_layered_generation_limit", DefaultUserLayeredGenerationLimit)
}

func EnsureAnonymousGenerationQuota(anonymousID string, layered bool) error {
	if anonymousID == "" {
		return ErrGenerationQuotaExceeded
	}
	if err := ensureAnonymousTotalQuota(anonymousID); err != nil {
		return err
	}
	if layered {
		return ensureAnonymousLayeredQuota(anonymousID)
	}
	return nil
}

func EnsureUserGenerationQuota(userID int64, layered bool) error {
	if err := ensureUserTotalQuota(userID); err != nil {
		return err
	}
	if layered {
		return ensureUserLayeredQuota(userID)
	}
	return nil
}

func ensureAnonymousTotalQuota(anonymousID string) error {
	var count int64
	if err := model.DB.Model(&model.Generation{}).
		Where("anonymous_id = ? AND is_deleted = ?", anonymousID, false).
		Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(GuestGenerationLimit()) {
		return ErrGenerationQuotaExceeded
	}
	return nil
}

func ensureAnonymousLayeredQuota(anonymousID string) error {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("anonymous_id = ? AND layered = ?", anonymousID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(GuestLayeredGenerationLimit()) {
		return ErrLayeredGenerationQuotaExceeded
	}
	return nil
}

func ensureUserTotalQuota(userID int64) error {
	var count int64
	if err := model.DB.Model(&model.Generation{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(UserGenerationLimit()) {
		return ErrGenerationQuotaExceeded
	}
	return nil
}

func ensureUserLayeredQuota(userID int64) error {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND layered = ?", userID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(UserLayeredGenerationLimit()) {
		return ErrLayeredGenerationQuotaExceeded
	}
	return nil
}

func positiveIntSetting(key string, fallback int) int {
	value := strings.TrimSpace(model.GetSettingValue(key, strconv.Itoa(fallback)))
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
