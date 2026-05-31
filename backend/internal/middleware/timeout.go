package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 为请求注入带超时的 context 并替换到 *http.Request 上，
// 使下游所有 DB 操作（WithContext）与外部调用都能感知超时/取消并及时返回
// （见 BACKEND.md §7.3 context 传播）。
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
