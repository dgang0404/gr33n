package naturalfarming

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/naturalfarmingpack"
)

// ApplyPack — POST /farms/{id}/naturalfarming/apply-pack
func (h *Handler) ApplyPack(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmScope(w, r, h.q, farmID, farmauthz.ScopeNFPackApply, "insufficient role to apply natural farming packs") {
		return
	}
	var body struct {
		PackKey string `json:"pack_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	packKey := strings.TrimSpace(body.PackKey)
	if packKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "pack_key is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	result, err := naturalfarmingpack.ApplyPackFromRepo(ctx, h.q, farmID, packKey)
	if err != nil {
		if strings.Contains(err.Error(), "unknown switchover pack") {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"pack_key": packKey,
			"apply":    result.Apply,
			"error":    err.Error(),
		})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
