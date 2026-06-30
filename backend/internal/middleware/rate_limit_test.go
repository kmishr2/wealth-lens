package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitRejectsRequestsAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	router.Use(RateLimit(2, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for requestNumber, expectedStatus := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != expectedStatus {
			t.Fatalf("request %d status = %d, want %d", requestNumber+1, response.Code, expectedStatus)
		}
		if requestNumber == 2 && response.Header().Get("Retry-After") == "" {
			t.Fatal("limited response missing Retry-After")
		}
	}
}

func TestRateLimitSeparatesClientIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(RateLimit(1, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, address := range []string{"192.0.2.1:1000", "192.0.2.2:1000"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = address
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != http.StatusNoContent {
			t.Fatalf("client %s status = %d", address, response.Code)
		}
	}
}
