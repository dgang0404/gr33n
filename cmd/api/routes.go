package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	devicehandler "gr33n-api/internal/handler/device"
	farmhandler   "gr33n-api/internal/handler/farm"
	sensorhandler "gr33n-api/internal/handler/sensor"
	zonehandler   "gr33n-api/internal/handler/zone"
	"gr33n-api/internal/httputil"
	db "gr33n-api/internal/db"
)

func registerRoutes(mux *http.ServeMux, pool *pgxpool.Pool) {
	farm   := farmhandler.NewHandler(pool)
	zone   := zonehandler.NewHandler(pool)
	device := devicehandler.NewHandler(pool)
	sensor := sensorhandler.NewHandler(pool)

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httputil.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gr33n-api"})
	})

	// Units
	mux.HandleFunc("GET /units", func(w http.ResponseWriter, r *http.Request) {
		q := db.New(pool)
		units, err := q.ListAllUnits(r.Context())
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.WriteJSON(w, http.StatusOK, units)
	})

	// Farms
	mux.HandleFunc("GET /farms/{id}",          farm.Get)
	mux.HandleFunc("GET /farms/{id}/zones",    zone.ListByFarm)
	mux.HandleFunc("GET /farms/{id}/devices",  device.ListByFarm)
	mux.HandleFunc("GET /farms/{id}/sensors",  sensor.ListByFarm)

	// Sensors
	mux.HandleFunc("GET /sensors/{id}",                 sensor.Get)
	mux.HandleFunc("POST /farms/{id}/sensors",          sensor.Create)
	mux.HandleFunc("DELETE /sensors/{id}",              sensor.Delete)

	// Devices
	mux.HandleFunc("GET /devices/{id}",          device.Get)
	mux.HandleFunc("POST /farms/{id}/devices",   device.Create)
	mux.HandleFunc("PATCH /devices/{id}/status", device.UpdateStatus)
	mux.HandleFunc("DELETE /devices/{id}",       device.Delete)

	// Zones
	mux.HandleFunc("GET /zones/{id}",          zone.Get)
	mux.HandleFunc("POST /farms/{id}/zones",   zone.Create)
	mux.HandleFunc("DELETE /zones/{id}",       zone.Delete)
}
