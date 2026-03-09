package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"auth-service/internal/transport/ratelimit"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

func RateLimitMiddleware(limiter *ratelimit.RouteLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.Allow(ip, r.URL.Path) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
			r.Header.Set("X-Request-Id", reqID)
		}
		rw.Header().Set("X-Request-Id", reqID)

		next.ServeHTTP(rw, r)

		span := trace.SpanFromContext(r.Context())
		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration", time.Since(start),
			"status", rw.statusCode,
			"trace_id", traceID,
			"request_id", reqID,
		}

		msg := "HTTP Request"

		if rw.statusCode >= 400 && len(rw.body) > 0 {
			var errResp struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(rw.body, &errResp); err == nil && errResp.Message != "" {
				msg = errResp.Message
			}
		}

		if rw.statusCode >= 500 {
			slog.Error(msg, attrs...)
		} else if rw.statusCode >= 400 {
			slog.Warn(msg, attrs...)
		} else {
			slog.Info(msg, attrs...)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode >= 400 {
		if len(rw.body) < 1024 {
			rw.body = append(rw.body, b...)
		}
	}
	return rw.ResponseWriter.Write(b)
}
