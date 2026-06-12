---
name: Phase 82 — Guardian plant intelligence & cultivator crop library
overview: >
  Make Guardian as smart as possible about plants: full cultivator crop library
  (≥25 profiles + field guides), live plant-context fusion (cycle + sensors +
  fertigation + light + grow_advisor), symptom/deficiency guides, substrate-aware
  watering, zero-chunk guardrails, multi-crop lookup, vision grounding, and
  honest unsupported-crop handling.
todos:
  - id: ws0-ops-prereqs
    content: "WS0: Ops — rag-ingest gate; LLM + vision model floor; bootstrap checklist"
    status: partial
  - id: ws1-zero-chunk-guardrail
    content: "WS1: Handler — zero-chunk policy; no fake [n] citations; strip orphan refs"
    status: completed
  - id: ws2-ui-honesty
    content: "WS2: UI — farm context · 0 doc chunks label; warning banner"
    status: completed
  - id: ws3-read-tool-widening
    content: "WS3: lookup_crop_targets — multi-crop compare; alias registry from YAML"
    status: completed
  - id: ws4a-catalog-source
    content: "WS4a: data/crop_library.yaml — crops, substrates, watering_style, cousin_of, unsupported"
    status: completed
  - id: ws4b-tier-a-profiles
    content: "WS4b: Tier A profiles + guides (eggplant, cucumber, kale, spinach, cilantro, microgreens, missing 4)"
    status: completed
  - id: ws4c-tier-b-profiles
    content: "WS4c: Tier B — zucchini, green_bean, mint, parsley, blueberry, hemp, broccoli, melon, arugula"
    status: completed
  - id: ws4d-field-guides
    content: "WS4d: Per-crop guides + deficiency/symptom guides; manifest + re-ingest"
    status: partial
  - id: ws4e-unsupported-registry
    content: "WS4e: Unsupported — ramps, mushrooms, fruit trees, in-ground root crops; cousin suggestions"
    status: completed
  - id: ws4f-ui-picker
    content: "WS4f: Profile picker — grouped, searchable, substrate hint"
    status: completed
  - id: ws5-follow-up-chips
    content: "WS5: Crop-aware follow-up chips (guardianFollowUps.js)"
    status: partial
  - id: ws7-plant-context-bundle
    content: "WS7: plant_context_bundle — fuse cycle, profile, sensors, fertigation, light, grow_advisor"
    status: deferred
  - id: ws8-substrate-watering
    content: "WS8: Substrate-aware watering from YAML — wet/dry, runoff, constant-feed"
    status: completed
  - id: ws9-symptom-deficiency
    content: "WS9: Symptom/deficiency RAG + intents; vision synergy (Phase 67)"
    status: deferred
  - id: ws10-stage-transitions
    content: "WS10: Stage transitions — flip, harvest, bolt, rebloom; grow_advisor on my room/my grow"
    status: partial
  - id: ws11-environment-reconcile
    content: "WS11: Live vs target — EC/VPD/DLI/photoperiod delta; comfort band conflicts; site_weather DLI"
    status: partial
  - id: ws6-docs-tests
    content: "WS6: architecture §7.0ag; smoke_phase82; phase-82-closure; OC-82"
    status: completed
isProject: false
---

# Phase 82 — Guardian plant intelligence & cultivator crop library

## Status

**Partially shipped** — catalog + multi-crop grounding + zero-chunk guardrail on `main`. Plant context bundle (WS7) and full target-vs-actual (WS11) deferred to Phase 97+.

**Closure:** [`phase-82-closure.md`](phase-82-closure.md) · **OC-82** (via Phase 110)

---

## The one job

> **Guardian answers plant questions like an experienced grower who knows your room — what you're growing, what stage it's in, what the sensors say, what the crop profile targets, and what the leaves might mean — without inventing numbers or fake doc citations.**

---

## Plant intelligence stack (what exists vs what Phase 82 adds)

