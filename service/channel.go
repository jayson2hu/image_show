package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jayson2hu/image-show/config"
	"github.com/jayson2hu/image-show/model"
)

const envFallbackChannelName = "env:SUB2API_BASE_URL"

type ChannelUse struct {
	ID   *int64
	Name string
}

type ChannelError struct {
	Channel ChannelUse
	Err     error
}

func (e ChannelError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e ChannelError) Unwrap() error {
	return e.Err
}

func SelectChannels() ([]model.Channel, error) {
	var channels []model.Channel
	if model.DB != nil {
		if err := model.DB.Where("status = ?", 1).Find(&channels).Error; err != nil {
			return nil, err
		}
	}
	if len(channels) == 0 && config.AppConfig != nil && config.AppConfig.Sub2APIBaseURL != "" {
		channels = append(channels, model.Channel{
			Name:    envFallbackChannelName,
			BaseURL: config.AppConfig.Sub2APIBaseURL,
			Status:  1,
			Weight:  1,
		})
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("no enabled sub2api channels")
	}
	return weightedShuffle(channels), nil
}

func GenerateImageViaChannels(prompt, quality, size, userIP string, options ImageOptions) (*ImageGenerationResult, error) {
	if config.AppConfig != nil && config.AppConfig.MockSub2API {
		result, err := NewSub2APIClient("http://mock", "", nil).GenerateImage(prompt, quality, size, userIP, options)
		if result != nil {
			result.Channel = ChannelUse{Name: "mock"}
		}
		return result, err
	}

	channels, err := SelectChannels()
	if err != nil {
		return nil, err
	}
	var lastErr error
	var lastChannel ChannelUse
	for _, channel := range channels {
		headers := map[string]string{}
		if channel.Headers != "" {
			_ = json.Unmarshal([]byte(channel.Headers), &headers)
		}
		client := NewSub2APIClient(channel.BaseURL, channel.APIKey, headers)
		result, err := client.GenerateImage(prompt, quality, size, userIP, options)
		if err == nil {
			result.Channel = channelUse(channel)
			return result, nil
		}
		lastErr = err
		lastChannel = channelUse(channel)
	}
	return nil, withChannelError(lastChannel, lastErr)
}

func EditImageViaChannels(prompt, quality, size, userIP string, imageData []byte, filename, contentType string, options ImageOptions) (*ImageGenerationResult, error) {
	if config.AppConfig != nil && config.AppConfig.MockSub2API {
		result, err := NewSub2APIClient("http://mock", "", nil).EditImage(prompt, quality, size, userIP, imageData, filename, contentType, options)
		if result != nil {
			result.Channel = ChannelUse{Name: "mock"}
		}
		return result, err
	}

	channels, err := SelectChannels()
	if err != nil {
		return nil, err
	}
	var lastErr error
	var lastChannel ChannelUse
	for _, channel := range channels {
		headers := map[string]string{}
		if channel.Headers != "" {
			_ = json.Unmarshal([]byte(channel.Headers), &headers)
		}
		client := NewSub2APIClient(channel.BaseURL, channel.APIKey, headers)
		result, err := client.EditImage(prompt, quality, size, userIP, imageData, filename, contentType, options)
		if err == nil {
			result.Channel = channelUse(channel)
			return result, nil
		}
		lastErr = err
		lastChannel = channelUse(channel)
	}
	return nil, withChannelError(lastChannel, lastErr)
}

func withChannelError(channel ChannelUse, err error) error {
	if err == nil {
		return nil
	}
	var channelErr ChannelError
	if errors.As(err, &channelErr) {
		return err
	}
	return ChannelError{Channel: channel, Err: err}
}

func channelUse(channel model.Channel) ChannelUse {
	if channel.ID == 0 {
		return ChannelUse{Name: channel.Name}
	}
	id := channel.ID
	return ChannelUse{ID: &id, Name: channel.Name}
}

func weightedShuffle(channels []model.Channel) []model.Channel {
	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	pool := append([]model.Channel(nil), channels...)
	result := make([]model.Channel, 0, len(pool))
	for len(pool) > 0 {
		total := 0
		for _, channel := range pool {
			if channel.Weight > 0 {
				total += channel.Weight
			}
		}
		if total == 0 {
			result = append(result, pool...)
			break
		}
		pick := source.Intn(total)
		acc := 0
		index := 0
		for i, channel := range pool {
			if channel.Weight <= 0 {
				continue
			}
			acc += channel.Weight
			if pick < acc {
				index = i
				break
			}
		}
		result = append(result, pool[index])
		pool = append(pool[:index], pool[index+1:]...)
	}
	return result
}
