package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

type sendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Code        string `json:"code" binding:"required,len=6"`
	AnonymousID string `json:"anonymous_id"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func SendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := service.SendVerificationCode(req.Email); err != nil {
		if errors.Is(err, service.ErrVerificationTooFrequent) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "please wait before requesting another code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification code"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	result, err := service.Register(req.Email, req.Password, req.Code, common.GetRealIP(c), req.AnonymousID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRegisterDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "registration is disabled"})
		case errors.Is(err, service.ErrInvalidVerificationCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verification code"})
		case errors.Is(err, service.ErrEmailExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	result, err := service.Login(req.Email, req.Password, common.GetRealIP(c), c.GetHeader("User-Agent"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user model.User
	if err := model.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
