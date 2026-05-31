package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"warden/pkg/errcode"
	"warden/pkg/response"
)

// tokenBucket 是一个简单的令牌桶限流器：以 refillPerSec 速率补充令牌，
// 桶容量为 capacity，每个请求消耗一个令牌。
type tokenBucket struct {
	mu           sync.Mutex
	capacity     float64
	tokens       float64
	refillPerSec float64
	last         time.Time
}

func newTokenBucket(rps, burst int) *tokenBucket {
	return &tokenBucket{
		capacity:     float64(burst),
		tokens:       float64(burst),
		refillPerSec: float64(rps),
		last:         time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	b.tokens += elapsed * b.refillPerSec
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimit 全局令牌桶限流。rps 为每秒补充速率，burst 为桶容量（突发上限）。
func RateLimit(rps, burst int) gin.HandlerFunc {
	bucket := newTokenBucket(rps, burst)
	return func(c *gin.Context) {
		if !bucket.allow() {
			response.Fail(c, errcode.ErrTooManyReq)
			c.Abort()
			return
		}
		c.Next()
	}
}
