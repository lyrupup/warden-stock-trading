package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 为请求注入带超时的 context 并替换到 *http.Request 上，
// 使下游所有 DB 操作（WithContext）与外部调用都能感知超时/取消并及时返回
// （见 BACKEND.md §7.3 context 传播）。
//
// 不传 overrides 等价于全局统一超时；传入时按 gin 路由模式（c.FullPath()）
// 命中前缀的最长匹配生效，便于「粗筛/回测/AI 流式」等耗时接口按业务覆盖默认值。
func Timeout(def time.Duration, overrides ...map[string]time.Duration) gin.HandlerFunc {
	var table map[string]time.Duration
	if len(overrides) > 0 {
		table = overrides[0]
	}
	return func(c *gin.Context) {
		d := resolveTimeout(c.FullPath(), def, table)
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// resolveTimeout 选择最长前缀匹配的覆盖项，命中不到时退回默认值。
// 之所以用前缀匹配而非精确匹配：M2 后续会有 :id/screen, :id/screen/:taskId 等同族慢接口，
// 统一前缀 "/api/strategies/:id/screen" 即可一次覆盖整族，避免漏配。
func resolveTimeout(fullPath string, def time.Duration, table map[string]time.Duration) time.Duration {
	if fullPath == "" || len(table) == 0 {
		return def
	}
	best := def
	bestLen := -1
	for prefix, d := range table {
		if d <= 0 {
			continue
		}
		if prefix == fullPath || (len(fullPath) >= len(prefix) && fullPath[:len(prefix)] == prefix) {
			if len(prefix) > bestLen {
				best = d
				bestLen = len(prefix)
			}
		}
	}
	return best
}
