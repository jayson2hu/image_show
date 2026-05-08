package controller_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jayson2hu/image-show/model"
)

func TestAccountOverviewRequiresAuth(t *testing.T) {
	engine := setupAuthTest(t)
	rec := adminRequest(engine, http.MethodGet, "/api/account/overview", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountOverviewReturnsCurrentUserAssets(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	otherID := userID + 100
	now := time.Now()
	expiry := now.Add(30 * 24 * time.Hour)
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"username":       "account-user",
		"avatar_url":     "https://example.com/avatar.png",
		"credits":        8.5,
		"credits_expiry": expiry,
		"last_login_at":  now,
		"last_login_ip":  "1.2.3.4",
	}).Error; err != nil {
		t.Fatalf("update user: %v", err)
	}
	if err := model.DB.Create(&[]model.CreditLog{
		{UserID: userID, Type: 5, Amount: 10, Balance: 10, Remark: "gift", CreatedAt: now.Add(-time.Minute)},
		{UserID: otherID, Type: 5, Amount: 99, Balance: 99, Remark: "other", CreatedAt: now},
	}).Error; err != nil {
		t.Fatalf("create credit logs: %v", err)
	}
	if err := model.DB.Create(&[]model.Generation{
		{UserID: &userID, Prompt: "owned success", Size: "1024x1024", Status: 3, ImageURL: "https://example.com/1.png", CreditsCost: 1, CreatedAt: now.Add(-2 * time.Minute)},
		{UserID: &userID, Prompt: "owned failed", Size: "1536x1024", Status: 4, ErrorMsg: "upstream 503", CreditsCost: 2, CreatedAt: now.Add(-time.Minute)},
		{UserID: &otherID, Prompt: "other private", Size: "1024x1024", Status: 3, CreatedAt: now},
	}).Error; err != nil {
		t.Fatalf("create generations: %v", err)
	}
	announcement := model.Announcement{Title: "notice", Content: "hello", Status: 1, Target: "user", NotifyMode: "silent", CreatedAt: now}
	if err := model.DB.Create(&announcement).Error; err != nil {
		t.Fatalf("create announcement: %v", err)
	}
	if err := model.DB.Create(&model.LoginLog{UserID: userID, IP: "5.6.7.8", Method: "email", Success: true, CreatedAt: now.Add(time.Minute)}).Error; err != nil {
		t.Fatalf("create login log: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/account/overview", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("overview status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, sensitive := range []string{"password_hash", "PasswordHash", "api_key", "wechat_server_token"} {
		if strings.Contains(body, sensitive) {
			t.Fatalf("overview leaked sensitive field %s: %s", sensitive, body)
		}
	}
	var resp struct {
		User struct {
			ID       int64   `json:"id"`
			Username string  `json:"username"`
			Email    string  `json:"email"`
			Credits  float64 `json:"credits"`
		} `json:"user"`
		Credits struct {
			RecentLogs []model.CreditLog `json:"recent_logs"`
		} `json:"credits"`
		Creations struct {
			Total       int64 `json:"total"`
			Completed   int64 `json:"completed"`
			Failed      int64 `json:"failed"`
			RecentItems []struct {
				Prompt string `json:"prompt"`
			} `json:"recent_items"`
		} `json:"creations"`
		Announcements struct {
			UnreadCount int `json:"unread_count"`
			RecentItems []struct {
				Title string `json:"title"`
			} `json:"recent_items"`
		} `json:"announcements"`
		Security struct {
			LatestLogin *struct {
				Method string `json:"method"`
				IP     string `json:"ip"`
			} `json:"latest_login"`
		} `json:"security"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if resp.User.ID != userID || resp.User.Username != "account-user" || resp.User.Credits != 8.5 {
		t.Fatalf("unexpected user response: %+v", resp.User)
	}
	if len(resp.Credits.RecentLogs) != 1 || resp.Credits.RecentLogs[0].Remark != "gift" {
		t.Fatalf("unexpected credit logs: %+v", resp.Credits.RecentLogs)
	}
	if resp.Creations.Total != 2 || resp.Creations.Completed != 1 || resp.Creations.Failed != 1 {
		t.Fatalf("unexpected creation summary: %+v", resp.Creations)
	}
	if len(resp.Creations.RecentItems) != 2 {
		t.Fatalf("unexpected recent creations: %+v", resp.Creations.RecentItems)
	}
	for _, item := range resp.Creations.RecentItems {
		if item.Prompt == "other private" {
			t.Fatalf("overview returned other user generation: %+v", resp.Creations.RecentItems)
		}
	}
	if resp.Announcements.UnreadCount != 1 || len(resp.Announcements.RecentItems) != 1 || resp.Announcements.RecentItems[0].Title != "notice" {
		t.Fatalf("unexpected announcements: %+v", resp.Announcements)
	}
	if resp.Security.LatestLogin == nil || resp.Security.LatestLogin.Method != "email" || resp.Security.LatestLogin.IP != "5.6.7.8" {
		t.Fatalf("unexpected security summary: %+v", resp.Security)
	}
}

func TestAccountOverviewLimitsRecentWorksPreview(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	userID := tokenUserID(t, token)
	now := time.Now()
	items := make([]model.Generation, 0, 5)
	for i := 0; i < 5; i++ {
		items = append(items, model.Generation{
			UserID:    &userID,
			Prompt:    "recent",
			Size:      "1024x1024",
			Status:    3,
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}
	if err := model.DB.Create(&items).Error; err != nil {
		t.Fatalf("create generations: %v", err)
	}

	rec := adminRequest(engine, http.MethodGet, "/api/account/overview", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("overview status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Creations struct {
			Total       int64         `json:"total"`
			RecentItems []interface{} `json:"recent_items"`
		} `json:"creations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if resp.Creations.Total != 5 || len(resp.Creations.RecentItems) != 3 {
		t.Fatalf("expected total 5 and 3 recent items, got %+v", resp.Creations)
	}
}

func TestAccountOverviewEmptyState(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	rec := adminRequest(engine, http.MethodGet, "/api/account/overview", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("overview status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Credits struct {
			RecentLogs []model.CreditLog `json:"recent_logs"`
		} `json:"credits"`
		Creations struct {
			Total       int64         `json:"total"`
			RecentItems []interface{} `json:"recent_items"`
		} `json:"creations"`
		Announcements struct {
			UnreadCount int           `json:"unread_count"`
			RecentItems []interface{} `json:"recent_items"`
		} `json:"announcements"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if resp.Credits.RecentLogs == nil || resp.Creations.RecentItems == nil || resp.Announcements.RecentItems == nil {
		t.Fatalf("expected empty arrays, got %+v", resp)
	}
	if resp.Creations.Total != 0 || resp.Announcements.UnreadCount != 0 {
		t.Fatalf("unexpected empty summary: %+v", resp)
	}
}

func TestAccountProfileUpdate(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	rec := adminJSON(engine, http.MethodPut, "/api/account/profile", map[string]interface{}{
		"username":   "新的昵称",
		"avatar_url": "https://example.com/avatar.png",
		"role":       10,
		"credits":    999,
		"email":      "changed@example.com",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile update status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		User struct {
			ID        int64   `json:"id"`
			Username  string  `json:"username"`
			AvatarURL string  `json:"avatar_url"`
			Email     string  `json:"email"`
			Role      int     `json:"role"`
			Credits   float64 `json:"credits"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode profile update: %v", err)
	}
	if resp.User.Username != "新的昵称" || resp.User.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("unexpected profile response: %+v", resp.User)
	}
	if resp.User.Role == 10 || resp.User.Credits == 999 || resp.User.Email == "changed@example.com" {
		t.Fatalf("sensitive fields should not be updated: %+v", resp.User)
	}
	var user model.User
	if err := model.DB.First(&user, resp.User.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Username != "新的昵称" || user.AvatarURL != "https://example.com/avatar.png" || user.Role == 10 || user.Credits == 999 {
		t.Fatalf("unexpected stored user: %+v", user)
	}
}

func TestAccountProfileRejectsInvalidAvatarURL(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	rec := adminJSON(engine, http.MethodPut, "/api/account/profile", map[string]string{
		"username":   "user",
		"avatar_url": "ftp://example.com/avatar.png",
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad avatar url 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountProfileRequiresAuth(t *testing.T) {
	engine := setupAuthTest(t)
	rec := adminJSON(engine, http.MethodPut, "/api/account/profile", map[string]string{
		"username": "user",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountAvatarUpload(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	rec := avatarUploadRequest(engine, token, "avatar.png", []byte{0x89, 0x50, 0x4e, 0x47})
	if rec.Code != http.StatusOK {
		t.Fatalf("avatar upload status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		AvatarURL string `json:"avatar_url"`
		User      struct {
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode avatar upload: %v", err)
	}
	if !strings.HasPrefix(resp.AvatarURL, "/uploads/avatars/") || resp.User.AvatarURL != resp.AvatarURL {
		t.Fatalf("unexpected avatar response: %+v", resp)
	}
	var user model.User
	if err := model.DB.First(&user, "avatar_url = ?", resp.AvatarURL).Error; err != nil {
		t.Fatalf("avatar url not stored: %v", err)
	}
}

func TestAccountAvatarUploadRequiresAuth(t *testing.T) {
	engine := setupAuthTest(t)
	rec := avatarUploadRequest(engine, "", "avatar.png", []byte("png"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountAvatarUploadRejectsInvalidType(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	rec := avatarUploadRequest(engine, token, "avatar.exe", []byte("bad"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid type 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountAvatarUploadRejectsOversize(t *testing.T) {
	engine := setupAuthTest(t)
	token := createTokenForRole(t, 1)
	if err := model.DB.Create(&model.Setting{Key: "avatar_max_size_mb", Value: "1"}).Error; err != nil {
		t.Fatalf("create avatar size setting: %v", err)
	}
	rec := avatarUploadRequest(engine, token, "avatar.png", bytes.Repeat([]byte("a"), 1024*1024+1))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected oversize 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func avatarUploadRequest(engine http.Handler, token, filename string, content []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("avatar", filename)
	_, _ = part.Write(content)
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/account/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
