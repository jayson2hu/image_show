package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

type adminStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

type adminRoleRequest struct {
	Role int `json:"role" binding:"required"`
}

type adminCreditRequest struct {
	Amount float64 `json:"amount" binding:"required"`
	Remark string  `json:"remark"`
}

func AdminUsers(c *gin.Context) {
	page, pageSize := pagination(c)
	query := model.DB.Model(&model.User{})
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("email LIKE ? OR username LIKE ?", like, like)
	}
	if status := c.Query("status"); status != "" {
		if parsed, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", parsed)
		}
	}

	var total int64
	var users []model.User
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
		return
	}
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": total, "page": page, "pageSize": pageSize})
}

func AdminUpdateUserStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req adminStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != 1 && req.Status != 2) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminUpdateUserRole(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req adminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Role < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", id).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminUserGenerations(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	page, pageSize := pagination(c)
	query := model.DB.Model(&model.Generation{}).Where("user_id = ?", id)

	var total int64
	var items []model.Generation
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count generations"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list generations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func AdminTopupCredits(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req adminCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}
	operatorID := c.GetInt64("userID")
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "admin topup"
	}
	if err := service.AdminTopup(id, operatorID, req.Amount, remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to top up credits"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminCreditLogs(c *gin.Context) {
	page, pageSize := pagination(c)
	query := model.DB.Model(&model.CreditLog{})
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if logType := c.Query("type"); logType != "" {
		if parsed, err := strconv.Atoi(logType); err == nil {
			query = query.Where("type = ?", parsed)
		}
	}
	query = applyTimeRange(c, query)

	var total int64
	var items []model.CreditLog
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count credit logs"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credit logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}
