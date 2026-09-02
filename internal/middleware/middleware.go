package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered", "error", recovered, "path", r.URL.Path)
				writeAPIError(w, r, http.StatusInternalServerError, "internal_server_error", "Error interno del servidor")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", remoteIP(r),
		)
	})
}

func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(allowedOrigin)
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Type, If-None-Match, X-API-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders(environment string) func(http.Handler) http.Handler {
	production := strings.EqualFold(strings.TrimSpace(environment), "production")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			if production {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireHTTPS(environment string) func(http.Handler) http.Handler {
	production := strings.EqualFold(strings.TrimSpace(environment), "production")
	return func(next http.Handler) http.Handler {
		if !production {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			forwarded := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]))
			if r.TLS == nil && forwarded != "https" {
				writeAPIError(w, r, http.StatusUpgradeRequired, "https_required", "HTTPS es obligatorio en producción")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type rateEntry struct {
	window time.Time
	count  int
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	entries map[string]rateEntry
}

func RateLimit(perMinute int, apiKey string) func(http.Handler) http.Handler {
	if perMinute < 1 {
		perMinute = 60
	}
	limiter := &rateLimiter{limit: perMinute, entries: make(map[string]rateEntry)}
	configuredKey := strings.TrimSpace(apiKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions || (configuredKey != "" && constantTimeEqual(r.Header.Get("X-API-Key"), configuredKey)) {
				next.ServeHTTP(w, r)
				return
			}
			allowed, retryAfter := limiter.allow(remoteIP(r), time.Now())
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				writeAPIError(w, r, http.StatusTooManyRequests, "rate_limited", "Límite de solicitudes excedido")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *rateLimiter) allow(key string, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, exists := l.entries[key]
	if !exists || now.Sub(entry.window) >= time.Minute {
		l.entries[key] = rateEntry{window: now, count: 1}
		return true, 0
	}
	if entry.count >= l.limit {
		retry := int(time.Until(entry.window.Add(time.Minute)).Seconds())
		if retry < 1 {
			retry = 1
		}
		return false, retry
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}

func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	type handlerResult struct {
		panicValue any
		panicked   bool
	}

	return func(next http.Handler) http.Handler {
		if timeout <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			recorder := httptest.NewRecorder()
			result := make(chan handlerResult, 1)
			go func() {
				res := handlerResult{}
				defer func() {
					if recovered := recover(); recovered != nil {
						res.panicValue = recovered
						res.panicked = true
					}
					result <- res
				}()
				next.ServeHTTP(recorder, r.WithContext(ctx))
			}()
			select {
			case res := <-result:
				if res.panicked {
					panic(res.panicValue)
				}
				copyHeaders(w.Header(), recorder.Header())
				w.WriteHeader(recorder.Code)
				_, _ = io.Copy(w, recorder.Body)
			case <-ctx.Done():
				writeAPIError(w, r, http.StatusGatewayTimeout, "timeout", "Tiempo de ejecución excedido")
			}
		})
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		dst.Del(key)
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

type MetricsRegistry struct {
	mu        sync.Mutex
	requests  uint64
	errors    uint64
	themes    map[string]uint64
	latencies []float64
}

func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{themes: make(map[string]uint64), latencies: make([]float64, 0, 1024)}
}

func (m *MetricsRegistry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		if r.URL.Path == "/metrics" {
			return
		}
		m.observe(r.URL.Path, recorder.status, time.Since(start).Seconds())
	})
}

func (m *MetricsRegistry) observe(path string, status int, latency float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests++
	if status >= 400 {
		m.errors++
	}
	if strings.HasPrefix(path, "/api/v1/theme/") && path != "/api/v1/theme/types" && path != "/api/v1/theme/custom" {
		typeValue := strings.TrimPrefix(path, "/api/v1/theme/")
		if !strings.Contains(typeValue, "/") && typeValue != "" {
			m.themes[typeValue]++
		}
	}
	m.latencies = append(m.latencies, latency)
	if len(m.latencies) > 2048 {
		copy(m.latencies, m.latencies[len(m.latencies)-1024:])
		m.latencies = m.latencies[:1024]
	}
}

func (m *MetricsRegistry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "Método no permitido")
			return
		}
		m.mu.Lock()
		requests, errorsCount := m.requests, m.errors
		themes := make(map[string]uint64, len(m.themes))
		for key, value := range m.themes {
			themes[key] = value
		}
		latencies := append([]float64(nil), m.latencies...)
		m.mu.Unlock()

		errorRate := 0.0
		if requests > 0 {
			errorRate = float64(errorsCount) / float64(requests)
		}
		sort.Float64s(latencies)
		p95 := 0.0
		if len(latencies) > 0 {
			index := int(float64(len(latencies)-1) * 0.95)
			p95 = latencies[index]
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, "# HELP eyex_http_requests_total Total de solicitudes HTTP observadas.\n# TYPE eyex_http_requests_total counter\neyex_http_requests_total %d\n", requests)
		fmt.Fprintf(w, "# HELP eyex_http_errors_total Total de respuestas HTTP con status >= 400.\n# TYPE eyex_http_errors_total counter\neyex_http_errors_total %d\n", errorsCount)
		fmt.Fprintf(w, "# HELP eyex_http_error_rate Fraccion acumulada de solicitudes con error.\n# TYPE eyex_http_error_rate gauge\neyex_http_error_rate %.6f\n", errorRate)
		keys := make([]string, 0, len(themes))
		for key := range themes {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Fprintln(w, "# HELP eyex_theme_requests_total Solicitudes por tipo de tema.")
		fmt.Fprintln(w, "# TYPE eyex_theme_requests_total counter")
		for _, key := range keys {
			fmt.Fprintf(w, "eyex_theme_requests_total{type=%q} %d\n", key, themes[key])
		}
		fmt.Fprintln(w, "# HELP eyex_request_latency_p95_seconds Percentil 95 de latencia de requests observados.")
		fmt.Fprintln(w, "# TYPE eyex_request_latency_p95_seconds gauge")
		fmt.Fprintf(w, "eyex_request_latency_p95_seconds %.6f\n", p95)
	})
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) || len(b) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func remoteIP(r *http.Request) string {
	if host := strings.TrimSpace(r.RemoteAddr); host != "" {
		if addrPort, err := netip.ParseAddrPort(host); err == nil {
			return addrPort.Addr().String()
		}
		if addr, err := netip.ParseAddr(host); err == nil {
			return addr.String()
		}
		return host
	}
	return "unknown"
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, status int, code, spanish string) {
	message := spanish
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Accept-Language"))), "en") {
		translations := map[string]string{
			"Error interno del servidor":         "Internal server error",
			"HTTPS es obligatorio en producción": "HTTPS is required in production",
			"Límite de solicitudes excedido":     "Request limit exceeded",
			"Tiempo de ejecución excedido":       "Request execution timed out",
			"Método no permitido":                "Method not allowed",
		}
		if translated, ok := translations[spanish]; ok {
			message = translated
		}
	}
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
