// Phase 145 WS3 — RAG retrieval guardrails for agronomy prompts.

package farmguardian

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
)

var offTopicDocPathMarkers = []string{
	"endocrine", "wildlife", "typha", "lake-erie", "lake_erie",
	"biosorption", "aquatic-ecosystem", "hormonal",
}

var agronomyDocPathMarkers = []string{
	"crop-", "nutrition", "fertigation", "nutrient", "leafy", "lettuce",
	"hydro", "irrigation", "water-quality", "ec-", "ph-",
}

// RAGFilterResult is the post-retrieval chunk list and optional debug note.
type RAGFilterResult struct {
	Chunks []db.SearchRagNearestNeighborsFilteredRow
	Note   string
}

var agronomyIntentLongMarkers = []string{
	"leafy green", "lettuce", "kale", "spinach", "crop",
	"fertigation", "nutrient", "hydro", "ms/cm",
}

var agronomyIntentShortTokens = []string{"ec", "ph"}

// AgronomyQueryIntent detects EC/pH / crop / nutrient field-guide prompts.
func AgronomyQueryIntent(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	for _, m := range agronomyIntentLongMarkers {
		if strings.Contains(q, m) {
			return true
		}
	}
	for _, w := range tokenizeWords(q) {
		for _, tok := range agronomyIntentShortTokens {
			if w == tok {
				return true
			}
		}
	}
	return false
}

// RAGRetrieveLimit expands fetch size before agronomy filtering trims back to topK.
func RAGRetrieveLimit(query string, topK int) int {
	if topK <= 0 {
		topK = RAGTopK
	}
	if AgronomyQueryIntent(query) {
		over := topK * 2
		if over < topK+4 {
			over = topK + 4
		}
		if over > 16 {
			over = 16
		}
		return over
	}
	return topK
}

// FilterRAGChunksForToolPlan applies agronomy filtering then walk_farm guardrails.
// When walk_farm is planned, platform_doc chunks are excluded — snapshot + read tools carry the answer.
func FilterRAGChunksForToolPlan(query string, plan ToolPlan, chunks []db.SearchRagNearestNeighborsFilteredRow, limit int) RAGFilterResult {
	res := FilterRAGChunks(query, chunks, limit)
	if !toolPlanIncludes(plan, "walk_farm") {
		return res
	}
	filtered := make([]db.SearchRagNearestNeighborsFilteredRow, 0, len(res.Chunks))
	dropped := 0
	for _, c := range res.Chunks {
		if c.SourceType == "platform_doc" {
			dropped++
			continue
		}
		filtered = append(filtered, c)
	}
	note := res.Note
	if dropped > 0 {
		suffix := fmt.Sprintf("walk_farm_filter: excluded %d platform_doc chunk(s)", dropped)
		if note != "" {
			note += "; " + suffix
		} else {
			note = suffix
		}
	}
	return RAGFilterResult{Chunks: filtered, Note: note}
}

func toolPlanIncludes(plan ToolPlan, toolID string) bool {
	for _, id := range plan.ToolIDs {
		if id == toolID {
			return true
		}
	}
	return false
}

// FilterRAGChunks reorders and trims retrieved chunks for agronomy queries.
func FilterRAGChunks(query string, chunks []db.SearchRagNearestNeighborsFilteredRow, limit int) RAGFilterResult {
	if len(chunks) == 0 || !AgronomyQueryIntent(query) {
		return RAGFilterResult{Chunks: chunks}
	}
	if limit <= 0 {
		limit = len(chunks)
	}

	var preferred, neutral, demoted []db.SearchRagNearestNeighborsFilteredRow
	for _, c := range chunks {
		switch {
		case chunkOffTopic(c):
			demoted = append(demoted, c)
		case chunkAgronomyPreferred(c):
			preferred = append(preferred, c)
		default:
			neutral = append(neutral, c)
		}
	}

	out := append(append([]db.SearchRagNearestNeighborsFilteredRow{}, preferred...), neutral...)
	if len(out) < limit {
		need := limit - len(out)
		if need > len(demoted) {
			need = len(demoted)
		}
		out = append(out, demoted[:need]...)
	}
	if len(out) > limit {
		out = out[:limit]
	}

	dropped := len(demoted)
	if len(out) >= limit && dropped > 0 {
		// Count demoted chunks actually excluded from the final slice.
		includedDemoted := 0
		for _, c := range out {
			for _, d := range demoted {
				if c.ID == d.ID {
					includedDemoted++
					break
				}
			}
		}
		dropped = len(demoted) - includedDemoted
	}

	if maxFG := ragMaxFieldGuideChunks(); maxFG > 0 {
		before := countFieldGuide(out)
		out = capFieldGuideChunks(out, maxFG)
		if capped := before - countFieldGuide(out); capped > 0 {
			dropped += capped
		}
	}

	note := ""
	if dropped > 0 {
		note = fmt.Sprintf("agronomy_filter: excluded %d off-topic or excess field_guide chunks", dropped)
	} else if len(preferred) > 0 {
		note = "agronomy_filter: preferred on-topic chunks"
	}
	return RAGFilterResult{Chunks: out, Note: note}
}

func ragMaxFieldGuideChunks() int {
	raw := strings.TrimSpace(os.Getenv("GUARDIAN_RAG_MAX_CHUNKS_FIELD_GUIDE"))
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func countFieldGuide(chunks []db.SearchRagNearestNeighborsFilteredRow) int {
	n := 0
	for _, c := range chunks {
		if c.SourceType == "field_guide" {
			n++
		}
	}
	return n
}

func capFieldGuideChunks(chunks []db.SearchRagNearestNeighborsFilteredRow, max int) []db.SearchRagNearestNeighborsFilteredRow {
	if max <= 0 || len(chunks) == 0 {
		return chunks
	}
	out := make([]db.SearchRagNearestNeighborsFilteredRow, 0, len(chunks))
	fg := 0
	for _, c := range chunks {
		if c.SourceType == "field_guide" {
			if fg >= max {
				continue
			}
			fg++
		}
		out = append(out, c)
	}
	return out
}

func chunkOffTopic(c db.SearchRagNearestNeighborsFilteredRow) bool {
	corpus := chunkDocPath(c) + " " + strings.ToLower(c.ContentText)
	for _, m := range offTopicDocPathMarkers {
		if strings.Contains(corpus, m) {
			return true
		}
	}
	for _, m := range offTopicCitationMarkers {
		if strings.Contains(corpus, m) {
			return true
		}
	}
	return false
}

func chunkAgronomyPreferred(c db.SearchRagNearestNeighborsFilteredRow) bool {
	if c.SourceType == "platform_doc" {
		return true
	}
	dp := chunkDocPath(c)
	for _, m := range agronomyDocPathMarkers {
		if strings.Contains(dp, m) {
			return true
		}
	}
	return false
}

func chunkDocPath(c db.SearchRagNearestNeighborsFilteredRow) string {
	if len(c.Metadata) > 0 {
		var meta map[string]any
		if json.Unmarshal(c.Metadata, &meta) == nil {
			if dp, ok := meta["doc_path"].(string); ok {
				return strings.ToLower(strings.TrimSpace(dp))
			}
		}
	}
	return parseDocPathFromContent(c.ContentText)
}

func parseDocPathFromContent(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(line, "doc_path:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "doc_path:"))
		}
	}
	return ""
}
