package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

type announcementRequest struct {
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Status    int    `json:"status"`
	SortOrder int    `json:"sort_order"`
}

func ActiveAnnouncement(c *gin.Context) {
	var item model.Announcement
	err := model.DB.
		Where("status = ?", 1).
		Order("sort_order ASC, updated_at DESC, id DESC").
		First(&item).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"item": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
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
	return model.Announcement{
		Title:     title,
		Content:   content,
		Status:    status,
		SortOrder: req.SortOrder,
	}, true
}