| Capability | Shipped today | Phase 82 adds |
|------------|---------------|---------------|
| EC/pH/VPD/DLI targets | Phase 64 `lookup_crop_targets` (7 crops) | **≥25 crops** + aliases + multi-crop compare |
| Active grow science | Phase 62 `grow_advisor` (narrow intent) | **WS10** — fires on "my room / this grow" |
| Leaf photo hypotheses | Phase 67 vision + crop profile block | **WS9** — deficiency RAG + symptom intents |
| Live sensors + fertigation | `summarize_zone`, `summarize_zone_fertigation` (keyword-gated) | **WS7** — auto-bundle on plant questions |
| Lighting vs target | `summarize_zone_lighting` (partial) | **WS11** — photoperiod + DLI vs profile |
| Outdoor supplemental light | Phase 66 `site_weather` | Wire into **WS11** for greenhouse crops |
| Cycle cost / yield context | Phase 28 snapshot analytics | Keep; link in **WS7** when "how's my grow" |
| "How wet should it be?" | ❌ model guesses | **WS8** — substrate + watering_style in YAML |
| Compare cannabis vs orchid | ❌ broken (incident) | **WS3** multi-crop + **WS4** library |
| Unknown crop | ❌ cannabis-shaped defaults | **WS4e** unsupported + **cousin_of** suggestion |
| Fake citations at 0 chunks | ❌ incident | **WS1** + **WS2** |
| Conversation loop | started `guardianFollowUps.js` | **WS5** crop-aware chips |

---

## Incident (2026-06 — operator chat)

**Question (paraphrased):** *How should we water and feed cannabis vs eggplant vs ramps vs orchid? What lighting cycles and programs does each want? How wet do they like it?*

**Guardian response metadata:** `llama3.1:8b · grounded · 0 chunks · 2735 tok`

### Three stacked failures

| # | Symptom | Root cause |
|---|---------|------------|
| **1** | Fake references `[1]–[5]` ("Plant Nutrient Requirements", "Cycle Guidelines…") | RAG returned **0 chunks** but handler still appended synthesis citation rules (`Answer using ONLY numbered sources…`) with an **empty source list**. Small model invented citations anyway. `BuildCitations` dropped invalid refs server-side, but UI showed uncited prose that *looked* authoritative. |
| **2** | Wrong numbers — "1–2% EC", "1–2 weeks veg, 4–6 weeks flower" | **`lookup_crop_targets` did not fire** — intent regex requires `ec`, `ph`, `photoperiod`, etc.; operator said "feed" and "lighting" not "EC". Model free-formed from training data. EC must be **mS/cm** per Phase 64 — never percentages. |
| **3** | Ramps treated like indoor hydro (12/12, generic EC) | **No ramps profile or field guide**; model filled gap with cannabis-shaped defaults. Ramps (*Allium tricoccum*) are woodland spring ephemerals — wrong domain for gr33n fertigation cycles. |

### What should have happened

| Crop | Structured (Phase 64 today) | After Phase 82 |
|------|----------------------------|----------------|
| Cannabis | ✅ profile + guide | unchanged |
| Orchid | ✅ `phalaenopsis` + guide | alias `orchid` → phalaenopsis |
| Eggplant | ❌ missing | ✅ Tier A profile + guide |
| Cucumber, kale, herbs… | ❌ missing | ✅ Tier A/B library |
| Ramps | ❌ model guessed | ✅ **unsupported registry** — no fake EC/photoperiod |

---

## Crop library scope (Phase 64 → 82)

Phase 64 shipped **7** built-in profiles. Real operators grow dozens of crops across indoor rooms, greenhouses, and herb benches. Phase 82 expands the **offline bundled library** — not an agronomy authority, but curated starting points with cited sources (same boundary as Phase 64).

### Coverage tiers

| Tier | Ship in Phase 82 | Structured profile | Field guide RAG | Guardian behaviour |
|------|------------------|--------------------|-----------------|-------------------|
| **Existing (7)** | ✅ keep | cannabis, tomato, pepper, lettuce, phalaenopsis, basil, strawberry | 3 guides today → add missing 4 | cite profiles |
| **A — common indoor/greenhouse** | ✅ required | eggplant, cucumber, kale, spinach, cilantro, microgreens | one `crop-*.md` each | cite profiles + RAG |
| **B — frequent add-ons** | ✅ required | zucchini, green_bean, mint, parsley, blueberry, hemp, broccoli, melon, arugula | one guide each | cite profiles |
| **C — greenhouse / ornamental** | ✅ if time | rose, sunflower, hops, succulents, houseplant | brief guides | category defaults |
| **Unsupported** | ✅ required | **no profile** | `crop-unsupported-*.md` narrative | plain block + **cousin_of** suggestion |

### Full cultivator catalog (target state)

**Fruiting vegetables (high EC, long photoperiod)**

