package farmguardian

import (
	"fmt"
	"strings"
)

// ImpactSummary returns ordered, plain-language lines describing what confirming a
// proposal will do — the "if you Confirm, this will…" block (Phase 34 WS4). The
// first line leads with the most impactful/irreversible effect for high-risk tools.
// Operator-supplied facts are appended, explicitly labeled as operator-stated.
func ImpactSummary(toolID string, args map[string]any, facts []OperatorFact) []string {
	steps := impactSteps(toolID, args)
	for _, f := range facts {
		label := f.Label
		if label == "" {
			label = operatorFactLabel(f)
		}
		steps = append(steps, "Assumes "+label)
	}
	return steps
}

func impactSteps(toolID string, args map[string]any) []string {
	switch toolID {
	case "apply_grow_setup_pack":
		return setupPackImpact(args)
	case "patch_fertigation_program":
		return patchFertigationImpact(args)
	case "create_plant":
		name := argString(args, "display_name")
		return []string{joinName("Create plant", name) + " (editable later)"}
	case "create_crop_cycle":
		name := argString(args, "name")
		return []string{joinName("Start crop cycle", name) + " in the selected zone (no harvest data yet)"}
	case "create_fertigation_program":
		return []string{joinName("Create fertigation program", argString(args, "name")) +
			fertigationHints(args) + " — no run triggered now"}
	case "create_lighting_program":
		preset := argString(args, "preset_key")
		line := "Create lighting program"
		if preset != "" {
			line += " from preset " + preset
		}
		if z := argString(args, "zone_id"); z != "" {
			line += " for zone " + z
		}
		return []string{line + " — ON/OFF schedules and actuator actions will be created"}
	case "create_task_from_alert":
		return createTaskFromAlertImpact(args)
	case "create_task":
		return []string{joinName("Create task", argString(args, "title")) + " (reversible — you can complete or delete it)"}
	case "update_cycle_stage":
		return []string{"Update the crop cycle growth stage to " + argString(args, "current_stage") + " (reversible)"}
	case "ack_alert":
		return []string{"Acknowledge the alert (marks it handled; reversible)"}
	case "mark_alert_read":
		return []string{"Mark the alert as read (reversible)"}
	case "patch_schedule":
		name := argString(args, "schedule_name")
		if active, ok := args["is_active"].(bool); ok && !active {
			return []string{joinName("Pause schedule", name) + " — no automatic runs until re-enabled"}
		}
		if active, ok := args["is_active"].(bool); ok && active {
			return []string{joinName("Enable schedule", name) + " — automatic runs resume"}
		}
		return []string{joinName("Update schedule", name) + " (reversible)"}
	case "patch_rule":
		name := argString(args, "rule_name")
		if active, ok := args["is_active"].(bool); ok && !active {
			return []string{joinName("Pause automation rule", name) + " — it will stop firing until re-enabled"}
		}
		if active, ok := args["is_active"].(bool); ok && active {
			return []string{joinName("Enable automation rule", name) + " — it will resume firing when conditions match"}
		}
		return []string{joinName("Update automation rule", name) + " (reversible)"}
	case "enqueue_actuator_command":
		cmd := argString(args, "command")
		name := argString(args, "actuator_name")
		line := "Queue a hardware command"
		if cmd != "" {
			line = fmt.Sprintf("Queue %q", cmd)
		}
		if name != "" {
			line += " for " + name
		}
		return []string{line + " — the Pi fires the relay on its next poll"}
	case "apply_bootstrap_template":
		return []string{"Apply a farm bootstrap template — creates zones, schedules, and starter config"}
	default:
		return nil
	}
}

