package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"gr33n-api/internal/farmguardian"
)

// ConfirmVerificationInput bundles confirm result + proposal metadata for DB checks.
type ConfirmVerificationInput struct {
	FixtureID string
	Tool      string
	Args      map[string]any
	Result    map[string]any
}

// ConfirmAndVerifyProposal confirms one pending proposal and verifies its DB side effect.
func ConfirmAndVerifyProposal(ctx context.Context, client *APIClient, fixtureID, proposalID string) error {
	proposalID = strings.TrimSpace(proposalID)
	if proposalID == "" {
		return fmt.Errorf("%s: empty proposal_id", fixtureID)
	}
	pending, err := client.FetchPendingProposals(ctx)
	if err != nil {
		return err
	}
	var prop PendingProposal
	found := false
	for _, p := range pending {
		if strings.TrimSpace(p.ProposalID) == proposalID {
			prop = p
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%s: proposal %s not in pending queue before confirm", fixtureID, proposalID)
	}
	log.Printf("confirm: %s proposal=%s tool=%s", fixtureID, proposalID, prop.Tool)
	cr, err := client.ConfirmProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("%s confirm %s: %w", fixtureID, proposalID, err)
	}
	if err := VerifyConfirmSideEffect(ctx, client, ConfirmVerificationInput{
		FixtureID: fixtureID,
		Tool:      prop.Tool,
		Args:      prop.Args,
		Result:    cr.Result,
	}); err != nil {
		return fmt.Errorf("%s post-confirm: %w", fixtureID, err)
	}
	log.Printf("confirm: %s verified DB side effect", fixtureID)
	return nil
}

// ConfirmAndVerifyPassedProposals confirms each passed write-intent proposal and
// asserts the expected DB side effect via confirm result + follow-up GETs.
func ConfirmAndVerifyPassedProposals(
	ctx context.Context,
	client *APIClient,
	fixtures []Question,
	scores []farmguardian.EvalQuestionScore,
) error {
	targets := ProposalConfirmTargets(fixtures, scores)
	if len(targets) == 0 {
		return fmt.Errorf("no passed write-intent proposals to confirm")
	}
	for _, t := range targets {
		if err := ConfirmAndVerifyProposal(ctx, client, t.FixtureID, t.ProposalID); err != nil {
			return err
		}
	}
	return nil
}

// VerifyConfirmSideEffect checks tool result and optional list GETs for write-intent fixtures.
func VerifyConfirmSideEffect(ctx context.Context, client *APIClient, in ConfirmVerificationInput) error {
	if in.Result == nil {
		return fmt.Errorf("empty confirm result")
	}
	switch in.FixtureID {
	case "write-ack":
		return verifyWriteAck(ctx, client, in)
	case "write-feed":
		return verifyWriteFeed(ctx, client, in)
	case "write-schedule":
		return verifyWriteSchedule(ctx, client, in)
	case "write-task":
		return verifyWriteTask(ctx, client, in)
	default:
		return fmt.Errorf("no confirm verification for fixture %q", in.FixtureID)
	}
}

