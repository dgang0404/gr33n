package main

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	db "gr33n-api/internal/db"
)

func registerRoutes(mux *http.ServeMux, pool *pgxpool.Pool) {
	queries := db.New(pool)

	// ── Health check ───────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", handleHealth(pool))

	// ── Units (reference data) ─────────────────────────────────────────────────
	mux.HandleFunc("GET /units", handleListUnits(queries))

	// ── Farms ──────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /farms/{id}", handleGetFarm(queries))

	// ── Zones ──────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /farms/{id}/zones", handleListZones(queries))

	// ── Devices ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /farms/{id}/devices", handleListDevices(queries))

	// ── Sensors ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /farms/{id}/sensors", handleListSensors(queries))

	// ── Sensor readings ────────────────────────────────────────────────────────
	mux.HandleFunc("GET /sensors/{id}/readings/latest", handleLatestReading(queries))
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func handleHealth(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy", "error": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gr33n-api"})
	}
}

func handleListUnits(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		units, err := q.ListAllUnits(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, units)
	}
}

func handleGetFarm(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		writeJSON(w, http.StatusOK, map[string]string{"farm_id": id, "status": "handler ready"})
	}
}

func handleListZones(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		writeJSON(w, http.StatusOK, map[string]string{"farm_id": id, "status": "handler ready"})
	}
}

func handleListDevices(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		writeJSON(w, http.StatusOK, map[string]string{"farm_id": id, "status": "handler ready"})
	}
}

func handleListSensors(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		writeJSON(w, http.StatusOK, map[string]string{"farm_id": id, "status": "handler ready"})
	}
}

func handleLatestReading(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		writeJSON(w, http.StatusOK, map[string]string{"sensor_id": id, "status": "handler ready"})
	}
}

// ── Helper ────────────────────────────────────────────────────────────────────
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
