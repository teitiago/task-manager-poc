//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestTraceRequestIDMiddleware Validates that the context will receive the x-request-id
func TestTraceRequestIDMiddleware(t *testing.T) {
	router := gin.New()
	router.GET("/", RequestIDMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, c.MustGet(RequestID).(string))
	})

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Body.String() == "" {
		t.Errorf("expected request id response")
	}
}

// TestResponseIDMiddleware Validates that the response will have the request id
func TestResponseIDMiddleware(t *testing.T) {
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}, ResponseIDMiddleware())

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	responseHeader := w.Header().Get(RequestID)

	if responseHeader == "" {
		t.Errorf("expected request id response")
	}
}

// TestIntegrateRequestResponseMiddleware Validates that both request and response keep the same request id
func TestIntegrateRequestResponseMiddleware(t *testing.T) {

	router := gin.New()
	router.GET("/", RequestIDMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, c.MustGet(RequestID).(string))
	}, ResponseIDMiddleware())

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	responseHeader := w.Header().Get(RequestID)

	if responseHeader != w.Body.String() {
		t.Errorf("request and response middleware mismatch, %v, %v", responseHeader, w.Body.String())
	}

}
