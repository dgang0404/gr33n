// Phase 211.06 WS3 — detect invent / roleplay / instruction-soup for one regenerate.

package farmguardian

import (
	"strings"
)

// inventStubOpenings are case-insensitive prefixes that mean the model collapsed
// into apology, roleplay, or prompt echo instead of answering.
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
	"it seems like there's a misunderstanding",
	"it seems like there is a misunderstanding",
	"it seems to be a grim",
}

// inventStubBodyMarkers catch collapse mid-answer (phi3 often apologizes after a
// nonsense opener rather than at byte 0).
var inventStubBodyMarkers = []string{
	"i apologize",
	"i apologise",
	"misunderstanding in the instructions",
	"provided document is an example",
	"textbook section",
	"write an extensive essay",
	"write an essay",
	"documentary filmography",
	"vectorization_text",
	"context vectorization",
	"the above context",
	"the above documentary",
	"@japanese-specifically",
}

// AnswerLooksLikeInventStub reports apology / roleplay / instruction-soup that
// should trigger one regenerate before persist (Phase 211.06).
func AnswerLooksLikeInventStub(answer string) bool {
	a := strings.TrimSpace(answer)
	if a == "" {
		return false
	}
	lower := strings.ToLower(a)
	head := lower
	if len(head) > 400 {
		head = head[:400]
	}
	for _, p := range inventStubOpenings {
		if strings.HasPrefix(lower, p) || strings.Contains(head, "\n"+p) {
			return true
		}
	}
	// Mid-body collapse in the opening window (not just line-start).
	for _, m := range inventStubBodyMarkers {
		if strings.Contains(head, m) {
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