| crop_key | display_name | Notes |
|----------|--------------|-------|
| `cannabis` | Cannabis | existing; hemp aliases here or separate lower-EC profile |
| `tomato` | Tomato | existing |
| `pepper` | Pepper (bell/chili) | existing |
| `eggplant` | Eggplant | solanaceous; hand-pollination indoors |
| `cucumber` | Cucumber | vining; higher humidity than tomato |
| `zucchini` | Zucchini / summer squash | similar to cucumber, fruiting EC |
| `green_bean` | Green bean | moderate EC; warm |
| `strawberry` | Strawberry | existing |
| `blueberry` | Blueberry | acidic pH band; lower volume than tomato |
| `broccoli` | Broccoli | cool brassica; lower EC than fruiting |
| `melon` | Melon / cantaloupe | warm vining; high transpiration |

**Leafy greens & fast crops (low EC, cool)**

| crop_key | display_name | Notes |
|----------|--------------|-------|
| `lettuce` | Lettuce / leafy greens | existing |
| `kale` | Kale | slightly higher EC than lettuce |
| `spinach` | Spinach | cool-season; bolt notes |
| `arugula` | Arugula / rocket | fast turnover; bolt in heat |
| `microgreens` | Microgreens | very low EC; 10–14 day cycle; shallow wet/dry |

**Herbs (warm, moderate EC, continuous harvest)**

| crop_key | display_name | Notes |
|----------|--------------|-------|
| `basil` | Basil | existing |
| `cilantro` | Cilantro / coriander | bolts in heat — stage notes |
| `mint` | Mint | aggressive roots; container note |
| `parsley` | Parsley | biennial herb baseline |

**Epiphytes & ornamentals (low EC, RH, special watering)**

| crop_key | display_name | Notes |
|----------|--------------|-------|
| `phalaenopsis` | Orchid (Phalaenopsis) | existing; aliases: orchid |
| `succulent` | Succulents (general) | optional Tier C; dry-down not constant wet |
| `houseplant` | Houseplant (general) | optional Tier C; conservative defaults |

**Industrial / dual-use**

| crop_key | display_name | Notes |
|----------|--------------|-------|
| `hemp` | Hemp (fiber/seed) | separate profile or alias; not flower EC curve |

**Greenhouse / cut flower / specialty (Tier C)**

| crop_key | display_name | Notes |
|----------|--------------|-------|
| `rose` | Rose (cut flower) | long photoperiod; moderate EC |
| `sunflower` | Sunflower | short cycle; high light |
| `hops` | Hops (bines) | long veg; very different from cannabis flower |
| `succulent` | Succulents (general) | dry-down; never constant wet |
| `houseplant` | Houseplant (general) | conservative defaults; many aliases |

**Unsupported — honest “not a gr33n automation crop”**

| Mention | Reason | Guardian behaviour |
|---------|--------|-------------------|
| `ramps`, wild leek | woodland spring ephemeral | unsupported + suggest foraged woodland docs only |
| `mushroom`, fungi | different substrate/domain | unsupported → husbandry module |
| fruit trees, grape, apple | scale / years / root zone | unsupported; greenhouse seedling caveat only |
| `carrot`, `potato`, in-ground root crops | deep soil / field scale | unsupported; cousin `lettuce` or `tomato` for hydro only |
| `ginseng`, woodland medicinals | multi-year shade | unsupported |
| aquaponics fish | animals domain | Phase 20 animals |

---

## Design principle (unchanged from Phase 64)

| Layer | Holds | Guardian must |
|-------|-------|---------------|
| **Structured DB** (`crop_profiles` + stages) | EC, pH, VPD, DLI, photoperiod in **mS/cm** | State numbers **only** from `lookup_crop_targets` output |
| **RAG `field_guide`** | Why orchids want low EC, cannabis flush, etc. | Cite ingested chunks; **0 chunks → no doc citations** |
| **LLM** | Synthesis, comparisons, plain language | **Forbidden** to invent targets, fake `[n]` refs, or "% EC" |

Persona rule already exists (`CropTargetsGroundingRule` in `readtools_crop.go`) — enforcement when tools + RAG don't run is the gap this phase closes.

---

## WS0 — Ops prerequisites (ship first, no code required)

Document in [local-operator-bootstrap.md](../local-operator-bootstrap.md) and [operator-tour.md](../operator-tour.md):

1. **Field guides must be ingested** before crop Q&A is trustworthy:
   ```bash
   make rag-ingest-field-guides    # crop-*.md in field-guide manifest
   make rag-ingest-platform-docs   # operator how-to (optional for crop compare)
   ```
   Requires `EMBEDDING_API_KEY` (or LAN embedding endpoint). Verify via Farm Knowledge search or chat turn showing `context_count > 0`.

