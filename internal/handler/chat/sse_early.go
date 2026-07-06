package chat

import (
	"encoding/json"
	"net/http"
)

type sseEmitter func(eventType string, payload any) bool

func phaseStatus(phase, message string) map[string]string {
	return map[string]string{"phase": phase, "message": message}
}

func openSSE(w http.ResponseWriter) (http.Flusher, sseEmitter, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	emit := func(eventType string, payload any) bool {
		b, _ := json.Marshal(payload)
		_, werr := w.Write([]byte("event: " + eventType + "\ndata: " + string(b) + "\n\n"))
		if werr != nil {
			return false
		}
		flusher.Flush()
		return true
	}
	return flusher, emit, true
}

func writeChatBusySSE(w http.ResponseWriter, emit sseEmitter) {
	if emit == nil {
		fl, newEmit, ok := openSSE(w)
		if !ok {
			writeChatBusyJSON(w)
			return
		}
		emit = newEmit
		_ = fl
	}
	emit("error", map[string]string{
		"error":      "Guardian is answering another farm counsel question — wait for it to finish.",
		"error_code": "chat_busy",
	})
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	if fl, ok := w.(http.Flusher); ok {
		fl.Flush()
	}
}

func writeChatBusyJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":      "Guardian is answering another farm counsel question — wait for it to finish.",
		"error_code": "chat_busy",
	})
}
