package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

type accountUserResponse struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	AvatarURL     string     `json:"avatar_url"`
	Role          int        `json:"role"`
	Status        int        `json:"status"`
	Credits       float64    `json:"credits"`
	CreditsExpiry *time.Time `json:"credits_expiry"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastLoginIP   string     `json:"last_login_ip"`
}

type accountGenerationResponse struct {
	ID          int64     `json:"id"`
	Prompt      string    `json:"prompt"`
	Mode        string    `json:"mode"`
	Quality     string    `json:"quality"`
	Size        string    `json:"size"`
	Status      int       `json:"status"`
	ImageURL    string    `json:"image_url"`
	ErrorMsg    string    `json:"error_msg"`
	CreditsCost float64   `json:"credits_cost"`
	CreatedAt   time.Time `json:"created_at"`
}

type accountAnnouncementResponse struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	NotifyMode string     `json:"notify_mode"`
	Target     string     `json:"target"`
	SortOrder  int        `json:"sort_order"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type accountLoginResponse struct {
	Method    string    `json:"method"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

type updateAccountProfileRequest struct {
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

func AccountOverview(c *gin.Context) {
	userID := c.GetInt64("userID")
	role := c.GetInt("role")
	var user model.User
	if err := model.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	recentCreditLogs, err := accountRecentCreditLogs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credit logs"})
		return
	}
	creationSummary, err := accountCreationSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load creations"})
		return
	}
	announcementSummary, err := accountAnnouncementSummary(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load announcements"})
		return
	}
	securitySummary, err := accountSecuritySummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load security summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          accountUserFromModel(user),
		"credits":       gin.H{"recent_logs": recentCreditLogs},
		"creations":     creationSummary,
		"announcements": announcementSummary,
		"security":      securitySummary,
	})
}

func UpdateAccountProfile(c *gin.Context) {
	userID := c.GetInt64("userID")
	var req updateAccountProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	username := strings.TrimSpace(req.Username)
	avatarURL := strings.TrimSpace(req.AvatarURL)
	if len([]rune(username)) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is too long"})
		return
	}
	if len(avatarURL) > 512 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar_url is too long"})
		return
	}
	if avatarURL != "" && !strings.HasPrefix(avatarURL, "http://") && !strings.HasPrefix(avatarURL, "https://") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar_url must start with http:// or https://"})
		return
	}

	var user model.User
	if err := model.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if err := model.DB.Model(&user).Updates(map[string]interface{}{
		"username":   username,
		"avatar_url": avatarURL,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	user.Username = username
	user.AvatarURL = avatarURL
	c.JSON(http.StatusOK, gin.H{"user": accountUserFromModel(user)})
}

func accountUserFromModel(user model.User) accountUserResponse {
	return accountUserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		Status:        user.Status,
		Credits:       user.Credits,
		CreditsExpiry: user.CreditsExpiry,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		LastLoginAt:   user.LastLoginAt,
		LastLoginIP:   user.LastLoginIP,
	}
}

func accountRecentCreditLogs(userID int64) ([]model.CreditLog, error) {
	var logs []model.CreditLog
	err := model.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(5).
		Find(&logs).Error
	return logs, err
}

func accountCreationSummary(userID int64) (gin.H, error) {
	base := model.DB.Model(&model.Generation{}).Where("user_id = ? AND is_deleted = ?", userID, false)
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var completed int64
	if err := base.Session(&gorm.Session{}).Where("status = ?", 3).Count(&completed).Error; err != nil {
		return nil, err
	}
	var failed int64
	if err := base.Session(&gorm.Session{}).Where("status = ?", 4).Count(&failed).Error; err != nil {
		return nil, err
	}
	var latest model.Generation
	var latestAt *time.Time
	if err := base.Session(&gorm.Session{}).Order("created_at DESC, id DESC").First(&latest).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	} else {
		value := latest.CreatedAt
		latestAt = &value
	}
	var items []model.Generation
	if err := base.Session(&gorm.Session{}).
		Order("created_at DESC, id DESC").
		Limit(6).
		Find(&items).Error; err != nil {
		return nil, err
	}
	recentItems := make([]accountGenerationResponse, 0, len(items))
	for _, item := range items {
		recentItems = append(recentItems, accountGenerationResponse{
			ID:          item.ID,
			Prompt:      item.Prompt,
			Mode:        item.Mode,
			Quality:     item.Quality,
			Size:        item.Size,
			Status:      item.Status,
			ImageURL:    item.ImageURL,
			ErrorMsg:    item.ErrorMsg,
			CreditsCost: item.CreditsCost,
			CreatedAt:   item.CreatedAt,
		})
	}
	return gin.H{
		"total":        total,
		"completed":    completed,
		"failed":       failed,
		"latest_at":    latestAt,
		"recent_items": recentItems,
	}, nil
}

func accountAnnouncementSummary(userID int64, role int) (gin.H, error) {
	var items []model.Announcement
	if err := model.DB.
		Scopes(activeAnnouncementScope(time.Now())).
		Scopes(announcementTargetScope(userID, role)).
		Order("sort_order ASC, created_at DESC, id DESC").
		Limit(5).
		Find(&items).Error; err != nil {
		return nil, err
	}
	readMap := map[int64]time.Time{}
	if len(items) > 0 {
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		var reads []model.AnnouncementRead
		if err := model.DB.Where("user_id = ? AND announcement_id IN ?", userID, ids).Find(&reads).Error; err != nil {
			return nil, err
		}
		for _, read := range reads {
			readMap[read.AnnouncementID] = read.ReadAt
		}
	}
	recentItems := make([]accountAnnouncementResponse, 0, len(items))
	var unreadCount int
	for _, item := range items {
		var readAt *time.Time
		if value, ok := readMap[item.ID]; ok {
			v := value
			readAt = &v
		} else {
			unreadCount++
		}
		recentItems = append(recentItems, accountAnnouncementResponse{
			ID:         item.ID,
			Title:      item.Title,
			Content:    item.Content,
			NotifyMode: item.NotifyMode,
			Target:     item.Target,
			SortOrder:  item.SortOrder,
			ReadAt:     readAt,
			CreatedAt:  item.CreatedAt,
		})
	}
	return gin.H{"unread_count": unreadCount, "recent_items": recentItems}, nil
}

func accountSecuritySummary(userID int64) (gin.H, error) {
	var login model.LoginLog
	if err := model.DB.
		Where("user_id = ? AND success = ?", userID, true).
		Order("created_at DESC, id DESC").
		First(&login).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return gin.H{"latest_login": nil}, nil
		}
		return nil, err
	}
	return gin.H{"latest_login": accountLoginResponse{Method: login.Method, IP: login.IP, CreatedAt: login.CreatedAt}}, nil
}
