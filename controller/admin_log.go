package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

func AdminGenerationLogs(c *gin.Context) {
	page, pageSize := pagination(c)
	query := model.DB.Model(&model.Generation{})
	if status := c.Query("status"); status != "" {
		if parsed, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", parsed)
		}
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	query = applyTimeRange(c, query)

	var total int64
	var items []model.Generation
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count generation logs"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list generation logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func AdminLoginLogs(c *gin.Context) {
	page, pageSize := pagination(c)
	query := model.DB.Model(&model.LoginLog{})
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	query = applyTimeRange(c, query)

	var total int64
	var items []model.LoginLog
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count login logs"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list login logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func AdminDeleteGenerationLogs(c *gin.Context) {
	before, ok := parseBefore(c)
	if !ok {
		return
	}
	result := model.DB.Where("created_at < ?", before).Delete(&model.Generation{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete generation logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": result.RowsAffected})
}

func AdminDeleteLoginLogs(c *gin.Context) {
	before, ok := parseBefore(c)
	if !ok {
		return
	}
	result := model.DB.Where("created_at < ?", before).Delete(&model.LoginLog{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete login logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": result.RowsAffected})
}

func pagination(c *gin.Context) (int, int) {
	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func applyTimeRange(c *gin.Context, db *gorm.DB) *gorm.DB {
	if start := c.Query("start"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			db = db.Where("created_at >= ?", parsed)
		}
	}
	if end := c.Query("end"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			db = db.Where("created_at <= ?", parsed)
		}
	}
	return db
}

func parseBefore(c *gin.Context) (time.Time, bool) {
	value := c.Query("before")
	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "before is required"})
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before"})
		return time.Time{}, false
	}
	return parsed, true
}
