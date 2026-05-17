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
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 200
)

type createMessageRequest struct {
	Prompt     string `json:"prompt" binding:"required,max=4000"`
	Size       string `json:"size" binding:"required"`
	StyleID    string `json:"style_id"`
	SceneID    string `json:"scene_id"`
	Layered    bool   `json:"layered"`
	LayerCount int    `json:"layer_count"`
}

func ListMessages(c *gin.Context) {
	conversation, ok := loadOwnedConversation(c)
	if !ok {
		return
	}
	limit := parseConversationLimit(c.DefaultQuery("limit", ""))
	if limit == defaultConversationLimit {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	var messages []model.Message
	if err := model.DB.
		Where("conversation_id = ?", conversation.ID).
		Order("created_at ASC").
		Order("id ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list messages failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": messages})
}

func CreateMessage(c *gin.Context) {
	conversation, ok := loadOwnedConversation(c)
	if !ok {
		return
	}

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		createMultipartMessage(c, conversation)
		return
	}
	createJSONMessage(c, conversation)
}

func createJSONMessage(c *gin.Context, conversation model.Conversation) {
	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	realSize, ok := enabledImageSizeValue(req.Size)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image size"})
		return
	}
	req.Size = realSize

	userID := currentUserID(c)
	message := model.Message{
		ConversationID: conversation.ID,
		UserID:         userID,
		Prompt:         req.Prompt,
		TaskKind:       "text2img",
		Size:           req.Size,
		StyleID:        strings.TrimSpace(req.StyleID),
		SceneID:        strings.TrimSpace(req.SceneID),
		Layered:        req.Layered,
		LayerCount:     normalizeLayerCount(req.LayerCount),
	}
	if err := model.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create message failed"})
		return
	}

	generation, err := service.CreateGenerationForMessage(req.Prompt, standardImageQuality, req.Size, common.GetRealIP(c), &userID, "", &message.ID, fixedGenerationImageOptions())
	if err != nil {
		_ = model.DB.Delete(&message).Error
		handleMessageGenerationError(c, err)
		return
	}
	linkMessageGeneration(c, conversation, &message, generation)
}

func createMultipartMessage(c *gin.Context, conversation model.Conversation) {
	if err := c.Request.ParseMultipartForm(15 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart request"})
		return
	}
	req := createMessageRequest{
		Prompt:     strings.TrimSpace(c.PostForm("prompt")),
		Size:       c.PostForm("size"),
		StyleID:    c.PostForm("style_id"),
		SceneID:    c.PostForm("scene_id"),
		Layered:    c.PostForm("layered") == "true",
		LayerCount: parseIntForm(c.PostForm("layer_count")),
	}
	if req.Prompt == "" || len(req.Prompt) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	realSize, ok := enabledImageSizeValue(req.Size)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image size"})
		return
	}
	req.Size = realSize

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file must be smaller than 10MB"})
		return
	}
	imageData, err := io.ReadAll(io.LimitReader(file, 10<<20+1))
	if err != nil || len(imageData) == 0 || len(imageData) > 10<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image file"})
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(imageData)
	}
	if !isSupportedEditImageType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image file type"})
		return
	}

	userID := currentUserID(c)
	message := model.Message{
		ConversationID: conversation.ID,
		UserID:         userID,
		Prompt:         req.Prompt,
		AttachmentName: header.Filename,
		AttachmentSize: header.Size,
		TaskKind:       "img2img_generic",
		Size:           req.Size,
		StyleID:        strings.TrimSpace(req.StyleID),
		SceneID:        strings.TrimSpace(req.SceneID),
		Layered:        req.Layered,
		LayerCount:     normalizeLayerCount(req.LayerCount),
	}
	if err := model.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create message failed"})
		return
	}

	generation, err := service.CreateImageEditForMessage(req.Prompt, standardImageQuality, req.Size, common.GetRealIP(c), &userID, "", imageData, header.Filename, contentType, &message.ID, fixedGenerationImageOptions())
	if err != nil {
		_ = model.DB.Delete(&message).Error
		handleMessageGenerationError(c, err)
		return
	}
	message.AttachmentURL = generation.SourceImageURL
	message.AttachmentKey = generation.SourceR2Key
	linkMessageGeneration(c, conversation, &message, generation)
}

func linkMessageGeneration(c *gin.Context, conversation model.Conversation, message *model.Message, generation *model.Generation) {
	message.GenerationID = &generation.ID
	updates := map[string]interface{}{
		"generation_id": message.GenerationID,
		"updated_at":    time.Now(),
	}
	if message.AttachmentURL != "" {
		updates["attachment_url"] = message.AttachmentURL
	}
	if message.AttachmentKey != "" {
		updates["attachment_key"] = message.AttachmentKey
	}
	if err := model.DB.Model(message).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "link generation failed"})
		return
	}

	var messageCount int64
	_ = model.DB.Model(&model.Message{}).Where("conversation_id = ?", conversation.ID).Count(&messageCount).Error
	conversationUpdates := map[string]interface{}{
		"msg_count":   int(messageCount),
		"last_msg_at": message.CreatedAt,
		"is_layered":  message.Layered,
	}
	if conversation.MsgCount == 0 {
		conversationUpdates["title"] = normalizeConversationTitle(message.Prompt)
	}
	if err := model.DB.Model(&conversation).Updates(conversationUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update conversation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": message, "generation_id": generation.ID})
}

func handleMessageGenerationError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInsufficientCredits) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient_credits", "message": "insufficient credits"})
		return
	}
	if errors.Is(err, service.ErrCreditsExpired) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "credits_expired", "message": "credits expired"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create generation"})
}

func normalizeLayerCount(value int) int {
	if value < 0 {
		return 0
	}
	if value > 20 {
		return 20
	}
	return value
}

func parseIntForm(value string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	return n
}
