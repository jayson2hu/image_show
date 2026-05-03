package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	cfg := r2SettingsFromConfig()
	if cfg.endpoint == "" || cfg.accessKey == "" || cfg.secretKey == "" || cfg.bucket == "" {
		return nil, nil
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.accessKey, cfg.secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.endpoint)
		o.UsePathStyle = true
	})
	return &R2Client{client: client, bucket: cfg.bucket, cdnURL: strings.TrimRight(cfg.publicURL, "/")}, nil
}

type r2Settings struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	publicURL string
}

func r2SettingsFromConfig() r2Settings {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	return r2Settings{
		endpoint:  model.GetSettingValue("r2_endpoint", cfg.R2Endpoint),
		accessKey: model.GetSettingValue("r2_access_key", cfg.R2AccessKey),
		secretKey: model.GetSettingValue("r2_secret_key", cfg.R2SecretKey),
		bucket:    model.GetSettingValue("r2_bucket", cfg.R2Bucket),
		publicURL: model.GetSettingValue("r2_public_url", cfg.R2PublicURL),
	}
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

func StoreGeneratedImage(generationID int64, result *ImageGenerationResult, targetSize string) (imageURL string, r2Key string, err error) {
	if result.URL != "" {
		if !ShouldResizeImage(targetSize) {
			return result.URL, "", nil
		}
		data, contentType, err := downloadImage(result.URL)
		if err != nil {
			return "", "", err
		}
		data, contentType, err = ResizeImageBytes(data, targetSize)
		if err != nil {
			return "", "", err
		}
		return storeImageBytes(generationID, data, contentType)
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
	if ShouldResizeImage(targetSize) {
		data, contentType, err = ResizeImageBytes(data, targetSize)
		if err != nil {
			return "", "", err
		}
	}

	return storeImageBytes(generationID, data, contentType)
}

func storeImageBytes(generationID int64, data []byte, contentType string) (imageURL string, r2Key string, err error) {
	r2Client, err := NewR2ClientFromConfig()
	if err != nil {
		return "", "", err
	}
	if r2Client == nil {
		return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), "", nil
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

func StoreSourceImage(generationID int64, data []byte, contentType string) (imageURL string, r2Key string, err error) {
	r2Client, err := NewR2ClientFromConfig()
	if err != nil {
		return "", "", err
	}
	if r2Client == nil {
		return "", "", nil
	}
	var generation model.Generation
	if err := model.DB.First(&generation, generationID).Error; err != nil {
		return "", "", err
	}
	key := BuildSourceR2Key(&generation, contentType)
	if err := r2Client.Upload(key, data, contentType); err != nil {
		return "", "", err
	}
	url, err := r2Client.GeneratePresignedURL(key, time.Hour)
	if err != nil {
		return "", "", err
	}
	return url, key, nil
}

func downloadImage(sourceURL string) ([]byte, string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(sourceURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("download generated image status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 25*1024*1024))
	if err != nil {
		return nil, "", err
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return data, contentType, nil
}

func ProviderImageSize(requested string) string {
	width, height, ok := ParseImageSize(requested)
	if !ok {
		return "1024x1024"
	}
	if width == 1536 && height == 1024 {
		return "1536x1024"
	}
	if width == 1024 && height == 1536 {
		return "1024x1536"
	}
	if height > width {
		return "1024x1536"
	}
	if width > height {
		return "1536x1024"
	}
	return "1024x1024"
}

func ShouldResizeImage(targetSize string) bool {
	return ProviderImageSize(targetSize) != targetSize
}

func ResizeImageBytes(data []byte, targetSize string) ([]byte, string, error) {
	width, height, ok := ParseImageSize(targetSize)
	if !ok {
		return data, http.DetectContentType(data), nil
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	dst := resizeNearest(src, width, height)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 88}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "image/jpeg", nil
}

func resizeNearest(src image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	for y := 0; y < height; y++ {
		srcY := bounds.Min.Y + y*srcHeight/height
		for x := 0; x < width; x++ {
			srcX := bounds.Min.X + x*srcWidth/width
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func ParseImageSize(size string) (int, int, bool) {
	parts := strings.Split(size, "x")
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
	return width, height, width > 0 && height > 0
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

func BuildSourceR2Key(generation *model.Generation, contentType string) string {
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
	return fmt.Sprintf("sources/%s/%s/%s-%d.%s", tier, month, owner, generation.ID, imageExtension(contentType))
}

func imageExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	default:
		return "png"
	}
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
