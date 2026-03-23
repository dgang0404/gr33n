package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	authhandler   "gr33n-api/internal/handler/auth"
	devicehandler "gr33n-api/internal/handler/device"
	farmhandler   "gr33n-api/internal/handler/farm"
	sensorhandler "gr33n-api/internal/handler/sensor"
	taskhandler   "gr33n-api/internal/handler/task"
	zonehandler   "gr33n-api/internal/handler/zone"
	db            "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

func registerRoutes(mux *http.ServeMux, pool *pgxpool.Pool, adminUser string, adminHash []byte, hashFilePath string) {
	farm   := farmhandler.NewHandler(pool)
	zone   := zonehandler.NewHandler(pool)
	device := devicehandler.NewHandler(pool)
	sensor := sensorhandler.NewHandler(pool)
	task   := taskhandler.NewHandler(pool)
	auth   := authhandler.NewHandler(adminUser, adminHash, hashFilePath, IssueToken)

	// ── Public ───────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httputil.WriteJSON(w, http.StatusServiceUnavailable,
				map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		httputil.WriteJSON(w, http.StatusOK,
			map[string]string{"status": "ok", "service": "gr33n-api"})
	})
	mux.HandleFunc("POST /auth/login", auth.Login)

	// ── Pi routes — API key required ─────────────────────────────────────────
	mux.Handle("POST /sensors/{id}/readings", requireAPIKey(http.HandlerFunc(sensor.PostReading)))
	mux.Handle("PATCH /devices/{id}/status",  requireAPIKey(http.HandlerFunc(device.UpdateStatus)))

	// ── Dashboard routes — JWT required ──────────────────────────────────────
	jwt := requireJWT

	// Auth — password change (JWT protected so you must be logged in)
	mux.Handle("PATCH /auth/password", jwt(http.HandlerFunc(auth.ChangePassword)))

	// Units
	mux.HandleFunc("GET /units", func(w http.ResponseWriter, r *http.Request) {
		units, err := db.New(pool).ListAllUnits(r.Context())
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.WriteJSON(w, http.StatusOK, units)
	})

	// Farms
	mux.Handle("GET /farms/{id}",          jwt(http.HandlerFunc(farm.Get)))
	mux.Handle("GET /farms/{id}/zones",    jwt(http.HandlerFunc(zone.ListByFarm)))
	mux.Handle("GET /farms/{id}/devices",  jwt(http.HandlerFunc(device.ListByFarm)))
	mux.Handle("GET /farms/{id}/sensors",  jwt(http.HandlerFunc(sensor.ListByFarm)))
	mux.Handle("GET /farms/{id}/tasks",    jwt(http.HandlerFunc(task.ListByFarm)))

	// Sensors
	mux.Handle("GET /sensors/{id}",                 jwt(http.HandlerFunc(sensor.Get)))
	mux.Handle("POST /farms/{id}/sensors",           jwt(http.HandlerFunc(sensor.Create)))
	mux.Handle("DELETE /sensors/{id}",              jwt(http.HandlerFunc(sensor.Delete)))
	mux.Handle("GET /sensors/{id}/readings/latest", jwt(http.HandlerFunc(sensor.LatestReading)))

	// Devices
	mux.Handle("GET /devices/{id}",        jwt(http.HandlerFunc(device.Get)))
	mux.Handle("POST /farms/{id}/devices", jwt(http.HandlerFunc(device.Create)))
	mux.Handle("DELETE /devices/{id}",     jwt(http.HandlerFunc(device.Delete)))

	// Zones
	mux.Handle("GET /zones/{id}",          jwt(http.HandlerFunc(zone.Get)))
	mux.Handle("POST /farms/{id}/zones",   jwt(http.HandlerFunc(zone.Create)))
	mux.Handle("DELETE /zones/{id}",       jwt(http.HandlerFunc(zone.Delete)))

	// Tasks
	mux.Handle("PATCH /tasks/{id}/status", jwt(http.HandlerFunc(task.UpdateStatus)))
}
