package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/service"
)

func AdminMonitorSummary(c *gin.Context) {
	summary, err := service.GetMonitorSummary(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load monitor summary"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func AdminCheckMonitorAlert(c *gin.Context) {
	result, err := service.CheckMonitorAlert(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check monitor alert"})
		return
	}
	c.JSON(http.StatusOK, result)
}
