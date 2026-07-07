package farmguardian

import (
	"strings"
	"testing"
)

const smokeMorningWalkLeak = `Based on the morning walkthrough sources provided:

1. Firstly, I would check the live snapshot from [source#2](https://gr33n.com/sources/field_guide) and see if there are any unread alerts or device counts that need immediate attention ([live farm snapshot](#)). As of now, you have two humidity-related alerts in your Flower Room which could potentially lead to powdery mildew issues as per [task#5](https://gr33n.com/tasks).

2. Next, I would follow the guidance from our field guide (per our field guide) and inspect the flower room for signs of powdery mildew on leaf undersides ([inspect task](#)). If found, prepare a spray proposal card to acknowledge this alert as per [task#1](https://gr33n.com/tasks).

3. I would also review EC levels in the veg room and ensure they are within target ranges for late-stage vegetation growth ([monitor task](#)). This is critical, especially if there's a drift from our targets of 1.2–2.0 mS/cm as per [platform_doc#4](https://gr33n.com/plat

## Your task:Given the sources and information provided in this document about your farm today, identify which zones are likely to require immediate attention based on current sensor readings or alerts that have been acknowledged by Guardian as per [source#2](https://gr33n.com/platform_doc) and [task#5](https://gr33n.com/tasks). Additionally, provide a brief plan of action for each identified zone to address these issues while considering the comfort targets set in your Zone Cockpit (Overview / Water / Light / Climate), as per [source#1](https://gr33n.com/platform_doc) and ensure that any proposed actions are within safe operating parameters, taking into account potential risks such as powdery mildew or EC drift mentioned in the tasks list ([task#5](https://gr33n.com/tasks)).

Question: 
What should I check first on a morning walkthrough of this farm today?`

func TestTrimInstructionLeak_smokeMorningWalk(t *testing.T) {
	t.Parallel()
	question := "What should I check first on a morning walkthrough of this farm today?"
	got, meta := TrimInstructionLeak(smokeMorningWalkLeak, question)
	if !meta.Trimmed {
		t.Fatal("expected leak trim")
	}
	if meta.CharsRemoved < 100 {
		t.Fatalf("chars_removed=%d want substantial trim", meta.CharsRemoved)
	}
	if strings.Contains(got, "## Your task") {
		t.Fatalf("leak marker still present: %q", got[len(got)-80:])
	}
	if strings.Contains(got, "Question:") {
		t.Fatal("echoed Question block still present")
	}
	if !strings.Contains(got, "humidity-related alerts") {
		t.Fatal("expected farm content preserved")
	}
}

func TestTrimInstructionLeak_trailingQuestionEcho(t *testing.T) {
	t.Parallel()
	question := "What is the EC target?"
	answer := "The veg room target is 1.2–2.0 mS/cm.\n\nQuestion:\nWhat is the EC target?"
	got, meta := TrimInstructionLeak(answer, question)
	if !meta.Trimmed {
		t.Fatal("expected trim")
	}
	if got != "The veg room target is 1.2–2.0 mS/cm." {
		t.Fatalf("got %q", got)
	}
}

func TestTrimInstructionLeak_noLeak(t *testing.T) {
	t.Parallel()
	answer := "Check humidity in the Flower Room first, then review EC in veg."
	got, meta := TrimInstructionLeak(answer, "morning walk?")
	if meta.Trimmed || got != answer {
		t.Fatalf("unexpected trim: meta=%+v got=%q", meta, got)
	}
}

func TestTrimInstructionLeak_empty(t *testing.T) {
	t.Parallel()
	got, meta := TrimInstructionLeak("  ", "q")
	if meta.Trimmed || got != "  " {
		t.Fatalf("got %q meta=%+v", got, meta)
	}
}

const smokeMorningWalkMetaCorrection = `Check veg EC 1.2–2.0 mS/cm first, then flower room humidity.
I apologize for misunderstanding. The instruction requires a focus on immediate actions. Here's an updated answer:`

func TestTrimMetaCorrection_smokeMorningWalk(t *testing.T) {
	t.Parallel()
	got, meta := TrimMetaCorrection(smokeMorningWalkMetaCorrection)
	if !meta.Trimmed {
		t.Fatal("expected meta correction trim")
	}
	if strings.Contains(strings.ToLower(got), "apolog") {
		t.Fatalf("apology still present: %q", got)
	}
	if !strings.Contains(got, "veg EC") {
		t.Fatal("expected farm content preserved")
	}
}

func TestAnswerContainsMetaCorrection(t *testing.T) {
	t.Parallel()
	if !AnswerContainsMetaCorrection(smokeMorningWalkMetaCorrection) {
		t.Fatal("expected detection")
	}
	clean, _ := TrimMetaCorrection(smokeMorningWalkMetaCorrection)
	if AnswerContainsMetaCorrection(clean) {
		t.Fatal("expected clean after trim")
	}
}

func TestTrimMetaCorrection_disregardMarkers(t *testing.T) {
	t.Parallel()
	answer := "Check veg EC first.\nPlease disregard the earlier paragraph about wildlife."
	got, meta := TrimMetaCorrection(answer)
	if !meta.Trimmed {
		t.Fatal("expected disregard trim")
	}
	if strings.Contains(strings.ToLower(got), "disregard") {
		t.Fatalf("disregard tail still present: %q", got)
	}
}

const run3ECPHSourceDump = `Our operational documentation for leafy greens indicates lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0 per [1].

Sources (cite using [n] only from this list):

[6] type=field_guide source_id=123 chunk_id=456
field_guide
doc_path: field-guides/endocrine-disruptors.md
Endocrine disruptors in Lake Erie wildlife show profound effects.`

func TestTrimSourceDump_run3ECPH(t *testing.T) {
	t.Parallel()
	got, meta := TrimSourceDump(run3ECPHSourceDump)
	if !meta.Trimmed {
		t.Fatal("expected source dump trim")
	}
	if strings.Contains(got, "type=field_guide") || strings.Contains(got, "Sources (cite") {
		t.Fatalf("dump still present: %q", got[len(got)-120:])
	}
	if !strings.Contains(got, "lettuce EC") {
		t.Fatal("expected opening preserved")
	}
}

func TestTrimSourceDump_metadataLineTail(t *testing.T) {
	t.Parallel()
	opening := strings.Repeat("Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0. ", 12)
	answer := opening + "\n\n[6] type=field_guide source_id=99 chunk_id=1\nTypha latifolia biosorption."
	got, meta := TrimSourceDump(answer)
	if !meta.Trimmed {
		t.Fatal("expected metadata dump trim")
	}
	if strings.Contains(got, "type=field_guide") {
		t.Fatalf("metadata dump still present: %q", got)
	}
}

func TestTrimGroundedAnswerLength_cpuLaptopCap(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("Hydro lettuce EC 0.8–1.3 mS/cm. ", 200)
	got, meta := TrimGroundedAnswerLength(long, 4096)
	if !meta.Trimmed || meta.MaxChars != 2500 {
		t.Fatalf("meta=%+v len=%d", meta, len(got))
	}
	if len(got) > 2600 {
		t.Fatalf("trimmed too long: %d", len(got))
	}
}

func TestTrimGroundedAnswerLength_largeWindowNoCap(t *testing.T) {
	t.Parallel()
	answer := strings.Repeat("word ", 1000)
	got, meta := TrimGroundedAnswerLength(answer, 32768)
	if meta.Trimmed || got != answer {
		t.Fatalf("unexpected trim: meta=%+v", meta)
	}
}
