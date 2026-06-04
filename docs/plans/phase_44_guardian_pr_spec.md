---
name: "Phase 44 — Guardian PR spec (wizards first, setup starters second)"
overview: >
  Implementation spec for Phase 44 Guardian slice: in-app wizards own linear setup;
  conversation starters and optional setup-mode persona only guide chat. Reuses
  apply_grow_setup_pack matcher; bootstrap template via wizard API (not chat-first).
parent_plan: phase_44_getting_started_edge_wizard.plan.md
status: planned
---

# Phase 44 — Guardian PR spec (wizards first, setup starters second)

**Parent:** [phase_44_getting_started_edge_wizard.plan.md](phase_44_getting_started_edge_wizard.plan.md)

**Not in this doc:** NL → PR when matchers miss → [phase_46_guardian_llm_tool_proposals.plan.md](phase_46_guardian_llm_tool_proposals.plan.md)

**Prerequisites:** [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) · Phase 15 bootstrap API · Phase 32 `apply_grow_setup_pack` · Phase 37 field procedures for Pi copy

---

## 1. UX principle (non-negotiable)

| Priority | Path | Why |
|----------|------|-----|
| **1 — Primary** | **Wizards** (farm / zone / device / first-run checklist) | Linear steps, preview, one API call per step — no chat literacy required |
| **2 — Secondary** | **Guardian starters + setup-mode hints** | Questions, grow-setup pack PR, procedure help — for operators who prefer chat |
| **3 — Last resort** | Phase **46** LLM tool proposals | Only when matchers miss and message is clearly a write |

**Do not** make Guardian the only way to create a farm, zone, or device. **Do not** auto-Confirm from starter chips.

---

## 2. What Phase 44 adds to Guardian

| Deliverable | Type | Outcome |
|-------------|------|---------|
| **First-run / setup starters** | UI (`guardianStarters.js`) | Chips on Dashboard checklist, wizard footers, empty zone |
| **Setup-mode persona** | Handler optional | System hint when farm has 0 zones or `?setup=1` |
| **Existing matchers** | Go (no change required v1) | `apply_grow_setup_pack`, `ack_alert` — starters send phrases matchers already understand |
| **Bootstrap template** | Wizard **POST**, not chat-first | `apply_bootstrap_template` remains **admin + high-tier PR** — wizard calls API directly |
| **No new Confirm tools** | — | Device register, zone POST stay wizard/API |

---

## 3. Wizards vs Guardian (ownership matrix)

| Job | Wizard (WS) | Guardian |
|-----|-------------|----------|
| Apply farm template | WS1 `POST …/bootstrap-template` | Starter: “What does indoor veg template include?” (read/advice only) |
| Add zone | WS2 `POST …/zones` | Starter: “What should I name my first grow room?” |
| Register Pi / API key | WS3 device flow | Starter: `start procedure wire-pi-relay-light` or link to field guide |
| First comfort band | WS5 links to Phase 42 Targets | Starter: “What humidity band should I set first?” (advice; patch matchers = 42) |
| First grow (plant + program) | Optional wizard step OR manual Plants | **Starter → `apply_grow_setup_pack` PR** when zone empty |
| Ack morning alert | Alerts UI / zone cockpit (40) | Starter: “Acknowledge alert #N” → `ack_alert` |

---

## 4. Conversation starters

### 4.1 Surfaces

| Surface key | Where | Max chips |
|-------------|-------|-----------|
| `first_run_dashboard` | Dashboard getting-started checklist (WS5) | 3 |
| `farm_setup_wizard` | Last step footer (“Need help?”) | 2 |
| `zone_wizard` | After zone created / empty zone list | 3 |
| `device_wizard` | Device register step | 2 |
| `empty_zone_grow` | Zone with no active cycle (cockpit 40) | 3 |
| `setup_mode_chat` | Guardian drawer when setup mode active | 4 |

### 4.2 Chip table (evaluate priority order)

