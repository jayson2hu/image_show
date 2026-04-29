package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

var (
	ErrCaptchaRequired = errors.New("captcha required")
	ErrCaptchaInvalid  = errors.New("captcha invalid")

	turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	captchaHTTPClient  = &http.Client{Timeout: 8 * time.Second}
)

type CaptchaConfig struct {
	Enabled bool   `json:"enabled"`
	SiteKey string `json:"site_key"`
}

type turnstileResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

func GetCaptchaConfig() CaptchaConfig {
	enabled := model.GetSettingValue("captcha_enabled", "false") == "true"
	siteKey := model.GetSettingValue("turnstile_site_key", configValue(func(c *config.Config) string { return c.TurnstileSiteKey }))
	secret := model.GetSettingValue("turnstile_secret", configValue(func(c *config.Config) string { return c.TurnstileSecret }))
	if siteKey == "" || secret == "" {
		enabled = false
	}
	return CaptchaConfig{Enabled: enabled, SiteKey: siteKey}
}

func VerifyCaptcha(token, ip string) error {
	cfg := GetCaptchaConfig()
	if !cfg.Enabled {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return ErrCaptchaRequired
	}
	secret := model.GetSettingValue("turnstile_secret", configValue(func(c *config.Config) string { return c.TurnstileSecret }))
	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if ip != "" {
		form.Set("remoteip", ip)
	}
	resp, err := captchaHTTPClient.PostForm(turnstileVerifyURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var parsed turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}
	if !parsed.Success {
		return ErrCaptchaInvalid
	}
	return nil
}

func configValue(read func(*config.Config) string) string {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	return read(cfg)
}

func SetCaptchaVerifyURLForTest(rawURL string) func() {
	original := turnstileVerifyURL
	turnstileVerifyURL = rawURL
	return func() { turnstileVerifyURL = original }
}
