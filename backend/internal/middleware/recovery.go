// Package middleware 提供 Gin 中间件。装配顺序见 BACKEND.md §7.3：
// Recovery → Logger → CORS → RateLimit → Timeout(context) → Auth → 业务路由。
package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"warden/pkg/errcode"
	"warden/pkg/response"
)

// Recovery 捕获 panic，记录堆栈并返回统一的内部错误响应，避免进程崩溃。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()),
				)
				response.Fail(c, errcode.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
}
