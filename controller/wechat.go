package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/service"
	"gorm.io/gorm"
)

type weChatBindRequest struct {
	Code string `json:"code" binding:"required"`
}

func WeChatQRCode(c *gin.Context) {
	c.JSON(http.StatusOK, service.WeChatQRCode())
}

func WeChatCallback(c *gin.Context) {
	result, err := service.WeChatLogin(c.Query("code"), common.GetRealIP(c), c.GetHeader("User-Agent"))
	if err != nil {
		writeWeChatError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func WeChatStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"mode":    "new-api-code",
		"message": "submit the WeChat verification code to /api/auth/wechat/callback?code=...",
	})
}

func WeChatBind(c *gin.Context) {
	var req weChatBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := service.BindWeChat(c.GetInt64("userID"), req.Code); err != nil {
		writeWeChatError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func WeChatUnbind(c *gin.Context) {
	if err := service.UnbindWeChat(c.GetInt64("userID")); err != nil {
		writeWeChatError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func writeWeChatError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWeChatDisabled), errors.Is(err, service.ErrRegisterDisabled):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrWeChatCodeInvalid), errors.Is(err, service.ErrWeChatAlreadyBound), errors.Is(err, service.ErrWeChatNotBound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrWeChatNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUserDisabled):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "wechat login failed"})
	}
}
