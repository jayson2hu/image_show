package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/controller"
	"github.com/jayson2hu/image-show/middleware"
)

func Register(r *gin.Engine) {
	r.Use(gin.Logger(), gin.Recovery(), middleware.RealIP())

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
	api.POST("/generations", middleware.OptionalAuth(), middleware.GenerationRateLimit(), controller.CreateGeneration)
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

	registerWebRoutes(r)
}
