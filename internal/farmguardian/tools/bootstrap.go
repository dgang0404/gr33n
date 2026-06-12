package tools

import (
	"context"
	"errors"
	"fmt"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmbootstrap"
	"gr33n-api/internal/platform/bootstraptemplates"
)

func execApplyBootstrapTemplate(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	template, err := stringFromArgs(args, "template")
	if err != nil {
		return nil, err
	}
	if farmbootstrap.IsBlankChoice(template) {
		return nil, errors.New("template is required")
	}
	if !bootstraptemplates.Current().IsValid(template) {
		return nil, fmt.Errorf("unknown template %q — use a key from GET /platform/bootstrap-templates", template)
	}
	q, ok := deps.Q.(*db.Queries)
	if !ok {
		return nil, errors.New("bootstrap requires database queries")
	}
	out, err := q.ApplyFarmBootstrapTemplate(ctx, deps.FarmID, template)
	if err != nil {
		return nil, err
	}
	if errKey, _ := out["error"].(string); errKey == "unknown_template" {
		return nil, fmt.Errorf("unknown template %q", template)
	} else if errKey == "farm_not_found" {
		return nil, errors.New("farm not found")
	}
	return map[string]any{
		"farm_id":   deps.FarmID,
		"template":  template,
		"bootstrap": out,
	}, nil
}
