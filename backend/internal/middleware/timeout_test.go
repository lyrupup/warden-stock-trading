package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestResolveTimeout(t *testing.T) {
	def := 10 * time.Second
	override := map[string]time.Duration{
		"/api/strategies/screen/preview": 30 * time.Second,
		"/api/strategies/:id/screen":     60 * time.Second,
	}

	cases := []struct {
		name string
		full string
		want time.Duration
	}{
		{"空 fullPath 退默认", "", def},
		{"无匹配退默认", "/api/market/indices", def},
		{"精确匹配 preview", "/api/strategies/screen/preview", 30 * time.Second},
		{"前缀匹配 :id/screen", "/api/strategies/:id/screen", 60 * time.Second},
		{"前缀匹配 :id/screen/latest", "/api/strategies/:id/screen/latest", 60 * time.Second},
		{"前缀匹配 :id/screen/:taskId", "/api/strategies/:id/screen/:taskId", 60 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveTimeout(tc.full, def, override)
			if got != tc.want {
				t.Fatalf("resolveTimeout(%q) = %v, want %v", tc.full, got, tc.want)
			}
		})
	}
}

func TestTimeoutMiddleware_InjectsDeadline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(Timeout(50*time.Millisecond, map[string]time.Duration{
		"/slow": 500 * time.Millisecond,
	}))
	r.GET("/fast", func(c *gin.Context) {
		dl, ok := c.Request.Context().Deadline()
		if !ok {
			t.Fatalf("/fast 缺失 deadline")
		}
		if time.Until(dl) > 60*time.Millisecond {
			t.Fatalf("/fast 应使用默认 50ms，但剩余 %v", time.Until(dl))
		}
		c.Status(http.StatusOK)
	})
	r.GET("/slow", func(c *gin.Context) {
		dl, ok := c.Request.Context().Deadline()
		if !ok {
			t.Fatalf("/slow 缺失 deadline")
		}
		if time.Until(dl) < 400*time.Millisecond {
			t.Fatalf("/slow 应使用覆盖 500ms，但剩余 %v", time.Until(dl))
		}
		c.Status(http.StatusOK)
	})

	for _, path := range []string{"/fast", "/slow"} {
		req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(context.Background())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, w.Code)
		}
	}
}
