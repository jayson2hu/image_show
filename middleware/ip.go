package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.GetHeader("CF-Connecting-IP")
		if ip == "" {
			ip = c.GetHeader("X-Real-IP")
		}
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
			if idx := strings.Index(ip, ","); idx != -1 {
				ip = strings.TrimSpace(ip[:idx])
			}
		}
		if ip == "" {
			ip = c.ClientIP()
		}
		c.Set("real_ip", ip)
		c.Next()
	}
}