2. **Model floor for crop science:** `llama3.1:8b` is insufficient for citation discipline. Recommend **≥14B** local (e.g. `llama3.1:70b`, `qwen2.5:14b`) or operator-chosen cloud model. Document in `.env.example` comment on `LLM_MODEL`.

3. **"grounded" label clarification:** With `farm_id` set, `grounded=true` means **live farm snapshot attached**, not "RAG found documents". WS2 fixes UI copy; WS0 documents for operators until then.

---

## WS1 — Zero-chunk handler guardrail

**Files:** [`internal/handler/chat/handler.go`](../../internal/handler/chat/handler.go), [`internal/rag/synthesis/synthesis.go`](../../internal/rag/synthesis/synthesis.go), optional [`internal/rag/synthesis/guardian.go`](../../internal/rag/synthesis/guardian.go)

### Behaviour

When `farm_id` is set and `len(chunks) == 0` after retrieval (embedder configured, query ran, no matches):

| Today | After WS1 |
|-------|-----------|
| Append `GuardianRAGInstructions(chunks)` + `BuildUserMessage(question, [])` → empty "Sources:" list + cite-only system prompt | **Skip** synthesis citation block |
| Model invents `[1]…[5]` | Inject **`ZeroChunkGuardBlock`** persona text |
| `BuildCitations` returns `[]` (refs out of range) but answer keeps fake brackets | **Post-process:** strip `\[\d+\]` when `len(chunks)==0` before persist + SSE `done` |

### `ZeroChunkGuardBlock` (draft)

```
No indexed documentation matched this question (0 RAG chunks).
- Do NOT use [n] citation brackets.
- Do NOT state EC, pH, VPD, DLI, or photoperiod numbers unless lookup_crop_targets
  results appear above in this system prompt.
- For each crop mentioned: if lookup_crop_targets returned a profile, use those
  mS/cm values; if not, say you have no built-in profile and offer Start grow / Plants.
- For crops outside gr33n support (e.g. woodland ephemerals), say so plainly.
```

### Edge cases

- **Embedder nil (offline):** retrieval skipped — same guard when `grounded && len(chunks)==0` (farm snapshot only).
- **Retrieval error on local Ollama:** existing Phase 37 degrade path unchanged; guard applies when retrieval succeeds with 0 rows.
- **Non-streaming + streaming:** both paths share helper `applyZeroChunkPolicy(system, user, chunks, answer)`.

### Tests

- `handler_test.go`: farm_id + mock embedder returns 0 rows → answer has no `[1]`; system prompt contains ZeroChunkGuardBlock.
- `synthesis_test.go`: `StripOrphanCitationRefs(answer, 0)` removes bracket refs.

---

## WS2 — UI honesty & operator warning

**Files:** [`ui/src/components/GuardianChatPanel.vue`](../../ui/src/components/GuardianChatPanel.vue)

| Change | Detail |
|--------|--------|
| Metadata line | `grounded · 0 chunks` → **`farm context · 0 doc chunks`** (or split: farm snapshot icon + doc chunk count separately) |
| Warning banner | When `t.grounded && t.context_count === 0` and assistant text matches `\[\d+\]` or `\d+\s*%\s*ec` → amber strip: *"No indexed docs matched — numbers may be unreliable. Run field-guide ingest or assign crop profiles."* |
| Empty citations | If `context_count === 0`, hide citation list UI — don't show "References" header |

Optional: link to `/operator-guide` § Guardian + Farm Knowledge.

---

## WS3 — Read-tool widening & multi-crop lookup

**Files:** [`internal/farmguardian/readtools_crop.go`](../../internal/farmguardian/readtools_crop.go), tests in [`readtools_crop_test.go`](../../internal/farmguardian/readtools_crop_test.go)

Extends Phase 73 WS4 scope for crop-specific natural language.

### WS3a — Broaden `lookupCropTargetsIntent`

Add triggers:

- **Feeding / water:** `feed`, `water`, `watering`, `wet`, `moisture`, `irrigation`, `fertigation`, `how wet`
- **Light:** `light`, `lighting`, `hours`, `dli`, `photoperiod`, `cycle`, `program`
- **Compare:** `compare`, `vs`, `versus`, `difference between`, `each plant`
- **Crop names:** resolved via **alias registry** (WS4a), not a hardcoded Go slice

Fire `lookup_crop_targets` on the incident question without requiring literal "EC" or "pH".

### WS3b — Multi-crop render

