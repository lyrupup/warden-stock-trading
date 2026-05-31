// Package response 提供统一响应封装（见 BACKEND.md §1.4）。
//
//	{ "code": 0, "message": "ok", "data": ... }
//
// code=0 表示成功；分页 data={list,total,page,size}。
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"warden/pkg/errcode"
)

// Response 统一响应体。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据结构。
type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// Success 返回成功响应（code=0）。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// PageSuccess 返回分页成功响应。
func PageSuccess(c *gin.Context, list interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    PageData{List: list, Total: total, Page: page, Size: size},
	})
}

// Fail 返回失败响应。识别 *errcode.AppError 取业务码与 HTTP 状态码，
// 其余错误统一按内部错误处理，避免泄露内部细节。
func Fail(c *gin.Context, err error) {
	var appErr *errcode.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, Response{Code: appErr.Code, Message: appErr.Message})
		return
	}
	c.JSON(errcode.ErrInternal.HTTPStatus, Response{
		Code:    errcode.ErrInternal.Code,
		Message: errcode.ErrInternal.Message,
	})
}
