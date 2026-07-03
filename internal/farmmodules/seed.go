package farmmodules

import (
	"context"

	db "gr33n-api/internal/db"
)

const (
	SchemaCrops           = "gr33ncrops"
	SchemaNaturalFarming  = "gr33nnaturalfarming"
	SchemaAnimals         = "gr33nanimals"
	SchemaAquaponics      = "gr33naquaponics"
)

var defaultModules = []struct {
	Schema  string
	Enabled bool
}{
	{SchemaCrops, true},
	{SchemaNaturalFarming, true},
	{SchemaAnimals, false},
	{SchemaAquaponics, false},
}

// SeedDefaults inserts default module rows for a new farm (idempotent).
func SeedDefaults(ctx context.Context, q db.Querier, farmID int64) error {
	for _, m := range defaultModules {
		if err := q.SeedFarmActiveModule(ctx, db.SeedFarmActiveModuleParams{
			FarmID:           farmID,
			ModuleSchemaName: m.Schema,
			IsEnabled:        m.Enabled,
		}); err != nil {
			return err
		}
	}
	return nil
}
