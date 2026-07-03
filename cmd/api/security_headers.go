package main

import (
	"net/http"
	"os"
	"strings"
)

// securityHeadersMiddleware adds baseline HTTP security headers on every response.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	hsts := strings.EqualFold(strings.TrimSpace(os.Getenv("SECURITY_HSTS_ENABLED")), "true")
	cspReportOnly := strings.TrimSpace(os.Getenv("SECURITY_CSP_REPORT_ONLY"))
	if cspReportOnly == "" {
		cspReportOnly = "default-src 'self'; frame-ancestors 'none'"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy-Report-Only", cspReportOnly)
		if hsts {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func wrapHTTPMiddleware(mux http.Handler) http.Handler {
	return corsMiddleware(securityHeadersMiddleware(mux))
}
