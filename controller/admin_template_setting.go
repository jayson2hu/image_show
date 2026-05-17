package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm/clause"
)

type promptTemplateRequest struct {
	Category         string `json:"category" binding:"required"`
	Label            string `json:"label" binding:"required"`
	Prompt           string `json:"prompt"`
	Icon             string `json:"icon"`
	RecommendedRatio string `json:"recommended_ratio"`
	Description      string `json:"description"`
	SortOrder        int    `json:"sort_order"`
	Status           int    `json:"status"`
}

type settingsRequest struct {
	Items map[string]string `json:"items" binding:"required"`
}

var adminSettingDefaults = map[string]string{
	"site_title":                         "来看看巴",
	"site_about":                         "把想法变成一张好图。",
	"seo_title":                          "来看看巴 - AI 图片生成",
	"seo_keywords":                       "AI图片生成,图片生成,AI绘画",
	"seo_description":                    "输入提示词，选择合适比例，持续查看生成进度，直到作品完成。",
	"register_enabled":                   "true",
	"register_email_domain_allowlist":    "",
	"wechat_auth_enabled":                "false",
	"wechat_server_address":              "",
	"wechat_server_token":                "",
	"wechat_qrcode_url":                  "",
	"ip_blacklist":                       "",
	"captcha_enabled":                    "false",
	"turnstile_site_key":                 "",
	"turnstile_secret":                   "",
	"monitor_daily_credit_threshold":     "0",
	"monitor_alert_last_date":            "",
	"register_gift_credits":              "10",
	"guest_free_credits":                 "5",
	"guest_generation_limit":             "5",
	"guest_layered_generation_limit":     "1",
	"user_generation_limit":              "100",
	"user_layered_generation_limit":      "10",
	"greeting_text":                      "",
	"credit_cost_square":                 "1",
	"credit_cost_portrait":               "2",
	"credit_cost_story":                  "2",
	"credit_cost_landscape":              "2",
	"credit_cost_widescreen":             "2",
	"credit_exhausted_message":           "额度已用完，可以注册账号获取新用户积分；如需人工开通或咨询套餐，请联系管理员。",
	"credit_exhausted_wechat_qrcode_url": "",
	"credit_exhausted_qq":                "",
	"manual_recharge_enabled":            "true",
	"manual_recharge_wechat_id":          "",
	"manual_recharge_wechat_qrcode_url":  "",
	"manual_recharge_qq":                 "",
	"manual_recharge_note":               "请联系管理员人工充值，并备注账号邮箱和需要开通的套餐。",
	"avatar_storage_driver":              "local",
	"avatar_max_size_mb":                 "2",
	"avatar_allowed_types":               "jpg,jpeg,png,webp",
	"image_model":                        "gpt-image-2",
	"enabled_image_sizes":                defaultEnabledImageSizes,
	"r2_endpoint":                        "",
	"r2_access_key":                      "",
	"r2_secret_key":                      "",
	"r2_bucket":                          "image-show",
	"r2_public_url":                      "",
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
	if !validPromptTemplateRequest(req) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
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
	if !validPromptTemplateRequest(req) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	if err := model.DB.Model(&model.PromptTemplate{}).Where("id = ?", id).Updates(promptTemplateUpdateMap(req)).Error; err != nil {
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
		"site_title":                         model.GetSettingValue("site_title", adminSettingDefaults["site_title"]),
		"site_about":                         model.GetSettingValue("site_about", adminSettingDefaults["site_about"]),
		"seo_title":                          model.GetSettingValue("seo_title", adminSettingDefaults["seo_title"]),
		"seo_keywords":                       model.GetSettingValue("seo_keywords", adminSettingDefaults["seo_keywords"]),
		"seo_description":                    model.GetSettingValue("seo_description", adminSettingDefaults["seo_description"]),
		"register_enabled":                   model.GetSettingValue("register_enabled", "true"),
		"register_email_domain_allowlist":    model.GetSettingValue("register_email_domain_allowlist", ""),
		"wechat_auth_enabled":                model.GetSettingValue("wechat_auth_enabled", "false"),
		"wechat_server_address":              model.GetSettingValue("wechat_server_address", ""),
		"wechat_server_token":                model.GetSettingValue("wechat_server_token", ""),
		"wechat_qrcode_url":                  model.GetSettingValue("wechat_qrcode_url", ""),
		"ip_blacklist":                       model.GetSettingValue("ip_blacklist", ""),
		"captcha_enabled":                    model.GetSettingValue("captcha_enabled", "false"),
		"turnstile_site_key":                 model.GetSettingValue("turnstile_site_key", ""),
		"turnstile_secret":                   model.GetSettingValue("turnstile_secret", ""),
		"monitor_daily_credit_threshold":     model.GetSettingValue("monitor_daily_credit_threshold", "0"),
		"monitor_alert_last_date":            model.GetSettingValue("monitor_alert_last_date", ""),
		"register_gift_credits":              model.GetSettingValue("register_gift_credits", "10"),
		"guest_free_credits":                 model.GetSettingValue("guest_free_credits", adminSettingDefaults["guest_free_credits"]),
		"guest_generation_limit":             model.GetSettingValue("guest_generation_limit", adminSettingDefaults["guest_generation_limit"]),
		"guest_layered_generation_limit":     model.GetSettingValue("guest_layered_generation_limit", adminSettingDefaults["guest_layered_generation_limit"]),
		"user_generation_limit":              model.GetSettingValue("user_generation_limit", adminSettingDefaults["user_generation_limit"]),
		"user_layered_generation_limit":      model.GetSettingValue("user_layered_generation_limit", adminSettingDefaults["user_layered_generation_limit"]),
		"greeting_text":                      model.GetSettingValue("greeting_text", adminSettingDefaults["greeting_text"]),
		"credit_cost_square":                 model.GetSettingValue("credit_cost_square", adminSettingDefaults["credit_cost_square"]),
		"credit_cost_portrait":               model.GetSettingValue("credit_cost_portrait", adminSettingDefaults["credit_cost_portrait"]),
		"credit_cost_story":                  model.GetSettingValue("credit_cost_story", adminSettingDefaults["credit_cost_story"]),
		"credit_cost_landscape":              model.GetSettingValue("credit_cost_landscape", adminSettingDefaults["credit_cost_landscape"]),
		"credit_cost_widescreen":             model.GetSettingValue("credit_cost_widescreen", adminSettingDefaults["credit_cost_widescreen"]),
		"credit_exhausted_message":           model.GetSettingValue("credit_exhausted_message", "额度已用完，可以注册账号获取新用户积分；如需人工开通或咨询套餐，请联系管理员。"),
		"credit_exhausted_wechat_qrcode_url": model.GetSettingValue("credit_exhausted_wechat_qrcode_url", ""),
		"credit_exhausted_qq":                model.GetSettingValue("credit_exhausted_qq", ""),
		"manual_recharge_enabled":            model.GetSettingValue("manual_recharge_enabled", adminSettingDefaults["manual_recharge_enabled"]),
		"manual_recharge_wechat_id":          model.GetSettingValue("manual_recharge_wechat_id", adminSettingDefaults["manual_recharge_wechat_id"]),
		"manual_recharge_wechat_qrcode_url":  model.GetSettingValue("manual_recharge_wechat_qrcode_url", adminSettingDefaults["manual_recharge_wechat_qrcode_url"]),
		"manual_recharge_qq":                 model.GetSettingValue("manual_recharge_qq", adminSettingDefaults["manual_recharge_qq"]),
		"manual_recharge_note":               model.GetSettingValue("manual_recharge_note", adminSettingDefaults["manual_recharge_note"]),
		"avatar_storage_driver":              model.GetSettingValue("avatar_storage_driver", adminSettingDefaults["avatar_storage_driver"]),
		"avatar_max_size_mb":                 model.GetSettingValue("avatar_max_size_mb", adminSettingDefaults["avatar_max_size_mb"]),
		"avatar_allowed_types":               model.GetSettingValue("avatar_allowed_types", adminSettingDefaults["avatar_allowed_types"]),
		"image_model":                        model.GetSettingValue("image_model", "gpt-image-2"),
		"enabled_image_sizes":                enabledImageSizesSettingValue(),
		"r2_endpoint":                        model.GetSettingValue("r2_endpoint", ""),
		"r2_access_key":                      model.GetSettingValue("r2_access_key", ""),
		"r2_secret_key":                      model.GetSettingValue("r2_secret_key", ""),
		"r2_bucket":                          model.GetSettingValue("r2_bucket", "image-show"),
		"r2_public_url":                      model.GetSettingValue("r2_public_url", ""),
	}
	for _, item := range items {
		if item.Key == "enabled_image_sizes" {
			values[item.Key] = normalizeEnabledImageSizesSetting(item.Value)
			continue
		}
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
		Category:         req.Category,
		Label:            req.Label,
		Prompt:           req.Prompt,
		Icon:             req.Icon,
		RecommendedRatio: req.RecommendedRatio,
		Description:      req.Description,
		SortOrder:        req.SortOrder,
		Status:           status,
	}
}

func promptTemplateUpdateMap(req promptTemplateRequest) map[string]interface{} {
	template := promptTemplateFromRequest(req)
	return map[string]interface{}{
		"category":          template.Category,
		"label":             template.Label,
		"prompt":            template.Prompt,
		"icon":              template.Icon,
		"recommended_ratio": template.RecommendedRatio,
		"description":       template.Description,
		"sort_order":        template.SortOrder,
		"status":            template.Status,
	}
}

func validPromptTemplateRequest(req promptTemplateRequest) bool {
	if strings.TrimSpace(req.Category) == "scene" && strings.TrimSpace(req.Label) == "自由创作" {
		return true
	}
	return strings.TrimSpace(req.Prompt) != ""
}
