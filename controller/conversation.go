package controller

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

const (
	defaultConversationLimit = 20
	maxConversationLimit     = 100
	defaultConversationTitle = "新会话"
)

type conversationTitleRequest struct {
	Title string `json:"title"`
}

func ListConversations(c *gin.Context) {
	userID := currentUserID(c)
	limit := parseConversationLimit(c.Query("limit"))

	query := model.DB.Where("user_id = ? AND is_deleted = ?", userID, false)
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		query = query.Where("title LIKE ?", "%"+q+"%")
	}
	if cursor := strings.TrimSpace(c.Query("cursor")); cursor != "" {
		cursorID, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil || cursorID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
		var cursorConversation model.Conversation
		if err := model.DB.
			Where("id = ? AND user_id = ? AND is_deleted = ?", cursorID, userID, false).
			First(&cursorConversation).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
		query = query.Where("last_msg_at < ? OR (last_msg_at = ? AND id < ?)", cursorConversation.LastMsgAt, cursorConversation.LastMsgAt, cursorConversation.ID)
	}

	var conversations []model.Conversation
	if err := query.
		Order("last_msg_at DESC").
		Order("id DESC").
		Limit(limit + 1).
		Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list conversations failed"})
		return
	}

	nextCursor := ""
	if len(conversations) > limit {
		nextCursor = strconv.FormatInt(conversations[limit-1].ID, 10)
		conversations = conversations[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"items": conversations, "next_cursor": nextCursor})
}

func CreateConversation(c *gin.Context) {
	userID := currentUserID(c)
	var req conversationTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	now := time.Now()
	conversation := model.Conversation{
		UserID:    userID,
		Title:     normalizeConversationTitle(req.Title),
		LastMsgAt: now,
	}
	if err := model.DB.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create conversation failed"})
		return
	}
	c.JSON(http.StatusCreated, conversation)
}

func GetConversation(c *gin.Context) {
	conversation, ok := loadOwnedConversation(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, conversation)
}

func RenameConversation(c *gin.Context) {
	conversation, ok := loadOwnedConversation(c)
	if !ok {
		return
	}
	var req conversationTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	title := normalizeConversationTitle(req.Title)
	if title == defaultConversationTitle {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if err := model.DB.Model(&conversation).Update("title", title).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rename conversation failed"})
		return
	}
	conversation.Title = title
	c.JSON(http.StatusOK, conversation)
}

func DeleteConversation(c *gin.Context) {
	conversation, ok := loadOwnedConversation(c)
	if !ok {
		return
	}
	if err := model.DB.Model(&conversation).Update("is_deleted", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete conversation failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func loadOwnedConversation(c *gin.Context) (model.Conversation, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return model.Conversation{}, false
	}
	var conversation model.Conversation
	err = model.DB.Where("id = ? AND user_id = ? AND is_deleted = ?", id, currentUserID(c), false).First(&conversation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return model.Conversation{}, false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load conversation failed"})
		return model.Conversation{}, false
	}
	return conversation, true
}

func currentUserID(c *gin.Context) int64 {
	value, _ := c.Get("userID")
	userID, _ := value.(int64)
	return userID
}

func normalizeConversationTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return defaultConversationTitle
	}
	if len([]rune(title)) > 128 {
		return string([]rune(title)[:128])
	}
	return title
}

func parseConversationLimit(raw string) int {
	limit := defaultConversationLimit
	if raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxConversationLimit {
		return maxConversationLimit
	}
	return limit
}
