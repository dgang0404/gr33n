package platform

import (
	"net/http"

	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/domainenums"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GET /platform/domain-enums
func (h *Handler) ListDomainEnums(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, domainenums.All())
}
