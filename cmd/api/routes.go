package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/ai"
	automationworker "gr33n-api/internal/automation"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/filestorage"
	actuatorhandler "gr33n-api/internal/handler/actuator"
	alerthandler "gr33n-api/internal/handler/alert"
	animalhandler "gr33n-api/internal/handler/animal"
	aquaponicshandler "gr33n-api/internal/handler/aquaponics"
	audithandler "gr33n-api/internal/handler/audit"
	authhandler "gr33n-api/internal/handler/auth"
	automationhandler "gr33n-api/internal/handler/automation"
	chathandler "gr33n-api/internal/handler/chat"
	fieldguideshandler "gr33n-api/internal/handler/fieldguides"
	commonscataloghandler "gr33n-api/internal/handler/commonscatalog"
	costhandler "gr33n-api/internal/handler/cost"
	cropcyclehandler "gr33n-api/internal/handler/cropcycle"
	cropprofilehandler "gr33n-api/internal/handler/cropprofile"
	guardianhandler "gr33n-api/internal/handler/guardian"
	devicehandler "gr33n-api/internal/handler/device"
	devicecmdhandler "gr33n-api/internal/handler/devicecmd"
	farmhandler "gr33n-api/internal/handler/farm"
	fertigationhandler "gr33n-api/internal/handler/fertigation"
	fileattachhandler "gr33n-api/internal/handler/fileattach"
	nfhandler "gr33n-api/internal/handler/naturalfarming"
	organizationhandler "gr33n-api/internal/handler/organization"
	planthandler "gr33n-api/internal/handler/plants"
	profilehandler "gr33n-api/internal/handler/profile"
	raghandler "gr33n-api/internal/handler/rag"
	recipehandler "gr33n-api/internal/handler/recipe"
	sensorhandler "gr33n-api/internal/handler/sensor"
	setpointhandler "gr33n-api/internal/handler/setpoint"
	ssehandler "gr33n-api/internal/handler/sse"
	taskhandler "gr33n-api/internal/handler/task"
	weatherhandler "gr33n-api/internal/handler/weather"
	lightinghandler "gr33n-api/internal/handler/lighting"
	zonehandler "gr33n-api/internal/handler/zone"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/pushnotify"
)