func createTaskFromAlertImpact(args map[string]any) []string {
	subject := argString(args, "alert_subject")
	title := argString(args, "title")
	src := argString(args, "alert_source_type")
	suffix := " (reversible — you can complete or delete it)"
	if src == "inventory_low_stock" {
		name := lowStockInputFromSubject(subject)
		if name == "" {
			name = title
		}
		line := "Create refill task"
		if name != "" {
			line += " for " + name
		}
		if subject != "" {
			line += " — " + subject
		}
		return []string{line + " — restock in Supplies hub; Guardian cannot change batch quantities" + suffix}
	}
	if subject != "" {
		return []string{joinName("Create task from alert", subject) + suffix}
	}
	return []string{joinName("Create task", title) + suffix}
}

func setupPackImpact(args map[string]any) []string {
	steps := []string{}
	zone := argString(args, "zone_name")
	if plant, ok := args["plant"].(map[string]any); ok {
		if name := argString(plant, "display_name"); name != "" {
			steps = append(steps, joinName("Create plant", name))
		}
	}
	if cycle, ok := args["cycle"].(map[string]any); ok {
		name := argString(cycle, "name")
		line := "Start a crop cycle"
		if name != "" {
			line += " “" + name + "”"
		}
		if zone != "" {
			line += " in " + zone
		}
		if stage := argString(cycle, "current_stage"); stage != "" {
			line += " (stage " + stage + ")"
		}
		steps = append(steps, line)
	}
	if program, ok := args["program"].(map[string]any); ok {
		name := argString(program, "name")
		line := "Create fertigation program"
		if name != "" {
			line += " “" + name + "”"
		}
		line += fertigationHints(program) + " — no run triggered now"
		steps = append(steps, line)
	}
	if task, ok := args["optional_task"].(map[string]any); ok {
		if title := argString(task, "title"); title != "" {
			steps = append(steps, joinName("Create follow-up task", title))
		}
	}
	return steps
}

func patchFertigationImpact(args map[string]any) []string {
	parts := []string{}
	if v, ok := argFloat(args, "ec_target_id"); ok {
		_ = v // ec target is an id ref; surfaced generically below
	}
	if v, ok := argFloat(args, "total_volume_liters"); ok {
		parts = append(parts, "volume → "+formatLiters(v))
	}
	if v, ok := argFloat(args, "ec_trigger_low"); ok {
		parts = append(parts, "EC target → "+formatEC(v))
	}
	if active, ok := args["is_active"].(bool); ok {
		if active {
			parts = append(parts, "set active")
		} else {
			parts = append(parts, "set inactive")
		}
	}
	if v, ok := args["irrigation_only"].(bool); ok && v {
		parts = append(parts, "water-only irrigation")
	}
	if len(parts) == 0 {
		return []string{"Update the feeding plan (no run triggered now)"}
	}
	return []string{"Update feeding plan: " + strings.Join(parts, ", ") + " (no run triggered now)"}
}

func fertigationHints(program map[string]any) string {
	hints := []string{}
	if v, ok := argFloat(program, "total_volume_liters"); ok {
		hints = append(hints, formatLiters(v))
	}
	if v, ok := argFloat(program, "ec_trigger_low"); ok {
		hints = append(hints, "EC "+formatEC(v))
	}
	lo, okLo := argFloat(program, "ph_trigger_low")
	hi, okHi := argFloat(program, "ph_trigger_high")
	if okLo && okHi {
		hints = append(hints, "pH "+formatPH(lo)+"–"+formatPH(hi))
	} else if okLo {
		hints = append(hints, "pH low "+formatPH(lo))
	}
	if len(hints) == 0 {
		return ""
	}
	return " (" + strings.Join(hints, " / ") + ")"
}

func operatorFactLabel(f OperatorFact) string {
	val := fmt.Sprintf("%v", f.Value)
	base := strings.TrimSpace(f.Field + " " + val)
	if base == "" {
		base = "operator-stated value"
	}
	return base + " (operator-stated, not measured)"
}

func joinName(verb, name string) string {
	if strings.TrimSpace(name) == "" {
		return verb
	}
	return verb + " " + name
}

func argString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func argFloat(m map[string]any, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	switch n := m[key].(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
