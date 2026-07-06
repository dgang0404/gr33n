---
name: Phase 141 — Guardian feedback review workflow
overview: >
  Farm admins review thumbs-down turns in Settings, export CSV for agronomy triage,
  and pair with QA smoke archives. Closes Phase 134 WS5 and the human quality loop
  before manual guardian-qa-smoke walkthrough.
todos:
  - id: ws1-settings-card
    content: "WS1: GuardianSettingsFeedbackReviewCard — down-vote queue, since 7d/30d, CSV download"
    status: completed
  - id: ws2-runbook
    content: "WS2: docs/guardian-feedback-review-runbook.md — triage guide + smoke pairing"
    status: completed
  - id: ws3-qa-json-prompt
    content: "WS3: QARunArchive.feedback_review_prompt in guardian_qa_runs JSON"
    status: completed
  - id: ws4-docs-hygiene
    content: "WS4: Mark 134 WS5 shipped; ci-guardian-qa + bootstrap links"
    status: completed
  - id: ws5-tests
    content: "WS5: phase-141-closure vitest; phase-129-134 plan hygiene test"
    status: completed
isProject: false
---

# Phase 141 — Guardian feedback review workflow

**Status:** **Shipped.** · **Depends on:** [134](phase_134_guardian_answer_feedback.plan.md), [140](phase_140_guardian_qa_settings.plan.md)

**Next arc:** [142 Virtual Pi field validation](phase_142_virtual_pi_field_validation.plan.md)

---

## Acceptance

- [x] Farm admin sees down-vote queue in Settings
- [x] CSV export from Settings matches API export
- [x] Runbook documents post-smoke review order
- [x] QA archive JSON includes feedback_review_prompt
