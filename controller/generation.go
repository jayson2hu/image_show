package controller

import (
	"encoding/json"
	"errors"
	"fmt"
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

type createGenerationRequest struct {
	Prompt            string `json:"prompt" binding:"required,max=4000"`
	Quality           string `json:"quality" binding:"omitempty,oneof=low medium high"`
	Size              string `json:"size" binding:"required"`
	AnonymousID       string `json:"anonymous_id"`
	CaptchaToken      string `json:"captcha_token"`
	OutputFormat      string `json:"output_format"`
	Background        string `json:"background"`
	OutputCompression *int   `json:"output_compression"`
}

const standardImageQuality = "medium"

type imageSizeOption struct {
	Value      string  `json:"value"`
	Label      string  `json:"label"`
	Ratio      string  `json:"ratio"`
	CreditCost float64 `json:"credit_cost"`
}

type batchDeleteGenerationsRequest struct {
	IDs      []int64 `json:"ids" binding:"required"`
	DeleteR2 bool    `json:"delete_r2"`
}

func GenerationOptions(c *gin.Context) {
	sizes := enabledImageSizes()
	c.JSON(http.StatusOK, gin.H{"sizes": sizes, "size_options": buildImageSizeOptions(sizes)})
}

func ListGenerations(c *gin.Context) {
	page, pageSize := pagination(c)
	userID := c.GetInt64("userID")
	query := model.DB.Model(&model.Generation{}).Where("user_id = ? AND is_deleted = ?", userID, false)

	var total int64
	var items []model.Generation
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count generations"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list generations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func GenerationDetail(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	userID := c.GetInt64("userID")
	var generation model.Generation
	if err := model.DB.Where("id = ? AND user_id = ? AND is_deleted = ?", id, userID, false).First(&generation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generation not found"})
		return
	}
	url, err := service.RefreshImageURL(&generation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh image url"})
		return
	}
	generation.ImageURL = url
	c.JSON(http.StatusOK, gin.H{"item": generation})
}

func DeleteGeneration(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	userID := c.GetInt64("userID")
	result := model.DB.Model(&model.Generation{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_deleted", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete generation"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "generation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func CancelGeneration(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	userID := c.GetInt64("userID")
	refunded, err := service.CancelGeneration(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel generation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "refunded": refunded})
}

func CreateGeneration(c *gin.Context) {
	var req createGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := service.VerifyCaptcha(req.CaptchaToken, common.GetRealIP(c)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	req.Quality = standardImageQuality
	if !isEnabledImageSize(req.Size) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image size"})
		return
	}
	if !validateImageOutputOptions(c, req.OutputFormat, req.Background, req.OutputCompression) {
		return
	}
	options := service.ImageOptions{OutputFormat: req.OutputFormat, Background: req.Background, OutputCompression: req.OutputCompression}

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
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":   "free_trial_exhausted",
				"message": "免费试用次数已用完，注册登录后可获得更多积分",
			})
			return
		}
		req.AnonymousID = anonymousID
		generation, err := service.CreateGeneration(req.Prompt, req.Quality, req.Size, common.GetRealIP(c), nil, req.AnonymousID, options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create generation"})
			return
		}
		service.MarkTrialUsed(anonymousID)
		c.JSON(http.StatusOK, gin.H{"id": generation.ID, "status": generation.Status, "anonymous_id": anonymousID})
		return
	}

	generation, err := service.CreateGeneration(req.Prompt, req.Quality, req.Size, common.GetRealIP(c), userID, req.AnonymousID, options)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientCredits) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient_credits", "message": "积分不足，请充值后继续"})
			return
		}
		if errors.Is(err, service.ErrCreditsExpired) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "credits_expired", "message": "积分已过期，请重新购买"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create generation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": generation.ID, "status": generation.Status})
}

