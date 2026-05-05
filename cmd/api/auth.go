package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gr33n-api/internal/authctx"
	"gr33n-api/internal/httputil"
)

// package-level vars set by main() before registerRoutes()
var (
	piAPIKey   string
	jwtSecret  []byte
	corsOrigin string
	authMode   string // "dev" | "auth_test" | "production"
	authDebug  bool   // AUTH_DEBUG_LOG=true: log auth rejection reasons (no secrets, no JWT body)
)

type contextKey string

const claimsKey contextKey = "claims"

// isDevAuthBypass reports whether auth bypass is active.
// Requires BOTH: binary compiled with `-tags dev` AND AUTH_MODE=dev at runtime.
func isDevAuthBypass() bool {
	return devBypassAllowed && authMode == "dev"
}

// ── API Key middleware (Pi → API) ────────────────────────────────────────────
// Protects POST /sensors/{id}/readings, POST /sensors/readings/batch,
// PATCH /devices/{id}/status, and other Pi-only routes.
// Pi sends:  X-API-Key: <PI_API_KEY>
func requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isDevAuthBypass() {
			ctx := authctx.WithFarmAuthzSkip(r.Context(), true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		key := r.Header.Get("X-API-Key")
		if key == "" {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "missing_x_api_key", "path", r.URL.Path)
			}
			httputil.WriteError(w, http.StatusUnauthorized, "X-API-Key required")
			return
		}
		if key != piAPIKey {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "invalid_x_api_key", "path", r.URL.Path)
			}
			httputil.WriteError(w, http.StatusForbidden, "invalid API key")
			return
		}
		// Same trust boundary as requireJWTOrPiEdge: a valid PI_API_KEY may
		// call farm-scoped handlers (e.g. RecordEvent → RequireFarmMemberOrPiEdge)
		// without a dashboard JWT.
		next.ServeHTTP(w, r.WithContext(authctx.WithPiEdgeAuth(r.Context())))
	})
}

// requireJWTOrPiEdge allows dashboard JWT or the shared Pi API key (no JWT).
// Used for GET /farms/{id}/devices so edge gateways can poll pending_command in production.
func requireJWTOrPiEdge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isDevAuthBypass() {
			ctx := authctx.WithFarmAuthzSkip(r.Context(), true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		key := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if key != "" {
			if key != piAPIKey {
				if authDebug {
					slog.Warn("auth_rejected", "reason", "invalid_x_api_key", "path", r.URL.Path)
				}
				httputil.WriteError(w, http.StatusForbidden, "invalid API key")
				return
			}
			next.ServeHTTP(w, r.WithContext(authctx.WithPiEdgeAuth(r.Context())))
			return
		}
		requireJWT(next).ServeHTTP(w, r)
	})
}

// ── JWT middleware (Dashboard → API) ─────────────────────────────────────────
// Protects all dashboard routes.
// Vue sends:  Authorization: Bearer <token>
func requireJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isDevAuthBypass() {
			ctx := authctx.WithFarmAuthzSkip(r.Context(), true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		var tokenStr string
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		} else if q := r.URL.Query().Get("token"); q != "" {
			tokenStr = q
		} else {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "missing_bearer_or_query_token", "path", r.URL.Path)
			}
			httputil.WriteError(w, http.StatusUnauthorized, "Authorization: Bearer <token> required")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		}, jwt.WithExpirationRequired())

		if err != nil || !token.Valid {
			if authDebug {
				if err != nil {
					slog.Warn("auth_rejected", "reason", "jwt_invalid", "path", r.URL.Path, "err", err.Error())
				} else {
					slog.Warn("auth_rejected", "reason", "jwt_invalid", "path", r.URL.Path)
				}
			}
			httputil.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, token.Claims)
		if mc, ok := token.Claims.(jwt.MapClaims); ok {
			if uidStr, exists := mc["user_id"]; exists {
				if s, ok := uidStr.(string); ok {
					if uid, err := uuid.Parse(s); err == nil {
						ctx = authctx.WithUserID(ctx, uid)
					}
				}
			}
			if email, exists := mc["email"]; exists {
				if s, ok := email.(string); ok {
					ctx = authctx.WithEmail(ctx, s)
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IssueToken is called by the login handler to mint a signed JWT.
// exp: duration until expiry (e.g. 24 * time.Hour)
func IssueToken(username string, exp time.Duration, extra map[string]any) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(exp).Unix(),
	}
	for k, v := range extra {
		claims[k] = v
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}
