package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/controller"
	"github.com/jayson2hu/image-show/middleware"
)

func Register(r *gin.Engine) {
	r.Use(gin.Logger(), gin.Recovery(), middleware.SecurityHeaders(), middleware.RealIP(), middleware.IPBlacklist())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	auth := api.Group("/auth")
	auth.POST("/send-code", controller.SendCode)
	auth.POST("/register", controller.Register)
	auth.POST("/login", controller.Login)
	auth.GET("/me", middleware.AuthRequired(), controller.Me)
	auth.GET("/wechat/qrcode", controller.WeChatQRCode)
	auth.GET("/wechat/callback", controller.WeChatCallback)
	auth.GET("/wechat/status", controller.WeChatStatus)
	auth.POST("/wechat/bind", middleware.AuthRequired(), controller.WeChatBind)
	auth.DELETE("/wechat/bind", middleware.AuthRequired(), controller.WeChatUnbind)
	api.GET("/generations", middleware.AuthRequired(), controller.ListGenerations)
	api.POST("/generations", middleware.OptionalAuth(), middleware.GenerationRateLimit(), controller.CreateGeneration)
	api.GET("/generations/:id", middleware.AuthRequired(), controller.GenerationDetail)
	api.DELETE("/generations/:id", middleware.AuthRequired(), controller.DeleteGeneration)
	api.POST("/generations/:id/cancel", middleware.AuthRequired(), controller.CancelGeneration)
	api.GET("/generations/:id/stream", controller.StreamGeneration)
	api.GET("/prompt-templates", controller.PromptTemplates)
	credits := api.Group("/credits", middleware.AuthRequired())
	credits.GET("/balance", controller.CreditBalance)
	credits.GET("/logs", controller.CreditLogs)
	admin := api.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())
	admin.GET("/logs/generations", controller.AdminGenerationLogs)
	admin.GET("/logs/logins", controller.AdminLoginLogs)
	admin.DELETE("/logs/generations", controller.AdminDeleteGenerationLogs)
	admin.DELETE("/logs/logins", controller.AdminDeleteLoginLogs)
	admin.GET("/channels", controller.AdminChannels)
	admin.POST("/channels", controller.AdminCreateChannel)
	admin.PUT("/channels/:id", controller.AdminUpdateChannel)
	admin.DELETE("/channels/:id", controller.AdminDeleteChannel)
	admin.POST("/channels/:id/test", controller.AdminTestChannel)
	admin.GET("/users", controller.AdminUsers)
	admin.PUT("/users/:id/status", controller.AdminUpdateUserStatus)
	admin.PUT("/users/:id/role", controller.AdminUpdateUserRole)
	admin.GET("/users/:id/generations", controller.AdminUserGenerations)
	admin.POST("/users/:id/credits", controller.AdminTopupCredits)
	admin.GET("/credits/logs", controller.AdminCreditLogs)
	admin.GET("/prompt-templates", controller.AdminPromptTemplates)
	admin.POST("/prompt-templates", controller.AdminCreatePromptTemplate)
	admin.PUT("/prompt-templates/:id", controller.AdminUpdatePromptTemplate)
	admin.DELETE("/prompt-templates/:id", controller.AdminDeletePromptTemplate)
	admin.GET("/settings", controller.AdminSettings)
	admin.PUT("/settings", controller.AdminUpdateSettings)
	admin.GET("/generations", controller.AdminGenerations)
	admin.DELETE("/generations/batch", controller.AdminBatchDeleteGenerations)

	registerWebRoutes(r)
}
