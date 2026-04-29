package service

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonitorSummary struct {
	Date             string  `json:"date"`
	GenerationCount  int64   `json:"generation_count"`
	CompletedCount   int64   `json:"completed_count"`
	FailedCount      int64   `json:"failed_count"`
	CreditsConsumed  float64 `json:"credits_consumed"`
	NewUsers         int64   `json:"new_users"`
	PaidOrderCount   int64   `json:"paid_order_count"`
	PaidOrderAmount  float64 `json:"paid_order_amount"`
	AlertThreshold   float64 `json:"alert_threshold"`
	AlertTriggered   bool    `json:"alert_triggered"`
	AlertAlreadySent bool    `json:"alert_already_sent"`
}

type MonitorAlertResult struct {
	Triggered bool `json:"triggered"`
	Sent      bool `json:"sent"`
	Skipped   bool `json:"skipped"`
}

func GetMonitorSummary(day time.Time) (*MonitorSummary, error) {
	start := dayStart(day)
	end := start.Add(24 * time.Hour)
	summary := &MonitorSummary{Date: start.Format("2006-01-02")}

	if err := countBetween(model.DB.Model(&model.Generation{}), start, end, &summary.GenerationCount); err != nil {
		return nil, err
	}
	if err := countBetween(model.DB.Model(&model.Generation{}).Where("status = ?", 3), start, end, &summary.CompletedCount); err != nil {
		return nil, err
	}
	if err := countBetween(model.DB.Model(&model.Generation{}).Where("status = ?", 4), start, end, &summary.FailedCount); err != nil {
		return nil, err
	}
	if err := countBetween(model.DB.Model(&model.User{}), start, end, &summary.NewUsers); err != nil {
		return nil, err
	}
	if err := countBetween(model.DB.Model(&model.Order{}).Where("status = ?", OrderStatusPaid), start, end, &summary.PaidOrderCount); err != nil {
		return nil, err
	}
	if err := model.DB.Model(&model.Order{}).Where("status = ? AND paid_at >= ? AND paid_at < ?", OrderStatusPaid, start, end).
		Select("COALESCE(SUM(amount), 0)").Scan(&summary.PaidOrderAmount).Error; err != nil {
		return nil, err
	}
	var consumed float64
	if err := model.DB.Model(&model.CreditLog{}).Where("type = ? AND created_at >= ? AND created_at < ?", 2, start, end).
		Select("COALESCE(SUM(amount), 0)").Scan(&consumed).Error; err != nil {
		return nil, err
	}
	if consumed < 0 {
		consumed = -consumed
	}
	summary.CreditsConsumed = consumed
	summary.AlertThreshold = monitorThreshold()
	summary.AlertTriggered = summary.AlertThreshold > 0 && summary.CreditsConsumed >= summary.AlertThreshold
	summary.AlertAlreadySent = model.GetSettingValue("monitor_alert_last_date", "") == summary.Date
	return summary, nil
}

func CheckMonitorAlert(day time.Time) (*MonitorAlertResult, error) {
	summary, err := GetMonitorSummary(day)
	if err != nil {
		return nil, err
	}
	result := &MonitorAlertResult{Triggered: summary.AlertTriggered}
	if !summary.AlertTriggered {
		result.Skipped = true
		return result, nil
	}
	if summary.AlertAlreadySent {
		result.Skipped = true
		return result, nil
	}
	if err := sendMonitorAlert(summary); err != nil {
		return nil, err
	}
	if err := upsertSetting("monitor_alert_last_date", summary.Date); err != nil {
		return nil, err
	}
	result.Sent = true
	return result, nil
}

func countBetween(query *gorm.DB, start, end time.Time, total *int64) error {
	return query.Where("created_at >= ? AND created_at < ?", start, end).Count(total).Error
}

func monitorThreshold() float64 {
	raw := model.GetSettingValue("monitor_daily_credit_threshold", "0")
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func sendMonitorAlert(summary *MonitorSummary) error {
	var admins []model.User
	if err := model.DB.Where("role >= ? AND status = ? AND email <> ''", 10, 1).Find(&admins).Error; err != nil {
		return err
	}
	recipients := make([]string, 0, len(admins))
	for _, admin := range admins {
		recipients = append(recipients, admin.Email)
	}
	body := fmt.Sprintf("Image Show daily credit usage alert\nDate: %s\nCredits consumed: %.2f\nThreshold: %.2f\nGenerations: %d\n",
		summary.Date, summary.CreditsConsumed, summary.AlertThreshold, summary.GenerationCount)
	return sendPlainMail(recipients, "Image Show monitor alert", body)
}

func sendPlainMail(to []string, subject, body string) error {
	if len(to) == 0 {
		return nil
	}
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		log.Printf("monitor alert mail skipped: %s", body)
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUser
	}
	msg := []byte("To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")
	return smtp.SendMail(addr, auth, from, to, msg)
}

func upsertSetting(key, value string) error {
	return model.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&model.Setting{Key: key, Value: value}).Error
}

func dayStart(day time.Time) time.Time {
	y, m, d := day.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, day.Location())
}
