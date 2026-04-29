package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
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
	return fmt.Sprintf("generations/%s/%s/%d.png", owner, month, generation.ID)
}

func sanitizeKeyPart(value string) string {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "-")
	return replacer.Replace(value)
}
