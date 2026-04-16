package insertcommonsschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SchemaVersion must match the farm-side sender and Insert Commons receiver.
// When evolving the contract, add a new constant (e.g. v2), keep the receiver accepting old
// versions during a documented cutover window, then migrate farms — see insert-commons-pipeline-runbook.md.
const SchemaVersion = "gr33n.insert_commons.v1"

// PackageVersionV1 wraps an ingest payload for file export / archival (Phase 14 WS2).
const PackageVersionV1 = "gr33n.insert_commons.package.v1"

// allowedIngestRootFields is the canonical top-level shape for ingest JSON. Extra keys are rejected
// so the pipeline stays scrubbed and forward-compatible only via explicit schema bumps.
var allowedIngestRootFields = map[string]struct{}{
	"schema_version": {},
	"generated_at":   {},
	"farm_pseudonym": {},
	"farm_profile":   {},
	"aggregates":     {},
	"privacy":        {},
}

// ValidatePayload checks Insert Commons v1 JSON (receiver and pre-flight on the sender).
func ValidatePayload(raw []byte) (farmPseudo string, genAt time.Time, err error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", time.Time{}, fmt.Errorf("invalid json: %w", err)
	}
	for k := range root {
		if _, ok := allowedIngestRootFields[k]; !ok {
			return "", time.Time{}, fmt.Errorf("unknown top-level field %q", k)
		}
	}
	required := []string{"schema_version", "generated_at", "farm_pseudonym", "farm_profile", "aggregates", "privacy"}
	for _, k := range required {
		if _, ok := root[k]; !ok {
			return "", time.Time{}, fmt.Errorf("missing required field: %s", k)
		}
	}
	var ver string
	if err := json.Unmarshal(root["schema_version"], &ver); err != nil {
		return "", time.Time{}, errors.New("invalid schema_version")
	}
	if strings.TrimSpace(ver) != SchemaVersion {
		return "", time.Time{}, fmt.Errorf("unsupported schema_version %q (expected %s)", ver, SchemaVersion)
	}
	var genStr string
	if err := json.Unmarshal(root["generated_at"], &genStr); err != nil {
		return "", time.Time{}, errors.New("invalid generated_at")
	}
	genStr = strings.TrimSpace(genStr)
	var parseErr error
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		genAt, parseErr = time.Parse(layout, genStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		return "", time.Time{}, errors.New("generated_at must be RFC3339 or RFC3339Nano")
	}
	if genStr == "" {
		return "", time.Time{}, errors.New("generated_at is empty")
	}
	now := time.Now().UTC()
	if genAt.After(now.Add(10 * time.Minute)) {
		return "", time.Time{}, errors.New("generated_at is too far in the future")
	}
	if genAt.Before(now.Add(-365 * 24 * time.Hour)) {
		return "", time.Time{}, errors.New("generated_at is too old")
	}

	if err := json.Unmarshal(root["farm_pseudonym"], &farmPseudo); err != nil {
		return "", time.Time{}, errors.New("invalid farm_pseudonym")
	}
	farmPseudo = strings.TrimSpace(farmPseudo)
	if farmPseudo == "" {
		return "", time.Time{}, errors.New("farm_pseudonym is required")
	}

	var fp map[string]any
	if err := json.Unmarshal(root["farm_profile"], &fp); err != nil {
		return "", time.Time{}, errors.New("farm_profile must be an object")
	}
	for _, k := range []string{"scale_tier", "timezone_bucket", "currency", "operational_status"} {
		if _, ok := fp[k]; !ok {
			return "", time.Time{}, fmt.Errorf("farm_profile missing %q", k)
		}
	}

	var agg map[string]any
	if err := json.Unmarshal(root["aggregates"], &agg); err != nil {
		return "", time.Time{}, errors.New("aggregates must be an object")
	}
	for _, k := range []string{"costs", "tasks", "devices"} {
		v, ok := agg[k]
		if !ok {
			return "", time.Time{}, fmt.Errorf("aggregates missing %q", k)
		}
		if _, isObj := v.(map[string]any); !isObj {
			return "", time.Time{}, fmt.Errorf("aggregates.%s must be an object", k)
		}
	}
	costs, _ := agg["costs"].(map[string]any)
	if _, ok := costs["totals"]; !ok {
		return "", time.Time{}, errors.New("aggregates.costs missing totals")
	}
	if _, ok := costs["totals"].(map[string]any); !ok {
		return "", time.Time{}, errors.New("aggregates.costs.totals must be an object")
	}
	if rawBC, ok := costs["by_category"]; !ok {
		return "", time.Time{}, errors.New("aggregates.costs missing by_category")
	} else if _, isArr := rawBC.([]any); !isArr {
		return "", time.Time{}, errors.New("aggregates.costs.by_category must be an array")
	}
	tasks, _ := agg["tasks"].(map[string]any)
	if _, ok := tasks["by_status"]; !ok {
		return "", time.Time{}, errors.New("aggregates.tasks missing by_status")
	}
	if _, ok := tasks["by_status"].(map[string]any); !ok {
		return "", time.Time{}, errors.New("aggregates.tasks.by_status must be an object")
	}
	devices, _ := agg["devices"].(map[string]any)
	if _, ok := devices["by_status"]; !ok {
		return "", time.Time{}, errors.New("aggregates.devices missing by_status")
	}
	if _, ok := devices["by_status"].(map[string]any); !ok {
		return "", time.Time{}, errors.New("aggregates.devices.by_status must be an object")
	}

	var priv map[string]any
	if err := json.Unmarshal(root["privacy"], &priv); err != nil {
		return "", time.Time{}, errors.New("privacy must be an object")
	}
	rawPII, ok := priv["includes_pii"]
	if !ok {
		return "", time.Time{}, errors.New("privacy.includes_pii is required")
	}
	switch rawPII.(type) {
	case bool:
	default:
		return "", time.Time{}, errors.New("privacy.includes_pii must be a boolean")
	}

	return farmPseudo, genAt.UTC(), nil
}
