package device

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/hardware"
	"gr33n-api/internal/httputil"
)

// GetPiConfig — GET /devices/{id}/pi-config
// Query: base_url (optional LAN API URL for the Pi).
// Returns application/json with yaml + filename, or raw YAML when Accept contains text/yaml or ?format=yaml.
func (h *Handler) GetPiConfig(w http.ResponseWriter, r *http.Request) {
	deviceID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	device, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, device.FarmID) {
		return
	}

	baseURL := strings.TrimSpace(r.URL.Query().Get("base_url"))
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

	yamlBytes, err := hardware.GeneratePiConfigYAML(hardware.PiConfigOptions{
		FarmID:     device.FarmID,
		DeviceID:   device.ID,
		DeviceUID:  derefStr(device.DeviceUid),
		DeviceName: device.Name,
		BaseURL:    baseURL,
		Sensors:    mapSensorsForPi(sensors),
		Actuators:  mapActuatorsForPi(actuators),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	runtimeCfg, err := hardware.BuildPiRuntimeConfig(hardware.PiConfigOptions{
		FarmID:     device.FarmID,
		DeviceID:   device.ID,
		DeviceUID:  derefStr(device.DeviceUid),
		Sensors:    mapSensorsForPi(sensors),
		Actuators:  mapActuatorsForPi(actuators),
	}, device.ConfigVersion, device.Config)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	configSHA := hardware.PiRuntimeConfigWiringSHA256(runtimeCfg)

	filename := piConfigFilename(device)
	wantYAML := r.URL.Query().Get("format") == "yaml" ||
		strings.Contains(r.Header.Get("Accept"), "text/yaml") ||
		r.URL.Query().Get("download") == "1"

	if wantYAML {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(yamlBytes)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"yaml":           string(yamlBytes),
		"filename":       filename,
		"config_sha256":  configSHA,
	})
}

func mapSensorsForPi(rows []db.Gr33ncoreSensor) []hardware.PiConfigSensor {
	out := make([]hardware.PiConfigSensor, len(rows))
	for i, s := range rows {
		out[i] = hardware.PiConfigSensor{
			ID:                     s.ID,
			SensorType:             s.SensorType,
			ReadingIntervalSeconds: s.ReadingIntervalSeconds,
			Config:                 s.Config,
			CalibrationData:        s.CalibrationData,
		}
	}
	return out
}

func mapActuatorsForPi(rows []db.Gr33ncoreActuator) []hardware.PiConfigActuator {
	out := make([]hardware.PiConfigActuator, len(rows))
	for i, a := range rows {
		out[i] = hardware.PiConfigActuator{
			ID:                 a.ID,
			ActuatorType:       a.ActuatorType,
			DeviceID:           a.DeviceID,
			HardwareIdentifier: a.HardwareIdentifier,
			Config:             a.Config,
		}
	}
	return out
}

func piConfigFilename(device db.Gr33ncoreDevice) string {
	uid := strings.TrimSpace(derefStr(device.DeviceUid))
	if uid != "" {
		safe := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, uid)
		return fmt.Sprintf("config-%s.yaml", safe)
	}
	return fmt.Sprintf("config-device-%d.yaml", device.ID)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
