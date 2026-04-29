package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func CreditBalance(c *gin.Context) {
	userID := c.GetInt64("userID")
	balance, err := service.GetBalance(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balance"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func CreditLogs(c *gin.Context) {
	userID := c.GetInt64("userID")
	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	var logs []model.CreditLog
	var total int64
	query := model.DB.Model(&model.CreditLog{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": logs, "total": total, "page": page, "pageSize": pageSize})
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
