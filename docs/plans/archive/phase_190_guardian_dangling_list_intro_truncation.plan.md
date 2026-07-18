---
name: Phase 190 — Guardian dangling list-intro truncation detection + completion budget bump
overview: >
  Part 3 of the Phase 188 audit. Phase 148's TruncatedAnswerTailNote only
  catches a word glued to trailing digits (a mid-token cutoff). Three of the
  20 live turns end the answer on a colon that promises a list or a
  confirmation and never deliver it — a different truncation shape entirely.
  One of the three hit exactly LLM_MAX_TOKENS=1024 completion tokens (a real
  budget cutoff); the other two stopped well under budget (the small model
  chose to stop generating early) — so this phase both adds detection for the
  shape and raises the default budget enough to remove the confirmed-cutoff
  case, while flagging the harder early-stop case for visibility rather than
  claiming to fix it outright.
findings: >
  Turn id=2 ("What should I check first on a morning walkthrough...") stops
  after exactly one checklist item (859/1024 completion tokens — not a budget
  cutoff, the model just stopped). Turn id=5 ("Set the feed volume to 0.3
  liters...") ends "Once you confirm this setup:" with nothing after
  (585/1024 tokens — also not budget). Turn id=12 ("Create a task to refill
  calcium nitrate...") ends "...while refilling calcium nitrate:" at exactly
  1024/1024 completion tokens — a real cutoff.
todos:
  - id: ws1-dangling-intro-note
    content: "WS1: DanglingListIntroNote in answer_accuracy.go — flags an answer whose trimmed tail ends in ':' with no following content, wired into AnswerAccuracyNote"
    status: completed
  - id: ws2-budget-bump
    content: "WS2: raise maxTokensFromEnv default 1024 -> 1536 (LLM_MAX_TOKENS override unchanged, still capped 8192)"
    status: completed
  - id: ws3-tests-docs
    content: "WS3: Go tests (fixed budget-cutoff + early-stop fixtures) + phase-190-closure.test.js + docs"
    status: completed
isProject: false
---
