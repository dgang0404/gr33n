---
name: Phase 189 — Guardian inline source-metadata + placeholder-citation redaction
overview: >
  Part 2 of the Phase 188 answer-quality audit. TrimSourceDump (Phase 143)
  only strips trailing block-level "Sources:" dumps. Live turns show the model
  weaving the same raw RAG bookkeeping into the middle of a sentence instead —
  "(field_guide source id=8, chunk id=66)", "field_guide source_id=17
  chunk_id=18" — and, separately, echoing the citation-format instruction's
  own [n] placeholder literally instead of substituting a real number —
  "source[n]", "(source:[5])" used in place of a missing numeric target. Adds
  a new sanitizer for both, wired into the same finalize pipeline as the
  other Phase 143 hygiene passes.
findings: >
  Turn id=6 ("Acknowledge the highest severity unread alert") — "field_guide
  source id=8, chunk id=66" and "field_guide source id=10" leak inline. Turn
  id=18 — "source_id=17 chunk_id=18" leaks inline. Turn id=15 ("Please revise
  — use 0.3 L instead of 0.5") — "a low EC of approximately [2] mS/cm" and
  "(source: [2])" use the citation marker itself as the missing numeric value
  because the cited field-guide chunk's own placeholder dashes ("~—–—
  mS/cm... use lookup_crop_targets for numeric targets") were echoed instead
  of calling lookup_crop_targets.
todos:
  - id: ws1-inline-metadata
    content: "WS1: RedactInlineSourceMetadata — strip '(field_guide source id=N[, chunk id=N])' / 'source_id=N chunk_id=N' inline mentions, collapsing resulting double spaces/empty parens"
    status: completed
  - id: ws2-placeholder-citation
    content: "WS2: RedactPlaceholderCitationMarkers — strip literal 'source[n]' / 'source_id=[n]' / 'chunk_id=[n]' / '(source:[n])' (literal letter n, not a digit)"
    status: completed
  - id: ws3-wire-pipeline
    content: "WS3: wire both into sanitizeAssistantAnswer + answerHygiene + TurnDebug fields, Go tests per function"
    status: completed
  - id: ws4-closure-docs
    content: "WS4: phase-189-closure.test.js + docs updates"
    status: completed
isProject: false
---