Today `resolveCropProfileContext` returns **first** matching crop key only.

**New:** `renderLookupCropTargetsMulti` — detect all crop keys + aliases in question; for each:

1. Resolve alias → canonical `crop_key` via registry
2. Load builtin profile via `GetCropProfileByKey`
3. Append stage summary (compare questions: `early_veg` + `early_flower` defaults)
4. If key in **unsupported registry**: append plain block — **no EC/photoperiod numbers**
5. If key unknown and not unsupported: *"No built-in profile for X — clone from nearest cousin in Plants or request a profile."*

Cap at **6 profiles + 2 unsupported mentions** per turn (prompt budget).

### WS3c — Central alias registry (shared with UI)

Replace scattered string slices in `readtools_crop.go` with generated or loaded map:

```yaml
# data/crop_library.yaml (WS4a)
aliases:
  orchid: phalaenopsis
  aubergine: eggplant
  weed: cannabis
  marijuana: cannabis
  coriander: cilantro
  wild_leek: ramps   # → unsupported, not profile
unsupported:
  - ramps
  - mushroom
  - fruit_tree
```

Go: `farmguardian/crop_library.go` — `ResolveCropKey(mention string) (key string, unsupported bool)`.

### WS3d — `grow_advisor` on compare questions

Do **not** auto-fire grow_advisor for multi-crop hypotheticals (no active cycle). Only `lookup_crop_targets` multi + RAG narrative.

### Tests

- Incident question → intent true; multi block has cannabis + phalaenopsis mS/cm
- `ramps` → unsupported block, no photoperiod
- `cucumber vs tomato` → both profiles
- `aubergine` → resolves to eggplant profile

---

## WS4 — Cultivator crop library (profiles + guides + aliases)

Single source of truth drives SQL seed, Guardian aliases, UI picker, and RAG manifest.

### WS4a — Canonical catalog (`data/crop_library.yaml`)

**New file** — versioned YAML listing every crop the platform knows about:

```yaml
version: 2
crops:
  - key: eggplant
    display_name: Eggplant
    category: fruiting
    substrate: coco / rockwool slab
    watering_style: pulse_to_dryback        # constant_feed | pulse_dryback | top_water_drydown | mist_epiphyte
    runoff_pct_target: 10-20
    cousin_of: tomato                       # for unknown-crop suggestions
    aliases: [aubergine]
    stages: [...]
unsupported:
  - key: ramps
    aliases: [wild_leek, allium_tricoccum]
    reason: "Woodland spring ephemeral — not indoor fertigation"
    cousin_of: null
  - key: mushroom
    reason: "Different production domain"
    cousin_of: null
  - key: in_ground_root
    aliases: [carrot, potato, sweet_potato]
    reason: "Field / deep soil crops — gr33n targets hydroponic & container"
    cousin_of: lettuce
```

**Tooling:**

- `scripts/generate-crop-seed.sql.sh` — YAML → idempotent migration patch (same pattern as Phase 64 migration)
- CI check: YAML stage rows match `growth_stage_enum`; EC in mS/cm only
- Bump `crop_profiles.version` on curated edits

### WS4b — Tier A profiles (ship first)

Add built-in profiles + stages (minimum 2–3 stages each):

| crop_key | Model from | Distinctive targets |
|----------|------------|---------------------|
| `eggplant` | tomato −10% EC | hand-pollinate; warm |
| `cucumber` | tomato | higher RH; vining DLI |
| `kale` | lettuce +0.2 EC | cooler ok |
| `spinach` | lettuce | bolt temp note |
| `cilantro` | basil | bolt / cool preference |
| `microgreens` | lettuce −30% EC | 10–14 d cycle; shallow moisture |

**Also add missing field guides for existing 7:** pepper, lettuce, basil, strawberry (profiles exist; RAG narrative missing today).

### WS4c — Tier B profiles

| crop_key | Model from | Notes |
|----------|------------|-------|
| `zucchini` | cucumber | fruiting squash |
| `green_bean` | pepper | moderate EC |
| `mint` | basil | root containment note |
| `parsley` | basil | slightly cooler |
| `blueberry` | strawberry | pH 4.5–5.5 band |
| `hemp` | cannabis veg stages | **or** alias to cannabis with persona note — document choice in YAML |

| `hemp` | cannabis veg stages | document separate vs alias in YAML |
| `broccoli` | kale + cooler temps | brassica bolt |
| `melon` | cucumber | higher transpiration |
| `arugula` | lettuce | fast crop; heat bolt |

