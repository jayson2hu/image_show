package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
	"github.com/jayson2hu/image-show/model"
	"github.com/jayson2hu/image-show/service"
)

type memoryWindow struct {
	mu     sync.Mutex
	events map[string][]time.Time
}

var localRateLimit = &memoryWindow{events: make(map[string][]time.Time)}

func GenerationRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := common.GetRealIP(c)
		if !allowRate(c, "imgshow:rl:ip:"+ip, 20, time.Minute) {
			tooMany(c, 60)
			return
		}
		if userID, exists := c.Get("userID"); exists {
			if id, ok := userID.(int64); ok {
				if !allowRate(c, "imgshow:rl:user:"+strconv.FormatInt(id, 10), 10, time.Minute) {
					tooMany(c, 60)
					return
				}
			}
		}
		if !allowDailyBudget(c) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "daily budget exhausted"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func allowRate(c *gin.Context, key string, limit int, window time.Duration) bool {
	client := service.RedisClient()
	if client == nil {
		return localRateLimit.allow(key, limit, window)
	}
	ctx := c.Request.Context()
	now := time.Now().UnixMilli()
	min := now - window.Milliseconds()
	pipe := client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(min, 10))
	pipe.ZAdd(ctx, key, service.RedisZ(now))
	count := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return localRateLimit.allow(key, limit, window)
	}
	return count.Val() <= int64(limit)
}

func allowDailyBudget(c *gin.Context) bool {
	value := model.GetSettingValue("daily_budget", "")
	if value == "" || value == "0" {
		return true
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return true
	}
	key := "imgshow:budget:" + time.Now().Format("2006-01-02")
	client := service.RedisClient()
	if client == nil {
		return localRateLimit.allow(key, limit, 24*time.Hour)
	}
	ctx := c.Request.Context()
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return localRateLimit.allow(key, limit, 24*time.Hour)
	}
	_ = client.Expire(ctx, key, 24*time.Hour).Err()
	return count <= int64(limit)
}

func (m *memoryWindow) allow(key string, limit int, window time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-window)
	events := m.events[key]
	kept := events[:0]
	for _, event := range events {
		if event.After(cutoff) {
			kept = append(kept, event)
		}
	}
	kept = append(kept, now)
	m.events[key] = kept
	return len(kept) <= limit
}

func tooMany(c *gin.Context, retryAfter int) {
	c.Header("Retry-After", strconv.Itoa(retryAfter))
	c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
	c.Abort()
}

func ResetRateLimitForTest() {
	localRateLimit.mu.Lock()
	localRateLimit.events = make(map[string][]time.Time)
	localRateLimit.mu.Unlock()
}