func CreateImageEdit(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(55 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart request"})
		return
	}
	req := createGenerationRequest{
		Prompt:       strings.TrimSpace(c.PostForm("prompt")),
		Quality:      standardImageQuality,
		Size:         c.PostForm("size"),
		AnonymousID:  c.PostForm("anonymous_id"),
		CaptchaToken: c.PostForm("captcha_token"),
		OutputFormat: c.PostForm("output_format"),
		Background:   c.PostForm("background"),
	}
	if compression := strings.TrimSpace(c.PostForm("output_compression")); compression != "" {
		value, err := strconv.Atoi(compression)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "output_compression must be between 0 and 100"})
			return
		}
		req.OutputCompression = &value
	}
	if req.Prompt == "" || len(req.Prompt) > 4000 || !isValidQuality(req.Quality) || req.Size == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := service.VerifyCaptcha(req.CaptchaToken, common.GetRealIP(c)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if !isEnabledImageSize(req.Size) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image size"})
		return
	}
	if !validateImageOutputOptions(c, req.OutputFormat, req.Background, req.OutputCompression) {
		return
	}
	options := service.ImageOptions{OutputFormat: req.OutputFormat, Background: req.Background, OutputCompression: req.OutputCompression}
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > 50<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file must be smaller than 50MB"})
		return
	}
	imageData, err := io.ReadAll(io.LimitReader(file, 50<<20+1))
	if err != nil || len(imageData) == 0 || len(imageData) > 50<<20 {
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

	var userID *int64
	if value, exists := c.Get("userID"); exists {
		if id, ok := value.(int64); ok {
			userID = &id
		}
	}
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "please login to edit images"})
		return
	}

	generation, err := service.CreateImageEdit(req.Prompt, req.Quality, req.Size, common.GetRealIP(c), userID, req.AnonymousID, imageData, header.Filename, contentType, options)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientCredits) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient_credits", "message": "积分不足，请充值后继续"})
			return
		}
		if errors.Is(err, service.ErrCreditsExpired) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "credits_expired", "message": "积分已过期，请重新购买"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create image edit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": generation.ID, "status": generation.Status})
}

const defaultEnabledImageSizes = "1280x720,720x1280,1024x1024,1536x1024,1024x1536,1920x1080,1080x1920,2048x2048"
const legacyDefaultEnabledImageSizes = "1280x720,720x1280,1024x1024,1536x1024,1024x1536"

func enabledImageSizes() []string {
	value := model.GetSettingValue("enabled_image_sizes", defaultEnabledImageSizes)
	if normalizeImageSizesValue(value) == legacyDefaultEnabledImageSizes {
		value = defaultEnabledImageSizes
	}
	parts := strings.Split(value, ",")
	sizes := make([]string, 0, len(parts))
	for _, part := range parts {
		size := strings.TrimSpace(part)
		if size != "" {
			sizes = append(sizes, size)
		}
	}
	if len(sizes) == 0 {
		return []string{"1024x1024"}
	}
	return sizes
}

func normalizeImageSizesValue(value string) string {
	parts := strings.Split(value, ",")
	sizes := make([]string, 0, len(parts))
	for _, part := range parts {
		size := strings.TrimSpace(part)
		if size != "" {
			sizes = append(sizes, size)
		}
	}
	return strings.Join(sizes, ",")
}

func isValidQuality(quality string) bool {
	return quality == "low" || quality == "medium" || quality == "high"
}

func isSupportedEditImageType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	return contentType == "image/png" || contentType == "image/jpeg" || contentType == "image/webp"
}

func validateImageOutputOptions(c *gin.Context, outputFormat, background string, outputCompression *int) bool {
	if !isValidOutputFormat(outputFormat) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid output_format, must be png, jpeg, or webp"})
		return false
	}
	if !isValidBackground(background) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid background, must be opaque or transparent"})
		return false
	}
	if !isValidOutputCompression(outputCompression) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "output_compression must be between 0 and 100"})
		return false
	}
	if background == "transparent" && outputFormat == "jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transparent background is not supported with jpeg format"})
		return false
	}
	return true
}