| Priority | Condition | Chip label | Message sent |
|----------|-----------|------------|--------------|
| 1 | Farm has 0 zones | Add my first grow room | `I'm setting up a new farm — what should I do first after creating a zone?` |
| 2 | Zone exists, no active cycle, no duplicate plant | Start a grow in {zoneName} | `Add my philodendron to {zoneName} with a light fertigation program` |
| 3 | Unread alert on dashboard | Handle this alert | `Acknowledge alert #{id}: {subject}` |
| 4 | Device wizard step | Wire Pi checklist | `start procedure wire-pi-relay-light` |
| 5 | Farm setup wizard | Compare templates | `What's the difference between indoor veg and greenhouse climate bootstrap templates?` |
| 6 | Device offline in snapshot | Why is my Pi offline? | `start procedure diagnose-pi-offline` |
| 7 | (fallback setup) | What does setup mode do? | `I'm in farm setup — walk me through zones, device, and comfort targets in order` |

**Behavior:** Open drawer → `prefilledMessage` → operator sends manually.

### 4.3 Anti-patterns

| Bad | Good |
|-----|------|
| Chip: “Apply bootstrap template” (auto PR) | Wizard button: **Apply template** with preview list |
| Chip: “Create zone Flower Room” (silent write) | Wizard: name + type fields → POST |
| Generic “What's the status of my farm?” | Job-shaped setup question |

---

## 5. Setup-mode persona (WS4)

### 5.1 When active

| Signal | Setup mode ON |
|--------|----------------|
| `GET /v1/chat` query `setup=1` | yes |
| Farm snapshot: `zone_count == 0` | yes (first open of drawer after 44) |
| First-run checklist incomplete (client flag) | optional extend |

### 5.2 System hint (append to grounded prompt)

Short bullet block (persona file or handler):

- Prefer **wizards** linked in UI over inventing config in chat.
- For **first grow** in an empty zone, may propose **`apply_grow_setup_pack`** only when matcher rules pass (Phase 32).
- **`apply_bootstrap_template`** — tell operator to use **Farm setup wizard** or Settings; do not promise chat can apply unless user is farm admin and explicitly asks.
- Pi wiring: cite **field procedures** and print URLs (Phase 37).

### 5.3 Acceptance

- [ ] Setup mode does not auto-insert proposals without user message
- [ ] Zero-zone farm: Guardian mentions checklist order (zone → device → targets)

---

## 6. PR cards in Phase 44

### 6.1 In scope (existing tools)

| Tool | How operator reaches it |
|------|-------------------------|
| `apply_grow_setup_pack` | Empty-zone starter or typed grow-setup phrase ([proposals_setup_pack.go](../../internal/farmguardian/proposals_setup_pack.go)) |
| `ack_alert` | Dashboard/alert starter ([proposals.go](../../internal/farmguardian/proposals.go)) |
| `create_lighting_program` | Optional starter: “Set up 18/6 veg lights in {zone}” — matcher exists in registry; add phrase in 44 only if sit-in needs |

### 6.2 Out of scope (wizard/API only)

| Tool | Reason |
|------|--------|
| `apply_bootstrap_template` | Admin PR; wizard uses direct POST with RBAC |
| `create_plant` / `create_crop_cycle` / `create_fertigation_program` alone | Prefer setup pack or wizard steps |
| Device / API key creation | No registry tool |

---

## 7. Workstream mapping

| Parent WS | Guardian slice |
|-----------|----------------|
| WS1–WS3, WS5 | Wizards primary; starters on footers/checklist |
| WS4 | §5 setup-mode persona |
| WS6 | operator-tour §8 + §6g + architecture §7.0j |
| **WS8** | This spec |

---

## 8. Definition of done (Guardian slice)

- [ ] Starters on first-run checklist, wizards, empty zone — no generic status chips
- [ ] Setup-mode hint when `zone_count == 0` or `?setup=1`
- [ ] Grow-setup starter triggers existing setup-pack matcher (manual send)
- [ ] Bootstrap **not** promoted via starter chips (wizard only)
- [ ] operator-tour §6g + §8
- [ ] No dependency on Phase 46

---

## Related

| Doc | Use |
|-----|-----|
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Setup pack matcher rules |
| [phase_15_farm_onboarding.plan.md](phase_15_farm_onboarding.plan.md) | Bootstrap API |
| [pi-integration-guide.md](../pi-integration-guide.md) | Device wizard embed |
