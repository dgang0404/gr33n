package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteChatBusyJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeChatBusyJSON(rec)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "chat_busy") {
		t.Fatalf("body %s", rec.Body.String())
	}
}
