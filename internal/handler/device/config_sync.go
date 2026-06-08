package device

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/hardware"
	"gr33n-api/internal/httputil"
)

// GetConfigByUID — GET /devices/by-uid/{device_uid}/config
// Pi edge auth (X-API-Key). Returns runtime wiring JSON for one edge device.
func (h *Handler) GetConfigByUID(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimSpace(r.PathValue("device_uid"))
	if uid == "" {
		httputil.WriteError(w, http.StatusBadRequest, "device_uid is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	device, err := h.q.GetDeviceByUID(ctx, &uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequirePiEdgeDeviceScope(w, r, device.ID) {
		return
	}

	sensors, err := h.q.ListSensorsByFarm(ctx, device.FarmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list sensors")
		return
	}
	actuators, err := h.q.ListActuatorsByFarm(ctx, device.FarmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list actuators")
		return
	}

	cfg, err := hardware.BuildPiRuntimeConfig(hardware.PiConfigOptions{
		FarmID:     device.FarmID,
		DeviceID:   device.ID,
		DeviceUID:  derefStr(device.DeviceUid),
		DeviceName: device.Name,
		Sensors:    mapSensorsForPi(sensors),
		Actuators:  mapActuatorsForPi(actuators),
	}, device.ConfigVersion, device.Config)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cfg)
}

// GetConfigVersionByUID — GET /devices/by-uid/{device_uid}/config/version
func (h *Handler) GetConfigVersionByUID(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimSpace(r.PathValue("device_uid"))
	if uid == "" {
		httputil.WriteError(w, http.StatusBadRequest, "device_uid is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.GetDeviceByUID(ctx, &uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequirePiEdgeDeviceScope(w, r, device.ID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]int32{
		"config_version": device.ConfigVersion,
	})
}