func verifyWriteAck(ctx context.Context, client *APIClient, in ConfirmVerificationInput) error {
	if in.Tool != "ack_alert" && in.Tool != "mark_alert_read" {
		return fmt.Errorf("write-ack: unexpected tool %q", in.Tool)
	}
	if ack, ok := boolFromAny(in.Result["is_acknowledged"]); in.Tool == "ack_alert" && (!ok || !ack) {
		return fmt.Errorf("write-ack: confirm result missing is_acknowledged=true (got %#v)", in.Result)
	}
	alertID, err := int64FromAny(in.Args["alert_id"])
	if err != nil {
		return fmt.Errorf("write-ack: %w", err)
	}
	alerts, err := client.ListFarmAlerts(ctx)
	if err != nil {
		return err
	}
	for _, a := range alerts {
		if int64FromMap(a, "id") == alertID {
			if in.Tool == "ack_alert" {
				if !boolFromMap(a, "is_acknowledged") {
					return fmt.Errorf("write-ack: alert %d not acknowledged in GET /farms/.../alerts", alertID)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("write-ack: alert %d not found in farm alerts list", alertID)
}

func verifyWriteFeed(ctx context.Context, client *APIClient, in ConfirmVerificationInput) error {
	if in.Tool != "patch_fertigation_program" {
		return fmt.Errorf("write-feed: unexpected tool %q", in.Tool)
	}
	programID, err := int64FromAny(in.Args["program_id"])
	if err != nil {
		return fmt.Errorf("write-feed: %w", err)
	}
	wantVol, err := float64FromAny(in.Args["total_volume_liters"])
	if err != nil {
		return fmt.Errorf("write-feed: proposal args missing total_volume_liters")
	}
	programs, err := client.ListFertigationPrograms(ctx)
	if err != nil {
		return err
	}
	for _, p := range programs {
		if int64FromMap(p, "id") != programID {
			continue
		}
		got, err := float64FromAny(p["total_volume_liters"])
		if err != nil {
			return fmt.Errorf("write-feed: program %d missing total_volume_liters in API", programID)
		}
		if math.Abs(got-wantVol) > 0.05 {
			return fmt.Errorf("write-feed: program %d volume %.3f want ~%.3f L", programID, got, wantVol)
		}
		return nil
	}
	return fmt.Errorf("write-feed: program %d not found", programID)
}

func verifyWriteSchedule(ctx context.Context, client *APIClient, in ConfirmVerificationInput) error {
	if in.Tool != "patch_schedule" {
		return fmt.Errorf("write-schedule: unexpected tool %q", in.Tool)
	}
	scheduleID, err := int64FromAny(in.Args["schedule_id"])
	if err != nil {
		return fmt.Errorf("write-schedule: %w", err)
	}
	wantActive, hasActive := boolFromAny(in.Args["is_active"])
	if hasActive && wantActive {
		return fmt.Errorf("write-schedule: expected pause proposal (is_active=false)")
	}
	if active, ok := boolFromAny(in.Result["is_active"]); ok && active {
		return fmt.Errorf("write-schedule: confirm result still active")
	}
	schedules, err := client.ListFarmSchedules(ctx)
	if err != nil {
		return err
	}
	for _, s := range schedules {
		if int64FromMap(s, "id") == scheduleID {
			if boolFromMap(s, "is_active") {
				return fmt.Errorf("write-schedule: schedule %d still active after confirm", scheduleID)
			}
			return nil
		}
	}
	return fmt.Errorf("write-schedule: schedule %d not found", scheduleID)
}

func verifyWriteTask(ctx context.Context, client *APIClient, in ConfirmVerificationInput) error {
	switch in.Tool {
	case "create_task", "create_task_from_alert":
	default:
		return fmt.Errorf("write-task: unexpected tool %q", in.Tool)
	}
	taskID, err := int64FromAny(in.Result["task_id"])
	if err != nil {
		return fmt.Errorf("write-task: confirm result missing task_id: %w", err)
	}
	tasks, err := client.ListFarmTasks(ctx)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if int64FromMap(t, "id") == taskID {
			return nil
		}
	}
	return fmt.Errorf("write-task: task %d not found in GET /farms/.../tasks", taskID)
}

func (c *APIClient) getJSON(ctx context.Context, path string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+path, nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s HTTP %d: %s", path, resp.StatusCode, truncate(string(body), 200))
	}
	var out any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *APIClient) ListFarmAlerts(ctx context.Context) ([]map[string]any, error) {
	if c.FarmID <= 0 {
		return nil, fmt.Errorf("farm_id required")
	}
	raw, err := c.getJSON(ctx, fmt.Sprintf("/farms/%d/alerts", c.FarmID))
	if err != nil {
		return nil, err
	}
	return sliceOfMaps(raw), nil
}

func (c *APIClient) ListFertigationPrograms(ctx context.Context) ([]map[string]any, error) {
	if c.FarmID <= 0 {
		return nil, fmt.Errorf("farm_id required")
	}
	raw, err := c.getJSON(ctx, fmt.Sprintf("/farms/%d/fertigation/programs", c.FarmID))
	if err != nil {
		return nil, err
	}
	return sliceOfMaps(raw), nil
}

func (c *APIClient) ListFarmSchedules(ctx context.Context) ([]map[string]any, error) {
	if c.FarmID <= 0 {
		return nil, fmt.Errorf("farm_id required")
	}
	raw, err := c.getJSON(ctx, fmt.Sprintf("/farms/%d/schedules", c.FarmID))
	if err != nil {
		return nil, err
	}
	return sliceOfMaps(raw), nil
}

func (c *APIClient) ListFarmTasks(ctx context.Context) ([]map[string]any, error) {
	if c.FarmID <= 0 {
		return nil, fmt.Errorf("farm_id required")
	}
	raw, err := c.getJSON(ctx, fmt.Sprintf("/farms/%d/tasks", c.FarmID))
	if err != nil {
		return nil, err
	}
	return sliceOfMaps(raw), nil
}

func sliceOfMaps(raw any) []map[string]any {
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func int64FromMap(m map[string]any, key string) int64 {
	v, _ := int64FromAny(m[key])
	return v
}

func boolFromMap(m map[string]any, key string) bool {
	v, _ := boolFromAny(m[key])
	return v
}

func int64FromAny(v any) (int64, error) {
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case json.Number:
		return n.Int64()
	default:
		return 0, fmt.Errorf("invalid int64 %#v", v)
	}
}

func float64FromAny(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int64:
		return float64(n), nil
	case int:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	case string:
		return strconv.ParseFloat(n, 64)
	default:
		return 0, fmt.Errorf("invalid float64 %#v", v)
	}
}

func boolFromAny(v any) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	default:
		return false, false
	}
}
