package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

func SiteConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"site_title":      model.GetSettingValue("site_title", adminSettingDefaults["site_title"]),
		"site_about":      model.GetSettingValue("site_about", adminSettingDefaults["site_about"]),
		"seo_title":       model.GetSettingValue("seo_title", adminSettingDefaults["seo_title"]),
		"seo_keywords":    model.GetSettingValue("seo_keywords", adminSettingDefaults["seo_keywords"]),
		"seo_description": model.GetSettingValue("seo_description", adminSettingDefaults["seo_description"]),
	})
}
