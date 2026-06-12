package platform

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/devicetaxonomy"
	"gr33n-api/internal/platform/domainenums"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GET /platform/domain-enums
func (h *Handler) ListDomainEnums(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, domainenums.All())
}

// GET /platform/device-taxonomy
func (h *Handler) ListDeviceTaxonomy(w http.ResponseWriter, r *http.Request) {
	payload, err := devicetaxonomy.LoadPayload(r.Context(), h.pool)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, payload)
}
