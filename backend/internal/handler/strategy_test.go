package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"warden/internal/dto/response"
	"warden/internal/handler"
	"warden/internal/middleware"
	"warden/internal/mock"
)

// TestStrategyHandler_RunScreen 校验粗筛端点：注入 userID、调 Service、返回 taskId。
func TestStrategyHandler_RunScreen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockStrategyService(ctrl)
	svc.EXPECT().RunScreen(gomock.Any(), uint(1), uint(7), gomock.Any()).Return(uint(99), nil)

	h := handler.NewStrategyHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint(1))
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/strategies/7/screen",
		strings.NewReader(`{"universe":{"type":"codes","codes":["600519"]}}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.RunScreen(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			TaskID string `json:"taskId"`
		} `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Code)
	assert.Equal(t, "99", body.Data.TaskID)
}

// TestStrategyHandler_Catalog 返回因子目录。
func TestStrategyHandler_Catalog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockStrategyService(ctrl)
	svc.EXPECT().Catalog().Return(nil)

	h := handler.NewStrategyHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/strategies/indicators/catalog", nil)

	h.Catalog(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestStrategyHandler_ScreenResult 返回粗筛结果。
func TestStrategyHandler_ScreenResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockStrategyService(ctrl)
	svc.EXPECT().GetScreenResult(gomock.Any(), uint(1), uint(7), uint(99)).
		Return(&response.ScreenResult{ID: 99, StrategyID: 7, Status: 2, MatchedCount: 1}, nil)

	h := handler.NewStrategyHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint(1))
	c.Params = gin.Params{{Key: "id", Value: "7"}, {Key: "taskId", Value: "99"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/strategies/7/screen/99", nil)

	h.ScreenResult(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
