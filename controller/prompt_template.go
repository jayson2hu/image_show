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
		{Category: "style", Label: "写实", Prompt: "写实摄影风格，细节丰富，自然光影，真实材质，高质量商业摄影", SortOrder: 1, Status: 1},
		{Category: "style", Label: "动漫", Prompt: "动漫插画风格，清晰线稿，高饱和色彩，精致角色设计，干净背景", SortOrder: 2, Status: 1},
		{Category: "style", Label: "幻想", Prompt: "幻想艺术风格，史诗氛围，电影级构图，丰富层次，强烈空间感", SortOrder: 3, Status: 1},
		{Category: "style", Label: "赛博朋克", Prompt: "赛博朋克风格，霓虹灯光，未来城市质感，高对比光影，雨夜氛围", SortOrder: 4, Status: 1},
		{Category: "style", Label: "水彩", Prompt: "水彩画风格，柔和笔触，温暖色调，纸张纹理，轻盈通透", SortOrder: 5, Status: 1},
		{Category: "style", Label: "抽象", Prompt: "抽象艺术风格，流动光影，紫蓝渐变，几何节奏，现代视觉表达", SortOrder: 6, Status: 1},
		{Category: "sample", Label: "幻想风景", Prompt: "沙漠中的神秘传送门，远处有漂浮的古代遗迹，超现实主义场景，金色夕阳，电影级构图，4K 高清细节", SortOrder: 20, Status: 1},
		{Category: "sample", Label: "赛博朋克城市", Prompt: "未来城市夜景，湿润街道反射霓虹灯，密集高楼与飞行交通，赛博朋克风格，强烈蓝紫色光影", SortOrder: 21, Status: 1},
		{Category: "sample", Label: "水彩小屋", Prompt: "森林中的小木屋，清晨薄雾，温暖阳光穿过树叶，柔和水彩画风格，安静治愈氛围", SortOrder: 22, Status: 1},
		{Category: "sample", Label: "抽象艺术", Prompt: "流动的光影和透明几何结构，紫蓝渐变，细腻颗粒质感，现代抽象艺术海报", SortOrder: 23, Status: 1},
	}
}
