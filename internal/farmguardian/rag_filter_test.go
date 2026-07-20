package farmguardian

import (
	"encoding/json"
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func metaDocPath(path string) json.RawMessage {
	b, _ := json.Marshal(map[string]string{"doc_path": path, "module": "field_guide"})
	return b
}

func TestAgronomyQueryIntent(t *testing.T) {
	t.Parallel()
	if !AgronomyQueryIntent("What EC and pH targets for leafy greens?") {
		t.Fatal("expected agronomy intent")
	}
	if AgronomyQueryIntent("What should I check on a morning walkthrough?") {
		t.Fatal("morning walk should not be agronomy intent")
	}
}

func TestRAGRetrieveLimit_overfetchAgronomy(t *testing.T) {
	t.Parallel()
	if got := RAGRetrieveLimit("EC targets for lettuce", 4); got < 8 {
		t.Fatalf("want overfetch >= 8, got %d", got)
	}
	if got := RAGRetrieveLimit("morning walkthrough", 4); got != 4 {
		t.Fatalf("non-agronomy should stay 4, got %d", got)
	}
}

func TestFilterRAGChunksForToolPlan_excludesPlatformDocOnWalkFarm(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "platform_doc", Metadata: metaDocPath("docs/operator-tour.md"), ContentText: "aria-live on Needs attention strip"},
		{ID: 2, SourceType: "alert_notification", ContentText: "Humidity high — Flower Room"},
		{ID: 3, SourceType: "field_guide", Metadata: metaDocPath("field-guides/crop-lettuce-nutrition.md"), ContentText: "Lettuce EC"},
	}
	plan := ToolPlan{ToolIDs: []string{"walk_farm", "summarize_device_health"}}
	res := FilterRAGChunksForToolPlan("morning walkthrough", plan, chunks, 3)
	if len(res.Chunks) != 2 {
		t.Fatalf("want 2 chunks after platform_doc drop, got %d", len(res.Chunks))
	}
	for _, c := range res.Chunks {
		if c.SourceType == "platform_doc" {
			t.Fatalf("platform_doc should be excluded on walk_farm, got id=%d", c.ID)
		}
	}
	if !strings.Contains(res.Note, "walk_farm_filter") {
		t.Fatalf("expected walk_farm_filter note, got %q", res.Note)
	}
}

func TestFilterRAGChunks_dropsOffTopicFromTop3(t *testing.T) {
	t.Parallel()
	query := "What does our operational documentation say about EC and pH targets for leafy greens here?"
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "field_guide", Metadata: metaDocPath("field-guides/crop-lettuce-nutrition.md"), ContentText: "Lettuce EC 0.8–1.3 mS/cm"},
		{ID: 2, SourceType: "field_guide", Metadata: metaDocPath("field-guides/endocrine-disruptors-lake-erie.md"), ContentText: "Endocrine disruptors in Lake Erie wildlife"},
		{ID: 3, SourceType: "platform_doc", Metadata: metaDocPath("docs/fertigation-basics.md"), ContentText: "pH 5.5–6.0 for leafy greens"},
		{ID: 4, SourceType: "field_guide", Metadata: metaDocPath("field-guides/crop-spinach-nutrition.md"), ContentText: "Spinach EC targets"},
		{ID: 5, SourceType: "field_guide", Metadata: metaDocPath("field-guides/wildlife-typha-biosorption.md"), ContentText: "Typha latifolia biosorption wetlands"},
	}
	res := FilterRAGChunks(query, chunks, 3)
	if len(res.Chunks) != 3 {
		t.Fatalf("want 3 chunks, got %d", len(res.Chunks))
	}
	for _, c := range res.Chunks {
		if chunkOffTopic(c) {
			t.Fatalf("off-topic chunk in top-3: id=%d path=%q", c.ID, chunkDocPath(c))
		}
	}
	if res.Note == "" {
		t.Fatal("expected filter note")
	}
}

func TestFilterRAGChunks_nonAgronomyPassthrough(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "field_guide", Metadata: metaDocPath("field-guides/endocrine-disruptors.md"), ContentText: "endocrine"},
	}
	res := FilterRAGChunks("morning walkthrough alerts", chunks, 3)
	if len(res.Chunks) != 1 || res.Chunks[0].ID != 1 {
		t.Fatalf("non-agronomy should pass through: %+v", res)
	}
	if res.Note != "" {
		t.Fatalf("unexpected note %q", res.Note)
	}
}

func TestFilterRAGChunks_capsFieldGuide(t *testing.T) {
	t.Setenv("GUARDIAN_RAG_MAX_CHUNKS_FIELD_GUIDE", "2")
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "field_guide", Metadata: metaDocPath("field-guides/crop-lettuce-nutrition.md")},
		{ID: 2, SourceType: "field_guide", Metadata: metaDocPath("field-guides/crop-kale-nutrition.md")},
		{ID: 3, SourceType: "field_guide", Metadata: metaDocPath("field-guides/crop-spinach-nutrition.md")},
		{ID: 4, SourceType: "platform_doc", Metadata: metaDocPath("docs/platform.md")},
	}
	res := FilterRAGChunks("EC for leafy greens", chunks, 4)
	if countFieldGuide(res.Chunks) != 2 {
		t.Fatalf("want 2 field_guide chunks, got %d (%+v)", countFieldGuide(res.Chunks), res.Chunks)
	}
}

func TestChunkDocPath_fromContentHeader(t *testing.T) {
	t.Parallel()
	c := db.SearchRagNearestNeighborsFilteredRow{
		ContentText: "field_guide\ndoc_path: field-guides/crop-lettuce-nutrition.md\n\nLettuce EC",
	}
	if got := chunkDocPath(c); got != "field-guides/crop-lettuce-nutrition.md" {
		t.Fatalf("got %q", got)
	}
}
