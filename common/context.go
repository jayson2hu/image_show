package common

import "github.com/gin-gonic/gin"

func GetRealIP(c *gin.Context) string {
	if value, exists := c.Get("real_ip"); exists {
		if ip, ok := value.(string); ok && ip != "" {
			return ip
		}
	}
	return c.ClientIP()
}
