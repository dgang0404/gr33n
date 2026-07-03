package automation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	commontypes "gr33n-api/internal/platform/commontypes"
	"gr33n-api/internal/systemlog"
)

const deviceOfflineSourceType = "device_offline"

// WithOfflineThreshold sets how long after last_heartbeat an online device is
// marked offline (default 900s when unset).
func WithOfflineThreshold(seconds int64) WorkerOption {
	return func(w *Worker) {
		if seconds > 0 {
			w.offlineAfterSeconds = seconds
		}
	}
}

func (w *Worker) offlineThresholdSeconds() int64 {
	if w.offlineAfterSeconds > 0 {
		return w.offlineAfterSeconds
	}
	return 900
}

// TickDeviceHealth marks stale online devices offline and raises farm alerts.
func (w *Worker) TickDeviceHealth(ctx context.Context) {
	threshold := w.offlineThresholdSeconds()
	rows, err := w.q.MarkStaleDevicesOffline(ctx, threshold)
	if err != nil {
		slog.Warn("device health sweep failed", "err", err)
		return
	}
	for _, d := range rows {
		systemlog.Submit(ctx, w.q, systemlog.FarmIDPtr(d.FarmID), commontypes.LogLevelWarning,
			"device_health", fmt.Sprintf("Device %q marked offline (stale heartbeat)", d.Name),
			map[string]any{"device_id": d.ID, "device_uid": d.DeviceUid})
		name := d.Name
		subject := fmt.Sprintf("Device offline: %s", name)
		body := fmt.Sprintf("Device %q (id %d) stopped reporting heartbeats and was marked offline.", name, d.ID)
		severity := db.Gr33ncoreNotificationPriorityEnumMedium
		sourceType := deviceOfflineSourceType
		sourceID := d.ID
		alert, err := w.q.CreateAlert(ctx, db.CreateAlertParams{
			FarmID:                    d.FarmID,
			RecipientUserID:           pgtype.UUID{},
			TriggeringEventSourceType: &sourceType,
			TriggeringEventSourceID:   &sourceID,
			Severity:                  &severity,
			SubjectRendered:           &subject,
			MessageTextRendered:       &body,
		})
		if err != nil {
			slog.Warn("device offline alert failed", "device_id", d.ID, "err", err)
			continue
		}
		if w.notifier != nil {
			w.notifier.DispatchFarmAlert(ctx, alert)
		}
	}
	if len(rows) > 0 {
		slog.Info("device health sweep marked devices offline", "count", len(rows), "threshold_seconds", threshold)
	}
}
