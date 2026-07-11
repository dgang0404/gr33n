package synthesis

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
)

const systemPrompt = `You are an assistant for a farm operator using the gr33n platform. Answer the user's question using ONLY the numbered sources below. Every substantive factual claim MUST include an inline citation using square brackets and the source number, for example [1] or [2]. If you mention information from source 3, write [3] immediately after that information. If the sources do not contain enough information to answer, say so clearly and do not invent facts. Do not cite a number that does not exist.

Answer the question in at most four short paragraphs, then stop. Do NOT append a "Sources:" list or echo raw source metadata lines (for example type=field_guide, doc_path=, chunk_id=). Use only [n] citations from the provided list — never invent citation lines. If sources conflict or seem off-topic for the question, say so briefly and stop; do not elaborate on unrelated chunks. Keep the answer concise and operational.`

var bracketRefRE = regexp.MustCompile(`\[(\d+)\]`)

// BuildUserMessage formats the operator question and numbered chunk sources for the chat model.
func BuildUserMessage(question string, chunks []db.SearchRagNearestNeighborsFilteredRow) string {
	var b strings.Builder
	b.WriteString("Question:\n")
	b.WriteString(strings.TrimSpace(question))
	b.WriteString("\n\nSources (cite using [n] only from this list):\n\n")
	for i := range chunks {
		n := i + 1
		ch := chunks[i]
		b.WriteString(fmt.Sprintf("[%d] type=%s source_id=%d chunk_id=%d\n",
			n, ch.SourceType, ch.SourceID, ch.ID))
		b.WriteString(strings.TrimSpace(ch.ContentText))
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// SystemPrompt returns the fixed system instruction for synthesis.
func SystemPrompt() string { return systemPrompt }

// RefNumbersInAnswer extracts unique bracket citation numbers from model output (e.g. [1], [12]).
func RefNumbersInAnswer(answer string) []int {
	matches := bracketRefRE.FindAllStringSubmatch(answer, -1)
	seen := make(map[int]struct{})
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		v, err := strconv.Atoi(m[1])
		if err != nil || v < 1 {
			continue
		}
		seen[v] = struct{}{}
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// Citation is one grounded reference returned to the client.
type Citation struct {
	Ref        int    `json:"ref"`
	ChunkID    int64  `json:"chunk_id"`
	SourceType string `json:"source_type"`
	SourceID   int64  `json:"source_id"`
	Route      string `json:"route,omitempty"`
	Excerpt    string `json:"excerpt"`
}

// BuildCitations maps 1-based ref indices to chunk rows (invalid refs skipped).
func BuildCitations(answer string, chunks []db.SearchRagNearestNeighborsFilteredRow) []Citation {
	refs := RefNumbersInAnswer(answer)
	var out []Citation
	for _, ref := range refs {
		if ref < 1 || ref > len(chunks) {
			continue
		}
		ch := chunks[ref-1]
		ex := ch.ContentText
		if len(ex) > 400 {
			ex = ex[:400] + "…"
		}
		out = append(out, Citation{
			Ref:        ref,
			ChunkID:    ch.ID,
			SourceType: ch.SourceType,
			SourceID:   ch.SourceID,
			Excerpt:    ex,
		})
	}
	return out
}

// StripOrphanCitationRefs removes [n] brackets when no chunks exist (Phase 82 WS1).
func StripOrphanCitationRefs(answer string, chunkCount int) string {
	if chunkCount > 0 || answer == "" {
		return answer
	}
	return bracketRefRE.ReplaceAllString(answer, "")
}
