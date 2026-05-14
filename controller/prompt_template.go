package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

type sceneTemplateResponse struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Icon             string  `json:"icon"`
	Description      string  `json:"description"`
	PromptTemplate   string  `json:"prompt_template"`
	RecommendedRatio string  `json:"recommended_ratio"`
	CreditCost       float64 `json:"credit_cost"`
}

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

func GenerationScenes(c *gin.Context) {
	var templates []model.PromptTemplate
	if model.DB != nil {
		_ = model.DB.Where("status = ? AND category = ?", 1, "scene").Order("sort_order ASC, id ASC").Find(&templates).Error
	}
	if len(templates) == 0 {
		for _, item := range defaultPromptTemplates() {
			if item.Category == "scene" {
				templates = append(templates, item)
			}
		}
	}
	items := make([]sceneTemplateResponse, 0, len(templates))
	for _, template := range templates {
		ratio := template.RecommendedRatio
		if ratio == "" {
			ratio = "square"
		}
		items = append(items, sceneTemplateResponse{
			ID:               template.ID,
			Name:             template.Label,
			Icon:             template.Icon,
			Description:      template.Description,
			PromptTemplate:   template.Prompt,
			RecommendedRatio: ratio,
			CreditCost:       service.CostForSize(ratio),
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func defaultPromptTemplates() []model.PromptTemplate {
	return []model.PromptTemplate{
		{Category: "style", Label: "写实", Prompt: "写实摄影风格，细节丰富，自然光影，真实材质，高质量商业摄影", SortOrder: 1, Status: 1},
		{Category: "style", Label: "动漫", Prompt: "动漫插画风格，清晰线稿，高饱和色彩，精致角色设计，干净背景", SortOrder: 2, Status: 1},
		{Category: "style", Label: "幻想", Prompt: "幻想艺术风格，史诗氛围，电影级构图，丰富层次，强烈空间感", SortOrder: 3, Status: 1},
		{Category: "style", Label: "赛博朋克", Prompt: "赛博朋克风格，霓虹灯光，未来城市质感，高对比光影，雨夜氛围", SortOrder: 4, Status: 1},
		{Category: "style", Label: "水彩", Prompt: "水彩画风格，柔和笔触，温暖色调，纸张纹理，轻盈通透", SortOrder: 5, Status: 1},
		{Category: "style", Label: "抽象", Prompt: "抽象艺术风格，流动光影，紫蓝渐变，几何节奏，现代视觉表达", SortOrder: 6, Status: 1},
		{Category: "style", Label: "插画", Prompt: "现代商业插画风格，清晰轮廓，柔和配色，细腻纹理，画面干净有层次，适合封面、海报和内容配图", SortOrder: 7, Status: 1},
		{Category: "sample", Label: "幻想风景", Prompt: "沙漠中的神秘传送门，远处有漂浮的古代遗迹，超现实主义场景，金色夕阳，电影级构图，4K 高清细节", SortOrder: 20, Status: 1},
		{Category: "sample", Label: "赛博朋克城市", Prompt: "未来城市夜景，湿润街道反射霓虹灯，密集高楼与飞行交通，赛博朋克风格，强烈蓝紫色光影", SortOrder: 21, Status: 1},
		{Category: "sample", Label: "水彩小屋", Prompt: "森林中的小木屋，清晨薄雾，温暖阳光穿过树叶，柔和水彩画风格，安静治愈氛围", SortOrder: 22, Status: 1},
		{Category: "sample", Label: "抽象艺术", Prompt: "流动的光影和透明几何结构，紫蓝渐变，细腻颗粒质感，现代抽象艺术海报", SortOrder: 23, Status: 1},
		{Category: "scene", Label: "小红书封面", Icon: "📸", Description: "精致生活、穿搭、美食风格封面", Prompt: "小红书封面图，精致生活方式视觉，一眼能看懂主题，清晰大标题留白，明亮干净的构图，适合手机竖屏浏览", RecommendedRatio: "portrait_3_4", SortOrder: 40, Status: 1},
		{Category: "scene", Label: "商品展示图", Icon: "🛒", Description: "白底或场景化商品展示", Prompt: "电商商品展示图，主体突出，干净背景，真实材质，高级商业摄影光影，适合商品主图", RecommendedRatio: "square", SortOrder: 41, Status: 1},
		{Category: "scene", Label: "社交头像", Icon: "👤", Description: "精致人物或动漫风格头像", Prompt: "精致社交头像，主体居中，五官清晰，背景简洁，有辨识度，适合作为社交平台头像", RecommendedRatio: "square", SortOrder: 42, Status: 1},
		{Category: "scene", Label: "海报设计", Icon: "🎨", Description: "活动、促销、艺术创意海报", Prompt: "活动宣传海报视觉，主题突出，层次清晰，保留文字排版空间，适合促销活动和创意传播", RecommendedRatio: "portrait_3_4", SortOrder: 43, Status: 1},
		{Category: "scene", Label: "手机壁纸", Icon: "📷", Description: "风景、抽象、治愈系壁纸", Prompt: "高清手机壁纸画面，风景治愈氛围，视觉舒适，构图开阔，细节丰富，适合手机屏幕背景", RecommendedRatio: "story", SortOrder: 44, Status: 1},
		{Category: "scene", Label: "自由创作", Icon: "✨", Description: "不填充提示词，自由输入", Prompt: "", RecommendedRatio: "square", SortOrder: 45, Status: 1},
	}
}
