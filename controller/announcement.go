package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

type announcementRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Status     int    `json:"status"`
	NotifyMode string `json:"notify_mode"`
	SortOrder  int    `json:"sort_order"`
	StartsAt   string `json:"starts_at"`
	EndsAt     string `json:"ends_at"`
}

type userAnnouncementResponse struct {
	model.Announcement
	ReadAt *time.Time `json:"read_at"`
}

func ActiveAnnouncement(c *gin.Context) {
	var item model.Announcement
	err := model.DB.
		Scopes(activeAnnouncementScope(time.Now())).
		Order("sort_order ASC, updated_at DESC, id DESC").
		First(&item).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"item": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func UserAnnouncements(c *gin.Context) {
	userID := c.GetInt64("userID")
	var items []model.Announcement
	if err := model.DB.
		Scopes(activeAnnouncementScope(time.Now())).
		Order("sort_order ASC, created_at DESC, id DESC").
		Limit(20).
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list announcements"})
		return
	}
	readMap := map[int64]time.Time{}
	if userID > 0 && len(items) > 0 {
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		var reads []model.AnnouncementRead
		if err := model.DB.Where("user_id = ? AND announcement_id IN ?", userID, ids).Find(&reads).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list announcement reads"})
			return
		}
		for _, read := range reads {
			readMap[read.AnnouncementID] = read.ReadAt
		}
	}
	response := make([]userAnnouncementResponse, 0, len(items))
	for _, item := range items {
		var readAt *time.Time
		if value, ok := readMap[item.ID]; ok {
			v := value
			readAt = &v
		}
		response = append(response, userAnnouncementResponse{Announcement: item, ReadAt: readAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": response})
}

func MarkAnnouncementRead(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var count int64
	if err := model.DB.Model(&model.Announcement{}).Where("id = ?", id).Count(&count).Error; err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "announcement not found"})
		return
	}
	read := model.AnnouncementRead{AnnouncementID: id, UserID: userID, ReadAt: time.Now()}
	if err := model.DB.Where("announcement_id = ? AND user_id = ?", id, userID).Assign(read).FirstOrCreate(&read).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark announcement read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminAnnouncements(c *gin.Context) {
	var items []model.Announcement
	if err := model.DB.Order("sort_order ASC, id DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list announcements"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func AdminCreateAnnouncement(c *gin.Context) {
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	item, ok := announcementFromRequest(c, req)
	if !ok {
		return
	}
	if err := model.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create announcement"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func AdminUpdateAnnouncement(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	item, ok := announcementFromRequest(c, req)
	if !ok {
		return
	}
	if err := model.DB.Model(&model.Announcement{}).Where("id = ?", id).Updates(item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update announcement"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminDeleteAnnouncement(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := model.DB.Delete(&model.Announcement{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete announcement"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func announcementFromRequest(c *gin.Context, req announcementRequest) (model.Announcement, bool) {
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required"})
		return model.Announcement{}, false
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	if status != 1 && status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return model.Announcement{}, false
	}
	notifyMode := strings.TrimSpace(req.NotifyMode)
	if notifyMode == "" {
		notifyMode = "silent"
	}
	if notifyMode != "silent" && notifyMode != "popup" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notify mode"})
		return model.Announcement{}, false
	}
	startsAt, ok := parseOptionalAnnouncementTime(c, req.StartsAt, "starts_at")
	if !ok {
		return model.Announcement{}, false
	}
	endsAt, ok := parseOptionalAnnouncementTime(c, req.EndsAt, "ends_at")
	if !ok {
		return model.Announcement{}, false
	}
	if startsAt != nil && endsAt != nil && !endsAt.After(*startsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
		return model.Announcement{}, false
	}
	return model.Announcement{
		Title:      title,
		Content:    content,
		Status:     status,
		NotifyMode: notifyMode,
		SortOrder:  req.SortOrder,
		StartsAt:   startsAt,
		EndsAt:     endsAt,
	}, true
}

func parseOptionalAnnouncementTime(c *gin.Context, value string, field string) (*time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + field})
		return nil, false
	}
	return &parsed, true
}

func activeAnnouncementScope(now time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("status = ?", 1).
			Where("starts_at IS NULL OR starts_at <= ?", now).
			Where("ends_at IS NULL OR ends_at > ?", now)
	}
}
