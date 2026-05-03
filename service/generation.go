package service

import (
	"sync"
	"time"

	"github.com/jayson2hu/image-show/model"
)

type GenerationEvent struct {
	Status   int    `json:"status"`
	Message  string `json:"message"`
	ImageURL string `json:"image_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

type GenerationNotifier struct {
	mu       sync.RWMutex
	channels map[int64]map[chan GenerationEvent]struct{}
}

var Notifier = &GenerationNotifier{channels: make(map[int64]map[chan GenerationEvent]struct{})}

const generationStartDelay = 600 * time.Millisecond
const GenerationModeGenerate = "generate"
const GenerationModeEdit = "edit"

func (n *GenerationNotifier) Subscribe(id int64) chan GenerationEvent {
	ch := make(chan GenerationEvent, 8)
	n.mu.Lock()
	if n.channels[id] == nil {
		n.channels[id] = make(map[chan GenerationEvent]struct{})
	}
	n.channels[id][ch] = struct{}{}
	n.mu.Unlock()
	return ch
}

func (n *GenerationNotifier) Unsubscribe(id int64, ch chan GenerationEvent) {
	n.mu.Lock()
	if subscribers, ok := n.channels[id]; ok {
		delete(subscribers, ch)
		close(ch)
		if len(subscribers) == 0 {
			delete(n.channels, id)
		}
	}
	n.mu.Unlock()
}

func (n *GenerationNotifier) Publish(id int64, event GenerationEvent) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for ch := range n.channels[id] {
		select {
		case ch <- event:
		default:
		}
	}
}

func CreateGeneration(prompt, quality, size, ip string, userID *int64, anonymousID string) (*model.Generation, error) {
	cost := CostForSize(size)
	if userID != nil {
		balance, err := GetBalance(*userID)
		if err != nil {
			return nil, err
		}
		if balance < cost {
			return nil, ErrInsufficientCredits
		}
	}
	generation := &model.Generation{
		UserID:      userID,
		AnonymousID: anonymousID,
		Mode:        GenerationModeGenerate,
		Prompt:      prompt,
		Quality:     quality,
		Size:        size,
		CreditsCost: cost,
		Status:      0,
		IP:          ip,
	}
	if err := model.DB.Create(generation).Error; err != nil {
		return nil, err
	}
	if userID != nil {
		if err := Deduct(*userID, cost, generation.ID); err != nil {
			_ = model.DB.Delete(generation).Error
			return nil, err
		}
	}
	go func() {
		time.Sleep(generationStartDelay)
		runGeneration(generation.ID, prompt, quality, size, ip)
	}()
	return generation, nil
}

func CreateImageEdit(prompt, quality, size, ip string, userID *int64, anonymousID string, imageData []byte, filename, contentType string) (*model.Generation, error) {
	cost := CostForSize(size)
	if userID != nil {
		balance, err := GetBalance(*userID)
		if err != nil {
			return nil, err
		}
		if balance < cost {
			return nil, ErrInsufficientCredits
		}
	}
	generation := &model.Generation{
		UserID:      userID,
		AnonymousID: anonymousID,
		Mode:        GenerationModeEdit,
		Prompt:      prompt,
		Quality:     quality,
		Size:        size,
		CreditsCost: cost,
		Status:      0,
		IP:          ip,
	}
	if err := model.DB.Create(generation).Error; err != nil {
		return nil, err
	}
	sourceURL, sourceKey, err := StoreSourceImage(generation.ID, imageData, contentType)
	if err != nil {
		_ = model.DB.Delete(generation).Error
		return nil, err
	}
	if sourceURL != "" || sourceKey != "" {
		_ = model.DB.Model(&model.Generation{}).Where("id = ?", generation.ID).Updates(map[string]interface{}{
			"source_image_url": sourceURL,
			"source_r2_key":    sourceKey,
		}).Error
		generation.SourceImageURL = sourceURL
		generation.SourceR2Key = sourceKey
	}
	if userID != nil {
		if err := Deduct(*userID, cost, generation.ID); err != nil {
			_ = DeleteR2Object(sourceKey)
			_ = model.DB.Delete(generation).Error
			return nil, err
		}
	}
	go func() {
		time.Sleep(generationStartDelay)
		runImageEdit(generation.ID, prompt, quality, size, ip, imageData, filename, contentType)
	}()
	return generation, nil
}

func runGeneration(id int64, prompt, quality, size, ip string) {
	if isGenerationCancelled(id) {
		return
	}
	updateGenerationStatus(id, 1, "正在生成图片...", "", "")
	providerSize := ProviderImageSize(size)
	result, err := GenerateImageViaChannels(prompt, quality, providerSize, ip)
	if isGenerationCancelled(id) {
		return
	}
	if err != nil {
		refundGenerationCredits(id)
		updateGenerationStatus(id, 4, "生成失败", "", err.Error())
		return
	}

	updateGenerationStatus(id, 2, "正在保存图片...", "", "")
	imageURL, r2Key, err := StoreGeneratedImage(id, result, size)
	if isGenerationCancelled(id) {
		return
	}
	if err != nil {
		refundGenerationCredits(id)
		updateGenerationStatus(id, 4, "图片保存失败", "", err.Error())
		return
	}
	updateGenerationStatus(id, 3, "生成完成", imageURL, "")
	if r2Key != "" {
		_ = model.DB.Model(&model.Generation{}).Where("id = ? AND status <> ?", id, 5).Update("r2_key", r2Key).Error
	}
}

func runImageEdit(id int64, prompt, quality, size, ip string, imageData []byte, filename, contentType string) {
	if isGenerationCancelled(id) {
		return
	}
	updateGenerationStatus(id, 1, "正在编辑图片...", "", "")
	providerSize := ProviderImageSize(size)
	result, err := EditImageViaChannels(prompt, quality, providerSize, ip, imageData, filename, contentType)
	if isGenerationCancelled(id) {
		return
	}
	if err != nil {
		refundGenerationCredits(id)
		updateGenerationStatus(id, 4, "图片编辑失败", "", err.Error())
		return
	}

	updateGenerationStatus(id, 2, "正在保存图片...", "", "")
	imageURL, r2Key, err := StoreGeneratedImage(id, result, size)
	if isGenerationCancelled(id) {
		return
	}
	if err != nil {
		refundGenerationCredits(id)
		updateGenerationStatus(id, 4, "图片保存失败", "", err.Error())
		return
	}
	updateGenerationStatus(id, 3, "编辑完成", imageURL, "")
	if r2Key != "" {
		_ = model.DB.Model(&model.Generation{}).Where("id = ? AND status <> ?", id, 5).Update("r2_key", r2Key).Error
	}
}

func CancelGeneration(id, userID int64) (bool, error) {
	var generation model.Generation
	if err := model.DB.Where("id = ? AND user_id = ?", id, userID).First(&generation).Error; err != nil {
		return false, err
	}
	if generation.Status == 5 {
		return false, nil
	}
	refunded := false
	if generation.Status == 0 && generation.UserID != nil && generation.CreditsCost > 0 {
		if err := Refund(*generation.UserID, generation.CreditsCost, generation.ID); err != nil {
			return false, err
		}
		refunded = true
	}
	if err := model.DB.Model(&model.Generation{}).Where("id = ?", id).Update("status", 5).Error; err != nil {
		return false, err
	}
	Notifier.Publish(id, GenerationEvent{Status: 5, Message: "任务已取消"})
	return refunded, nil
}

func refundGenerationCredits(id int64) {
	var generation model.Generation
	if err := model.DB.First(&generation, id).Error; err != nil {
		return
	}
	if generation.UserID != nil && generation.CreditsCost > 0 {
		_ = Refund(*generation.UserID, generation.CreditsCost, generation.ID)
	}
}

func updateGenerationStatus(id int64, status int, message, imageURL, errMsg string) {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if imageURL != "" {
		updates["image_url"] = imageURL
	}
	if errMsg != "" {
		updates["error_msg"] = errMsg
	}
	_ = model.DB.Model(&model.Generation{}).Where("id = ? AND status <> ?", id, 5).Updates(updates).Error
	Notifier.Publish(id, GenerationEvent{
		Status:   status,
		Message:  message,
		ImageURL: imageURL,
		Error:    errMsg,
	})
}

func isGenerationCancelled(id int64) bool {
	var generation model.Generation
	if err := model.DB.Select("status").First(&generation, id).Error; err != nil {
		return false
	}
	return generation.Status == 5
}
