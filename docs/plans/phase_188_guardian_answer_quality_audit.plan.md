---
name: Phase 188 — Guardian answer quality audit + broadened instruction-leak detection
overview: >
  Read all 20 live conversation_turns rows in the dev DB (phi3:mini, farm 1) end
  to end and rated each Q&A. Found four concrete, reproducible defect classes not
  caught by the existing Phase 143/145/148/150/151/152 hygiene pipeline. This
  phase fixes the worst one (a severe off-topic hallucination that leaked a
  completely unrelated essay-writing prompt template into a farmer-facing
  answer) by broadening farmguardian.TrimInstructionLeak's marker set. Phases
  189-191 fix the remaining three.
findings: >
  Turn id=17 (session 9ca339ad…, "Before I confirm — which zone should this
  task refer to?") answered with a hallucinated essay prompt about "The Great
  Gatsby" / a fictional "Atonement (1930) by William Faulkner" — zero relation
  to farming. The leaked template used "## Instruction>" and a bare "Question"
  heading, neither of which TrimInstructionLeak's marker list ("## your task",
  "\nquestion:\n<echoed question>") recognizes, so it passed through unfiltered
  and un-flagged (no accuracy_note either).
todos:
  - id: ws1-broaden-markers
    content: "WS1: add '## instruction', '\\ndocument:\\n', and an essay-writing tell ('write an extensive essay', 'write a) essay') to leakCutIndex; treat a bare '\\nquestion\\n' heading (no colon, no echoed question match required) as a cut boundary when it's immediately followed by one of those markers"
    status: completed
  - id: ws2-tests
    content: "WS2: Go test using a redacted version of the live turn 17 answer as fixture; TestTrimInstructionLeak_offTopicEssayLeak"
    status: completed
  - id: ws3-closure-docs
    content: "WS3: phase-188-closure.test.js + docs/current-state.md + docs/farm-guardian-architecture.md note"
    status: completed
isProject: false
---