### WS4d — Field guides (RAG narrative)

**Per-crop guides** — one markdown per Tier A + B (+ Tier C if shipped):

```
crop-*-care.md / crop-*-nutrition.md   # per profile
```

**Cross-cutting plant intelligence guides (new):**

| Guide | Purpose |
|-------|---------|
| `crop-deficiency-patterns.md` | Interveinal yellowing, tip burn, purple stems — **by category** (fruiting / leafy / epiphyte); hypothesis not diagnosis |
| `crop-watering-substrates.md` | How wet for coco vs rockwool vs bark vs peat; ties to `watering_style` in YAML |
| `crop-stage-transitions.md` | When to flip, flush, harvest, rebloom, bolt — narrative companion to structured stages |
| `crop-unsupported-woodland.md` | Ramps, ginseng — why gr33n doesn't automate |

Update [`docs/rag/field-guide-manifest.yaml`](../rag/field-guide-manifest.yaml) — all files above.

**After merge:** `make rag-ingest-field-guides` on every farm with embedding configured.

### WS4e — Unsupported crop registry

**No structured targets** for unsupported keys — prevents llama from filling gaps with cannabis-shaped defaults.

Guardian block template:

```
lookup_crop_targets: {name} — not supported as a gr33n indoor fertigation crop.
{reason from YAML}. I can help with: {list nearest supported cousins by category}.
```

Detect mentions via alias registry (`ramps`, `wild leek`, `mushroom`, `shiitake`, `apple tree`, …).

### WS4f — UI profile picker

**Files:** Start-grow wizard, `/crop-profiles`, Plants assign profile

| Change | Detail |
|--------|--------|
| Grouped picker | Categories: Fruiting · Leafy · Herbs · Berries · Flower/epiphyte · Hemp |
| Search | Matches display_name + aliases (`aubergine` finds eggplant) |
| Unsupported | Not in picker — only surfaced via Guardian honest answer |
| Count badge | "22 crops" (or current count) in picker subtitle |

Detect mentions via alias registry. For unknown mentions with `cousin_of`:

> I don't have a profile for hops. Closest starting point: **cannabis vegetative** photoperiod targets — clone and adjust in Plants.

---

## WS7 — Plant context bundle (fuse live + reference data)

**Problem:** Plant questions today hit **one** read tool if a keyword matches. "How's my tomato doing?" should automatically pull everything relevant — not require the operator to say "EC VPD DLI summarize zone".

**New orchestrator:** `renderPlantContextBundle(ctx, farmID, question, ref, snap)` in `readtools_plant.go`.

When `shouldRunPlantContextBundleIntent(question, ref)` (broad — `my plant`, `this room`, `this grow`, crop names + `how`, `why`, `yellow`, `wilting`, `feed`, `water`, `light`, active zone/cycle ref):

| Block | Source | Always when |
|-------|--------|-------------|
| Active cycle + stage | snapshot / `GetActiveCropCycleForZone` | zone or cycle ref |
| `lookup_crop_targets` | WS3 multi or single | crop known |
| `grow_advisor` | Phase 62 | active cycle + sensors |
| `summarize_zone` | latest temp/RH/EC | zone resolved |
| `summarize_zone_fertigation` | last/next feed, program | zone resolved |
| `summarize_zone_lighting` | photoperiod, schedule | light-related or bundle default |
| Cycle analytics snippet | Phase 28 | "how's the run" phrasing |

**Persona rule:** Guardian must **lead with live vs target deltas** when bundle present (see WS11).

Cap total bundle at ~1.2k tokens; drop lowest-priority blocks first (analytics before targets).

---

## WS8 — Substrate-aware watering ("how wet?")

**Problem:** The incident question asked *"how wet do they like it?"* — no structured field exists today; the model guesses.

Add to each crop in `crop_library.yaml`:

| Field | Example values |
|-------|----------------|
| `substrate` | `coco`, `rockwool`, `orchid_bark`, `DWC`, `peat`, `soilless_mix` |
| `watering_style` | `constant_feed`, `pulse_dryback`, `top_water_drydown`, `mist_epiphyte`, `dry_down_succulent` |
| `moisture_guidance` | Plain operator sentence in YAML |
| `runoff_pct_target` | e.g. `10–20%` for fruiting hydro |

**Read tool output:**

```
lookup_crop_targets — tomato · watering
substrate: coco/rockwool
style: pulse to ~10–20% dryback between feeds
moisture: keep root zone evenly moist — not soggy; first inch dryback is normal
runoff: 10–20% per feed event
```

