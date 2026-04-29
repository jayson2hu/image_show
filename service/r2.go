package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

type R2Client struct {
	client *s3.Client
	bucket string
	cdnURL string
}

func NewR2ClientFromConfig() (*R2Client, error) {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if cfg.R2Endpoint == "" || cfg.R2AccessKey == "" || cfg.R2SecretKey == "" || cfg.R2Bucket == "" {
		return nil, nil
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2AccessKey, cfg.R2SecretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.R2Endpoint)
		o.UsePathStyle = true
	})
	return &R2Client{client: client, bucket: cfg.R2Bucket, cdnURL: strings.TrimRight(cfg.R2PublicURL, "/")}, nil
}

func (r *R2Client) Upload(key string, data []byte, contentType string) error {
	_, err := r.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

func (r *R2Client) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
	if r.cdnURL != "" {
		return r.cdnURL + "/" + strings.TrimLeft(key, "/"), nil
	}
	presigner := s3.NewPresignClient(r.client)
	result, err := presigner.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (r *R2Client) Delete(key string) error {
	_, err := r.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (r *R2Client) Copy(sourceKey, targetKey string) error {
	copySource := url.PathEscape(r.bucket + "/" + sourceKey)
	_, err := r.client.CopyObject(context.Background(), &s3.CopyObjectInput{
		Bucket:     aws.String(r.bucket),
		Key:        aws.String(targetKey),
		CopySource: aws.String(copySource),
	})
	return err
}

func StoreGeneratedImage(generationID int64, result *ImageGenerationResult) (imageURL string, r2Key string, err error) {
	if result.URL != "" {
		return result.URL, "", nil
	}
	if result.Base64Data == "" {
		return "", "", fmt.Errorf("empty image result")
	}

	data, err := base64.StdEncoding.DecodeString(result.Base64Data)
	if err != nil {
		return "", "", err
	}
	contentType := http.DetectContentType(data)
	if contentType == "application/octet-stream" {
		contentType = "image/png"
	}

	r2Client, err := NewR2ClientFromConfig()
	if err != nil {
		return "", "", err
	}
	if r2Client == nil {
		return "data:" + contentType + ";base64," + result.Base64Data, "", nil
	}

	var generation model.Generation
	if err := model.DB.First(&generation, generationID).Error; err != nil {
		return "", "", err
	}
	key := BuildR2Key(&generation)
	if err := r2Client.Upload(key, data, contentType); err != nil {
		return "", "", err
	}
	url, err := r2Client.GeneratePresignedURL(key, time.Hour)
	if err != nil {
		return "", "", err
	}
	return url, key, nil
}

func RefreshImageURL(generation *model.Generation) (string, error) {
	if generation.R2Key == "" {
		return generation.ImageURL, nil
	}
	r2Client, err := NewR2ClientFromConfig()
	if err != nil {
		return "", err
	}
	if r2Client == nil {
		return generation.ImageURL, nil
	}
	return r2Client.GeneratePresignedURL(generation.R2Key, time.Hour)
}

func DeleteR2Object(key string) error {
	if key == "" {
		return nil
	}
	r2Client, err := NewR2ClientFromConfig()
	if err != nil {
		return err
	}
	if r2Client == nil {
		return nil
	}
	return r2Client.Delete(key)
}

func PromoteUserGenerationsToPaid(userID int64) error {
	var generations []model.Generation
	if err := model.DB.Where("user_id = ? AND r2_key LIKE ?", userID, "generations/free/%").Find(&generations).Error; err != nil {
		return err
	}
	if len(generations) == 0 {
		return nil
	}
	r2Client, err := NewR2ClientFromConfig()
	if err != nil {
		return err
	}
	for _, generation := range generations {
		paidKey := strings.Replace(generation.R2Key, "generations/free/", "generations/paid/", 1)
		if paidKey == generation.R2Key {
			continue
		}
		if r2Client != nil {
			if err := r2Client.Copy(generation.R2Key, paidKey); err != nil {
				return err
			}
			if err := r2Client.Delete(generation.R2Key); err != nil {
				return err
			}
		}
		if err := model.DB.Model(&model.Generation{}).Where("id = ?", generation.ID).Update("r2_key", paidKey).Error; err != nil {
			return err
		}
	}
	return nil
}

func BuildR2Key(generation *model.Generation) string {
	owner := "anon"
	if generation.UserID != nil {
		owner = fmt.Sprintf("user-%d", *generation.UserID)
	} else if generation.AnonymousID != "" {
		owner = "anon-" + sanitizeKeyPart(generation.AnonymousID)
	}
	month := generation.CreatedAt.Format("2006-01")
	if month == "0001-01" {
		month = time.Now().Format("2006-01")
	}
	tier := "free"
	if generation.UserID != nil && userHasPaidCredit(*generation.UserID) {
		tier = "paid"
	}
	return fmt.Sprintf("generations/%s/%s/%s-%d.png", tier, month, owner, generation.ID)
}

func sanitizeKeyPart(value string) string {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "-")
	return replacer.Replace(value)
}

func userHasPaidCredit(userID int64) bool {
	if model.DB == nil {
		return false
	}
	var count int64
	if err := model.DB.Model(&model.CreditLog{}).Where("user_id = ? AND type IN ?", userID, []int{3, CreditLogTypePaymentTopup}).Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}
