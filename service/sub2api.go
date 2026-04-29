package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jayson2hu/image-show/config"
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
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Quality string `json:"quality"`
	Size    string `json:"size"`
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
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Sub2APIClient) GenerateImage(prompt, quality, size, userIP string) (*ImageGenerationResult, error) {
	if config.AppConfig != nil && config.AppConfig.MockSub2API {
		return mockImageResult(), nil
	}

	body, err := json.Marshal(imageGenerationRequest{
		Model:   "gpt-image-1",
		Prompt:  prompt,
		Quality: quality,
		Size:    size,
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
		return nil, fmt.Errorf("sub2api status %d: %s", resp.StatusCode, string(payload))
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
