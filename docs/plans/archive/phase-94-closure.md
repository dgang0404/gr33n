# Phase 94 — closure (OC-94)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_94_genetics_batch_ec_profiles.plan.md`](phase_94_genetics_batch_ec_profiles.plan.md)

**Depends on:** [Phase 87](phase_87_crop_knowledge_operator_closure.plan.md) (v1 runbook scope); [Phase 93](phase_93_plant_identity_vocabulary_cleanup.plan.md) (`variety_or_cultivar`).

**Closes:** Blind spot **#3** — farm-wide `crop_key` override vs per-variety EC (Blue Dream vs Wedding Cake).

---

## The one job (done)

> **Per-variety EC overrides** keyed by `plants.variety_or_cultivar` on a farm, resolved **above** farm `crop_key` Settings and below nothing — same chain for strip, hints, effective API, and Guardian `lookup_crop_targets`.

---

## Resolution order

```
1. genetics profile (farm + crop_key + variety slug) — if linked
2. farm crop_key override (Settings PUT)
3. platform builtin profile
```

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Runbook v1 vs v2 genetics scope | [`crop-knowledge-operator-runbook.md`](../crop-knowledge-operator-runbook.md) |
| **WS1** | `plant_genetics_profiles` link table + synthetic `genetics:{crop_key}:{slug}` profiles | migration `20260629_phase94_plant_genetics_profiles.sql` |
| **WS2** | Effective + genetics CRUD API | `GET …/effective`, `PUT/GET/DELETE …/genetics/{variety_slug}` |
| **WS3** | Plants — **Tune EC for this variety →** | `Plants.vue`, `getEffectiveCropProfile` in farm store |
| **WS4** | Guardian uses `ResolveProfileID` when variety set | `readtools_crop.go` |
| **WS5** | Precedence smoke | `smoke_phase94_test.go` |

---

## Operator behavior

| Surface | Genetics override |
|---------|-------------------|
| **Plants** | Link to Settings when `variety_or_cultivar` is set |
| **Zone strip / hints** | `GET …/crop-profiles/effective?crop_key=&variety=` |
| **Settings** | Farm `crop_key` table unchanged; genetics rows hidden from list |
| **Guardian** | `lookup_crop_targets` prefers genetics EC when active plant has variety + linked profile |

Without a genetics profile, behavior matches Phase 87 (farm-wide `crop_key` only).

---

## Automated tests

| Test | Path |
|------|------|
| Genetics > farm precedence | `cmd/api/smoke_phase94_test.go` |
| Slug / synthetic key helpers | `internal/cropprofile/effective_test.go` |

---

## OC-94

Phase 94 is **closed** when genetics EC overrides beat farm-wide `crop_key` targets on the effective API, zone UI, and Guardian read path.
