package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimitAndAPIKeyBypass(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := RateLimit(1, "secret")(next)
	req := httptest.NewRequest(http.MethodGet, "http://example.test/x", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	if rr := httptest.NewRecorder(); func() int { handler.ServeHTTP(rr, req); return rr.Code }() != 200 {
		t.Fatal("first request should pass")
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != 429 || !strings.Contains(rr.Body.String(), "rate_limited") {
		t.Fatalf("expected 429, got %d %s", rr.Code, rr.Body.String())
	}
	req2 := httptest.NewRequest(http.MethodGet, "http://example.test/x", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	req2.Header.Set("X-API-Key", "secret")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != 200 {
		t.Fatalf("api key should bypass limiter, got %d", rr2.Code)
	}
}

func TestRequireHTTPSProduction(t *testing.T) {
	h := RequireHTTPS("production")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	r := httptest.NewRequest(http.MethodGet, "http://example.test", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != 426 {
		t.Fatalf("expected 426, got %d", rr.Code)
	}
	r2 := httptest.NewRequest(http.MethodGet, "http://example.test", nil)
	r2.Header.Set("X-Forwarded-Proto", "https")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, r2)
	if rr2.Code != 200 {
		t.Fatalf("forwarded https should pass, got %d", rr2.Code)
	}
}

func TestRequestTimeout(t *testing.T) {
	h := RequestTimeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { time.Sleep(40 * time.Millisecond); w.WriteHeader(200) }))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://example.test", nil))
	if rr.Code != 504 || !strings.Contains(rr.Body.String(), "timeout") {
		t.Fatalf("expected timeout, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestRequestTimeoutPropagatesPanicToRecover(t *testing.T) {
	h := Recover(RequestTimeout(time.Second)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://example.test", nil))
	if rr.Code != http.StatusInternalServerError || !strings.Contains(rr.Body.String(), "internal_server_error") {
		t.Fatalf("expected recovered 500, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestMetrics(t *testing.T) {
	registry := NewMetricsRegistry()
	h := registry.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(400) }))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://example.test/api/v1/theme/protanopia", nil))
	metrics := httptest.NewRecorder()
	registry.Handler().ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "http://example.test/metrics", nil))
	if !strings.Contains(metrics.Body.String(), "eyex_http_requests_total 1") || !strings.Contains(metrics.Body.String(), `eyex_theme_requests_total{type="protanopia"} 1`) {
		t.Fatalf("unexpected metrics: %s", metrics.Body.String())
	}
}

func TestOperationalErrorRespectsAcceptLanguage(t *testing.T) {
	h := RequireHTTPS("production")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "http://example.test/api/v1/theme/types", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUpgradeRequired || !strings.Contains(rr.Body.String(), "HTTPS is required in production") {
		t.Fatalf("unexpected localized operational error: %d %s", rr.Code, rr.Body.String())
	}
}