func isValidOutputFormat(format string) bool {
	return format == "" || format == "png" || format == "jpeg" || format == "webp"
}

func isValidBackground(background string) bool {
	return background == "" || background == "opaque" || background == "transparent"
}

func isValidOutputCompression(compression *int) bool {
	return compression == nil || (*compression >= 0 && *compression <= 100)
}

func buildImageSizeOptions(sizes []string) []imageSizeOption {
	options := make([]imageSizeOption, 0, len(sizes))
	for _, size := range sizes {
		ratio := imageRatioLabel(size)
		label := ratio
		if ratio == "" {
			label = strings.Replace(size, "x", " x ", 1)
		}
		options = append(options, imageSizeOption{Value: size, Label: label, Ratio: ratio, CreditCost: service.CostForSize(size)})
	}
	return options
}

func imageRatioLabel(size string) string {
	width, height, ok := parseImageSize(size)
	if !ok {
		return ""
	}
	gcd := greatestCommonDivisor(width, height)
	if gcd == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", width/gcd, height/gcd)
}

func greatestCommonDivisor(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func isEnabledImageSize(size string) bool {
	for _, item := range enabledImageSizes() {
		if item == size {
			width, height, ok := parseImageSize(size)
			return ok && isGPTImage2CompatibleSize(width, height)
		}
	}
	return false
}

func isGPTImage2CompatibleSize(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	if width%16 != 0 || height%16 != 0 {
		return false
	}
	if width > 3840 || height > 3840 {
		return false
	}
	longSide := width
	shortSide := height
	if height > width {
		longSide = height
		shortSide = width
	}
	if longSide > shortSide*3 {
		return false
	}
	pixels := width * height
	return pixels >= 655360 && pixels <= 8294400
}

func parseImageSize(size string) (int, int, bool) {
	return service.ParseImageSize(size)
}

func CaptchaConfig(c *gin.Context) {
	c.JSON(http.StatusOK, service.GetCaptchaConfig())
}

func AdminGenerations(c *gin.Context) {
	page, pageSize := pagination(c)
	query := model.DB.Model(&model.Generation{})
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status := c.Query("status"); status != "" {
		if parsed, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", parsed)
		}
	}
	query = applyTimeRange(c, query)

	var total int64
	var items []model.Generation
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count generations"})
		return
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list generations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func AdminBatchDeleteGenerations(c *gin.Context) {
	var req batchDeleteGenerationsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.DeleteR2 {
		var generations []model.Generation
		if err := model.DB.Where("id IN ?", req.IDs).Find(&generations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load generations"})
			return
		}
		for _, generation := range generations {
			if err := service.DeleteR2Object(generation.R2Key); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete r2 object"})
				return
			}
		}
	}
	result := model.DB.Model(&model.Generation{}).Where("id IN ?", req.IDs).Update("is_deleted", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete generations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": result.RowsAffected})
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

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := service.Notifier.Subscribe(id)
	defer service.Notifier.Unsubscribe(id, ch)

	initialEvent := service.GenerationEvent{
		Status:   generation.Status,
		Message:  statusMessage(generation.Status),
		ImageURL: generation.ImageURL,
		Error:    generation.ErrorMsg,
	}
	sendSSE(c, initialEvent)
	if isTerminalGenerationStatus(initialEvent.Status) {
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-ch:
			sendSSE(c, event)
			if isTerminalGenerationStatus(event.Status) {
				return
			}
		case <-ticker.C:
			_, _ = fmt.Fprint(c.Writer, ":keepalive\n\n")
			c.Writer.Flush()
		}
	}
}

func isTerminalGenerationStatus(status int) bool {
	return status == 3 || status == 4 || status == 5
}

func sendSSE(c *gin.Context, event service.GenerationEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(c.Writer, "event:status\ndata:%s\n\n", payload)
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
