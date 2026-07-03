package pushnotify

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/notifyprefs"
)

type Dispatcher struct {
	q      *db.Queries
	client *messaging.Client
}

var noopLog sync.Once

// NewDispatcher wires FCM when FCM_SERVICE_ACCOUNT_JSON (inline JSON) or
// GOOGLE_APPLICATION_CREDENTIALS (file path) is set; otherwise DispatchFarmAlert is a no-op.
func NewDispatcher(pool *pgxpool.Pool) *Dispatcher {
	d := &Dispatcher{q: db.New(pool)}
	opts := firebaseClientOptions()
	if len(opts) == 0 {
		return d
	}
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		log.Printf("pushnotify: firebase app: %v", err)
		return d
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("pushnotify: messaging client: %v", err)
		return d
	}
	d.client = client
	return d
}

func firebaseClientOptions() []option.ClientOption {
	if j := strings.TrimSpace(os.Getenv("FCM_SERVICE_ACCOUNT_JSON")); j != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(j))}
	}
	path := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if path == "" {
		return nil
	}
	if st, err := os.Stat(path); err != nil || st.IsDir() {
		return nil
	}
	return []option.ClientOption{option.WithCredentialsFile(path)}
}

// DispatchFarmAlert sends a data+notification message to each registered device for farm
// owners, managers, and operators who opted in and whose min_priority allows this severity.
func (d *Dispatcher) DispatchFarmAlert(ctx context.Context, alert db.Gr33ncoreAlertsNotification) {
	if d.client == nil {
		noopLog.Do(func() {
			log.Printf("pushnotify: FCM disabled (set FCM_SERVICE_ACCOUNT_JSON or GOOGLE_APPLICATION_CREDENTIALS)")
		})
		d.recordDelivery(ctx, alert.ID, "push", false, "FCM not configured")
		return
	}
	userIDs, err := d.q.ListFarmPushNotifyMemberUserIDs(ctx, alert.FarmID)
	if err != nil {
		log.Printf("pushnotify: list notify members farm=%d: %v", alert.FarmID, err)
		return
	}
	for _, uid := range userIDs {
		p, err := d.q.GetProfileByUserID(ctx, uid)
		if err != nil {
			continue
		}
		np := notifyprefs.FromPreferencesJSON(p.Preferences)
		if !np.PushEnabled {
			continue
		}
		if !notifyprefs.AlertMeetsMinPriority(alert, np.MinPriority) {
			continue
		}
		tokens, err := d.q.ListPushTokensByUserID(ctx, uid)
		if err != nil || len(tokens) == 0 {
			continue
		}
		for _, t := range tokens {
			d.sendOne(ctx, alert, t.FcmToken, "farm_alert")
		}
	}
	if len(userIDs) == 0 {
		d.recordDelivery(ctx, alert.ID, "push", false, "no eligible recipients")
	}
}

// DispatchCatalogUpdate sends FCM for catalog version bump alerts when prefs allow.
func (d *Dispatcher) DispatchCatalogUpdate(ctx context.Context, alert db.Gr33ncoreAlertsNotification) {
	if d.client == nil {
		return
	}
	if alert.RecipientUserID.Valid {
		p, err := d.q.GetProfileByUserID(ctx, alert.RecipientUserID.Bytes)
		if err != nil {
			return
		}
		np := notifyprefs.FromPreferencesJSON(p.Preferences)
		if !np.PushEnabled || !np.CatalogUpdates {
			return
		}
	}
	d.sendCatalogOne(ctx, alert)
}

func (d *Dispatcher) sendCatalogOne(ctx context.Context, alert db.Gr33ncoreAlertsNotification) {
	if alert.RecipientUserID.Valid {
		tokens, err := d.q.ListPushTokensByUserID(ctx, alert.RecipientUserID.Bytes)
		if err != nil || len(tokens) == 0 {
			return
		}
		for _, t := range tokens {
			d.sendOne(ctx, alert, t.FcmToken, "catalog_update")
		}
	}
}

func (d *Dispatcher) sendOne(ctx context.Context, alert db.Gr33ncoreAlertsNotification, token string, kind string) {
	if kind == "" {
		kind = "farm_alert"
	}
	title := "Farm alert"
	if alert.SubjectRendered != nil && *alert.SubjectRendered != "" {
		title = *alert.SubjectRendered
	}
	body := ""
	if alert.MessageTextRendered != nil {
		body = *alert.MessageTextRendered
	}
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{Priority: "high"},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{"apns-priority": "10"},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{Sound: "default"},
			},
		},
		Data: map[string]string{
			"farm_id":  strconv.FormatInt(alert.FarmID, 10),
			"alert_id": strconv.FormatInt(alert.ID, 10),
			"kind":     kind,
		},
	}
	_, err := d.client.Send(ctx, msg)
	if err == nil {
		d.recordDelivery(ctx, alert.ID, "push", true, "")
		return
	}
	if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
		d.recordDelivery(ctx, alert.ID, "push", false, err.Error())
		if delErr := d.q.DeletePushTokenByFCMToken(ctx, token); delErr != nil {
			log.Printf("pushnotify: drop bad token: %v", delErr)
		}
		return
	}
	log.Printf("pushnotify: alert %d send: %v", alert.ID, err)
	d.recordDelivery(ctx, alert.ID, "push", false, err.Error())
}

func (d *Dispatcher) recordDelivery(ctx context.Context, alertID int64, channel string, ok bool, detail string) {
	alert, err := d.q.GetAlertNotificationByID(ctx, alertID)
	if err != nil {
		return
	}
	merged := mergeDeliveryAttempt(alert.DeliveryAttempts, channel, ok, detail)
	status := ""
	if !ok {
		status = "delivery_failed"
	}
	_ = d.q.UpdateAlertDeliveryAttempts(ctx, db.UpdateAlertDeliveryAttemptsParams{
		ID:               alertID,
		DeliveryAttempts: merged,
		Column3:          status,
	})
}

func mergeDeliveryAttempt(existing json.RawMessage, channel string, ok bool, detail string) json.RawMessage {
	root := map[string]any{}
	if len(existing) > 0 {
		_ = json.Unmarshal(existing, &root)
	}
	entry := map[string]any{
		"at":  time.Now().UTC().Format(time.RFC3339),
		"ok":  ok,
	}
	if detail != "" {
		entry["detail"] = detail
	}
	var list []any
	if raw, exists := root[channel]; exists {
		if arr, ok := raw.([]any); ok {
			list = arr
		}
	}
	list = append(list, entry)
	root[channel] = list
	out, err := json.Marshal(root)
	if err != nil {
		return existing
	}
	return out
}
