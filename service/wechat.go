package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

var (
	ErrWeChatDisabled      = errors.New("wechat login is disabled")
	ErrWeChatNotConfigured = errors.New("wechat server is not configured")
	ErrWeChatCodeInvalid   = errors.New("wechat code is invalid")
	ErrWeChatAlreadyBound  = errors.New("wechat account is already bound")
	ErrWeChatNotBound      = errors.New("wechat account is not bound")
)

type WeChatQRCodeInfo struct {
	Enabled   bool   `json:"enabled"`
	QRCodeURL string `json:"qrcode_url"`
	Mode      string `json:"mode"`
}

type weChatServerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func WeChatQRCode() WeChatQRCodeInfo {
	return WeChatQRCodeInfo{
		Enabled:   weChatEnabled(),
		QRCodeURL: weChatSetting("wechat_qrcode_url", weChatConfig().WeChatQRCode),
		Mode:      "new-api-code",
	}
}

func WeChatLogin(code, ip, userAgent string) (*AuthResult, error) {
	openID, err := WeChatOpenIDByCode(code)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = model.DB.Where("wechat_open_id = ?", openID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !model.RegisterEnabled() {
			return nil, ErrRegisterDisabled
		}
		user = model.User{
			Username:     "wechat_" + strconv.FormatInt(time.Now().UnixNano(), 36),
			Email:        weChatSyntheticEmail(openID),
			WechatOpenID: openID,
			Role:         1,
			Status:       1,
			LastLoginIP:  ip,
		}
		if err := model.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			return RegisterGift(tx, &user)
		}); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if user.Status != 1 {
		recordLoginLog(user.ID, ip, userAgent, "wechat", false)
		return nil, ErrUserDisabled
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ip
	if err := model.DB.Save(&user).Error; err != nil {
		return nil, err
	}
	recordLoginLog(user.ID, ip, userAgent, "wechat", true)

	token, err := GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func BindWeChat(userID int64, code string) error {
	openID, err := WeChatOpenIDByCode(code)
	if err != nil {
		return err
	}
	var count int64
	if err := model.DB.Model(&model.User{}).Where("wechat_open_id = ? AND id <> ?", openID, userID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrWeChatAlreadyBound
	}
	result := model.DB.Model(&model.User{}).Where("id = ?", userID).Update("wechat_open_id", openID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func UnbindWeChat(userID int64) error {
	result := model.DB.Model(&model.User{}).Where("id = ? AND wechat_open_id <> ''", userID).Update("wechat_open_id", "")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWeChatNotBound
	}
	return nil
}

func WeChatOpenIDByCode(code string) (string, error) {
	if !weChatEnabled() {
		return "", ErrWeChatDisabled
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return "", ErrWeChatCodeInvalid
	}
	cfg := weChatConfig()
	server := strings.TrimRight(weChatSetting("wechat_server_address", cfg.WeChatServer), "/")
	token := weChatSetting("wechat_server_token", cfg.WeChatToken)
	if server == "" || token == "" {
		return "", ErrWeChatNotConfigured
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/wechat/user?code=%s", server, url.QueryEscape(code)), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", token)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload weChatServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if !payload.Success || payload.Data == "" {
		if payload.Message != "" {
			return "", fmt.Errorf("%w: %s", ErrWeChatCodeInvalid, payload.Message)
		}
		return "", ErrWeChatCodeInvalid
	}
	return payload.Data, nil
}

func weChatEnabled() bool {
	return weChatSetting("wechat_auth_enabled", strconv.FormatBool(weChatConfig().WeChatEnabled)) == "true"
}

func weChatSetting(key, fallback string) string {
	return model.GetSettingValue(key, fallback)
}

func weChatConfig() *config.Config {
	if config.AppConfig == nil {
		return config.LoadConfig()
	}
	return config.AppConfig
}

func weChatSyntheticEmail(openID string) string {
	replacer := strings.NewReplacer("@", "_", ":", "_", "/", "_", "\\", "_", " ", "_")
	return "wechat_" + replacer.Replace(openID) + "@wechat.local"
}
