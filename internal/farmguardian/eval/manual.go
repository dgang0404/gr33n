package eval

import (
	"fmt"
	"strings"
)

// PrintManualChecklist writes UI validation steps for the given suite (Phase 131 WS6).
func PrintManualChecklist(suite string) {
	fixtures := FixturesForSuite(suite)
	if len(fixtures) == 0 {
		fmt.Printf("No fixtures for suite %q\n", suite)
		return
	}
	title := strings.TrimSpace(suite)
	if title == "" {
		title = "regression"
	}
	fmt.Printf("# Guardian manual checklist — %s\n\n", title)
	fmt.Println("Run from the dashboard at http://localhost:5173/chat (or Guardian drawer).")
	fmt.Println("One message at a time — wait for each answer before sending the next.")
	fmt.Println()

	for i, q := range fixtures {
		step := i + 1
		model := strings.TrimSpace(q.Model)
		if model == "" {
			model = "phi3:mini"
		}
		contextMode := "Farm counsel ON"
		if !q.Grounded {
			contextMode = "Quick chat (farm context OFF)"
		}
		fmt.Printf("## Step %d — %s\n", step, q.ID)
		fmt.Printf("- **Mode:** %s\n", contextMode)
		fmt.Printf("- **Model:** %s\n", model)
		if q.Grounded {
			fmt.Println("- **Farm:** gr33n Demo Farm (sidebar)")
		}
		fmt.Printf("- **Prompt:** %s\n", q.Prompt)
		if q.Grounded {
			fmt.Println("- **Wait:** Generating… may take many minutes on CPU; watch phase line in chat")
		} else {
			fmt.Println("- **Wait:** Usually faster without farm context")
		}
		fmt.Printf("- **Pass if:** %s\n", manualPassHint(q))
		fmt.Println()
	}
	fmt.Println("Automated equivalent:")
	fmt.Printf("  make guardian-qa-%s MODEL=phi3:mini\n", manualMakeTarget(suite))
}

func manualMakeTarget(suite string) string {
	switch strings.ToLower(strings.TrimSpace(suite)) {
	case "smoke":
		return "smoke"
	case "phase127", "phase128", "p128":
		return "phase127"
	default:
		return "regression"
	}
}

func manualPassHint(q Question) string {
	switch q.ID {
	case "smoke-cherry-forest":
		return "Answer mentions cherry, goldenrod, or blackberry; no timeout"
	case "smoke-morning-walk":
		return "Use Today → Morning check entry point; answer references alerts, zones, or devices; API log may show tool_id=walk_farm"
	case "smoke-unread-alerts":
		return "Answer summarizes seed/demo alerts; len > 40 chars"
	case "smoke-ec-ph":
		return "Citations present or answer mentions EC/pH targets"
	case "regression-cherry-goldenrod-jlf":
		return "Grounded JLF answer with dilution or catalog; extension-method goldenrod (not Cho recipe); API log may show tool_id=suggest_process_from_material"
	case "farm-devices", "p128-devices":
		return "Mentions snapshot device line or online/offline edge devices; no invented GPIO"
	case "farm-fert-schedule", "p128-fert-manual":
		return "Names Outdoor JLF or cites manual-only / schedule posture from snapshot"
	case "fg-demo-pi", "p128-demo-pi":
		return "Cites demo-farm-pi-layout or relay_1 / Veg Relay Controller"
	case "fg-fertigation-triage", "p128-fert-triage":
		return "Cites fertigation-troubleshooting (schedule, Pi, reservoir)"
	default:
		if q.ExpectCitation {
			return "Citations or [1] references in answer"
		}
		if q.ExpectDecline {
			return "Polite decline without inventing farm data"
		}
		if q.ExpectProposal {
			return "Guardian proposes a confirmable change"
		}
		if q.ExpectTool != "" {
			return fmt.Sprintf("API log contains tool_id=%s", q.ExpectTool)
		}
		return "Grounded farm-specific answer without invented zones"
	}
}
