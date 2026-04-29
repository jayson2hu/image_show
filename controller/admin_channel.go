package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

type channelRequest struct {
	Name    string `json:"name" binding:"required"`
	BaseURL string `json:"base_url" binding:"required"`
	APIKey  string `json:"api_key"`
	Headers string `json:"headers"`
	Status  int    `json:"status"`
	Weight  int    `json:"weight"`
	Remark  string `json:"remark"`
}

func AdminChannels(c *gin.Context) {
	var channels []model.Channel
	if err := model.DB.Order("id DESC").Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list channels"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": channels})
}

func AdminCreateChannel(c *gin.Context) {
	var req channelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	channel := channelFromRequest(req)
	if err := model.DB.Create(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
		return
	}
	c.JSON(http.StatusOK, channel)
}

func AdminUpdateChannel(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req channelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := channelFromRequest(req)
	if err := model.DB.Model(&model.Channel{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update channel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminDeleteChannel(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := model.DB.Delete(&model.Channel{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete channel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminTestChannel(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var channel model.Channel
	if err := model.DB.First(&channel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(channel.BaseURL, "/")+"/v1/models", nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel base_url"})
		return
	}
	if channel.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+channel.APIKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	c.JSON(http.StatusOK, gin.H{"ok": resp.StatusCode >= 200 && resp.StatusCode < 300, "status": resp.StatusCode})
}

func channelFromRequest(req channelRequest) model.Channel {
	status := req.Status
	if status == 0 {
		status = 1
	}
	weight := req.Weight
	if weight <= 0 {
		weight = 1
	}
	return model.Channel{
		Name:    req.Name,
		BaseURL: strings.TrimRight(req.BaseURL, "/"),
		APIKey:  req.APIKey,
		Headers: req.Headers,
		Status:  status,
		Weight:  weight,
		Remark:  req.Remark,
	}
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}
