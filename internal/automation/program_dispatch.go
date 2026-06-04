package automation

// programDispatchCtx carries schedule linkage and provenance for program-driven
// actuator events and device commands. Cron ticks set scheduleID; run-now sets
// manualRun so sources read operator/API instead of schedule_trigger.
type programDispatchCtx struct {
	scheduleID *int64
	manualRun  bool
}
