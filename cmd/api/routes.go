package main

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	automationworker "gr33n-api/internal/automation"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/filestorage"
	actuatorhandler "gr33n-api/internal/handler/actuator"
	alerthandler "gr33n-api/internal/handler/alert"
	audithandler "gr33n-api/internal/handler/audit"
	authhandler "gr33n-api/internal/handler/auth"
	automationhandler "gr33n-api/internal/handler/automation"
	commonscataloghandler "gr33n-api/internal/handler/commonscatalog"
	costhandler "gr33n-api/internal/handler/cost"
	cropcyclehandler "gr33n-api/internal/handler/cropcycle"
	devicehandler "gr33n-api/internal/handler/device"
	farmhandler "gr33n-api/internal/handler/farm"
	fertigationhandler "gr33n-api/internal/handler/fertigation"
	fileattachhandler "gr33n-api/internal/handler/fileattach"
	nfhandler "gr33n-api/internal/handler/naturalfarming"
	organizationhandler "gr33n-api/internal/handler/organization"
	profilehandler "gr33n-api/internal/handler/profile"
	recipehandler "gr33n-api/internal/handler/recipe"
	sensorhandler "gr33n-api/internal/handler/sensor"
	ssehandler "gr33n-api/internal/handler/sse"
	animalhandler "gr33n-api/internal/handler/animal"
	aquaponicshandler "gr33n-api/internal/handler/aquaponics"
	planthandler "gr33n-api/internal/handler/plants"
	raghandler "gr33n-api/internal/handler/rag"
	setpointhandler "gr33n-api/internal/handler/setpoint"
	taskhandler "gr33n-api/internal/handler/task"
	zonehandler "gr33n-api/internal/handler/zone"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/pushnotify"
)

func registerRoutes(mux *http.ServeMux, pool *pgxpool.Pool, worker *automationworker.Worker, pushDispatch *pushnotify.Dispatcher, adminUser string, adminHash []byte, hashFilePath string, fileStore filestorage.Store, fileCfg filestorage.Config, adminBindUserID uuid.UUID, adminBindEmail string) {
	farm := farmhandler.NewHandler(pool)
	org := organizationhandler.NewHandler(pool)
	audit := audithandler.NewHandler(pool)
	zone := zonehandler.NewHandler(pool)
	device := devicehandler.NewHandler(pool)
	actuator := actuatorhandler.NewHandler(pool)
	automation := automationhandler.NewHandler(pool, worker)
	sse := ssehandler.NewHandler(pool)
	if pushDispatch == nil {
		pushDispatch = pushnotify.NewDispatcher(pool)
	}
	sensor := sensorhandler.NewHandler(pool, sse, pushDispatch)
	task := taskhandler.NewHandler(pool)
	fertigation := fertigationhandler.NewHandler(pool)
	nf := nfhandler.NewHandler(pool)
	recipe := recipehandler.NewHandler(pool)
	cropcycle := cropcyclehandler.NewHandler(pool)
	rag := raghandler.NewHandler(pool)
	plants := planthandler.NewHandler(pool)
	animals := animalhandler.NewHandler(pool)
	aquaponics := aquaponicshandler.NewHandler(pool)
	alert := alerthandler.NewHandler(pool)
	prof := profilehandler.NewHandler(pool)
	setpoint := setpointhandler.NewHandler(pool)
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
	mux.HandleFunc("POST /auth/register", auth.Register)
	mux.HandleFunc("GET /auth/mode", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"mode": authMode})
	})

	// ── Pi routes — API key required ─────────────────────────────────────────
	mux.Handle("POST /sensors/{id}/readings", requireAPIKey(http.HandlerFunc(sensor.PostReading)))
	mux.Handle("POST /sensors/readings/batch", requireAPIKey(http.HandlerFunc(sensor.PostReadingsBatch)))
	mux.Handle("PATCH /devices/{id}/status", requireAPIKey(http.HandlerFunc(device.UpdateStatus)))
	mux.Handle("POST /actuators/{id}/events", requireAPIKey(http.HandlerFunc(actuator.RecordEvent)))
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
	mux.Handle("GET /farms/{id}/devices", requireJWTOrPiEdge(http.HandlerFunc(device.ListByFarm)))
	mux.Handle("GET /farms/{id}/actuators", jwt(http.HandlerFunc(actuator.ListByFarm)))
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

	// Stage-scoped setpoints (Phase 20.6 WS2)
	mux.Handle("GET /farms/{id}/setpoints", jwt(http.HandlerFunc(setpoint.List)))
	mux.Handle("POST /farms/{id}/setpoints", jwt(http.HandlerFunc(setpoint.Create)))
	mux.Handle("GET /setpoints/{id}", jwt(http.HandlerFunc(setpoint.Get)))
	mux.Handle("PUT /setpoints/{id}", jwt(http.HandlerFunc(setpoint.Update)))
	mux.Handle("DELETE /setpoints/{id}", jwt(http.HandlerFunc(setpoint.Delete)))

	// Sensors
	mux.Handle("GET /sensors/{id}", jwt(http.HandlerFunc(sensor.Get)))
	mux.Handle("POST /farms/{id}/sensors", jwt(http.HandlerFunc(sensor.Create)))
	mux.Handle("PUT /sensors/{id}", jwt(http.HandlerFunc(sensor.Update)))
	mux.Handle("DELETE /sensors/{id}", jwt(http.HandlerFunc(sensor.Delete)))
	mux.Handle("GET /sensors/{id}/readings/latest", jwt(http.HandlerFunc(sensor.LatestReading)))
	mux.Handle("GET /sensors/{id}/readings/stats", jwt(http.HandlerFunc(sensor.ReadingStats)))
	mux.Handle("GET /sensors/{id}/readings", jwt(http.HandlerFunc(sensor.ListReadings)))

	// Devices
	mux.Handle("GET /devices/{id}", jwt(http.HandlerFunc(device.Get)))
	mux.Handle("POST /farms/{id}/devices", jwt(http.HandlerFunc(device.Create)))
	mux.Handle("DELETE /devices/{id}", jwt(http.HandlerFunc(device.Delete)))
	mux.Handle("PATCH /actuators/{id}/state", jwt(http.HandlerFunc(actuator.UpdateState)))
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

	mux.Handle("GET /farms/{id}/crop-cycles", jwt(http.HandlerFunc(cropcycle.List)))
	mux.Handle("POST /farms/{id}/crop-cycles", jwt(http.HandlerFunc(cropcycle.Create)))
	mux.Handle("PATCH /crop-cycles/{id}/stage", jwt(http.HandlerFunc(cropcycle.UpdateStage)))
	mux.Handle("GET /crop-cycles/{id}", jwt(http.HandlerFunc(cropcycle.Get)))
	mux.Handle("PUT /crop-cycles/{id}", jwt(http.HandlerFunc(cropcycle.Update)))
	mux.Handle("DELETE /crop-cycles/{id}", jwt(http.HandlerFunc(cropcycle.Delete)))

	// Plants (crop tracking)
	mux.Handle("GET /farms/{id}/plants", jwt(http.HandlerFunc(plants.List)))
	mux.Handle("POST /farms/{id}/plants", jwt(http.HandlerFunc(plants.Create)))
	mux.Handle("GET /plants/{id}", jwt(http.HandlerFunc(plants.Get)))
	mux.Handle("PUT /plants/{id}", jwt(http.HandlerFunc(plants.Update)))
	mux.Handle("DELETE /plants/{id}", jwt(http.HandlerFunc(plants.Delete)))

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
