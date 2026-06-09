package chat

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gr33n-api/internal/httputil"
)

// TranscribeSTT — POST /v1/chat/stt (multipart audio) proxies to STT_BASE_URL for whisper.cpp LAN installs.
func (h *Handler) TranscribeSTT(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("STT_BASE_URL")), "/")
	if base == "" {
		httputil.WriteError(w, http.StatusNotImplemented, "local STT not configured (set STT_BASE_URL)")
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, _, err := r.FormFile("audio")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "audio file required")
		return
	}
	defer file.Close()
	audio, err := io.ReadAll(file)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "could not read audio")
		return
	}

	target := base + "/transcribe"
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(audio))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "stt request failed")
		return
	}
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "local STT unreachable")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		httputil.WriteError(w, http.StatusBadGateway, "local STT error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if len(body) == 0 {
		_, _ = w.Write([]byte(`{"text":""}`))
		return
	}
	if bytes.Contains(body, []byte(`"text"`)) {
		_, _ = w.Write(body)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"text": strings.TrimSpace(string(body))})
}
