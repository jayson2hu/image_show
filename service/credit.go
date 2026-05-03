package service

import (
	"errors"
	"time"

	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
)

var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrCreditsExpired      = errors.New("credits expired")
)

var QualityCost = map[string]float64{
	"low":    0.2,
	"medium": 1.0,
	"high":   4.0,
}

func CostForQuality(quality string) float64 {
	return QualityCost[quality]
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
	user.Credits = 3
	user.CreditsExpiry = &expiry
	if err := tx.Model(user).Updates(map[string]interface{}{"credits": user.Credits, "credits_expiry": user.CreditsExpiry}).Error; err != nil {
		return err
	}
	return tx.Create(&model.CreditLog{
		UserID:  user.ID,
		Type:    1,
		Amount:  3,
		Balance: 3,
		Remark:  "register gift",
	}).Error
}
