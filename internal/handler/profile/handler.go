package profile

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
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
	httputil.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	var body struct {
		FullName    *string `json:"full_name"`
		AvatarURL   *string `json:"avatar_url"`
		Preferences []byte  `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	q := db.New(h.pool)
	existing, err := q.GetProfileByUserID(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "profile not found")
		return
	}
	fullName := existing.FullName
	if body.FullName != nil {
		fullName = body.FullName
	}
	avatarURL := existing.AvatarUrl
	if body.AvatarURL != nil {
		avatarURL = body.AvatarURL
	}
	prefs := existing.Preferences
	if body.Preferences != nil {
		prefs = body.Preferences
	}
	updated, err := q.UpdateProfile(r.Context(), db.UpdateProfileParams{
		UserID:      uid,
		FullName:    fullName,
		AvatarUrl:   avatarURL,
		Role:        existing.Role,
		Preferences: prefs,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}

// PatchMyHourlyRate — PATCH /profile/hourly-rate (Phase 20.9 WS1)
//
// Body: `{ hourly_rate: number|null, currency: "USD"|null }`. Passing
// both fields null (or an empty body) clears the default rate and the
// labor autologger will emit no cost row for future closes until a
// snapshot is supplied per-log.
func (h *Handler) PatchMyHourlyRate(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "user_id not in token")
		return
	}
	var body struct {
		HourlyRate *float64 `json:"hourly_rate"`
		Currency   *string  `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var rateN pgtype.Numeric
	if body.HourlyRate != nil {
		if *body.HourlyRate < 0 {
			httputil.WriteError(w, http.StatusBadRequest, "hourly_rate must be >= 0")
			return
		}
		if err := rateN.Scan(strconv.FormatFloat(*body.HourlyRate, 'f', -1, 64)); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid hourly_rate")
			return
		}
	}
	var currency *string
	if body.Currency != nil {
		cur := strings.ToUpper(strings.TrimSpace(*body.Currency))
		if cur != "" {
			if len(cur) != 3 {
				httputil.WriteError(w, http.StatusBadRequest, "currency must be ISO 4217 (3 uppercase letters)")
				return
			}
			currency = &cur
		}
	}
	// Rate + currency must both be set or both be cleared — a lone
	// rate with no currency is useless to the autologger.
	if rateN.Valid != (currency != nil) {
		httputil.WriteError(w, http.StatusBadRequest, "hourly_rate and currency must be set (or cleared) together")
		return
	}
	q := db.New(h.pool)
	updated, err := q.UpdateProfileHourlyRate(r.Context(), db.UpdateProfileHourlyRateParams{
		UserID:             uid,
		HourlyRate:         rateN,
		HourlyRateCurrency: currency,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) GetFarmMembers(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmAdmin(w, r, q, farmID) {
		return
	}
	rows, err := q.GetFarmMembers(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.GetFarmMembersRow{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) AddFarmMember(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var body struct {
		Email      string `json:"email"`
		RoleInFarm string `json:"role_in_farm"`
		FullName   string `json:"full_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if body.RoleInFarm == "" {
		body.RoleInFarm = "viewer"
	}

	q := db.New(h.pool)
	if !farmauthz.RequireFarmAdmin(w, r, q, farmID) {
		return
	}

	authUser, err := q.GetAuthUserByEmail(r.Context(), &body.Email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Invite flow: create auth user with null password
		authUser, err = q.CreateAuthUser(r.Context(), db.CreateAuthUserParams{
			Email: &body.Email,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
			return
		}
		fullName := body.FullName
		if fullName == "" {
			fullName = body.Email
		}
		_, err = q.CreateProfile(r.Context(), db.CreateProfileParams{
			UserID:      authUser.ID,
			FullName:    &fullName,
			Email:       body.Email,
			Role:        commontypes.UserRoleEnum("user"),
			Preferences: []byte("{}"),
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to create profile: "+err.Error())
			return
		}
	}

	member, err := q.AddFarmMember(r.Context(), db.AddFarmMemberParams{
		FarmID:      farmID,
		UserID:      authUser.ID,
		RoleInFarm:  commontypes.FarmMemberRoleEnum(body.RoleInFarm),
		Permissions: []byte("{}"),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	mod := "gr33ncore"
	tbl := "farm_memberships"
	rid := member.UserID.String()
	role := string(member.RoleInFarm)
	auditlog.Submit(r.Context(), q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumCreateRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind":         "farm_member_added",
			"role_in_farm": role,
			"email":        body.Email,
		},
	})
	httputil.WriteJSON(w, http.StatusCreated, member)
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	uid, err := uuid.Parse(r.PathValue("uid"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var body struct {
		RoleInFarm string `json:"role_in_farm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmAdmin(w, r, q, farmID) {
		return
	}
	m, err := q.UpdateFarmMemberRole(r.Context(), db.UpdateFarmMemberRoleParams{
		FarmID:     farmID,
		UserID:     uid,
		RoleInFarm: commontypes.FarmMemberRoleEnum(body.RoleInFarm),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	mod := "gr33ncore"
	tbl := "farm_memberships"
	rid := uid.String()
	auditlog.Submit(r.Context(), q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumUpdateRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind":         "farm_member_role_changed",
			"role_in_farm": body.RoleInFarm,
		},
	})
	httputil.WriteJSON(w, http.StatusOK, m)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	uid, err := uuid.Parse(r.PathValue("uid"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmAdmin(w, r, q, farmID) {
		return
	}
	if err := q.RemoveFarmMember(r.Context(), db.RemoveFarmMemberParams{
		FarmID: farmID,
		UserID: uid,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	mod := "gr33ncore"
	tbl := "farm_memberships"
	rid := uid.String()
	auditlog.Submit(r.Context(), q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumDeleteRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details:        map[string]any{"kind": "farm_member_removed"},
	})
	w.WriteHeader(http.StatusNoContent)
}
