package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/common"
)

func TestRealIPHeaderPriority(t *testing.T) {
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.Header.Set("CF-Connecting-IP", "1.1.1.1")
	req.Header.Set("X-Real-IP", "2.2.2.2")
	req.Header.Set("X-Forwarded-For", "3.3.3.3, 4.4.4.4")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Body.String() != "1.1.1.1" {
		t.Fatalf("expected CF-Connecting-IP priority, got %q", rec.Body.String())
	}
}

func TestRealIPForwardedForFirstIP(t *testing.T) {
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.Header.Set("X-Forwarded-For", "3.3.3.3, 4.4.4.4")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Body.String() != "3.3.3.3" {
		t.Fatalf("expected first forwarded ip, got %q", rec.Body.String())
	}
}

func TestRealIPFallback(t *testing.T) {
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "5.5.5.5:1234"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Body.String() != "5.5.5.5" {
		t.Fatalf("expected remote addr fallback, got %q", rec.Body.String())
	}
}

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RealIP())
	r.GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, common.GetRealIP(c))
	})
	return r
}
