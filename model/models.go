package model

import "time"

type User struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string     `gorm:"size:64;index" json:"username"`
	Email         string     `gorm:"size:128;uniqueIndex" json:"email"`
	PasswordHash  string     `gorm:"size:256" json:"-"`
	WechatOpenID  string     `gorm:"size:128;index" json:"wechat_open_id"`
	Role          int        `gorm:"default:1" json:"role"`
	Status        int        `gorm:"default:1" json:"status"`
	Credits       float64    `gorm:"type:numeric;default:0" json:"credits"`
	CreditsExpiry *time.Time `json:"credits_expiry"`
	AvatarURL     string     `gorm:"size:512" json:"avatar_url"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastLoginIP   string     `gorm:"size:64" json:"last_login_ip"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Generation struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         *int64    `gorm:"index" json:"user_id"`
	AnonymousID    string    `gorm:"size:128;index" json:"anonymous_id"`
	Mode           string    `gorm:"size:16;default:generate;index" json:"mode"`
	Prompt         string    `gorm:"type:text" json:"prompt"`
	Quality        string    `gorm:"size:16" json:"quality"`
	Size           string    `gorm:"size:16" json:"size"`
	CreditsCost    float64   `gorm:"type:numeric" json:"credits_cost"`
	Status         int       `gorm:"default:0;index" json:"status"`
	R2Key          string    `gorm:"size:256" json:"r2_key"`
	ImageURL       string    `gorm:"size:512" json:"image_url"`
	SourceR2Key    string    `gorm:"size:256" json:"source_r2_key"`
	SourceImageURL string    `gorm:"size:512" json:"source_image_url"`
	ErrorMsg       string    `gorm:"size:512" json:"error_msg"`
	IP             string    `gorm:"size:64" json:"ip"`
	IsDeleted      bool      `gorm:"default:false;index" json:"is_deleted"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreditLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"index" json:"user_id"`
	Type       int       `gorm:"index" json:"type"`
	Amount     float64   `gorm:"type:numeric" json:"amount"`
	Balance    float64   `gorm:"type:numeric" json:"balance"`
	RelatedID  *int64    `json:"related_id"`
	Remark     string    `gorm:"size:256" json:"remark"`
	OperatorID *int64    `json:"operator_id"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

type LoginLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"index" json:"user_id"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:512" json:"user_agent"`
	Method    string    `gorm:"size:16" json:"method"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

type Channel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	BaseURL   string    `gorm:"size:256" json:"base_url"`
	APIKey    string    `gorm:"size:256" json:"api_key"`
	Headers   string    `gorm:"type:text" json:"headers"`
	Status    int       `gorm:"default:1" json:"status"`
	Weight    int       `gorm:"default:1" json:"weight"`
	Remark    string    `gorm:"size:256" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Setting struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PromptTemplate struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Category  string    `gorm:"size:32;index" json:"category"`
	Label     string    `gorm:"size:64" json:"label"`
	Prompt    string    `gorm:"type:text" json:"prompt"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Package struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	Credits   float64   `gorm:"type:numeric" json:"credits"`
	Price     float64   `gorm:"type:numeric" json:"price"`
	ValidDays int       `json:"valid_days"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo    string     `gorm:"size:64;uniqueIndex" json:"order_no"`
	UserID     int64      `gorm:"index" json:"user_id"`
	PackageID  int64      `gorm:"index" json:"package_id"`
	Amount     float64    `gorm:"type:numeric" json:"amount"`
	Status     int        `gorm:"default:0;index" json:"status"`
	PayMethod  string     `gorm:"size:32" json:"pay_method"`
	PayTradeNo string     `gorm:"size:128" json:"pay_trade_no"`
	PaidAt     *time.Time `json:"paid_at"`
	CreatedAt  time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type AnonymousIdentity struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AnonymousID     string     `gorm:"size:128;uniqueIndex" json:"anonymous_id"`
	Fingerprint     string     `gorm:"size:256" json:"fingerprint"`
	IP              string     `gorm:"size:64;index" json:"ip"`
	UserAgent       string     `gorm:"size:512" json:"user_agent"`
	FreeUsed        bool       `gorm:"default:false" json:"free_used"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	ClaimedByUserID *int64     `gorm:"index" json:"claimed_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
