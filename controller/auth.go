package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func SendCode(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "email registration is disabled"})
}

func Register(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "email registration is disabled"})
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
		if errors.Is(err, service.ErrUserDisabled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
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
