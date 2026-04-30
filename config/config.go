package config

import (
	"os"
	"strconv"

	"github.com/jayson2hu/image-show/common"
)

type Config struct {
	AppEnv           string
	Port             int
	DBDriver         string
	DatabaseDSN      string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	JWTSecret        string
	R2Endpoint       string
	R2AccessKey      string
	R2SecretKey      string
	R2Bucket         string
	R2PublicURL      string
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPassword     string
	SMTPFrom         string
	Sub2APIBaseURL   string
	ImageModel       string
	MockSub2API      bool
	WeChatEnabled    bool
	WeChatServer     string
	WeChatToken      string
	WeChatQRCode     string
	ServerAddress    string
	PayAddress       string
	EpayID           string
	EpayKey          string
	EpayPayMethods   string
	TurnstileSiteKey string
	TurnstileSecret  string
	AdminEmail       string
	AdminPassword    string
}

var AppConfig *Config

func LoadConfig() *Config {
	cfg := &Config{
		AppEnv:           getEnv("APP_ENV", common.EnvDevelopment),
		Port:             getEnvInt("PORT", 3000),
		DBDriver:         getEnv("DB_DRIVER", "sqlite"),
		DatabaseDSN:      getEnv("DATABASE_DSN", ""),
		RedisAddr:        getEnv("REDIS_ADDR", ""),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvInt("REDIS_DB", 1),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		R2Endpoint:       getEnv("R2_ENDPOINT", ""),
		R2AccessKey:      getEnv("R2_ACCESS_KEY", ""),
		R2SecretKey:      getEnv("R2_SECRET_KEY", ""),
		R2Bucket:         getEnv("R2_BUCKET", "image-show"),
		R2PublicURL:      getEnv("R2_PUBLIC_URL", ""),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnvInt("SMTP_PORT", 465),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:         getEnv("SMTP_FROM", ""),
		Sub2APIBaseURL:   getEnv("SUB2API_BASE_URL", "http://sub2api:8080"),
		ImageModel:       getEnv("IMAGE_MODEL", "gpt-image-1"),
		MockSub2API:      getEnvBool("MOCK_SUB2API", false),
		WeChatEnabled:    getEnvBool("WECHAT_AUTH_ENABLED", false),
		WeChatServer:     getEnv("WECHAT_SERVER_ADDRESS", ""),
		WeChatToken:      getEnv("WECHAT_SERVER_TOKEN", ""),
		WeChatQRCode:     getEnv("WECHAT_QRCODE_URL", ""),
		ServerAddress:    getEnv("SERVER_ADDRESS", ""),
		PayAddress:       getEnv("PAY_ADDRESS", ""),
		EpayID:           getEnv("EPAY_ID", ""),
		EpayKey:          getEnv("EPAY_KEY", ""),
		EpayPayMethods:   getEnv("EPAY_PAY_METHODS", "alipay,wxpay"),
		TurnstileSiteKey: getEnv("TURNSTILE_SITE_KEY", ""),
		TurnstileSecret:  getEnv("TURNSTILE_SECRET", ""),
	}
	defaultAdminEmail := ""
	defaultAdminPassword := ""
	if cfg.AppEnv != common.EnvProduction {
		defaultAdminEmail = "admin@image-show.local"
		defaultAdminPassword = "Admin123456"
	}
	cfg.AdminEmail = getEnv("ADMIN_EMAIL", defaultAdminEmail)
	cfg.AdminPassword = getEnv("ADMIN_PASSWORD", defaultAdminPassword)
	AppConfig = cfg
	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
