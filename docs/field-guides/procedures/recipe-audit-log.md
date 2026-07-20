# Natural farming recipe audit log (Phase 208 WS0)

Audited `db/seeds/master_seed.sql` and `gr33ncore._bootstrap_jadam_indoor_photoperiod_v1`
against **Youngsang Cho, *JADAM Organic Farming*, 2016** and KNF references (Cho Han-kyu / UH CTAHR).

| Change | Was | Now | Citation |
|--------|-----|-----|----------|
| JMS Soil Drench dilution | 1:500 | **1:10** | Cho 2016 — JMS soil application |
| JMS Foliar Spray dilution | 1:500 | **1:20 + JWA** | Cho 2016 — JMS foliar; JWA for coverage |
| JLF and JMS Combined (JMS part) | JMS 1:500 | **JMS 1:10** | Cho 2016 — combined tank same as soil drench |
| Combined recipe component `part_value` | 0.025 | **2.0** | Derived: (1/10)/(1/20) relative to JLF base |
| JLF General Soil Drench instructions | 1:20 only | **Start 1:100; 1:20 when tested** | Cho/FigJam conservative start |
| JMS preparation / storage | 3–7 days generic | **Peak foam 24–72 h; use within 6–12 h of peak** | Cho 2016 — JMS active window |
| JLF Spring description | "nitrogen-fixing plants" | **dynamic accumulator biomass** | Botany — nettle/comfrey are not N-fixers |
| JHS preparation | Simmer 1–3 h | **Boil 1 kg plant in 4–5 L water 4–5 h** | Cho 2016 — JHS method |
| JS input | Wettable sulfur 0.5% shortcut | **JS (JADAM Sulfur concentrate)** — caustic batch | Cho 2016 — real JADAM JS |
| JS Fungicide Spray | 0.5% wettable sulfur | **0.5–2 L concentrate per 500 L + JWA** | Cho 2016 — JS application band |
| FFJ bootstrap ingredients | fruit, sugar, **water** | **fruit + brown sugar only** | KNF standard — no water in FFJ |
| LAB, FPJ, FFJ, OHN, WCA, WCS `reference_source` | Cho 2016 | **KNF (Cho Han-kyu); often used with JADAM** | Tradition honesty — sugar-based KNF inputs |
| FAA input | missing | **FAA (Fish Amino Acid)** added | KNF standard — fish + sugar 1:1; dilute ~1:1000 |

**Not changed (intentional):** `Veg Daily JLF Program.dilution_ratio = '1:500'` on farm 1 remains for Phase 39 MixPlan EC demo — not a JMS recipe citation.

Migration: `db/migrations/20260720_phase208_ws0_recipe_audit.sql`
