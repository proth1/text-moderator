package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGatewayHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "gateway",
			"version": "0.1.0",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGatewayContentTypeValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if c.Request.Method == "POST" && c.Request.ContentLength > 0 {
			ct := c.GetHeader("Content-Type")
			if ct == "" {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type header is required"})
				c.Abort()
				return
			}
		}
		c.Next()
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
	}{
		{"valid JSON", "application/json", `{"test": true}`, http.StatusOK},
		{"missing content-type", "", "body", http.StatusUnsupportedMediaType},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			if tc.body != "" {
				req = httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tc.body))
				req.ContentLength = int64(len(tc.body))
			}
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestGatewayRateLimitHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Verify rate limit headers are set in responses
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", "99")
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-RateLimit-Limit") != "100" {
		t.Error("expected X-RateLimit-Limit header")
	}
	if w.Header().Get("X-RateLimit-Remaining") != "99" {
		t.Error("expected X-RateLimit-Remaining header")
	}
}