func registerRoutes(mux *http.ServeMux, pool *pgxpool.Pool, worker *automationworker.Worker, pushDispatch *pushnotify.Dispatcher, adminUser string, adminHash []byte, hashFilePath string, fileStore filestorage.Store, fileCfg filestorage.Config, adminBindUserID uuid.UUID, adminBindEmail string, aiCfg ai.Config) {
	farm := farmhandler.NewHandler(pool)
	weather := weatherhandler.NewHandler(pool)
	org := organizationhandler.NewHandler(pool)
	audit := audithandler.NewHandler(pool)
	zone := zonehandler.NewHandler(pool)
	device := devicehandler.NewHandler(pool)
	initPiEdgeAuth(db.New(pool))
	devicecmd := devicecmdhandler.NewHandler(pool)
	actuator := actuatorhandler.NewHandler(pool)
	automation := automationhandler.NewHandler(pool, worker)
	sse := ssehandler.NewHandler(pool)
	if pushDispatch == nil {
		pushDispatch = pushnotify.NewDispatcher(pool)
	}
	sensor := sensorhandler.NewHandler(pool, sse, pushDispatch)
	task := taskhandler.NewHandler(pool)
	fertigation := fertigationhandler.NewHandler(pool, worker)
	nf := nfhandler.NewHandler(pool)
	recipe := recipehandler.NewHandler(pool)
	cropcycle := cropcyclehandler.NewHandler(pool)
	rag := raghandler.NewHandler(pool, aiCfg.Enabled)
	aichat := chathandler.NewHandler(pool, aiCfg, fileStore)
	fieldGuides := fieldguideshandler.NewHandler("")
	plants := planthandler.NewHandler(pool)
	cropProfiles := cropprofilehandler.NewHandler(pool)
	guardianNudge := guardianhandler.NewHandler(pool)
	animals := animalhandler.NewHandler(pool)
	aquaponics := aquaponicshandler.NewHandler(pool)
	alert := alerthandler.NewHandler(pool)
	prof := profilehandler.NewHandler(pool)
	setpoint := setpointhandler.NewHandler(pool)
	lighting := lightinghandler.NewHandler(pool)
	auth := authhandler.NewHandler(adminUser, adminHash, hashFilePath, IssueToken, pool, adminBindUserID, adminBindEmail)

	if fileStore == nil {
		store, err := filestorage.NewStore(context.Background(), filestorage.Config{Backend: "local", LocalRoot: "./data/files"})
		if err != nil {
			log.Fatalf("file storage init: %v", err)
		}
		fileStore = store
	}
	cost := costhandler.NewHandler(pool, fileStore)
	files := fileattachhandler.NewHandler(pool, fileStore, fileCfg.DownloadURLTTL)
	commonsCatalog := commonscataloghandler.NewHandler(pool)

	jwtChain := func(h http.Handler) http.Handler {
		return requireJWT(withRequestLog("jwt", h))
	}
	piChain := func(h http.Handler) http.Handler {
		return requireAPIKey(withRequestLog("api_key", h))
	}
	jwtOrPiChain := func(h http.Handler) http.Handler {
		return requireJWTOrPiEdge(withRequestLog("jwt_or_pi", h))
	}
	jwt := jwtChain

	// ── Public ───────────────────────────────────────────────────────────────
	mux.Handle("GET /health", withRequestLog("public", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httputil.WriteJSON(w, http.StatusServiceUnavailable,
				map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		httputil.WriteJSON(w, http.StatusOK,
			map[string]string{"status": "ok", "service": "gr33n-api"})
	})))
	mux.Handle("POST /auth/login", withRequestLog("public", http.HandlerFunc(auth.Login)))
	mux.Handle("POST /auth/register", withRequestLog("public", http.HandlerFunc(auth.Register)))
	mux.Handle("GET /auth/mode", withRequestLog("public", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"mode": authMode})
	})))
	mux.Handle("GET /capabilities", withRequestLog("public", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sttLocal := strings.TrimSpace(os.Getenv("STT_BASE_URL")) != ""
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"ai_enabled":          aiCfg.Enabled,
			"vision_chat_enabled": ai.VisionConfigured(),
			"stt_local_enabled":   sttLocal,
		})
	})))

	// ── Pi routes — API key required ─────────────────────────────────────────
	mux.Handle("POST /sensors/{id}/readings", piChain(http.HandlerFunc(sensor.PostReading)))
	mux.Handle("POST /sensors/readings/batch", piChain(http.HandlerFunc(sensor.PostReadingsBatch)))
	mux.Handle("PATCH /devices/{id}/status", piChain(http.HandlerFunc(device.UpdateStatus)))
	mux.Handle("POST /actuators/{id}/events", piChain(http.HandlerFunc(actuator.RecordEvent)))
	mux.Handle("DELETE /devices/{id}/pending-command", piChain(http.HandlerFunc(device.ClearPendingCommand)))
	// Phase 39 WS1 — Pi polls queue head and acks after execution
	mux.Handle("GET /devices/{id}/commands/next", piChain(http.HandlerFunc(devicecmd.Next)))
	mux.Handle("POST /devices/{id}/commands/{cid}/ack", piChain(http.HandlerFunc(devicecmd.Ack)))
	// Phase 51 WS1 — Pi runtime config sync by device_uid (version before full config)
	mux.Handle("GET /devices/by-uid/{device_uid}/config/version", piChain(http.HandlerFunc(device.GetConfigVersionByUID)))
	mux.Handle("GET /devices/by-uid/{device_uid}/config", piChain(http.HandlerFunc(device.GetConfigByUID)))

	// ── Dashboard routes — JWT required ──────────────────────────────────────

	// Auth — password change (JWT protected so you must be logged in)
	mux.Handle("PATCH /auth/password", jwt(http.HandlerFunc(auth.ChangePassword)))

	// Phase 27 — Farm Guardian chat + session history
	mux.Handle("GET /v1/chat/health", jwt(http.HandlerFunc(aichat.GetHealth)))
	mux.Handle("POST /v1/chat", jwt(http.HandlerFunc(aichat.PostV1)))
	mux.Handle("POST /v1/chat/stt", jwt(http.HandlerFunc(aichat.TranscribeSTT)))
	mux.Handle("POST /v1/chat/confirm", jwt(http.HandlerFunc(aichat.PostConfirm)))
	mux.Handle("GET /v1/chat/proposals", jwt(http.HandlerFunc(aichat.ListProposals)))
	mux.Handle("GET /v1/chat/sessions", jwt(http.HandlerFunc(aichat.ListSessions)))
	mux.Handle("GET /v1/chat/sessions/{session_id}", jwt(http.HandlerFunc(aichat.GetSession)))
	mux.Handle("PATCH /v1/chat/sessions/{session_id}", jwt(http.HandlerFunc(aichat.PatchSession)))
	mux.Handle("DELETE /v1/chat/sessions/{session_id}", jwt(http.HandlerFunc(aichat.DeleteSession)))
	mux.Handle("POST /v1/chat/sessions/{session_id}/close", jwt(http.HandlerFunc(aichat.CloseSession)))
	mux.Handle("GET /farms/{id}/guardian-memory/recent", jwt(http.HandlerFunc(aichat.RecentMemory)))
	mux.Handle("GET /farms/{id}/guardian-memory/export", jwt(http.HandlerFunc(aichat.ExportMemory)))
	mux.Handle("DELETE /farms/{id}/guardian-memory", jwt(http.HandlerFunc(aichat.ClearMemory)))
	// Phase 28 WS5 — operator-facing token-usage dashboard
	mux.Handle("GET /v1/chat/usage", jwt(http.HandlerFunc(aichat.GetUsage)))

	// Phase 37 — field guide procedures (static; works without LLM for print/list)
	mux.Handle("GET /v1/field-guides/procedures", jwt(http.HandlerFunc(fieldGuides.ListProcedures)))
	mux.Handle("GET /v1/field-guides/procedures/{id}", jwt(http.HandlerFunc(fieldGuides.GetProcedure)))
	mux.Handle("GET /v1/field-guides/procedures/{id}/print", jwt(http.HandlerFunc(fieldGuides.PrintProcedure)))

	// Units
	mux.Handle("GET /units", jwt(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		units, err := db.New(pool).ListAllUnits(r.Context())
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.WriteJSON(w, http.StatusOK, units)
	})))

	// Commons catalog (gr33n_inserts — browse + per-farm import audit)
	mux.Handle("GET /commons/catalog", jwt(http.HandlerFunc(commonsCatalog.List)))
	mux.Handle("GET /commons/catalog/{slug}", jwt(http.HandlerFunc(commonsCatalog.GetBySlug)))
	mux.Handle("GET /farms/{id}/commons/catalog-imports", jwt(http.HandlerFunc(commonsCatalog.ListFarmImports)))
	mux.Handle("POST /farms/{id}/commons/catalog-imports", jwt(http.HandlerFunc(commonsCatalog.Import)))

	// Farms
	mux.Handle("GET /farms", jwt(http.HandlerFunc(farm.List)))
	mux.Handle("POST /farms", jwt(http.HandlerFunc(farm.Create)))
	mux.Handle("POST /farms/{id}/bootstrap-template", jwt(http.HandlerFunc(farm.ApplyFarmBootstrapTemplate)))
	mux.Handle("PUT /farms/{id}", jwt(http.HandlerFunc(farm.Update)))
	mux.Handle("PATCH /farms/{id}/site", jwt(http.HandlerFunc(farm.PatchSite)))
	mux.Handle("GET /farms/{id}/site-weather", jwt(http.HandlerFunc(weather.GetSiteWeather)))
	mux.Handle("POST /farms/{id}/weather/manual", jwt(http.HandlerFunc(weather.PostManual)))
	mux.Handle("PATCH /farms/{id}/organization", jwt(http.HandlerFunc(farm.SetOrganization)))
	mux.Handle("DELETE /farms/{id}", jwt(http.HandlerFunc(farm.Delete)))
	mux.Handle("GET /farms/{id}", jwt(http.HandlerFunc(farm.Get)))

	// Organizations (multi-farm tenant grouping)
	mux.Handle("POST /organizations", jwt(http.HandlerFunc(org.Create)))
	mux.Handle("GET /organizations", jwt(http.HandlerFunc(org.ListMine)))
	mux.Handle("GET /organizations/{id}", jwt(http.HandlerFunc(org.Get)))
	mux.Handle("PATCH /organizations/{id}", jwt(http.HandlerFunc(org.Update)))
	mux.Handle("GET /organizations/{id}/usage-summary", jwt(http.HandlerFunc(org.UsageSummary)))
	mux.Handle("GET /organizations/{id}/audit-events", jwt(http.HandlerFunc(audit.ListByOrganization)))
	mux.Handle("POST /organizations/{id}/members", jwt(http.HandlerFunc(org.AddMember)))
	mux.Handle("PATCH /farms/{id}/insert-commons/opt-in", jwt(http.HandlerFunc(farm.SetInsertCommonsOptIn)))
	mux.Handle("GET /farms/{id}/insert-commons/preview", jwt(http.HandlerFunc(farm.InsertCommonsPreview)))
	mux.Handle("POST /farms/{id}/insert-commons/sync", jwt(http.HandlerFunc(farm.InsertCommonsSync)))
	mux.Handle("GET /farms/{id}/insert-commons/sync-events", jwt(http.HandlerFunc(farm.InsertCommonsHistory)))
	mux.Handle("GET /farms/{id}/insert-commons/bundles", jwt(http.HandlerFunc(farm.ListInsertCommonsBundles)))
	mux.Handle("POST /farms/{id}/insert-commons/bundles/{bundle_id}/approve", jwt(http.HandlerFunc(farm.ApproveInsertCommonsBundleHTTP)))
	mux.Handle("POST /farms/{id}/insert-commons/bundles/{bundle_id}/reject", jwt(http.HandlerFunc(farm.RejectInsertCommonsBundleHTTP)))
	mux.Handle("POST /farms/{id}/insert-commons/bundles/{bundle_id}/deliver", jwt(http.HandlerFunc(farm.RetryInsertCommonsBundleDeliver)))
	mux.Handle("GET /farms/{id}/insert-commons/bundles/{bundle_id}/export", jwt(http.HandlerFunc(farm.ExportInsertCommonsBundle)))
	mux.Handle("GET /farms/{id}/audit-events", jwt(http.HandlerFunc(audit.ListByFarm)))
	mux.Handle("GET /farms/{id}/zones", jwt(http.HandlerFunc(zone.ListByFarm)))
	mux.Handle("GET /farms/{id}/devices", jwtOrPiChain(http.HandlerFunc(device.ListByFarm)))
	mux.Handle("GET /farms/{id}/actuators", jwt(http.HandlerFunc(actuator.ListByFarm)))
	mux.Handle("POST /farms/{id}/actuators", jwt(http.HandlerFunc(actuator.Create)))
	mux.Handle("GET /actuators/{id}", jwt(http.HandlerFunc(actuator.Get)))
	mux.Handle("POST /actuators/{id}/command", jwt(http.HandlerFunc(actuator.EnqueueCommand)))
	// Phase 39 WS1 — operator JWT routes for device command queue
	mux.Handle("POST /devices/{id}/commands", jwt(http.HandlerFunc(devicecmd.Enqueue)))
	mux.Handle("GET /devices/{id}/commands", jwt(http.HandlerFunc(devicecmd.List)))
	mux.Handle("GET /farms/{id}/sensors/readings/latest", jwt(http.HandlerFunc(sensor.LatestReadingsByFarm)))
	mux.Handle("GET /farms/{id}/sensors", jwt(http.HandlerFunc(sensor.ListByFarm)))
	mux.Handle("GET /farms/{id}/schedules", jwt(http.HandlerFunc(automation.ListSchedulesByFarm)))
	mux.Handle("POST /farms/{id}/schedules", jwt(http.HandlerFunc(automation.CreateSchedule)))
	mux.Handle("GET /farms/{id}/tasks", jwt(http.HandlerFunc(task.ListByFarm)))
	mux.Handle("POST /farms/{id}/tasks", jwt(http.HandlerFunc(task.Create)))
	mux.Handle("GET /farms/{id}/rag/search", jwt(http.HandlerFunc(rag.Search)))
	mux.Handle("POST /farms/{id}/rag/search", jwt(http.HandlerFunc(rag.Search)))
	mux.Handle("POST /farms/{id}/rag/answer", jwt(http.HandlerFunc(rag.Answer)))
	mux.Handle("GET /farms/{id}/automation/runs", jwt(http.HandlerFunc(automation.ListRunsByFarm)))
	mux.Handle("GET /farms/{id}/automation/rules", jwt(http.HandlerFunc(automation.ListAutomationRulesByFarm)))
	mux.Handle("POST /farms/{id}/automation/rules", jwt(http.HandlerFunc(automation.CreateAutomationRule)))
	// Phase 36 WS3 — greenhouse climate rule templates
	mux.Handle("POST /farms/{id}/automation/rule-templates/greenhouse", jwt(http.HandlerFunc(automation.ApplyGreenhouseRuleTemplates)))

	// Stage-scoped setpoints (Phase 20.6 WS2)
	mux.Handle("GET /farms/{id}/setpoints", jwt(http.HandlerFunc(setpoint.List)))
	mux.Handle("POST /farms/{id}/setpoints", jwt(http.HandlerFunc(setpoint.Create)))
	mux.Handle("GET /setpoints/{id}", jwt(http.HandlerFunc(setpoint.Get)))
	mux.Handle("PUT /setpoints/{id}", jwt(http.HandlerFunc(setpoint.Update)))
	mux.Handle("DELETE /setpoints/{id}", jwt(http.HandlerFunc(setpoint.Delete)))

	// Sensors
	mux.Handle("GET /sensors/{id}", jwt(http.HandlerFunc(sensor.Get)))
	mux.Handle("PATCH /sensors/{id}/wiring", jwt(http.HandlerFunc(sensor.PatchWiring)))
	mux.Handle("POST /farms/{id}/sensors", jwt(http.HandlerFunc(sensor.Create)))
	mux.Handle("PUT /sensors/{id}", jwt(http.HandlerFunc(sensor.Update)))
	mux.Handle("DELETE /sensors/{id}", jwt(http.HandlerFunc(sensor.Delete)))
	mux.Handle("GET /sensors/{id}/readings/latest", jwt(http.HandlerFunc(sensor.LatestReading)))
	mux.Handle("GET /sensors/{id}/readings/stats", jwt(http.HandlerFunc(sensor.ReadingStats)))
	mux.Handle("GET /sensors/{id}/readings", jwt(http.HandlerFunc(sensor.ListReadings)))

	// Devices
	mux.Handle("GET /devices/{id}", jwt(http.HandlerFunc(device.Get)))
	mux.Handle("GET /devices/{id}/pi-config", jwt(http.HandlerFunc(device.GetPiConfig)))
	mux.Handle("GET /devices/{id}/api-keys", jwt(http.HandlerFunc(device.ListAPIKeys)))
	mux.Handle("POST /devices/{id}/api-keys", jwt(http.HandlerFunc(device.IssueAPIKey)))
	mux.Handle("POST /devices/{id}/api-keys/{key_id}/revoke", jwt(http.HandlerFunc(device.RevokeAPIKey)))
	mux.Handle("POST /farms/{id}/devices", jwt(http.HandlerFunc(device.Create)))
	mux.Handle("DELETE /devices/{id}", jwt(http.HandlerFunc(device.Delete)))
	mux.Handle("PATCH /actuators/{id}/assign", jwt(http.HandlerFunc(actuator.UpdateAssignment)))
	mux.Handle("PATCH /actuators/{id}/state", jwt(http.HandlerFunc(actuator.UpdateState)))
	mux.Handle("PATCH /actuators/{id}/wiring", jwt(http.HandlerFunc(actuator.PatchWiring)))
	mux.Handle("GET /actuators/{id}/events", jwt(http.HandlerFunc(actuator.ListEvents)))
	mux.Handle("PATCH /schedules/{id}/active", jwt(http.HandlerFunc(automation.UpdateScheduleActive)))
	mux.Handle("PUT /schedules/{id}", jwt(http.HandlerFunc(automation.UpdateSchedule)))
	mux.Handle("DELETE /schedules/{id}", jwt(http.HandlerFunc(automation.DeleteSchedule)))
	mux.Handle("GET /automation/worker/health", jwt(http.HandlerFunc(automation.WorkerHealth)))

	// Automation rules (Phase 20 WS1)
	mux.Handle("GET /automation/rules/{id}", jwt(http.HandlerFunc(automation.GetAutomationRule)))
	mux.Handle("PUT /automation/rules/{id}", jwt(http.HandlerFunc(automation.UpdateAutomationRule)))
	mux.Handle("DELETE /automation/rules/{id}", jwt(http.HandlerFunc(automation.DeleteAutomationRule)))
	mux.Handle("PATCH /automation/rules/{id}/active", jwt(http.HandlerFunc(automation.UpdateAutomationRuleActive)))
	mux.Handle("GET /automation/rules/{id}/actions", jwt(http.HandlerFunc(automation.ListActionsByRule)))
	mux.Handle("POST /automation/rules/{id}/actions", jwt(http.HandlerFunc(automation.CreateActionForRule)))
	mux.Handle("PUT /automation/actions/{id}", jwt(http.HandlerFunc(automation.UpdateAction)))
	mux.Handle("DELETE /automation/actions/{id}", jwt(http.HandlerFunc(automation.DeleteAction)))

	// Zones
	mux.Handle("GET /zones/{id}", jwt(http.HandlerFunc(zone.Get)))
	mux.Handle("PUT /zones/{id}", jwt(http.HandlerFunc(zone.Update)))
	mux.Handle("POST /farms/{id}/zones", jwt(http.HandlerFunc(zone.Create)))
	mux.Handle("DELETE /zones/{id}", jwt(http.HandlerFunc(zone.Delete)))

	// Tasks
	mux.Handle("PUT /tasks/{id}", jwt(http.HandlerFunc(task.Update)))
	mux.Handle("DELETE /tasks/{id}", jwt(http.HandlerFunc(task.Delete)))
	mux.Handle("PATCH /tasks/{id}/status", jwt(http.HandlerFunc(task.UpdateStatus)))
	// Task labor log (Phase 20.95 WS1 + Phase 20.9 WS1 timer)
	mux.Handle("GET /tasks/{id}/labor", jwt(http.HandlerFunc(task.ListLabor)))
	mux.Handle("POST /tasks/{id}/labor", jwt(http.HandlerFunc(task.CreateLabor)))
	mux.Handle("POST /tasks/{id}/labor/start", jwt(http.HandlerFunc(task.StartLabor)))
	mux.Handle("POST /tasks/{id}/labor/stop", jwt(http.HandlerFunc(task.StopLabor)))
	mux.Handle("DELETE /labor/{id}", jwt(http.HandlerFunc(task.DeleteLabor)))
	// Phase 20.7 WS3: task-driven consumption ledger (autologged).
	mux.Handle("GET /farms/{id}/task-consumptions", jwt(http.HandlerFunc(task.ListFarmConsumptions)))
	mux.Handle("GET /tasks/{id}/consumptions", jwt(http.HandlerFunc(task.ListConsumptions)))
	mux.Handle("POST /tasks/{id}/consumptions", jwt(http.HandlerFunc(task.CreateConsumption)))
	mux.Handle("DELETE /consumptions/{id}", jwt(http.HandlerFunc(task.DeleteConsumption)))

	// Fertigation
	mux.Handle("GET /farms/{id}/fertigation/reservoirs", jwt(http.HandlerFunc(fertigation.ListReservoirsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/reservoirs", jwt(http.HandlerFunc(fertigation.CreateReservoir)))
	mux.Handle("PATCH /fertigation/reservoirs/{rid}", jwt(http.HandlerFunc(fertigation.UpdateReservoir)))
	mux.Handle("DELETE /fertigation/reservoirs/{rid}", jwt(http.HandlerFunc(fertigation.DeleteReservoir)))
	mux.Handle("GET /farms/{id}/fertigation/ec-targets", jwt(http.HandlerFunc(fertigation.ListEcTargetsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/ec-targets", jwt(http.HandlerFunc(fertigation.CreateEcTarget)))
	mux.Handle("GET /farms/{id}/fertigation/programs", jwt(http.HandlerFunc(fertigation.ListProgramsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/programs", jwt(http.HandlerFunc(fertigation.CreateProgram)))
	mux.Handle("POST /farms/{id}/fertigation/programs/{rid}/run-now", jwt(http.HandlerFunc(fertigation.RunProgramNow)))
	mux.Handle("PATCH /fertigation/programs/{rid}", jwt(http.HandlerFunc(fertigation.UpdateProgram)))
	mux.Handle("DELETE /fertigation/programs/{rid}", jwt(http.HandlerFunc(fertigation.DeleteProgram)))
	// Program-bound executable actions (Phase 20.9 WS4)
	mux.Handle("GET /fertigation/programs/{id}/actions", jwt(http.HandlerFunc(automation.ListActionsByProgram)))
	mux.Handle("POST /fertigation/programs/{id}/actions", jwt(http.HandlerFunc(automation.CreateActionForProgram)))
	mux.Handle("GET /farms/{id}/fertigation/events", jwt(http.HandlerFunc(fertigation.ListEventsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/events", jwt(http.HandlerFunc(fertigation.CreateEvent)))
	mux.Handle("GET /farms/{id}/fertigation/mixing-events", jwt(http.HandlerFunc(fertigation.ListMixingEventsByFarm)))
	mux.Handle("POST /farms/{id}/fertigation/mixing-events", jwt(http.HandlerFunc(fertigation.CreateMixingEvent)))
	mux.Handle("GET /farms/{id}/fertigation/mixing-events/{mid}/components", jwt(http.HandlerFunc(fertigation.ListMixingEventComponents)))
	// Phase 39 WS3 — mix_batch enqueue + preview
	mux.Handle("POST /farms/{id}/fertigation/mix-jobs", jwt(http.HandlerFunc(fertigation.EnqueueMixJob)))
	mux.Handle("GET /fertigation/programs/{rid}/mix-preview", jwt(http.HandlerFunc(fertigation.MixPreview)))
	// Phase 39 WS7 — Zone Water tab one-shot status
	mux.Handle("GET /fertigation/programs/{rid}/water-status", jwt(http.HandlerFunc(fertigation.WaterStatus)))
	// Phase 39 WS6 — set base water EC on reservoir
	mux.Handle("PATCH /fertigation/reservoirs/{rid}/base-water", jwt(http.HandlerFunc(fertigation.SetReservoirBaseWater)))

	mux.Handle("GET /farms/{id}/crop-cycles", jwt(http.HandlerFunc(cropcycle.List)))
	mux.Handle("POST /farms/{id}/crop-cycles", jwt(http.HandlerFunc(cropcycle.Create)))
	mux.Handle("PATCH /crop-cycles/{id}/stage", jwt(http.HandlerFunc(cropcycle.UpdateStage)))
	mux.Handle("GET /crop-cycles/{id}", jwt(http.HandlerFunc(cropcycle.Get)))
	mux.Handle("PUT /crop-cycles/{id}", jwt(http.HandlerFunc(cropcycle.Update)))
	mux.Handle("DELETE /crop-cycles/{id}", jwt(http.HandlerFunc(cropcycle.Delete)))
	// Phase 28 WS1 — crop cycle analytics. .csv suffix routes share the same
	// handler; the handler switches output mode on the URL path.
	mux.Handle("GET /crop-cycles/{id}/summary", jwt(http.HandlerFunc(cropcycle.Summary)))
	mux.Handle("GET /crop-cycles/{id}/summary.csv", jwt(http.HandlerFunc(cropcycle.Summary)))
	mux.Handle("GET /farms/{id}/crop-cycles/compare", jwt(http.HandlerFunc(cropcycle.Compare)))
	mux.Handle("GET /farms/{id}/crop-cycles/compare.csv", jwt(http.HandlerFunc(cropcycle.Compare)))

	// Plants (crop tracking)
	mux.Handle("GET /farms/{id}/plants", jwt(http.HandlerFunc(plants.List)))
	mux.Handle("POST /farms/{id}/plants", jwt(http.HandlerFunc(plants.Create)))
	mux.Handle("GET /plants/{id}", jwt(http.HandlerFunc(plants.Get)))
	mux.Handle("PUT /plants/{id}", jwt(http.HandlerFunc(plants.Update)))
	mux.Handle("DELETE /plants/{id}", jwt(http.HandlerFunc(plants.Delete)))

	// Crop knowledge base (Phase 64)
	mux.Handle("GET /farms/{id}/guardian-nudge", jwt(http.HandlerFunc(guardianNudge.Nudge)))
	mux.Handle("GET /farms/{id}/crop-profiles", jwt(http.HandlerFunc(cropProfiles.List)))
	mux.Handle("POST /farms/{id}/crop-profiles/import", jwt(http.HandlerFunc(cropProfiles.Import)))
	mux.Handle("GET /crop-profiles/{id}", jwt(http.HandlerFunc(cropProfiles.Get)))
	mux.Handle("POST /crop-profiles/{id}/clone", jwt(http.HandlerFunc(cropProfiles.Clone)))
	mux.Handle("GET /crop-profiles/{id}/export", jwt(http.HandlerFunc(cropProfiles.Export)))

	// Animal husbandry (Phase 20.8 WS2)
	mux.Handle("GET /farms/{id}/animal-groups", jwt(http.HandlerFunc(animals.ListGroups)))
	mux.Handle("POST /farms/{id}/animal-groups", jwt(http.HandlerFunc(animals.CreateGroup)))
	mux.Handle("GET /animal-groups/{id}", jwt(http.HandlerFunc(animals.GetGroup)))
	mux.Handle("PUT /animal-groups/{id}", jwt(http.HandlerFunc(animals.UpdateGroup)))
	mux.Handle("PATCH /animal-groups/{id}/archive", jwt(http.HandlerFunc(animals.ArchiveGroup)))
	mux.Handle("DELETE /animal-groups/{id}", jwt(http.HandlerFunc(animals.DeleteGroup)))
	mux.Handle("GET /animal-groups/{id}/lifecycle-events", jwt(http.HandlerFunc(animals.ListLifecycle)))
	mux.Handle("POST /animal-groups/{id}/lifecycle-events", jwt(http.HandlerFunc(animals.CreateLifecycle)))
	mux.Handle("DELETE /lifecycle-events/{id}", jwt(http.HandlerFunc(animals.DeleteLifecycle)))

	// Aquaponics loops (Phase 20.8 WS2)
	mux.Handle("GET /farms/{id}/aquaponics-loops", jwt(http.HandlerFunc(aquaponics.ListLoops)))
	mux.Handle("POST /farms/{id}/aquaponics-loops", jwt(http.HandlerFunc(aquaponics.CreateLoop)))
	mux.Handle("GET /aquaponics-loops/{id}", jwt(http.HandlerFunc(aquaponics.GetLoop)))
	mux.Handle("PUT /aquaponics-loops/{id}", jwt(http.HandlerFunc(aquaponics.UpdateLoop)))
	mux.Handle("DELETE /aquaponics-loops/{id}", jwt(http.HandlerFunc(aquaponics.DeleteLoop)))

	mux.Handle("GET /farms/{id}/costs/summary", jwt(http.HandlerFunc(cost.Summary)))
	mux.Handle("GET /farms/{id}/costs/export", jwt(http.HandlerFunc(cost.Export)))
	// Phase 20.7 WS6: per-crop-cycle cost lens (first RAG-precursor view).
	mux.Handle("GET /crop-cycles/{id}/cost-summary", jwt(http.HandlerFunc(cost.CropCycleSummary)))
	mux.Handle("GET /farms/{id}/finance/coa-mappings", jwt(http.HandlerFunc(cost.ListCoaMappings)))
	mux.Handle("PUT /farms/{id}/finance/coa-mappings", jwt(http.HandlerFunc(cost.UpsertCoaMappings)))
	mux.Handle("DELETE /farms/{id}/finance/coa-mappings", jwt(http.HandlerFunc(cost.ResetCoaMappingsAll)))
	mux.Handle("DELETE /farms/{id}/finance/coa-mappings/{category}", jwt(http.HandlerFunc(cost.ResetCoaMappingByCategory)))
	mux.Handle("GET /farms/{id}/costs", jwt(http.HandlerFunc(cost.List)))
	mux.Handle("POST /farms/{id}/costs", jwt(http.HandlerFunc(cost.Create)))
	mux.Handle("PUT /costs/{id}", jwt(http.HandlerFunc(cost.Update)))
	mux.Handle("DELETE /costs/{id}", jwt(http.HandlerFunc(cost.Delete)))
	// Farm energy prices (Phase 20.95 WS2)
	mux.Handle("GET /farms/{id}/energy-prices", jwt(http.HandlerFunc(cost.ListEnergyPrices)))
	mux.Handle("POST /farms/{id}/energy-prices", jwt(http.HandlerFunc(cost.CreateEnergyPrice)))
	mux.Handle("PUT /energy-prices/{id}", jwt(http.HandlerFunc(cost.UpdateEnergyPrice)))
	mux.Handle("DELETE /energy-prices/{id}", jwt(http.HandlerFunc(cost.DeleteEnergyPrice)))
	mux.Handle("POST /farms/{id}/cost-receipts", jwt(http.HandlerFunc(files.UploadCostReceipt)))
	mux.Handle("POST /zones/{id}/photos", jwt(http.HandlerFunc(files.UploadZonePhoto)))
	mux.Handle("GET /zones/{id}/photos", jwt(http.HandlerFunc(files.ListZonePhotos)))
	mux.Handle("DELETE /zones/{id}/photos/{attachment_id}", jwt(http.HandlerFunc(files.DeleteZonePhoto)))
	mux.Handle("GET /file-attachments/{id}/download", jwt(http.HandlerFunc(files.DownloadTarget)))
	mux.Handle("GET /file-attachments/{id}/content", jwt(http.HandlerFunc(files.Download)))

	// Natural farming
	mux.Handle("GET /farms/{id}/naturalfarming/inputs", jwt(http.HandlerFunc(nf.ListInputs)))
	mux.Handle("POST /farms/{id}/naturalfarming/inputs", jwt(http.HandlerFunc(nf.CreateInputDefinition)))
	mux.Handle("PUT /naturalfarming/inputs/{id}", jwt(http.HandlerFunc(nf.UpdateInputDefinition)))
	mux.Handle("DELETE /naturalfarming/inputs/{id}", jwt(http.HandlerFunc(nf.DeleteInputDefinition)))
	mux.Handle("GET /farms/{id}/naturalfarming/batches", jwt(http.HandlerFunc(nf.ListBatches)))
	mux.Handle("POST /farms/{id}/naturalfarming/batches", jwt(http.HandlerFunc(nf.CreateInputBatch)))
	mux.Handle("PUT /naturalfarming/batches/{id}", jwt(http.HandlerFunc(nf.UpdateInputBatch)))
	mux.Handle("DELETE /naturalfarming/batches/{id}", jwt(http.HandlerFunc(nf.DeleteInputBatch)))

	mux.Handle("GET /farms/{id}/naturalfarming/recipes", jwt(http.HandlerFunc(recipe.List)))
	mux.Handle("POST /farms/{id}/naturalfarming/recipes", jwt(http.HandlerFunc(recipe.Create)))
	mux.Handle("GET /naturalfarming/recipes/{id}/components", jwt(http.HandlerFunc(recipe.ListComponents)))
	mux.Handle("POST /naturalfarming/recipes/{id}/components", jwt(http.HandlerFunc(recipe.AddComponent)))
	mux.Handle("DELETE /naturalfarming/recipes/{id}/components/{iid}", jwt(http.HandlerFunc(recipe.RemoveComponent)))
	mux.Handle("GET /naturalfarming/recipes/{id}", jwt(http.HandlerFunc(recipe.Get)))
	mux.Handle("PUT /naturalfarming/recipes/{id}", jwt(http.HandlerFunc(recipe.Update)))
	mux.Handle("DELETE /naturalfarming/recipes/{id}", jwt(http.HandlerFunc(recipe.Delete)))

	// Phase 35 — lighting programs (photoperiod domain)
	mux.Handle("GET /lighting-programs/presets", jwt(http.HandlerFunc(lighting.ListPresets)))
	mux.Handle("GET /farms/{id}/lighting-programs", jwt(http.HandlerFunc(lighting.ListByFarm)))
	mux.Handle("POST /farms/{id}/lighting-programs", jwt(http.HandlerFunc(lighting.CreateProgram)))
	mux.Handle("POST /farms/{id}/lighting-programs/from-preset", jwt(http.HandlerFunc(lighting.CreateFromPreset)))
	mux.Handle("GET /lighting-programs/{pid}", jwt(http.HandlerFunc(lighting.Get)))
	mux.Handle("PATCH /lighting-programs/{pid}", jwt(http.HandlerFunc(lighting.Update)))
	mux.Handle("DELETE /lighting-programs/{pid}", jwt(http.HandlerFunc(lighting.Delete)))
	mux.Handle("POST /lighting-programs/{pid}/activate", jwt(http.HandlerFunc(lighting.Activate)))
	mux.Handle("POST /lighting-programs/{pid}/deactivate", jwt(http.HandlerFunc(lighting.Deactivate)))

	// Phase 35 WS3 — schedule-bound executable actions
	mux.Handle("GET /schedules/{id}/actions", jwt(http.HandlerFunc(automation.ListActionsBySchedule)))
	mux.Handle("POST /schedules/{id}/actions", jwt(http.HandlerFunc(automation.CreateActionForSchedule)))

	// Actuator events by schedule (for Schedules page event history)
	mux.Handle("GET /schedules/{id}/actuator-events", jwt(http.HandlerFunc(actuator.ListEventsBySchedule)))

	// Alerts
	mux.Handle("GET /farms/{id}/alerts", jwt(http.HandlerFunc(alert.ListByFarm)))
	mux.Handle("GET /farms/{id}/alerts/unread-count", jwt(http.HandlerFunc(alert.CountUnread)))
	mux.Handle("PATCH /alerts/{id}/read", jwt(http.HandlerFunc(alert.MarkRead)))
	mux.Handle("PATCH /alerts/{id}/acknowledge", jwt(http.HandlerFunc(alert.MarkAcknowledged)))
	mux.Handle("POST /alerts/{id}/create-task", jwt(http.HandlerFunc(alert.CreateTaskFromAlert)))

	// Profile & farm members
	mux.Handle("GET /profile", jwt(http.HandlerFunc(prof.GetMyProfile)))
	mux.Handle("PUT /profile", jwt(http.HandlerFunc(prof.UpdateMyProfile)))
	mux.Handle("PATCH /profile/hourly-rate", jwt(http.HandlerFunc(prof.PatchMyHourlyRate)))
	mux.Handle("GET /profile/notification-preferences", jwt(http.HandlerFunc(prof.GetNotificationPreferences)))
	mux.Handle("PATCH /profile/notification-preferences", jwt(http.HandlerFunc(prof.PatchNotificationPreferences)))
	mux.Handle("GET /profile/push-tokens", jwt(http.HandlerFunc(prof.ListMyPushTokens)))
	mux.Handle("POST /profile/push-tokens", jwt(http.HandlerFunc(prof.RegisterPushToken)))
	mux.Handle("DELETE /profile/push-tokens", jwt(http.HandlerFunc(prof.UnregisterPushToken)))
	mux.Handle("GET /farms/{id}/members", jwt(http.HandlerFunc(prof.GetFarmMembers)))
	mux.Handle("POST /farms/{id}/members", jwt(http.HandlerFunc(prof.AddFarmMember)))
	mux.Handle("PATCH /farms/{id}/members/{uid}/role", jwt(http.HandlerFunc(prof.UpdateMemberRole)))
	mux.Handle("DELETE /farms/{id}/members/{uid}", jwt(http.HandlerFunc(prof.RemoveMember)))

	// SSE — live sensor readings push
	mux.Handle("GET /farms/{id}/sensors/stream", jwt(http.HandlerFunc(sse.Stream)))
}
