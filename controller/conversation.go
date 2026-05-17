package controller

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
	"gorm.io/gorm"
)

const (
	defaultConversationLimit = 20
	maxConversationLimit     = 100
	defaultConversationTitle = "新对话"
)

type conversationTitleRequest struct {
	Title string `json:"title"`
}

type claimGuestConversationRequest struct {
	Title    string                     `json:"title"`
	Messages []claimGuestMessageRequest `json:"messages" binding:"required"`
}

type claimGuestMessageRequest struct {
	GenerationID int64  `json:"generation_id" binding:"required"`
	Prompt       string `json:"prompt"`
	TaskKind     string `json:"task_kind"`
	Size         string `json:"size"`
	StyleID      string `json:"style_id"`
	SceneID      string `json:"scene_id"`
	Layered      bool   `json:"layered"`
	LayerCount   int    `json:"layer_count"`
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

func ClaimGuestConversation(c *gin.Context) {
	userID := currentUserID(c)
	fingerprint := c.GetHeader("X-Fingerprint")
	if strings.TrimSpace(fingerprint) == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "fingerprint required"})
		return
	}

	var req claimGuestConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	anonymousID := service.TrialAnonymousID(common.GetRealIP(c), fingerprint)
	messageByGenerationID := make(map[int64]claimGuestMessageRequest, len(req.Messages))
	generationIDs := make([]int64, 0, len(req.Messages))
	for _, item := range req.Messages {
		if item.GenerationID <= 0 {
			continue
		}
		if _, exists := messageByGenerationID[item.GenerationID]; exists {
			continue
		}
		messageByGenerationID[item.GenerationID] = item
		generationIDs = append(generationIDs, item.GenerationID)
	}
	if len(generationIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var responseConversation model.Conversation
	responseMessages := []model.Message{}
	claimed := 0
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var generations []model.Generation
		if err := tx.
			Where("id IN ? AND anonymous_id = ? AND user_id IS NULL AND message_id IS NULL AND is_deleted = ?", generationIDs, anonymousID, false).
			Order("created_at ASC, id ASC").
			Find(&generations).Error; err != nil {
			return err
		}
		if len(generations) == 0 {
			return nil
		}

		now := time.Now()
		conversation := model.Conversation{
			UserID:    userID,
			Title:     normalizeConversationTitle(req.Title),
			LastMsgAt: now,
		}
		if conversation.Title == defaultConversationTitle {
			conversation.Title = normalizeConversationTitle(firstClaimPrompt(generations, messageByGenerationID))
		}
		if err := tx.Create(&conversation).Error; err != nil {
			return err
		}

		messages := make([]model.Message, 0, len(generations))
		isLayered := false
		for _, generation := range generations {
			source := messageByGenerationID[generation.ID]
			message := model.Message{
				ConversationID: conversation.ID,
				UserID:         userID,
				AnonymousID:    anonymousID,
				Prompt:         normalizeClaimPrompt(source.Prompt, generation.Prompt),
				TaskKind:       normalizeClaimTaskKind(source.TaskKind, generation.Mode),
				Size:           normalizeClaimValue(source.Size, generation.Size),
				StyleID:        strings.TrimSpace(source.StyleID),
				SceneID:        strings.TrimSpace(source.SceneID),
				Layered:        source.Layered,
				LayerCount:     normalizeLayerCount(source.LayerCount),
				GenerationID:   &generation.ID,
				CreatedAt:      generation.CreatedAt,
				UpdatedAt:      now,
			}
			if message.Layered {
				isLayered = true
			}
			if err := tx.Create(&message).Error; err != nil {
				return err
			}
			updateGeneration := tx.Model(&model.Generation{}).
				Where("id = ? AND anonymous_id = ? AND user_id IS NULL AND message_id IS NULL", generation.ID, anonymousID).
				Updates(map[string]interface{}{"user_id": userID, "message_id": message.ID})
			if updateGeneration.Error != nil {
				return updateGeneration.Error
			}
			if updateGeneration.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			messages = append(messages, message)
		}

		updates := map[string]interface{}{
			"msg_count":   len(messages),
			"last_msg_at": messages[len(messages)-1].CreatedAt,
			"is_layered":  isLayered,
		}
		if err := tx.Model(&conversation).Updates(updates).Error; err != nil {
			return err
		}
		conversation.MsgCount = len(messages)
		conversation.LastMsgAt = messages[len(messages)-1].CreatedAt
		conversation.IsLayered = isLayered
		responseConversation = conversation
		responseMessages = messages
		claimed = len(messages)
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "claim guest conversation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversation": responseConversation, "messages": responseMessages, "claimed": claimed})
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

func conversationTitleFromPrompt(prompt string) string {
	runes := []rune(strings.TrimSpace(prompt))
	if len(runes) == 0 {
		return defaultConversationTitle
	}
	if len(runes) > 12 {
		return string(runes[:12]) + "..."
	}
	return string(runes)
}

func firstClaimPrompt(generations []model.Generation, messages map[int64]claimGuestMessageRequest) string {
	for _, generation := range generations {
		if source, ok := messages[generation.ID]; ok && strings.TrimSpace(source.Prompt) != "" {
			return conversationTitleFromPrompt(source.Prompt)
		}
		if strings.TrimSpace(generation.Prompt) != "" {
			return conversationTitleFromPrompt(generation.Prompt)
		}
	}
	return defaultConversationTitle
}

func normalizeClaimPrompt(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func normalizeClaimValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func normalizeClaimTaskKind(value, generationMode string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	if generationMode == service.GenerationModeEdit {
		return "img2img_generic"
	}
	return "text2img"
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
