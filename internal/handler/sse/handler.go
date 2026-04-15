package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
)

type Handler struct {
	pool *pgxpool.Pool

	mu      sync.RWMutex
	clients map[chan struct{}]struct{}
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:    pool,
		clients: make(map[chan struct{}]struct{}),
	}
}

// Notify wakes all connected SSE clients so they re-query immediately.
func (h *Handler) Notify() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// GET /farms/{id}/sensors/stream — SSE endpoint that pushes latest readings.
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	wake := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[wake] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, wake)
		h.mu.Unlock()
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	farmID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if farmID == 0 {
		farmID = 1
	}

	q := db.New(h.pool)
	sendReadings := func() {
		sensors, err := q.ListSensorsByFarm(r.Context(), farmID)
		if err != nil {
			log.Printf("SSE: list sensors error: %v", err)
			return
		}
		readings := make(map[int64]any, len(sensors))
		for _, s := range sensors {
			rd, err := q.GetLatestReadingBySensor(r.Context(), s.ID)
			if err != nil {
				continue
			}
			readings[s.ID] = rd
		}
		data, _ := json.Marshal(readings)
		fmt.Fprintf(w, "event: readings\ndata: %s\n\n", data)
		flusher.Flush()
	}

	sendReadings()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			sendReadings()
		case <-wake:
			sendReadings()
		}
	}
}
