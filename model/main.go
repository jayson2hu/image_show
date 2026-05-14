package model

import (
	"fmt"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/jayson2hu/image-show/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

func InitDB() error {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	db, err := openDB(cfg.DBDriver, cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(
		&User{},
		&Generation{},
		&CreditLog{},
		&LoginLog{},
		&Channel{},
		&Setting{},
		&PromptTemplate{},
		&Announcement{},
		&AnnouncementRead{},
		&Package{},
		&Order{},
		&AnonymousIdentity{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	if err := seedDefaultPackages(db); err != nil {
		return err
	}
	if err := seedDefaultAdmin(db, cfg); err != nil {
		return err
	}
	if err := seedDefaultPromptTemplates(db); err != nil {
		return err
	}

	DB = db
	return nil
}

func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func seedDefaultPackages(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Package{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaults := []Package{
		{Name: "Starter Pack", Credits: 10, Price: 9.9, ValidDays: 30, SortOrder: 1, Status: 1},
		{Name: "Standard Pack", Credits: 50, Price: 39.9, ValidDays: 90, SortOrder: 2, Status: 1},
		{Name: "Pro Pack", Credits: 100, Price: 79.9, ValidDays: 180, SortOrder: 3, Status: 1},
	}
	return db.Create(&defaults).Error
}

func seedDefaultAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	admin := User{
		Email:        cfg.AdminEmail,
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         10,
		Status:       1,
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "email"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"username":      admin.Username,
			"password_hash": admin.PasswordHash,
			"role":          admin.Role,
			"status":        admin.Status,
			"updated_at":    time.Now(),
		}),
	}).Create(&admin).Error
}

func seedDefaultPromptTemplates(db *gorm.DB) error {
	defaults := []PromptTemplate{
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
	for _, item := range defaults {
		var count int64
		if err := db.Model(&PromptTemplate{}).
			Where("category = ? AND label = ?", item.Category, item.Label).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func openDB(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "", "sqlite":
		if err := os.MkdirAll("data", 0755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
		return gorm.Open(sqlite.Open("./data/image_show.db"), &gorm.Config{})
	case "postgres":
		if dsn == "" {
			return nil, fmt.Errorf("DATABASE_DSN is required when DB_DRIVER=postgres")
		}
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q", driver)
	}
}
