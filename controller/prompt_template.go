package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

func PromptTemplates(c *gin.Context) {
	var templates []model.PromptTemplate
	if model.DB != nil {
		_ = model.DB.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&templates).Error
	}
	if len(templates) == 0 {
		templates = defaultPromptTemplates()
	}
	c.JSON(http.StatusOK, gin.H{"items": templates})
}

func defaultPromptTemplates() []model.PromptTemplate {
	return []model.PromptTemplate{
		{Category: "default", Label: "电影感", Prompt: "电影感光影，高质量细节", SortOrder: 1, Status: 1},
		{Category: "default", Label: "产品摄影", Prompt: "干净背景，商业产品摄影风格", SortOrder: 2, Status: 1},
		{Category: "repair", Label: "修复模糊", Prompt: "修复模糊，提升主体清晰度", SortOrder: 3, Status: 1},
		{Category: "repair", Label: "提升清晰度", Prompt: "提升分辨率，保留自然纹理", SortOrder: 4, Status: 1},
		{Category: "style", Label: "增强色彩", Prompt: "增强色彩层次，保持真实自然", SortOrder: 5, Status: 1},
	}
}
