package farmauthz

import (
	"net/http"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/httputil"
)

// RequirePiEdgeDeviceScope allows legacy shared PI_API_KEY or a per-device key
// that matches resourceDeviceID. Denies when a device key targets another device.
func RequirePiEdgeDeviceScope(w http.ResponseWriter, r *http.Request, resourceDeviceID int64) bool {
	if authctx.FarmAuthzSkip(r.Context()) {
		return true
	}
	if keyDev, ok := authctx.DeviceKeyDeviceID(r.Context()); ok {
		if keyDev != resourceDeviceID {
			httputil.WriteError(w, http.StatusForbidden, "device key does not match this device")
			return false
		}
		return true
	}
	if authctx.PiEdgeAuth(r.Context()) {
		return true
	}
	httputil.WriteError(w, http.StatusUnauthorized, "edge authentication required")
	return false
}

// RequirePiEdgeResourceDevice scopes Pi edge auth to hardware linked to a device row.
func RequirePiEdgeResourceDevice(w http.ResponseWriter, r *http.Request, resourceDeviceID *int64) bool {
	if authctx.FarmAuthzSkip(r.Context()) {
		return true
	}
	if resourceDeviceID == nil || *resourceDeviceID == 0 {
		if authctx.DeviceKeyAuth(r.Context()) {
			httputil.WriteError(w, http.StatusForbidden, "device key cannot access unassigned hardware")
			return false
		}
		return authctx.PiEdgeAuth(r.Context())
	}
	return RequirePiEdgeDeviceScope(w, r, *resourceDeviceID)
}
