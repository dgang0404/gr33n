package profile

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/notifyprefs"
)

func validPushPlatform(p string) bool {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "android", "ios", "web":
		return true
	default:
		return false
	}
}

// RegisterPushToken upserts an FCM registration token for the current user.
func (h *Handler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	var body struct {
		Platform string `json:"platform"`
		FcmToken string `json:"fcm_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	body.FcmToken = strings.TrimSpace(body.FcmToken)
	if body.FcmToken == "" {
		httputil.WriteError(w, http.StatusBadRequest, "fcm_token is required")
		return
	}
	if !validPushPlatform(body.Platform) {
		httputil.WriteError(w, http.StatusBadRequest, "platform must be android, ios, or web")
		return
	}
	q := db.New(h.pool)
	row, err := q.UpsertUserPushToken(r.Context(), db.UpsertUserPushTokenParams{
		UserID:   uid,
		Platform: strings.ToLower(strings.TrimSpace(body.Platform)),
		FcmToken: body.FcmToken,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// UnregisterPushToken removes one FCM token for the current user.
func (h *Handler) UnregisterPushToken(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	var body struct {
		FcmToken string `json:"fcm_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	body.FcmToken = strings.TrimSpace(body.FcmToken)
	if body.FcmToken == "" {
		httputil.WriteError(w, http.StatusBadRequest, "fcm_token is required")
		return
	}
	q := db.New(h.pool)
	if err := q.DeleteUserPushToken(r.Context(), db.DeleteUserPushTokenParams{
		UserID:   uid,
		FcmToken: body.FcmToken,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetNotificationPreferences returns merged notification prefs (defaults + profile).
func (h *Handler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	q := db.New(h.pool)
	p, err := q.GetProfileByUserID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "profile not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, notifyprefs.FromPreferencesJSON(p.Preferences))
}

// PatchNotificationPreferences merges notify.* into profiles.preferences.
func (h *Handler) PatchNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	var body struct {
		PushEnabled *bool   `json:"push_enabled"`
		MinPriority *string `json:"min_priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.PushEnabled == nil && body.MinPriority == nil {
		httputil.WriteError(w, http.StatusBadRequest, "no fields to patch")
		return
	}
	q := db.New(h.pool)
	existing, err := q.GetProfileByUserID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "profile not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cur := notifyprefs.FromPreferencesJSON(existing.Preferences)
	if body.PushEnabled != nil {
		cur.PushEnabled = *body.PushEnabled
	}
	if body.MinPriority != nil {
		s := strings.ToLower(strings.TrimSpace(*body.MinPriority))
		if s != "low" && s != "medium" && s != "high" && s != "critical" {
			httputil.WriteError(w, http.StatusBadRequest, "min_priority must be low, medium, high, or critical")
			return
		}
		cur.MinPriority = s
	}
	mergedPrefs, err := notifyprefs.SetNotify(existing.Preferences, cur)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := q.UpdateProfile(r.Context(), db.UpdateProfileParams{
		UserID:      uid,
		FullName:    existing.FullName,
		AvatarUrl:   existing.AvatarUrl,
		Role:        existing.Role,
		Preferences: mergedPrefs,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, notifyprefs.FromPreferencesJSON(updated.Preferences))
}

// ListMyPushTokens returns registered device tokens for the current user (for debugging/settings).
func (h *Handler) ListMyPushTokens(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	q := db.New(h.pool)
	rows, err := q.ListPushTokensByUserID(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreUserPushToken{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}
