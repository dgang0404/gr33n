package device

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	commontypes "gr33n-api/internal/platform/commontypes"
)

// clientIP extracts the observed request IP, preferring a trusted reverse
// proxy header over the raw connection address (Phase 214 enterprise
// topology: API may sit behind a load balancer separate from the DB/site).
func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return xff
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// recordDeviceIPIfChanged compares the observed request IP against the
// device's last known IP and, on a change, updates devices.ip_address and
// appends a gr33ncore.device_ip_events row. Best-effort: logging failures
// here must never fail the heartbeat itself.
func (h *Handler) recordDeviceIPIfChanged(ctx context.Context, device db.Gr33ncoreDevice, r *http.Request) {
	observed := clientIP(r)
	if observed == "" {
		return
	}
	newIP := net.ParseIP(observed)
	if newIP == nil {
		return
	}
	if device.IpAddress != nil {
		if addr, ok := netip.AddrFromSlice(newIP); ok && addr.Unmap() == device.IpAddress.Unmap() {
			return
		}
	}
	_ = h.q.InsertDeviceIPEvent(ctx, db.InsertDeviceIPEventParams{
		DeviceID: device.ID,
		FarmID:   device.FarmID,
		OldIp:    device.IpAddress,
		NewIp:    newIP,
	})
	newAddr, ok := netip.AddrFromSlice(newIP)
	if !ok {
		return
	}
	_ = h.q.UpdateDeviceIPAddress(ctx, db.UpdateDeviceIPAddressParams{
		ID:        device.ID,
		IpAddress: &newAddr,
	})
}

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func NewHandlerWithQuerier(q db.Querier) *Handler {
	return &Handler{q: q}
}

// GET /farms/{id}/devices
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMemberOrPiEdge(w, r, h.q, farmID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	devices, err := h.q.ListDevicesByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, devices)
}

// GET /devices/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.GetDeviceByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "device not found")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, device.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, device)
}

// POST /farms/{id}/devices
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.CreateDeviceParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.FarmID = farmID
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.CreateDevice(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create device")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, device)
}

// PATCH /devices/{id}/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	if !farmauthz.RequirePiEdgeDeviceScope(w, r, id) {
		return
	}
	var body struct {
		Status            string  `json:"status"`
		LastConfigFetchAt *string `json:"last_config_fetch_at"`
		FirmwareVersion   *string `json:"firmware_version"`
		ClientVersion     *string `json:"client_version"`
		UptimeSeconds     *int64  `json:"uptime_seconds"`
		ConfigSHA256      *string `json:"config_sha256"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	lastFetch := ""
	if body.LastConfigFetchAt != nil {
		lastFetch = strings.TrimSpace(*body.LastConfigFetchAt)
	}
	configHash := ""
	if body.ConfigSHA256 != nil {
		configHash = strings.TrimSpace(*body.ConfigSHA256)
	}
	status := commontypes.DeviceStatusEnum(body.Status)

	hasTelemetry := body.FirmwareVersion != nil || body.ClientVersion != nil || body.UptimeSeconds != nil
	var device db.Gr33ncoreDevice
	var err2 error
	if hasTelemetry {
		fw := ""
		if body.FirmwareVersion != nil {
			fw = strings.TrimSpace(*body.FirmwareVersion)
		}
		cv := ""
		if body.ClientVersion != nil {
			cv = strings.TrimSpace(*body.ClientVersion)
		}
		uptime := int64(-1)
		if body.UptimeSeconds != nil {
			uptime = *body.UptimeSeconds
		}
		device, err2 = h.q.UpdateDeviceStatusTelemetry(ctx, db.UpdateDeviceStatusTelemetryParams{
			ID:      id,
			Status:  status,
			Column3: lastFetch,
			Column4: fw,
			Column5: cv,
			Column6: uptime,
			Column7: configHash,
		})
	} else {
		device, err2 = h.q.UpdateDeviceStatus(ctx, db.UpdateDeviceStatusParams{
			ID:      id,
			Status:  status,
			Column3: lastFetch,
			Column4: configHash,
		})
	}
	if err2 != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update device status")
		return
	}
	h.recordDeviceIPIfChanged(ctx, device, r)
	httputil.WriteJSON(w, http.StatusOK, device)
}

// DELETE /devices/{id}/pending-command
func (h *Handler) ClearPendingCommand(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	if !farmauthz.RequirePiEdgeDeviceScope(w, r, id) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.q.ClearDevicePendingCommand(ctx, id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to clear pending command")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /devices/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	d0, err := h.q.GetDeviceByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "device not found")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, d0.FarmID) {
		return
	}

	err = h.q.SoftDeleteDevice(ctx, db.SoftDeleteDeviceParams{ID: id})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete device")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
