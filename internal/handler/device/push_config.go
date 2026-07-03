package device

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// PushConfig — POST /devices/{id}/push-config
// Bumps config_version so platform-sync Pis reload wiring on their next poll (Phase 123).
func (h *Handler) PushConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.GetDeviceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, device.FarmID) {
		return
	}

	updated, err := h.q.BumpDeviceConfigVersion(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to bump config version")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"device_id":       updated.ID,
		"config_version":  updated.ConfigVersion,
		"message":         "Pi will reload platform wiring on its next config poll (typically within 30s).",
	})
}
