package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

type Sub2APIClient struct {
	BaseURL    string
	APIKey     string
	Headers    map[string]string
	HTTPClient *http.Client
}

type ImageGenerationResult struct {
	Base64Data string
	URL        string
}

type imageGenerationRequest struct {
	Model             string `json:"model"`
	Prompt            string `json:"prompt"`
	Quality           string `json:"quality"`
	Size              string `json:"size"`
	OutputFormat      string `json:"output_format,omitempty"`
	OutputCompression *int   `json:"output_compression,omitempty"`
	Background        string `json:"background,omitempty"`
}

type imageGenerationResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
}

func NewSub2APIClient(baseURL, apiKey string, headers map[string]string) *Sub2APIClient {
	return &Sub2APIClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Headers: headers,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				ForceAttemptHTTP2:   false,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

func (c *Sub2APIClient) GenerateImage(prompt, quality, size, userIP string, options ImageOptions) (*ImageGenerationResult, error) {
	if config.AppConfig != nil && config.AppConfig.MockSub2API {
		return mockImageResult(), nil
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		result, err := c.generateImageOnce(prompt, quality, size, userIP, options)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetryableSub2APIError(err) || attempt == 3 {
			break
		}
		time.Sleep(time.Duration(attempt) * 800 * time.Millisecond)
	}
	return nil, userFacingSub2APIError(lastErr)
}

func (c *Sub2APIClient) EditImage(prompt, quality, size, userIP string, imageData []byte, filename, contentType string, options ImageOptions) (*ImageGenerationResult, error) {
	if config.AppConfig != nil && config.AppConfig.MockSub2API {
		return mockImageResult(), nil
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		result, err := c.editImageOnce(prompt, quality, size, userIP, imageData, filename, contentType, options)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetryableSub2APIError(err) || attempt == 3 {
			break
		}
		time.Sleep(time.Duration(attempt) * 800 * time.Millisecond)
	}
	return nil, userFacingSub2APIError(lastErr)
}

func userFacingSub2APIError(err error) error {
	var statusErr sub2APIStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.statusCode {
		case cloudflareTimeoutStatus:
			return fmt.Errorf("sub2api upstream timeout: image generation exceeded upstream proxy timeout, please retry or switch channel")
		case http.StatusTooManyRequests:
			return fmt.Errorf("sub2api rate limited: please retry later or switch channel")
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return fmt.Errorf("sub2api upstream unavailable: please retry or switch channel")
		}
	}
	return err
}

func (c *Sub2APIClient) generateImageOnce(prompt, quality, size, userIP string, options ImageOptions) (*ImageGenerationResult, error) {
	body, err := json.Marshal(imageGenerationRequest{
		Model:             imageModel(),
		Prompt:            prompt,
		Quality:           quality,
		Size:              size,
		OutputFormat:      options.OutputFormat,
		OutputCompression: options.OutputCompression,
		Background:        options.Background,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	if userIP != "" {
		req.Header.Set("X-Real-IP", userIP)
		req.Header.Set("X-Forwarded-For", userIP)
	}
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, sub2APIStatusError{statusCode: resp.StatusCode, payload: strings.TrimSpace(string(payload))}
	}

	var parsed imageGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("sub2api returned empty image data")
	}
	return &ImageGenerationResult{
		Base64Data: parsed.Data[0].B64JSON,
		URL:        parsed.Data[0].URL,
	}, nil
}

func (c *Sub2APIClient) editImageOnce(prompt, quality, size, userIP string, imageData []byte, filename, contentType string, options ImageOptions) (*ImageGenerationResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("model", imageModel()); err != nil {
		return nil, err
	}
	if err := writer.WriteField("prompt", prompt); err != nil {
		return nil, err
	}
	if err := writer.WriteField("quality", quality); err != nil {
		return nil, err
	}
	if err := writer.WriteField("size", size); err != nil {
		return nil, err
	}
	if options.OutputFormat != "" {
		if err := writer.WriteField("output_format", options.OutputFormat); err != nil {
			return nil, err
		}
	}
	if options.Background != "" {
		if err := writer.WriteField("background", options.Background); err != nil {
			return nil, err
		}
	}
	if options.OutputCompression != nil {
		if err := writer.WriteField("output_compression", fmt.Sprintf("%d", *options.OutputCompression)); err != nil {
			return nil, err
		}
	}
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, escapeMultipartFilename(filename)))
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(imageData); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/images/edits", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	if userIP != "" {
		req.Header.Set("X-Real-IP", userIP)
		req.Header.Set("X-Forwarded-For", userIP)
	}
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, sub2APIStatusError{statusCode: resp.StatusCode, payload: strings.TrimSpace(string(payload))}
	}

	var parsed imageGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("sub2api returned empty image data")
	}
	return &ImageGenerationResult{
		Base64Data: parsed.Data[0].B64JSON,
		URL:        parsed.Data[0].URL,
	}, nil
}

func imageModel() string {
	fallback := "gpt-image-2"
	if config.AppConfig != nil && config.AppConfig.ImageModel != "" {
		fallback = config.AppConfig.ImageModel
	}
	return model.GetSettingValue("image_model", fallback)
}

func escapeMultipartFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "source.png"
	}
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(filename)
}

type sub2APIStatusError struct {
	statusCode int
	payload    string
}

const cloudflareTimeoutStatus = 524

func (e sub2APIStatusError) Error() string {
	if e.payload == "" {
		return fmt.Sprintf("sub2api status %d", e.statusCode)
	}
	return fmt.Sprintf("sub2api status %d: %s", e.statusCode, e.payload)
}

func isRetryableSub2APIError(err error) bool {
	if err == nil {
		return false
	}
	var statusErr sub2APIStatusError
	if errors.As(err, &statusErr) {
		return statusErr.statusCode == http.StatusTooManyRequests ||
			statusErr.statusCode == http.StatusBadGateway ||
			statusErr.statusCode == http.StatusServiceUnavailable ||
			statusErr.statusCode == http.StatusGatewayTimeout ||
			statusErr.statusCode == cloudflareTimeoutStatus
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unexpected eof") || strings.Contains(text, "connection reset") || strings.Contains(text, "timeout") || strings.Contains(text, "temporary")
}

func mockImageResult() *ImageGenerationResult {
	pixel := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
	return &ImageGenerationResult{Base64Data: base64.StdEncoding.EncodeToString(pixel)}
}
