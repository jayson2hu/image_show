package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/router"
	"github.com/jayson2hu/image-show/service"
)

func main() {
	common.LoadEnv()

	cfg := config.LoadConfig()
	if err := model.InitDB(); err != nil {
		log.Fatalf("init database: %v", err)
	}
	defer model.CloseDB()

	if cfg.AppEnv == common.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	stopOrderExpiry := make(chan struct{})
	service.StartOrderExpiryLoop(stopOrderExpiry)
	defer close(stopOrderExpiry)

	engine := gin.New()
	router.Register(engine)

	addr := fmt.Sprintf(":%d", cfg.Port)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("start server %s: %v", addr, err)
	}
}
