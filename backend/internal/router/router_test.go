package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"warden/config"
	"warden/internal/handler"
	"warden/internal/router"
)

func init() { gin.SetMode(gin.TestMode) }

// TestRouter_New_NoRouteConflict 确保策略静态路由（indicators/catalog、templates、
// screen/preview）与参数路由（:id、:taskId）共存不冲突、不 panic。
func TestRouter_New_NoRouteConflict(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Env = "test"
	cfg.RateLimit.RPS = 100
	cfg.RateLimit.Burst = 200
	cfg.Timeout.Seconds = 5

	var r *gin.Engine
	assert.NotPanics(t, func() {
		r = router.New(cfg, router.Handlers{
			Market:   handler.NewMarketHandler(nil),
			Strategy: handler.NewStrategyHandler(nil),
		})
	})

	// 健康检查可用。
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
