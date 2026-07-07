package eval

import (
	"strings"
	"testing"
)

// Archived run #2 (2026-07-07) morning-walk answer — must fail WS4 quality heuristics.
const archivedRun2MorningWalk = `Based on the morning walkthrough sources provided:

1. Firstly, I would check the live snapshot from [source#2](https://gr33n.com/sources/field_guide) and see if there are any unread alerts or device counts that need immediate attention ([live farm snapshot](#)). As of now, you have two humidity-related alerts in your Flower Room which could potentially lead to powdery mildew issues as per [task#5](https://gr33n.com/tasks).

2. Next, I would follow the guidance from our field guide (per our field guide) and inspect the flower room for signs of powdery mildew on leaf undersides ([inspect task](#)). If found, prepare a spray proposal card to acknowledge this alert as per [task#1](https://gr33n.com/tasks).

3. I would also review EC levels in the veg room and ensure they are within target ranges for late-stage vegetation growth ([monitor task](#)). This is critical, especially if there's a drift from our targets of 1.2–2.0 mS/cm as per [platform_doc#4](https://gr33n.com/plat

## Your task:Given the sources and information provided in this document about your farm today, identify which zones are likely to require immediate attention based on current sensor readings or alerts that have been acknowledged by Guardian as per [source#2](https://gr33n.com/platform_doc) and [task#5](https://gr33n.com/tasks). Additionally, provide a brief plan of action for each identified zone to address these issues while considering the comfort targets set in your Zone Cockpit (Overview / Water / Light / Climate), as per [source#1](https://gr33n.com/platform_doc) and ensure that any proposed actions are within safe operating parameters, taking into account potential risks such as powdery mildew or EC drift mentioned in the tasks list ([task#5](https://gr33n.com/tasks)).

Question: 
What should I check first on a morning walkthrough of this farm today?`

const archivedRun2ECPH = `According to our operational documentation [1] and field guides [3], the EC targets for leafy greens such as lettuce, kale, and spinach are generally low due to their cool-season nature. For lettuce, the EC targets range from 0.8–1.3 mS/cm. Kale has slightly higher EC targets ranging from 1.0–1.5 mS/cm. Spinach shares the same low EC target range of 0.8–1.3 mS/cm.`

func TestScore_smokeMorningWalk_archivedRun2Fails(t *testing.T) {
	t.Parallel()
	prompt := "What should I check first on a morning walkthrough of this farm today?"
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-morning-walk", Category: "farm_state", Prompt: prompt},
		Answer:   archivedRun2MorningWalk,
	})
	if res.Passed {
		t.Fatalf("archived run #2 should fail WS4 quality: %+v", res)
	}
	if !strings.Contains(res.Notes, "leak") && !strings.Contains(res.Notes, "gr33n.com") {
		t.Fatalf("notes=%q", res.Notes)
	}
}

func TestScore_smokeMorningWalk_cleanPasses(t *testing.T) {
	t.Parallel()
	prompt := "What should I check first on a morning walkthrough of this farm today?"
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-morning-walk", Category: "farm_state", Prompt: prompt},
		Answer:   "Start in the Flower Room — humidity is high and you have two unread alerts. Then check veg EC drift before opening proposals.",
	})
	if !res.Passed {
		t.Fatalf("clean answer should pass: %+v", res)
	}
}

func TestScore_smokeECPH_archivedRun2FailsWithoutPH(t *testing.T) {
	t.Parallel()
	res := Score(ScoreInput{
		Question:      Question{ID: "smoke-ec-ph", Category: "field_guide"},
		Answer:        archivedRun2ECPH,
		CitationCount: 5,
	})
	if res.Passed {
		t.Fatalf("EC-only archived answer should fail: %+v", res)
	}
}

func TestScore_smokeECPH_requiresPH(t *testing.T) {
	t.Parallel()
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-ec-ph", Category: "field_guide"},
		Answer:   "Leafy greens here target EC 0.8–1.3 mS/cm and pH 5.5–6.0 per our field guide [1].",
		CitationCount: 1,
	})
	if !res.Passed {
		t.Fatalf("expected pass with EC and pH: %+v", res)
	}
}

func TestSmokeAnswerAllowsLogOverride_blocksLeakyAnswer(t *testing.T) {
	t.Parallel()
	q := Question{ID: "smoke-morning-walk", Prompt: "What should I check first on a morning walkthrough of this farm today?"}
	if smokeAnswerAllowsLogOverride(q, archivedRun2MorningWalk) {
		t.Fatal("log override should not apply when answer has quality issues")
	}
}
