package model

import "time"

type User struct {
	ID            int64   `gorm:"primaryKey;autoIncrement"`
	Username      string  `gorm:"size:64;uniqueIndex"`
	Email         string  `gorm:"size:128;uniqueIndex"`
	PasswordHash  string  `gorm:"size:256" json:"-"`
	WechatOpenID  string  `gorm:"size:128;index"`
	Role          int     `gorm:"default:1"`
	Status        int     `gorm:"default:1"`
	Credits       float64 `gorm:"type:numeric;default:0"`
	CreditsExpiry *time.Time
	AvatarURL     string `gorm:"size:512"`
	LastLoginAt   *time.Time
	LastLoginIP   string `gorm:"size:64"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Generation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	UserID      *int64    `gorm:"index"`
	AnonymousID string    `gorm:"size:128;index"`
	Prompt      string    `gorm:"type:text"`
	Quality     string    `gorm:"size:16"`
	Size        string    `gorm:"size:16"`
	CreditsCost float64   `gorm:"type:numeric"`
	Status      int       `gorm:"default:0;index"`
	R2Key       string    `gorm:"size:256"`
	ImageURL    string    `gorm:"size:512"`
	ErrorMsg    string    `gorm:"size:512"`
	IP          string    `gorm:"size:64"`
	IsDeleted   bool      `gorm:"default:false;index"`
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
}

type CreditLog struct {
	ID         int64   `gorm:"primaryKey;autoIncrement"`
	UserID     int64   `gorm:"index"`
	Type       int     `gorm:"index"`
	Amount     float64 `gorm:"type:numeric"`
	Balance    float64 `gorm:"type:numeric"`
	RelatedID  *int64
	Remark     string `gorm:"size:256"`
	OperatorID *int64
	CreatedAt  time.Time `gorm:"index"`
}

type LoginLog struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	UserID    int64  `gorm:"index"`
	IP        string `gorm:"size:64"`
	UserAgent string `gorm:"size:512"`
	Method    string `gorm:"size:16"`
	Success   bool
	CreatedAt time.Time `gorm:"index"`
}

type Channel struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:64"`
	BaseURL   string `gorm:"size:256"`
	APIKey    string `gorm:"size:256"`
	Headers   string `gorm:"type:text"`
	Status    int    `gorm:"default:1"`
	Weight    int    `gorm:"default:1"`
	Remark    string `gorm:"size:256"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Setting struct {
	Key       string `gorm:"primaryKey;size:64"`
	Value     string `gorm:"type:text"`
	UpdatedAt time.Time
}

type PromptTemplate struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Category  string `gorm:"size:32;index"`
	Label     string `gorm:"size:64"`
	Prompt    string `gorm:"type:text"`
	SortOrder int    `gorm:"default:0"`
	Status    int    `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnonymousIdentity struct {
	ID              int64  `gorm:"primaryKey;autoIncrement"`
	AnonymousID     string `gorm:"size:128;uniqueIndex"`
	Fingerprint     string `gorm:"size:256"`
	IP              string `gorm:"size:64;index"`
	UserAgent       string `gorm:"size:512"`
	FreeUsed        bool   `gorm:"default:false"`
	LastUsedAt      *time.Time
	ClaimedByUserID *int64 `gorm:"index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
