package farmguardian

// SourceLabelingGroundingRule tells the model how to label live vs documented sources (Phase 133 WS1).
const SourceLabelingGroundingRule = `Source honesty (Phase 133):
- LIVE FARM DATA (snapshot block, read-tool results): phrase as "right now" / "on your farm today" / "currently".
- FIELD GUIDE [n] citations: phrase as "per our field guide" or "per the install guide".
- PLATFORM DOC [n] citations: phrase as "per platform docs" or "per the operator guide".
- Farm-note / operational [n] citations: phrase as "per a saved note" — never present note text as live sensor state.
- Never state a sensor value, alert count, or zone reading from RAG alone — cross-check snapshot/read tools or say you only have an older note.`

// GroundedHonestyPromptBlock returns rules appended to grounded system prompts.
func GroundedHonestyPromptBlock() string {
	return SourceLabelingGroundingRule
}
