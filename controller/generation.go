package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

type createGenerationRequest struct {
	Prompt      string `json:"prompt" binding:"required,max=4000"`
	Quality     string `json:"quality" binding:"required,oneof=low medium high"`
	Size        string `json:"size" binding:"required"`
	AnonymousID string `json:"anonymous_id"`
}

func CreateGeneration(c *gin.Context) {
	var req createGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var userID *int64
	if value, exists := c.Get("userID"); exists {
		if id, ok := value.(int64); ok {
			userID = &id
		}
	}
	if userID == nil {
		fingerprint := c.GetHeader("X-Fingerprint")
		if fingerprint == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "fingerprint required for free trial"})
			return
		}
		anonymousID, ok := service.CheckTrialEligible(common.GetRealIP(c), fingerprint)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "free trial used, please register"})
			return
		}
		req.Quality = "low"
		req.AnonymousID = anonymousID
		generation, err := service.CreateGeneration(req.Prompt, req.Quality, req.Size, common.GetRealIP(c), nil, req.AnonymousID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create generation"})
			return
		}
		service.MarkTrialUsed(anonymousID)
		c.JSON(http.StatusOK, gin.H{"id": generation.ID, "status": generation.Status, "anonymous_id": anonymousID})
		return
	}

	generation, err := service.CreateGeneration(req.Prompt, req.Quality, req.Size, common.GetRealIP(c), userID, req.AnonymousID)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientCredits) || errors.Is(err, service.ErrCreditsExpired) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient credits"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create generation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": generation.ID, "status": generation.Status})
}

func StreamGeneration(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid generation id"})
		return
	}

	var generation model.Generation
	if err := model.DB.First(&generation, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generation not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := service.Notifier.Subscribe(id)
	defer service.Notifier.Unsubscribe(id, ch)

	sendSSE(c, service.GenerationEvent{
		Status:   generation.Status,
		Message:  statusMessage(generation.Status),
		ImageURL: generation.ImageURL,
		Error:    generation.ErrorMsg,
	})

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-ch:
			sendSSE(c, event)
			if event.Status == 3 || event.Status == 4 || event.Status == 5 {
				return
			}
		case <-ticker.C:
			_, _ = fmt.Fprint(c.Writer, ":keepalive\n\n")
			c.Writer.Flush()
		}
	}
}

func sendSSE(c *gin.Context, event service.GenerationEvent) {
	c.SSEvent("status", event)
	c.Writer.Flush()
}

func statusMessage(status int) string {
	switch status {
	case 0:
		return "任务已创建"
	case 1:
		return "正在生成图片..."
	case 2:
		return "正在上传图片..."
	case 3:
		return "生成完成"
	case 4:
		return "生成失败，请重试"
	case 5:
		return "任务已取消"
	default:
		return "处理中"
	}
}
