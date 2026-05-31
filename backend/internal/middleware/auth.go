package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"warden/pkg/errcode"
	"warden/pkg/response"
)

// ContextUserIDKey 是注入到 gin.Context 的用户 ID 键。
const ContextUserIDKey = "user_id"

// defaultUserID 单用户模式下无 token 时注入的默认用户（见 BACKEND.md §2）。
const defaultUserID uint = 1

// Auth 解析 JWT 并注入 user_id。
//
//	singleUserMode=true：无 token 时自动注入默认用户 id=1，保证 V1 可直接使用；
//	多用户上线时把开关关掉即可，无需改业务代码。
func Auth(secret string, singleUserMode bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c)
		if token == "" {
			if singleUserMode {
				c.Set(ContextUserIDKey, defaultUserID)
				c.Next()
				return
			}
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		uid, err := parseToken(token, secret)
		if err != nil {
			if singleUserMode {
				// 单用户模式下容错：token 非法时仍以默认用户放行。
				c.Set(ContextUserIDKey, defaultUserID)
				c.Next()
				return
			}
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, uid)
		c.Next()
	}
}

// GetUserID 从 context 取当前用户 ID（见 BACKEND.md §2）。
func GetUserID(c *gin.Context) uint {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0
	}
	id, _ := v.(uint)
	return id
}

func extractBearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(h)
}

func parseToken(tokenStr, secret string) (uint, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errcode.ErrUnauthorized
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	switch v := claims["user_id"].(type) {
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	default:
		return 0, errcode.ErrUnauthorized
	}
}
