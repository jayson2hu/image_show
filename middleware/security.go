package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; connect-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'")
		c.Next()
	}
}

func IPBlacklist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.GetString("real_ip")
		if ip == "" {
			ip = c.ClientIP()
		}
		if blacklistedIP(ip) {
			c.JSON(http.StatusForbidden, gin.H{"error": "ip is blocked"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func blacklistedIP(ip string) bool {
	value := model.GetSettingValue("ip_blacklist", "")
	if value == "" {
		return false
	}
	for _, item := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	}) {
		if strings.TrimSpace(item) == ip {
			return true
		}
	}
	return false
}
