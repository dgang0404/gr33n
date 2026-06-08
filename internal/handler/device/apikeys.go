package device

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/deviceapikey"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	db "gr33n-api/internal/db"
)

type deviceAPIKeyView struct {
	ID         int64      `json:"id"`
	DeviceID   int64      `json:"device_id"`
	Label      *string    `json:"label"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	Active     bool       `json:"active"`
}

// ListAPIKeys — GET /devices/{id}/api-keys
func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	deviceID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deviceID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dev, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "device not found")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, dev.FarmID) {
		return
	}

	rows, err := h.q.ListDeviceAPIKeysByDevice(ctx, deviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list device keys")
		return
	}
	out := make([]deviceAPIKeyView, 0, len(rows))
	for _, row := range rows {
		out = append(out, deviceAPIKeyView{
			ID:         row.ID,
			DeviceID:   row.DeviceID,
			Label:      row.Label,
			CreatedAt:  row.CreatedAt,
			RevokedAt:  tsPtr(row.RevokedAt),
			LastUsedAt: tsPtr(row.LastUsedAt),
			Active:     !row.RevokedAt.Valid,
		})
	}
	active, _ := h.q.CountActiveDeviceAPIKeysByDevice(ctx, deviceID)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"keys":              out,
		"active_count":      active,
		"uses_legacy_auth":  active == 0,
	})
}

// IssueAPIKey — POST /devices/{id}/api-keys
func (h *Handler) IssueAPIKey(w http.ResponseWriter, r *http.Request) {
	deviceID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deviceID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	var body struct {
		Label *string `json:"label"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dev, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "device not found")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, dev.FarmID) {
		return
	}

	secret, err := deviceapikey.NewSecret()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate key")
		return
	}
	plaintext := deviceapikey.Format(deviceID, secret)
	hash, err := deviceapikey.Hash(plaintext)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to hash key")
		return
	}
	label := body.Label
	if label != nil {
		trimmed := strings.TrimSpace(*label)
		if trimmed == "" {
			label = nil
		} else {
			label = &trimmed
		}
	}
	row, err := h.q.InsertDeviceAPIKey(ctx, db.InsertDeviceAPIKeyParams{
		DeviceID: deviceID,
		KeyHash:  hash,
		Label:    label,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to store key")
		return
	}

	schema := "gr33ncore"
	table := "device_api_keys"
	recID := strconv.FormatInt(row.ID, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:       auditlog.FarmIDPtr(dev.FarmID),
		Action:       db.Gr33ncoreUserActionTypeEnumChangeSetting,
		TargetSchema: &schema,
		TargetTable:  &table,
		TargetRecordID: &recID,
		TargetDesc:   ptrStr("device API key issued"),
		Details: map[string]any{
			"device_id": deviceID,
			"key_id":    row.ID,
		},
	})

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"key": deviceAPIKeyView{
			ID:        row.ID,
			DeviceID:  row.DeviceID,
			Label:     row.Label,
			CreatedAt: row.CreatedAt,
			Active:    true,
		},
		"api_key":      plaintext,
		"show_once":    true,
		"header_name":  "X-Device-Key",
		"auth_scheme":  "Device",
	})
}

// RevokeAPIKey — POST /devices/{id}/api-keys/{key_id}/revoke
func (h *Handler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	deviceID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deviceID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	keyID, err := strconv.ParseInt(r.PathValue("key_id"), 10, 64)
	if err != nil || keyID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid key id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dev, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "device not found")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, dev.FarmID) {
		return
	}

	row, err := h.q.RevokeDeviceAPIKey(ctx, db.RevokeDeviceAPIKeyParams{
		ID:       keyID,
		DeviceID: deviceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "active key not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to revoke key")
		return
	}

	schema := "gr33ncore"
	table := "device_api_keys"
	recID := strconv.FormatInt(row.ID, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:       auditlog.FarmIDPtr(dev.FarmID),
		Action:       db.Gr33ncoreUserActionTypeEnumChangeSetting,
		TargetSchema: &schema,
		TargetTable:  &table,
		TargetRecordID: &recID,
		TargetDesc:   ptrStr("device API key revoked"),
		Details: map[string]any{
			"device_id": deviceID,
			"key_id":    row.ID,
		},
	})

	httputil.WriteJSON(w, http.StatusOK, deviceAPIKeyView{
		ID:         row.ID,
		DeviceID:   row.DeviceID,
		Label:      row.Label,
		CreatedAt:  row.CreatedAt,
		RevokedAt:  tsPtr(row.RevokedAt),
		LastUsedAt: tsPtr(row.LastUsedAt),
		Active:     false,
	})
}

func ptrStr(s string) *string { return &s }

func tsPtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
