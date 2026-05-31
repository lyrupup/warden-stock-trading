package errcode_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"warden/pkg/errcode"
)

func TestAppError_Error(t *testing.T) {
	e := errcode.New(12345, "测试错误", 400)
	assert.Contains(t, e.Error(), "12345")
	assert.Contains(t, e.Error(), "测试错误")
}

func TestAppError_WithMessage_PreservesCode(t *testing.T) {
	e := errcode.ErrInvalidParam.WithMessage("自定义提示")
	assert.Equal(t, errcode.ErrInvalidParam.Code, e.Code)
	assert.Equal(t, errcode.ErrInvalidParam.HTTPStatus, e.HTTPStatus)
	assert.Equal(t, "自定义提示", e.Message)
	// 不污染包级单例。
	assert.NotEqual(t, "自定义提示", errcode.ErrInvalidParam.Message)
}

func TestAppError_Wrap(t *testing.T) {
	base := errcode.ErrMarketProvider
	wrapped := base.Wrap(errors.New("connection refused"))
	assert.Equal(t, base.Code, wrapped.Code)
	assert.Contains(t, wrapped.Message, "connection refused")
	// Wrap(nil) 应返回自身（不改变 message）。
	same := errcode.ErrInternal.Wrap(nil)
	assert.Equal(t, errcode.ErrInternal.Message, same.Message)
}

func TestErrorIs_Identity(t *testing.T) {
	err := error(errcode.ErrNotFound)
	assert.ErrorIs(t, err, errcode.ErrNotFound)
	assert.NotErrorIs(t, err, errcode.ErrUnauthorized)
}

func TestSegments(t *testing.T) {
	// 段位校验（见 BACKEND.md §1.4）。
	assert.Equal(t, 1, errcode.ErrInternal.Code/10000)
	assert.Equal(t, 2, errcode.ErrMarketProvider.Code/10000)
	assert.Equal(t, 4, errcode.ErrUnauthorized.Code/10000)
}
