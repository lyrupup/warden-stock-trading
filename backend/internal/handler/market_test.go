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

	"warden/internal/dto/request"
	"warden/internal/handler"
	"warden/internal/middleware"
	"warden/internal/mock"
	"warden/internal/model"
)

func init() { gin.SetMode(gin.TestMode) }

// TestMarketHandler_AddWatchlist 校验 Handler 绑定参数、注入 userID 并组装统一响应。
func TestMarketHandler_AddWatchlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockMarketService(ctrl)
	svc.EXPECT().
		AddWatchlist(gomock.Any(), uint(1), gomock.AssignableToTypeOf(&request.AddWatchlistReq{})).
		Return(&model.WatchlistItem{BaseModel: model.BaseModel{ID: 9}, StockCode: "600519"}, nil)

	h := handler.NewMarketHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint(1))
	c.Request = httptest.NewRequest(http.MethodPost, "/api/market/watchlist",
		strings.NewReader(`{"stock_code":"600519"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.AddWatchlist(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.EqualValues(t, 0, body["code"])
}

// TestMarketHandler_Search_MissingKeyword 缺少 kw 返回参数错误。
func TestMarketHandler_Search_MissingKeyword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockMarketService(ctrl)
	h := handler.NewMarketHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/market/search", nil)

	h.Search(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
