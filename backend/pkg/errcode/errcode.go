// Package errcode 定义统一错误码与应用错误类型。
//
// 错误码段位（见 BACKEND.md §1.4）：
//
//	10xxx 通用    20xxx 行情    21xxx 策略    22xxx 持仓
//	23xxx 风控    24xxx 任务    25xxx AI      26xxx 配置    40xxx 鉴权
package errcode

import (
	"fmt"
	"net/http"
)

// AppError 是贯穿各层的业务错误类型，携带业务码、提示信息与建议的 HTTP 状态码。
// Service/Repository 返回 *AppError，Handler 经 response.Fail 统一转换为响应体。
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("errcode: code=%d message=%s", e.Code, e.Message)
}

// WithMessage 返回一个保留错误码、替换提示信息的副本，避免污染包级单例。
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{Code: e.Code, Message: msg, HTTPStatus: e.HTTPStatus}
}

// Wrap 在保留错误码的前提下，把底层错误信息拼接进提示，便于排查。
func (e *AppError) Wrap(err error) *AppError {
	if err == nil {
		return e
	}
	return &AppError{Code: e.Code, Message: fmt.Sprintf("%s: %v", e.Message, err), HTTPStatus: e.HTTPStatus}
}

// New 创建一个新的应用错误。
func New(code int, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

// 通用错误码 10xxx。
var (
	ErrInternal     = New(10001, "服务器内部错误", http.StatusInternalServerError)
	ErrInvalidParam = New(10002, "参数错误", http.StatusBadRequest)
	ErrNotFound     = New(10003, "资源不存在", http.StatusNotFound)
	ErrTimeout      = New(10004, "请求超时", http.StatusGatewayTimeout)
	ErrTooManyReq   = New(10005, "请求过于频繁", http.StatusTooManyRequests)
)

// 行情错误码 20xxx。
var (
	ErrMarketProvider  = New(20001, "行情数据源异常", http.StatusBadGateway)
	ErrWatchlistExists = New(20002, "该股票已在自选列表中", http.StatusConflict)
	ErrStockNotFound   = New(20003, "未找到该股票", http.StatusNotFound)
)

// 鉴权错误码 40xxx。
var (
	ErrUnauthorized = New(40001, "未登录或登录已过期", http.StatusUnauthorized)
	ErrForbidden    = New(40003, "无权访问该资源", http.StatusForbidden)
)