Guardian **must** quote this block for "how wet" questions — never "% EC" for moisture.

---

## WS9 — Symptom & deficiency intelligence

**Builds on Phase 67 vision** — same hypothesis band, richer reference text.

### WS9a — Symptom intent

Trigger on: `yellow`, `brown`, `spot`, `wilting`, `drooping`, `curl`, `tip burn`, `purple`, `deficien`, `nutrient`, `lockout`, `pest`, `mold`, `powdery`, `mildew`, `bugs`.

Auto-run:
1. **WS7 plant context bundle** (live EC/pH vs target)
2. RAG filter `source_type=field_guide` boost for `crop-deficiency-patterns.md`
3. If photo attached → existing `VisionContextBlock` + crop profile block

### WS9b — Answer shape (persona)

```
1. What I see (or what you described)
2. Hypothesis ranked (magnesium lockout vs natural senescence) — tied to live EC/pH if available
3. Next checks (runoff pH, lower/raise feed, inspect undersides)
4. Optional: propose create_task — not silent program changes
```

**Hard rule:** never certified diagnosis; never pesticide dose without operator label photo.

---

## WS10 — Stage transition intelligence

Widen `grow_advisor` + add `lookup_crop_targets` stage notes for:

| Transition | Crops | Signals |
|------------|-------|---------|
| Flip to flower | cannabis, hemp | height, node spacing, days in veg |
| Harvest window | cannabis, tomato, pepper | stage enum + days in flower/fruit |
| Bolt / go to seed | cilantro, spinach, arugula, lettuce | temp + daylength |
| Orchid rebloom | phalaenopsis | spike cut, temp drop |
| Flush | cannabis | pre-harvest EC taper from profile |

**Intent widening:** `shouldRunGrowAdvisorReadIntent` fires on `my room`, `this grow`, `my plant`, `how's it doing`, zone/cycle `context_ref` without requiring "VPD" keyword.

**Output includes** structured `days_in_stage` + profile `notes` field from YAML.

---

## WS11 — Target vs reality reconciliation

When WS7 bundle includes both **live readings** and **profile targets**, Guardian answers in delta form:

> Your Flower Room is at **EC 1.4 mS/cm** (target **1.6–1.8** for early flower — a little light). **VPD 1.35 kPa** (target **1.0–1.2** — slightly high). **Photoperiod 11.2 h** (target **12 h** — extend lighting schedule or check timer).

**Comfort band conflicts:** if zone comfort RH band ≠ crop profile RH, call it out:

> Comfort band says 50–60% RH but your cannabis early-flower profile wants 45–55% — crop stage usually wins; consider adjusting comfort or accepting higher mold risk.

**Greenhouse outdoor light (Phase 66):** when `site_weather` available + crop has `dli_target`, compare clear-sky DLI to target for supplemental light advice.

New read helper: `renderTargetVsActualBlock` — pure formatting, no LLM math.

---

## WS5 — Chat follow-up chips (conversation loop)

**Started:** [`ui/src/lib/guardianFollowUps.js`](../../ui/src/lib/guardianFollowUps.js) + [`GuardianChatPanel.vue`](../../ui/src/components/GuardianChatPanel.vue) — keyword-derived 2–3 chips after last assistant turn.

**Finish in this phase:**

- Chip styling consistent with [`GuardianStarterChips.vue`](../../ui/src/components/GuardianStarterChips.vue) (or intentionally subtler — document choice)
- Hide chips while streaming; clear on new user-typed send
- Extend detectors for multi-crop compare follow-ups ("Set up eggplant profile", "Show me orchid watering signs")
- [`ui/src/__tests__/guardian-follow-ups.test.js`](../../ui/src/__tests__/guardian-follow-ups.test.js) — keep green; add multi-crop case

---

## WS6 — Docs, architecture, closure

| Artifact | Content |
|----------|---------|
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | **§7.0y** — zero-chunk, multi-crop, unsupported; **§7.0z** — plant context bundle, substrate watering, deficiency band |
| [operator-tour.md](../operator-tour.md) | § Guardian plant intelligence — ingest, assign profile, photo + voice |
| [local-operator-bootstrap.md](../local-operator-bootstrap.md) | Guardian-ready: ingest all crop guides, model floor, vision model |
| `docs/crop-library-operator-guide.md` | New — what profiles cover, unsupported list, clone-to-customize |
| `cmd/api/smoke_phase82_test.go` | Zero-chunk; multi-crop; plant bundle logs; symptom intent |
| `ui/src/__tests__/phase-82-closure.test.js` | UI banner + picker groups |

