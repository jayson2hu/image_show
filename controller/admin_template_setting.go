package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm/clause"
)

type promptTemplateRequest struct {
	Category  string `json:"category" binding:"required"`
	Label     string `json:"label" binding:"required"`
	Prompt    string `json:"prompt" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status"`
}

type settingsRequest struct {
	Items map[string]string `json:"items" binding:"required"`
}

func AdminPromptTemplates(c *gin.Context) {
	var items []model.PromptTemplate
	if err := model.DB.Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list prompt templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func AdminCreatePromptTemplate(c *gin.Context) {
	var req promptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	template := promptTemplateFromRequest(req)
	if err := model.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create prompt template"})
		return
	}
	c.JSON(http.StatusOK, template)
}

func AdminUpdatePromptTemplate(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req promptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := promptTemplateFromRequest(req)
	if err := model.DB.Model(&model.PromptTemplate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update prompt template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminDeletePromptTemplate(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := model.DB.Delete(&model.PromptTemplate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete prompt template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminSettings(c *gin.Context) {
	var items []model.Setting
	if err := model.DB.Order("key ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list settings"})
		return
	}
	values := map[string]string{
		"register_enabled":               model.GetSettingValue("register_enabled", "true"),
		"wechat_auth_enabled":            model.GetSettingValue("wechat_auth_enabled", "false"),
		"wechat_server_address":          model.GetSettingValue("wechat_server_address", ""),
		"wechat_server_token":            model.GetSettingValue("wechat_server_token", ""),
		"wechat_qrcode_url":              model.GetSettingValue("wechat_qrcode_url", ""),
		"ip_blacklist":                   model.GetSettingValue("ip_blacklist", ""),
		"captcha_enabled":                model.GetSettingValue("captcha_enabled", "false"),
		"turnstile_site_key":             model.GetSettingValue("turnstile_site_key", ""),
		"turnstile_secret":               model.GetSettingValue("turnstile_secret", ""),
		"monitor_daily_credit_threshold": model.GetSettingValue("monitor_daily_credit_threshold", "0"),
		"monitor_alert_last_date":        model.GetSettingValue("monitor_alert_last_date", ""),
	}
	for _, item := range items {
		values[item.Key] = item.Value
	}
	c.JSON(http.StatusOK, gin.H{"items": values})
}

func AdminUpdateSettings(c *gin.Context) {
	var req settingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	settings := make([]model.Setting, 0, len(req.Items))
	for key, value := range req.Items {
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "setting key is required"})
			return
		}
		settings = append(settings, model.Setting{Key: key, Value: value})
	}
	if len(settings) > 0 {
		err := model.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).Create(&settings).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func promptTemplateFromRequest(req promptTemplateRequest) model.PromptTemplate {
	status := req.Status
	if status == 0 {
		status = 1
	}
	return model.PromptTemplate{
		Category:  req.Category,
		Label:     req.Label,
		Prompt:    req.Prompt,
		SortOrder: req.SortOrder,
		Status:    status,
	}
}
