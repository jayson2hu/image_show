package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/service"
)

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			if claims, err := service.ParseToken(strings.TrimPrefix(header, "Bearer ")); err == nil {
				c.Set("userID", claims.UserID)
				c.Set("role", claims.Role)
			}
		}
		c.Next()
	}
}
