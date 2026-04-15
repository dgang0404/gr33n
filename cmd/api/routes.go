package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	automationworker "gr33n-api/internal/automation"
	actuatorhandler "gr33n-api/internal/handler/actuator"
	automationhandler "gr33n-api/internal/handler/automation"
	authhandler   "gr33n-api/internal/handler/auth"
	devicehandler "gr33n-api/internal/handler/device"
	farmhandler   "gr33n-api/internal/handler/farm"
	fertigationhandler "gr33n-api/internal/handler/fertigation"
	nfhandler     "gr33n-api/internal/handler/naturalfarming"
	sensorhandler "gr33n-api/internal/handler/sensor"
	ssehandler    "gr33n-api/internal/handler/sse"
	taskhandler   "gr33n-api/internal/handler/task"
	zonehandler   "gr33n-api/internal/handler/zone"
	db            "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

func registerRoutes(mux *http.ServeMux, pool *pgxpool.Pool, worker *automationworker.Worker, adminUser string, adminHash []byte, hashFilePath string) {
	farm   := farmhandler.NewHandler(pool)
	zone   := zonehandler.NewHandler(pool)
	device := devicehandler.NewHandler(pool)
	actuator := actuatorhandler.NewHandler(pool)
	automation := automationhandler.NewHandler(pool, worker)
	sse    := ssehandler.NewHandler(pool)
	sensor := sensorhandler.NewHandler(pool, sse)
	task   := taskhandler.NewHandler(pool)
	fertigation := fertigationhandler.NewHandler(pool)
	nf     := nfhandler.NewHandler(pool)
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
	mux.HandleFunc("GET /auth/mode", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"mode": authMode})
	})

	// ── Pi routes — API key required ─────────────────────────────────────────
	mux.Handle("POST /sensors/{id}/readings",   requireAPIKey(http.HandlerFunc(sensor.PostReading)))
	mux.Handle("PATCH /devices/{id}/status",    requireAPIKey(http.HandlerFunc(device.UpdateStatus)))
	mux.Handle("POST /actuators/{id}/events",   requireAPIKey(http.HandlerFunc(actuator.RecordEvent)))
	mux.Handle("DELETE /devices/{id}/pending-command", requireAPIKey(http.HandlerFunc(device.ClearPendingCommand)))

	// ── Dashboard routes — JWT required ──────────────────────────────────────
	jwt := requireJWT

	// Auth — password change (JWT protected so you must be logged in)
	mux.Handle("PATCH /auth/password", jwt(http.HandlerFunc(auth.ChangePassword)))

	// Units
	mux.Handle("GET /units", jwt(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		units, err := db.New(pool).ListAllUnits(r.Context())
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.WriteJSON(w, http.StatusOK, units)
	})))

	// Farms
	mux.Handle("GET /farms",               jwt(http.HandlerFunc(farm.List)))
	mux.Handle("POST /farms",              jwt(http.HandlerFunc(farm.Create)))
	mux.Handle("PUT /farms/{id}",          jwt(http.HandlerFunc(farm.Update)))
	mux.Handle("DELETE /farms/{id}",       jwt(http.HandlerFunc(farm.Delete)))
	mux.Handle("GET /farms/{id}",          jwt(http.HandlerFunc(farm.Get)))
	mux.Handle("GET /farms/{id}/zones",    jwt(http.HandlerFunc(zone.ListByFarm)))
	mux.Handle("GET /farms/{id}/devices",  jwt(http.HandlerFunc(device.ListByFarm)))
	mux.Handle("GET /farms/{id}/actuators", jwt(http.HandlerFunc(actuator.ListByFarm)))
	mux.Handle("GET /farms/{id}/sensors",  jwt(http.HandlerFunc(sensor.ListByFarm)))
	mux.Handle("GET /farms/{id}/schedules", jwt(http.HandlerFunc(automation.ListSchedulesByFarm)))
	mux.Handle("GET /farms/{id}/tasks",    jwt(http.HandlerFunc(task.ListByFarm)))
	mux.Handle("GET /farms/{id}/automation/runs", jwt(http.HandlerFunc(automation.ListRunsByFarm)))

	// Sensors
	mux.Handle("GET /sensors/{id}",                 jwt(http.HandlerFunc(sensor.Get)))
	mux.Handle("POST /farms/{id}/sensors",           jwt(http.HandlerFunc(sensor.Create)))
	mux.Handle("DELETE /sensors/{id}",              jwt(http.HandlerFunc(sensor.Delete)))
	mux.Handle("GET /sensors/{id}/readings/latest", jwt(http.HandlerFunc(sensor.LatestReading)))

	// Devices
	mux.Handle("GET /devices/{id}",        jwt(http.HandlerFunc(device.Get)))
	mux.Handle("POST /farms/{id}/devices", jwt(http.HandlerFunc(device.Create)))
	mux.Handle("DELETE /devices/{id}",     jwt(http.HandlerFunc(device.Delete)))
	mux.Handle("PATCH /actuators/{id}/state", jwt(http.HandlerFunc(actuator.UpdateState)))
	mux.Handle("GET /actuators/{id}/events", jwt(http.HandlerFunc(actuator.ListEvents)))
	mux.Handle("PATCH /schedules/{id}/active", jwt(http.HandlerFunc(automation.UpdateScheduleActive)))
	mux.Handle("GET /automation/worker/health", jwt(http.HandlerFunc(automation.WorkerHealth)))

	// Zones
	mux.Handle("GET /zones/{id}",          jwt(http.HandlerFunc(zone.Get)))
	mux.Handle("PUT /zones/{id}",          jwt(http.HandlerFunc(zone.Update)))
	mux.Handle("POST /farms/{id}/zones",   jwt(http.HandlerFunc(zone.Create)))
	mux.Handle("DELETE /zones/{id}",       jwt(http.HandlerFunc(zone.Delete)))

	// Tasks
	mux.Handle("PATCH /tasks/{id}/status", jwt(http.HandlerFunc(task.UpdateStatus)))

	// Fertigation
	mux.Handle("GET /farms/{id}/fertigation/reservoirs", jwt(http.HandlerFunc(fertigation.ListReservoirsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/reservoirs", jwt(http.HandlerFunc(fertigation.CreateReservoir)))
	mux.Handle("PATCH /fertigation/reservoirs/{rid}", jwt(http.HandlerFunc(fertigation.UpdateReservoir)))
	mux.Handle("DELETE /fertigation/reservoirs/{rid}", jwt(http.HandlerFunc(fertigation.DeleteReservoir)))
	mux.Handle("GET /farms/{id}/fertigation/ec-targets", jwt(http.HandlerFunc(fertigation.ListEcTargetsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/ec-targets", jwt(http.HandlerFunc(fertigation.CreateEcTarget)))
	mux.Handle("GET /farms/{id}/fertigation/programs", jwt(http.HandlerFunc(fertigation.ListProgramsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/programs", jwt(http.HandlerFunc(fertigation.CreateProgram)))
	mux.Handle("PATCH /fertigation/programs/{rid}", jwt(http.HandlerFunc(fertigation.UpdateProgram)))
	mux.Handle("DELETE /fertigation/programs/{rid}", jwt(http.HandlerFunc(fertigation.DeleteProgram)))
	mux.Handle("GET /farms/{id}/fertigation/events", jwt(http.HandlerFunc(fertigation.ListEventsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/events", jwt(http.HandlerFunc(fertigation.CreateEvent)))

	// Natural farming
	mux.Handle("GET /farms/{id}/naturalfarming/inputs",  jwt(http.HandlerFunc(nf.ListInputs)))
	mux.Handle("POST /farms/{id}/naturalfarming/inputs", jwt(http.HandlerFunc(nf.CreateInputDefinition)))
	mux.Handle("PUT /naturalfarming/inputs/{id}",        jwt(http.HandlerFunc(nf.UpdateInputDefinition)))
	mux.Handle("DELETE /naturalfarming/inputs/{id}",     jwt(http.HandlerFunc(nf.DeleteInputDefinition)))
	mux.Handle("GET /farms/{id}/naturalfarming/batches", jwt(http.HandlerFunc(nf.ListBatches)))
	mux.Handle("POST /farms/{id}/naturalfarming/batches", jwt(http.HandlerFunc(nf.CreateInputBatch)))
	mux.Handle("PUT /naturalfarming/batches/{id}",       jwt(http.HandlerFunc(nf.UpdateInputBatch)))
	mux.Handle("DELETE /naturalfarming/batches/{id}",    jwt(http.HandlerFunc(nf.DeleteInputBatch)))

	// Actuator events by schedule (for Schedules page event history)
	mux.Handle("GET /schedules/{id}/actuator-events", jwt(http.HandlerFunc(actuator.ListEventsBySchedule)))

	// SSE — live sensor readings push
	mux.Handle("GET /farms/{id}/sensors/stream", jwt(http.HandlerFunc(sse.Stream)))
}
