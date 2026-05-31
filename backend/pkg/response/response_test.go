package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"warden/pkg/errcode"
	"warden/pkg/response"
)

func init() { gin.SetMode(gin.TestMode) }

func decodeBody(t *testing.T, b []byte) response.Response {
	t.Helper()
	var r response.Response
	assert.NoError(t, json.Unmarshal(b, &r))
	return r
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	response.Success(c, gin.H{"hello": "world"})

	assert.Equal(t, http.StatusOK, w.Code)
	r := decodeBody(t, w.Body.Bytes())
	assert.Equal(t, 0, r.Code)
	assert.Equal(t, "ok", r.Message)
	assert.NotNil(t, r.Data)
}

func TestPageSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	response.PageSuccess(c, []string{"a", "b"}, 2, 1, 20)

	assert.Equal(t, http.StatusOK, w.Code)
	r := decodeBody(t, w.Body.Bytes())
	assert.Equal(t, 0, r.Code)
	data := r.Data.(map[string]interface{})
	assert.EqualValues(t, 2, data["total"])
	assert.EqualValues(t, 1, data["page"])
	assert.EqualValues(t, 20, data["size"])
}

func TestFail_AppError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	response.Fail(c, errcode.ErrNotFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
	r := decodeBody(t, w.Body.Bytes())
	assert.Equal(t, errcode.ErrNotFound.Code, r.Code)
	assert.Equal(t, errcode.ErrNotFound.Message, r.Message)
}

func TestFail_GenericError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	response.Fail(c, errors.New("boom"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	r := decodeBody(t, w.Body.Bytes())
	assert.Equal(t, errcode.ErrInternal.Code, r.Code)
}

func TestFail_WrappedAppError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	response.Fail(c, errcode.ErrInvalidParam.Wrap(errors.New("field x")))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	r := decodeBody(t, w.Body.Bytes())
	assert.Equal(t, errcode.ErrInvalidParam.Code, r.Code)
}
