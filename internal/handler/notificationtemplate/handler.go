package notificationtemplate

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// GET /farms/{id}/notification-templates
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListNotificationTemplatesByFarm(r.Context(), &farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list templates")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreNotificationTemplate{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// POST /farms/{id}/notification-templates
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body struct {
		TemplateKey             string   `json:"template_key"`
		Description             *string  `json:"description"`
		SubjectTemplate         *string  `json:"subject_template"`
		BodyTemplateText        *string  `json:"body_template_text"`
		BodyTemplateHTML        *string  `json:"body_template_html"`
		DefaultDeliveryChannels []string `json:"default_delivery_channels"`
		DefaultPriority         *string  `json:"default_priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	key := strings.TrimSpace(body.TemplateKey)
	if key == "" {
		httputil.WriteError(w, http.StatusBadRequest, "template_key is required")
		return
	}
	channels := body.DefaultDeliveryChannels
	if len(channels) == 0 {
		channels = []string{"in_app", "email"}
	}
	var priority *db.Gr33ncoreNotificationPriorityEnum
	if body.DefaultPriority != nil && *body.DefaultPriority != "" {
		p := db.Gr33ncoreNotificationPriorityEnum(*body.DefaultPriority)
		priority = &p
	}
	row, err := h.q.CreateNotificationTemplate(r.Context(), db.CreateNotificationTemplateParams{
		FarmID:                  &farmID,
		TemplateKey:             key,
		Description:             body.Description,
		SubjectTemplate:         body.SubjectTemplate,
		BodyTemplateText:        body.BodyTemplateText,
		BodyTemplateHtml:        body.BodyTemplateHTML,
		DefaultDeliveryChannels: channels,
		DefaultPriority:         priority,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create template")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// PATCH /notification-templates/{id}
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid template id")
		return
	}
	t0, err := h.q.GetNotificationTemplateByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load template")
		return
	}
	if t0.FarmID == nil {
		httputil.WriteError(w, http.StatusForbidden, "system templates are read-only")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, *t0.FarmID) {
		return
	}
	var body struct {
		TemplateKey             *string  `json:"template_key"`
		Description             *string  `json:"description"`
		SubjectTemplate         *string  `json:"subject_template"`
		BodyTemplateText        *string  `json:"body_template_text"`
		BodyTemplateHTML        *string  `json:"body_template_html"`
		DefaultDeliveryChannels []string `json:"default_delivery_channels"`
		DefaultPriority         *string  `json:"default_priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	key := ""
	if body.TemplateKey != nil {
		key = strings.TrimSpace(*body.TemplateKey)
	}
	var channels []string
	if len(body.DefaultDeliveryChannels) > 0 {
		channels = body.DefaultDeliveryChannels
	}
	var priority *db.Gr33ncoreNotificationPriorityEnum
	if body.DefaultPriority != nil && *body.DefaultPriority != "" {
		p := db.Gr33ncoreNotificationPriorityEnum(*body.DefaultPriority)
		priority = &p
	}
	row, err := h.q.UpdateNotificationTemplate(r.Context(), db.UpdateNotificationTemplateParams{
		ID:                      id,
		Column2:                 key,
		Description:             body.Description,
		SubjectTemplate:         body.SubjectTemplate,
		BodyTemplateText:        body.BodyTemplateText,
		BodyTemplateHtml:        body.BodyTemplateHTML,
		DefaultDeliveryChannels: channels,
		DefaultPriority:         priority,
		FarmID:                  t0.FarmID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update template")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}
