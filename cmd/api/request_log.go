package main

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gr33n-api/internal/authctx"
)

// captureWriter records HTTP status for access logs and preserves optional
// interfaces (e.g. http.Flusher for SSE).
type captureWriter struct {
	http.ResponseWriter
	status int
}

func (c *captureWriter) WriteHeader(code int) {
	if c.status == 0 {
		c.status = code
	}
	c.ResponseWriter.WriteHeader(code)
}

func (c *captureWriter) Write(b []byte) (int, error) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	return c.ResponseWriter.Write(b)
}

func (c *captureWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func requestIDFor(r *http.Request, w http.ResponseWriter) string {
	if id := strings.TrimSpace(r.Header.Get("X-Request-ID")); id != "" {
		w.Header().Set("X-Request-ID", id)
		return id
	}
	id := uuid.New().String()
	w.Header().Set("X-Request-ID", id)
	return id
}

func farmIDFromPath(path string) string {
	const pfx = "/farms/"
	if !strings.HasPrefix(path, pfx) {
		return ""
	}
	rest := strings.TrimPrefix(path, pfx)
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 {
		return ""
	}
	return rest[:i]
}

// withRequestLog wraps a handler and emits one structured log line per request
// after completion (method, path, status, duration, optional farm_id and user).
func withRequestLog(authKind string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rid := requestIDFor(r, w)
		cw := &captureWriter{ResponseWriter: w}
		next.ServeHTTP(cw, r)
		status := cw.status
		if status == 0 {
			status = http.StatusOK
		}
		ms := time.Since(start).Milliseconds()

		attrs := []any{
			"request_id", rid,
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", ms,
			"auth", authKind,
		}
		if fid := farmIDFromPath(r.URL.Path); fid != "" {
			attrs = append(attrs, "farm_id", fid)
		}
		if uid, ok := authctx.UserID(r.Context()); ok {
			attrs = append(attrs, "user_id", uid.String())
		}
		if authctx.PiEdgeAuth(r.Context()) {
			attrs = append(attrs, "pi_edge", true)
		}

		slog.Info("request", attrs...)
	})
}
