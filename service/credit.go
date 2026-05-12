package service

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrCreditsExpired      = errors.New("credits expired")
)

const (
	CreditCostSquareKey     = "credit_cost_square"
	CreditCostPortraitKey   = "credit_cost_portrait"
	CreditCostStoryKey      = "credit_cost_story"
	CreditCostLandscapeKey  = "credit_cost_landscape"
	CreditCostWidescreenKey = "credit_cost_widescreen"
)

var defaultRatioCreditCosts = map[string]float64{
	"square":        1,
	"portrait_3_4":  2,
	"story":         2,
	"landscape_4_3": 2,
	"widescreen":    2,
}

var ratioCreditSettingKeys = map[string]string{
	"square":        CreditCostSquareKey,
	"portrait_3_4":  CreditCostPortraitKey,
	"story":         CreditCostStoryKey,
	"landscape_4_3": CreditCostLandscapeKey,
	"widescreen":    CreditCostWidescreenKey,
}

var pixelSizeRatioKeys = map[string]string{
	"1024x1024": "square",
	"1152x1536": "portrait_3_4",
	"1008x1792": "story",
	"1536x1152": "landscape_4_3",
	"1792x1008": "widescreen",
}

func CreditCostsByRatio() map[string]float64 {
	return map[string]float64{
		"square":     CostForRatio("square"),
		"portrait":   CostForRatio("portrait_3_4"),
		"story":      CostForRatio("story"),
		"landscape":  CostForRatio("landscape_4_3"),
		"widescreen": CostForRatio("widescreen"),
	}
}

func CostForRatio(ratio string) float64 {
	ratio = normalizeCreditRatio(ratio)
	defaultCost, ok := defaultRatioCreditCosts[ratio]
	if !ok {
		return 1
	}
	value := strings.TrimSpace(model.GetSettingValue(ratioCreditSettingKeys[ratio], strconv.FormatFloat(defaultCost, 'f', 0, 64)))
	cost, err := strconv.ParseFloat(value, 64)
	if err != nil || cost < 1 {
		return defaultCost
	}
	return math.Ceil(cost)
}

func CostForSize(size string) float64 {
	ratio := normalizeCreditRatio(size)
	if _, ok := defaultRatioCreditCosts[ratio]; ok {
		return CostForRatio(ratio)
	}
	width, height, ok := parseCreditImageSize(size)
	if !ok || width <= 0 || height <= 0 {
		return 1
	}
	basePixels := 1024.0 * 1024.0
	cost := float64(width*height) / basePixels
	if cost < 1 {
		cost = 1
	}
	return math.Ceil(cost)
}

func parseCreditImageSize(size string) (int, int, bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) != 2 {
		return 0, 0, false
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}
	return width, height, true
}

func normalizeCreditRatio(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "1:1", "square":
		return "square"
	case "3:4", "portrait", "portrait_3_4":
		return "portrait_3_4"
	case "9:16", "story":
		return "story"
	case "4:3", "landscape", "landscape_4_3":
		return "landscape_4_3"
	case "16:9", "widescreen":
		return "widescreen"
	default:
		if ratio, ok := pixelSizeRatioKeys[value]; ok {
			return ratio
		}
		return value
	}
}

func GetBalance(userID int64) (float64, error) {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		return 0, err
	}
	if user.CreditsExpiry != nil && time.Now().After(*user.CreditsExpiry) {
		return 0, nil
	}
	return user.Credits, nil
}

func Deduct(userID int64, amount float64, generationID int64) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		if user.CreditsExpiry != nil && time.Now().After(*user.CreditsExpiry) {
			return ErrCreditsExpired
		}
		if user.Credits < amount {
			return ErrInsufficientCredits
		}
		user.Credits -= amount
		if err := tx.Model(&user).Updates(map[string]interface{}{"credits": user.Credits}).Error; err != nil {
			return err
		}
		return tx.Create(&model.CreditLog{
			UserID:    user.ID,
			Type:      2,
			Amount:    -amount,
			Balance:   user.Credits,
			RelatedID: &generationID,
			Remark:    "generation consume",
		}).Error
	})
}

func Refund(userID int64, amount float64, generationID int64) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		user.Credits += amount
		if err := tx.Model(&user).Updates(map[string]interface{}{"credits": user.Credits}).Error; err != nil {
			return err
		}
		return tx.Create(&model.CreditLog{
			UserID:    user.ID,
			Type:      4,
			Amount:    amount,
			Balance:   user.Credits,
			RelatedID: &generationID,
			Remark:    "generation refund",
		}).Error
	})
}

func AdminTopup(userID, operatorID int64, amount float64, remark string) error {
	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		user.Credits += amount
		if err := tx.Model(&user).Updates(map[string]interface{}{"credits": user.Credits}).Error; err != nil {
			return err
		}
		return tx.Create(&model.CreditLog{
			UserID:     user.ID,
			Type:       3,
			Amount:     amount,
			Balance:    user.Credits,
			Remark:     remark,
			OperatorID: &operatorID,
		}).Error
	}); err != nil {
		return err
	}
	return PromoteUserGenerationsToPaid(userID)
}

func RegisterGift(tx *gorm.DB, user *model.User) error {
	expiry := time.Now().Add(7 * 24 * time.Hour)
	amount := RegisterGiftCredits()
	user.Credits = amount
	user.CreditsExpiry = &expiry
	if err := tx.Model(user).Updates(map[string]interface{}{"credits": user.Credits, "credits_expiry": user.CreditsExpiry}).Error; err != nil {
		return err
	}
	return tx.Create(&model.CreditLog{
		UserID:  user.ID,
		Type:    1,
		Amount:  amount,
		Balance: amount,
		Remark:  "register gift",
	}).Error
}

func RegisterGiftCredits() float64 {
	value := strings.TrimSpace(model.GetSettingValue("register_gift_credits", "10"))
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil || amount < 0 {
		return 10
	}
	return amount
}