**OC-82** closes when all WS DoD met.

---

## Definition of done

- [ ] **≥25 built-in crop profiles** (7 existing + Tier A + B + Tier C if shipped)
- [ ] **`crop_library.yaml`** drives seed, aliases, substrates, watering, unsupported, cousin_of
- [ ] **Field guides** for every profile + cross-cutting deficiency/watering/stage guides; re-ingest documented
- [ ] **Plant context bundle** fires on natural plant questions; answers use live vs target deltas
- [ ] **"How wet?"** answered from `watering_style` — never guessed moisture
- [ ] **Symptom questions** pull deficiency guide + live EC/pH; vision stays hypothesis-only
- [ ] **Stage transitions** (flip, harvest, bolt, rebloom) cite profile notes + grow_advisor
- [ ] Incident multi-crop compare + ramps unsupported — no fake citations at 0 chunks
- [ ] UI picker + follow-up chips + honesty banners
- [ ] Go smoke + Vitest; **OC-82** closed

---

## Suggested implementation order

1. **WS4a** — `crop_library.yaml` schema (include substrate/watering/cousin)
2. **WS4b–WS4c** — profiles + Tier guides
3. **WS3** — multi-crop lookup + alias registry
4. **WS7** — plant context bundle (biggest intelligence win)
5. **WS8** — substrate watering blocks
6. **WS11** — target vs actual formatter
7. **WS1** — zero-chunk guardrail
8. **WS9–WS10** — symptom + stage transitions
9. **WS4d–WS4f** — cross-cutting RAG guides + UI picker
10. **WS2, WS0, WS5, WS6** — polish + closure

---

## Out of scope

| Topic | Where |
|-------|--------|
| Enterprise seed pack, bootstrap script, site manifest hook, farm crop override UI, scheduled ingest | **[Phase 83](phase_83_enterprise_agronomy_seed_pack.plan.md)** (**shipped** — [`phase-83-closure.md`](phase-83-closure.md)) |
| Per-genetics ML / auto-tune from harvest | Phase 84+ — needs data volume + training pipeline |
| Certified pest ID / pesticide prescriptions | regulatory |
| Operator-uploaded PDF plant notes RAG | Phase 53 roadmap item |
| Proactive "EC drifting 3 days" nudges | Phase 61 nudge extension |
| Autonomous program creation from chat | still Confirm-gated proposals only |
| Every global crop species | unsupported + cousin_of covers long tail |

---

## Related

| Doc | Use |
|-----|-----|
| [phase_64_crop_knowledge_base.plan.md](phase_64_crop_knowledge_base.plan.md) | Structured targets source of truth |
| [phase_62_guardian_grow_advisor.plan.md](phase_62_guardian_grow_advisor.plan.md) | Active-cycle grow science |
| [phase_73_guardian_pr_discoverability.plan.md](phase_73_guardian_pr_discoverability.plan.md) | Read-tool reliability (overlapping WS4) |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | field_guide RAG + procedures |
| [phase_67_guardian_field_assistant.plan.md](phase_67_guardian_field_assistant.plan.md) | Vision + deficiency hypotheses |
| [phase_66_weather_site_context.plan.md](phase_66_weather_site_context.plan.md) | Outdoor DLI (WS11) |
| [phase_28_crop_intelligence_guardian_depth.md](phase_28_crop_intelligence_guardian_depth.md) | Cycle analytics in snapshot |
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | Guardian layers §7 |
| [phase_83_enterprise_agronomy_seed_pack.plan.md](phase_83_enterprise_agronomy_seed_pack.plan.md) | Enterprise bootstrap + override packs |
| [docs/rag/field-guide-manifest.yaml](../rag/field-guide-manifest.yaml) | Crop guide ingest list |
| [internal/farmguardian/readtools_crop.go](../../internal/farmguardian/readtools_crop.go) | lookup_crop_targets |
| [internal/handler/chat/handler.go](../../internal/handler/chat/handler.go) | Chat turn assembly |

---

## Using this in a new chat

> Read `docs/plans/phase_82_guardian_crop_grounding_hardening.plan.md`. Build full plant intelligence: `crop_library.yaml` (≥25 crops, substrates, watering), plant context bundle WS7, target-vs-actual WS11, deficiency guides WS9, zero-chunk guardrail WS1. Start WS4a + WS7. mS/cm only. Unsupported crops get cousin suggestions, not fake targets.
