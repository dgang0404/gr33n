// Phase 211.06 WS3 — detect invent / roleplay / instruction-soup openings for one regenerate.

package farmguardian

import (
	"strings"
)

// inventStubOpenings are case-insensitive prefixes / early markers that mean the
// model collapsed into apology, roleplay, or prompt echo instead of answering.
var inventStubOpenings = []string{
	"i apologize",
	"i apologise",
	"i'm sorry",
	"i am sorry",
	"sorry,",
	"sorry!",
	"you are a ",
	"you are an ",
	"as an ai",
	"as a language model",
	"as an assistant",
	"## your task",
	"## instruction",
	"question:",
	"here is a new question",
	"here's a new question",
}

// AnswerLooksLikeInventStub reports apology / roleplay / instruction-soup openings
// that should trigger one regenerate before persist (Phase 211.06).
func AnswerLooksLikeInventStub(answer string) bool {
	a := strings.TrimSpace(answer)
	if a == "" {
		return false
	}
	lower := strings.ToLower(a)
	head := lower
	if len(head) > 240 {
		head = head[:240]
	}
	for _, p := range inventStubOpenings {
		if strings.HasPrefix(lower, p) || strings.Contains(head, "\n"+p) {
			return true
		}
	}
	// Instruction soup: many prompt-template markers in a short answer.
	markers := 0
	for _, m := range []string{"## your task", "## instruction", "write an essay", "document:\n", "system:"} {
		if strings.Contains(lower, m) {
			markers++
		}
	}
	return markers >= 2
}

// InventStubRefuseMessage is the grounded fallback when regenerate still invents.
const InventStubRefuseMessage = "I don't have a reliable answer from farm records for that right now. Please rephrase, or check the field guides and Pending change requests."
