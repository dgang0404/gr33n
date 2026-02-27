package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// PathID extracts the numeric ID segment at position n (1-indexed) from a URL path.
// e.g. PathID("/farms/42/zones", 2) → 42
func PathID(path string, n int) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < n {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(parts[n-1], 10, 64)
}
