package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

func SiteConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"site_title":         model.GetSettingValue("site_title", adminSettingDefaults["site_title"]),
		"site_about":         model.GetSettingValue("site_about", adminSettingDefaults["site_about"]),
		"seo_title":          model.GetSettingValue("seo_title", adminSettingDefaults["seo_title"]),
		"seo_keywords":       model.GetSettingValue("seo_keywords", adminSettingDefaults["seo_keywords"]),
		"seo_description":    model.GetSettingValue("seo_description", adminSettingDefaults["seo_description"]),
		"register_enabled":   model.RegisterEnabled(),
		"credit_costs":       service.CreditCostsByRatio(),
		"greeting_text":      model.GetSettingValue("greeting_text", adminSettingDefaults["greeting_text"]),
		"guest_free_credits": guestFreeCredits(),
	})
}

func guestFreeCredits() int {
	value := strings.TrimSpace(model.GetSettingValue("guest_free_credits", adminSettingDefaults["guest_free_credits"]))
	credits, err := strconv.Atoi(value)
	if err != nil || credits < 0 {
		return 5
	}
	return credits
}
