# UI test ownership

Where to add tests when you change hot surfaces — avoids copy-paste
`readFileSync` blocks in `phase-*-closure.test.js`.

## Guardian chat panel

| Concern | Home |
|---------|------|
| Panel store, drawer mount, POST body | `ui/src/__tests__/guardian-panel.test.js` |
| Template anchors (`data-test`, wiring strings) | `ui/src/__tests__/guardian-chat-panel-source.test.js` |
| Citations, labels, links | `guardian-citation-links.test.js`, `guardian-citation-labels.test.js` |
| Phase-specific mount behavior | Keep in that phase closure **only** when it exercises runtime (e.g. Phase 197 pending chip mount) |

**Do not** add new `readFileSync(...GuardianChatPanel.vue)` blocks to phase closures.

## Today / Dashboard

| Concern | Home |
|---------|------|
| Workspace link helpers | `dashboard-workspace-links.test.js` |
| Hero arc wiring (166–177) | `today-excellence-arc.test.js` |
| Section order / a11y | `phase-177-today-a11y.test.js` |
| Component libs (filter, pulse, ask) | `farm-today-*.test.js` |

**Do not** re-assert `Dashboard.vue` imports in phase-166+ closures — extend `today-excellence-arc.test.js`.

## Phase closures

Phase `phase-N-closure.test.js` files lock **phase-specific** behavior:

- Go handler / migration wiring
- New lib exports unique to that phase
- Docs marked shipped

They are not the place to re-check template strings already covered by canonical modules above.

## Guardrails

`phase-202-closure.test.js` enforces:

- Phase-closure files mentioning `GuardianChatPanel` ≤ 9
- Phase-closure files mentioning `Dashboard.vue` ≤ 7
- Single `readFileSync` scanner for `GuardianChatPanel.vue`
- Dashboard scanners limited to `today-excellence-arc`, `dashboard-workspace-links`, `phase-177-today-a11y`

Run UI tests: `cd ui && npm test -- --run`
