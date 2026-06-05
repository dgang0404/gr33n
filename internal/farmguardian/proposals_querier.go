package farmguardian

import (
	"context"

	db "gr33n-api/internal/db"
)

// farmMatchQuerier is the narrow DB surface shared by feeding and comfort matchers.
type farmMatchQuerier interface {
	ListZonesByFarm(ctx context.Context, farmID int64) ([]db.Gr33ncoreZone, error)
	ListProgramsByFarm(ctx context.Context, farmID int64) ([]db.Gr33nfertigationProgram, error)
	ListSchedulesByFarm(ctx context.Context, farmID int64) ([]db.Gr33ncoreSchedule, error)
}

// comfortMatchQuerier adds automation rule + sensor lookups for Phase 42 matchers.
type comfortMatchQuerier interface {
	farmMatchQuerier
	ListAutomationRulesByFarm(ctx context.Context, farmID int64) ([]db.Gr33ncoreAutomationRule, error)
	ListSensorsByFarm(ctx context.Context, farmID int64) ([]db.Gr33ncoreSensor, error)
}
