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

type adminChannelResponse struct {
	model.Channel
	RecentSuccessCount int64   `json:"recent_success_count"`
	RecentFailedCount  int64   `json:"recent_failed_count"`
	RecentFailureRate  float64 `json:"recent_failure_rate"`
}

func AdminChannels(c *gin.Context) {
	var channels []model.Channel
	if err := model.DB.Order("id DESC").Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list channels"})
		return
	}
	stats, err := recentChannelGenerationStats(time.Now().Add(-24 * time.Hour))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel stats"})
		return
	}
	items := make([]adminChannelResponse, 0, len(channels))
	for _, channel := range channels {
		stat := stats[channel.ID]
		total := stat.success + stat.failed
		failureRate := 0.0
		if total > 0 {
			failureRate = float64(stat.failed) / float64(total)
		}
		items = append(items, adminChannelResponse{
			Channel:            channel,
			RecentSuccessCount: stat.success,
			RecentFailedCount:  stat.failed,
			RecentFailureRate:  failureRate,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
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
	testedAt := time.Now()
	errorSummary := ""
	if err != nil {
		errorSummary = sanitizeChannelTestError(err.Error())
		_ = model.DB.Model(&model.Channel{}).Where("id = ?", channel.ID).Updates(map[string]interface{}{
			"last_test_at":      testedAt,
			"last_test_success": false,
			"last_test_status":  0,
			"last_test_error":   errorSummary,
		}).Error
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	testOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !testOK {
		errorSummary = "upstream status " + strconv.Itoa(resp.StatusCode)
	}
	if err := model.DB.Model(&model.Channel{}).Where("id = ?", channel.ID).Updates(map[string]interface{}{
		"last_test_at":      testedAt,
		"last_test_success": testOK,
		"last_test_status":  resp.StatusCode,
		"last_test_error":   errorSummary,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record channel test"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": testOK, "status": resp.StatusCode})
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

func sanitizeChannelTestError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 240 {
		return value[:240]
	}
	return value
}

type channelGenerationStat struct {
	success int64
	failed  int64
}

func recentChannelGenerationStats(since time.Time) (map[int64]channelGenerationStat, error) {
	type row struct {
		ChannelID int64
		Status    int
		Count     int64
	}
	var rows []row
	err := model.DB.Model(&model.Generation{}).
		Select("channel_id, status, COUNT(*) as count").
		Where("channel_id IS NOT NULL AND status IN ? AND created_at >= ?", []int{3, 4}, since).
		Group("channel_id, status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	stats := make(map[int64]channelGenerationStat)
	for _, item := range rows {
		stat := stats[item.ChannelID]
		if item.Status == 3 {
			stat.success = item.Count
		}
		if item.Status == 4 {
			stat.failed = item.Count
		}
		stats[item.ChannelID] = stat
	}
	return stats, nil
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}
