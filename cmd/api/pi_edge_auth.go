package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/authctx"
	"gr33n-api/internal/authsecurity"
	"gr33n-api/internal/deviceapikey"

)

var (
	deviceKeyQ       db.Querier
	deviceKeyLimiter = deviceapikey.NewRateLimiter(120)
)

func initPiEdgeAuth(q db.Querier) {
	deviceKeyQ = q
}

func authenticatePiEdge(r *http.Request) (ctx context.Context, ok bool) {
	if isDevAuthBypass() {
		return authctx.WithFarmAuthzSkip(r.Context(), true), true
	}

	rawDevice := deviceapikey.ExtractFromRequest(
		r.Header.Get("X-Device-Key"),
		r.Header.Get("Authorization"),
	)
	if rawDevice != "" {
		devID, _, parsed := deviceapikey.Parse(rawDevice)
		if !parsed {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "invalid_device_key_format", "path", r.URL.Path)
			}
			return r.Context(), false
		}
		if deviceKeyQ == nil {
			return r.Context(), false
		}
		cctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		rows, err := deviceKeyQ.ListActiveDeviceAPIKeyHashesByDevice(cctx, devID)
		if err != nil || len(rows) == 0 {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "device_key_not_found", "path", r.URL.Path, "device_id", devID)
			}
			return r.Context(), false
		}
		var matchedID int64
		for _, row := range rows {
			if deviceapikey.Verify(rawDevice, row.KeyHash) {
				matchedID = row.ID
				break
			}
		}
		if matchedID == 0 {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "invalid_device_key", "path", r.URL.Path, "device_id", devID)
			}
			return r.Context(), false
		}
		if !deviceKeyLimiter.Allow(matchedID) {
			if authDebug {
				slog.Warn("auth_rejected", "reason", "device_key_rate_limited", "path", r.URL.Path, "key_id", matchedID)
			}
			return r.Context(), false
		}
		_ = deviceKeyQ.TouchDeviceAPIKeyLastUsed(cctx, matchedID)
		return authctx.WithDeviceKeyAuth(r.Context(), matchedID, devID), true
	}

	legacy := r.Header.Get("X-API-Key")
	if legacy == "" {
		if authDebug {
			slog.Warn("auth_rejected", "reason", "missing_edge_credentials", "path", r.URL.Path)
		}
		return r.Context(), false
	}
	if authsecurity.LegacyPiKeyDisabled() {
		if authDebug {
			slog.Warn("auth_rejected", "reason", "legacy_pi_key_disabled", "path", r.URL.Path)
		}
		return r.Context(), false
	}
	if legacy != piAPIKey {
		if authDebug {
			slog.Warn("auth_rejected", "reason", "invalid_x_api_key", "path", r.URL.Path)
		}
		return r.Context(), false
	}
	if authDebug {
		slog.Info("pi_edge_auth", "mode", "legacy_shared_key", "path", r.URL.Path)
	}
	return authctx.WithPiEdgeAuth(r.Context()), true
}
