package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger_WithBody(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(RequestLogger())
	engine.POST("/", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(body))
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"key":"value"}`, w.Body.String())
}

func TestRequestLogger_NilBody(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(RequestLogger())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
