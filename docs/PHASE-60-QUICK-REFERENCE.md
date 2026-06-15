# Phase 60 — Quick Reference & Overview

> **Phase 60: Pi Setup Wizard & Onboarding Clarity**  
> Eliminate every gap in the Pi → UI → Pi experience. Interactive 6-step wizard + comprehensive tooltips + network diagnostics.

---

## The Problem (from UX audit)

| What's hard today | Why it's hard | User impact |
|-------------------|---------------|------------|
| DIP switches | Reference table only; easy to mis-set | Hardware doesn't work; Pi can't be detected |
| Config generation | Manual SSH + YAML editing | YAML indentation errors; fields unclear |
| Network setup | Hidden in config file | Silent failure; hard to debug "is my Pi even connected?" |
| Understanding terms | No glossary (I²C, GPIO, channel, relay, stack) | Non-IT farmer is lost; context-switches to Google |
| Wiring overview | Click each actuator individually | No sense of completion; unclear if setup is done |

**Current clarity:** 7/10  
**Target clarity:** 9.5/10

---

## The Solution: 6-Step Wizard

```
┌─────────────────────────────────┐
│  STEP 1: Welcome & Checklist    │
│  □ Do you have Pi, HAT, power?  │
└─────────────────────────────────┘
               ↓
┌─────────────────────────────────┐
│ STEP 2: Register Pi on Farm     │
│ Name, UID, API key generation   │
└─────────────────────────────────┘
               ↓
┌─────────────────────────────────┐
│ STEP 3: Assign Relay Channels   │
│ Visual grid + DIP calculator    │
└─────────────────────────────────┘
               ↓
┌─────────────────────────────────┐
│ STEP 4: Network & API Config    │
│ Base URL + connectivity test    │
└─────────────────────────────────┘
               ↓
┌─────────────────────────────────┐
│ STEP 5: Download Config         │
│ YAML file (no manual editing)   │
└─────────────────────────────────┘
               ↓
┌─────────────────────────────────┐
│ STEP 6: Verify & Complete       │
│ Deployment checklist + test     │
└─────────────────────────────────┘
               ↓
            ✅ Done!
```

**Non-IT farmer: never leaves the wizard. Never opens a terminal. Never edits YAML by hand.**

---

## Key Features

### 1. **Comprehensive Tooltips**
Every unfamiliar term has inline help:
- I²C, GPIO, BCM pin, relay, channel, stack level, DIP switch, I²C address
- **Search glossary** by term or concept
- **Links to docs** for deeper learning

### 2. **Interactive DIP Calculator**
- Input: target channel (0–63)
- Output: which DIP switches to set (ON/OFF for ID0, ID1, ID2)
- Visual + text; matches physical board diagram
- Prevents mis-configuration

### 3. **Real-time Validation**
- Banner shows: ✓ complete, ⚠ warning, ✗ blocker
- "Next" button disabled until all blockers resolved
- Clear error messages (e.g., "Nutrient Pump A needs an actuator assigned")

### 4. **Network Test**
- Single button: "Test connectivity"
- Pi calls back to API (proves both can reach each other)
- Shows latency + pass/fail + friendly errors
- Helps diagnose firewall, DNS, network issues

### 5. **Config Download (No Manual Editing)**
- Click "Download config.yaml"
- YAML file is auto-generated, pre-filled, with comments
- User copies to Pi (scp one-liner provided)
- Zero YAML knowledge required

### 6. **Deployment Checklist**
- Final step: ✓ config copied, ✓ service restarted, ✓ wiring correct, etc.
- Optional: test one actuator (pump pulse for 1s)
- Clear success state

---

## Files & Docs Created

### 1. **Phase Plan**
📄 [`docs/plans/phase_60_pi_setup_wizard_ux.plan.md`](docs/plans/phase_60_pi_setup_wizard_ux.plan.md)
- Full spec: 10 workstreams, success criteria, closure contract
- Completely standalone (doesn't touch phases 50–59)

### 2. **UI Design Guide**
📄 [`docs/pi-setup-wizard-design-guide.md`](docs/pi-setup-wizard-design-guide.md)
- Mockups of all 6 steps
- Copy/tone guidance
- Validation states
- Mobile / accessibility notes
- Perfect for designers & frontend devs

### 3. **Glossary Data**
📄 [`ui/src/lib/phase-60-glossary.js`](ui/src/lib/phase-60-glossary.js)
- 15+ terms with definitions, context, links
- Searchable
- Ready to plug into components
- Covers: I²C, GPIO, BCM pin, relay, channel, stack, DIP, API key, config.yaml, etc.

### 4. **Implementation Checklist**
📄 [`docs/phase-60-implementation-checklist.md`](docs/phase-60-implementation-checklist.md)
- Component-by-component breakdown
- Files to create
- Pinia store schema
- API endpoints (Go code outline)
- Test strategy
- ~4300 LOC estimate, 4–6 weeks

---

## Terminology: What We're Adding

| Term | What is it? | Why matters? |
|------|-----------|-------------|
| **I²C** | Serial protocol (daisy-chain on 2 wires) | Lets you stack 8 relay cards on 2 GPIO pins |
| **GPIO** | Programmable pins on the Pi | Direct control of relays/sensors |
| **BCM pin** | Pin numbering scheme (GPIO 17, 27, etc.) | How you reference which pin you're using |
| **Relay** | Electric switch (on/off) | Controls high-power devices safely |
| **Channel** | Numbered slot for one relay (0–63) | How gr33n refers to relays; scalable numbering |
| **Stack level** | Physical position of card (0, 1, 2, …) | Determines DIP switches + I²C address |
| **DIP switch** | 3 tiny switches (ID0, ID1, ID2) | Sets unique address for each card; must match stack level |
| **I²C address** | Hex code (0x20–0x27) | How the Pi distinguishes cards; set by DIP switches |
| **API key** | Secret token (gdev_*) | Proves the Pi is trusted; like a password |
| **Config.yaml** | Bootstrap text file on the Pi | Tells Pi where API is + its farm_id (minimal, auto-generated) |
| **Network test** | Connectivity check | Proves Pi ↔ API can reach each other |

**All are explained in tooltips, searchable glossary, linked to docs.**

---

## Success Metrics

When Phase 60 ships:

✅ **Completeness:** Non-IT farmer never leaves `/pi-setup-wizard`; never uses terminal or text editor.

✅ **Clarity:** Every field has a tooltip; 15+ terms have glossary entries with links to docs.

✅ **Validation:** Real-time banner shows blockers; user can't submit incomplete setup.

✅ **Confidence:** Network test gives pass/fail + latency; user knows if config is correct before deploying.

✅ **Speed:** Experienced user: 5 min. First-timer: 15–20 min.

✅ **Testing:** E2E test covers full 6-step flow; unit tests for each component.

✅ **Documentation:** Phase plan + design guide + implementation checklist + glossary all ready.

---

## Scope: What Phase 60 Does & Doesn't

### ✅ What Phase 60 DOES

- 6-step interactive wizard (UI only; doesn't modify existing routes/APIs)
- Comprehensive tooltips & glossary
- Network diagnostics endpoint (`POST /devices/{id}/network-test`)
- Config download endpoint (`GET /devices/{id}/config/download`)
- DIP calculator (no manual math)
- Real-time validation
- E2E test coverage
- Design guide + implementation checklist

### ❌ What Phase 60 DOES NOT

- Modify phases 50–51 (platform sync, wiring model)
- Force migration of existing setups
- Add video/animations (can be Phase 61)
- Auto-SSH to Pi (user still copies files; scp one-liner provided)
- Mobile-first design (mobile layout is secondary)
- Localization (designed with i18n in mind, but English-only launch)

---

## How It Addresses Every Gap

| Gap from audit | How Phase 60 fixes it |
|----------------|----------------------|
| "DIP switches confusing" | Interactive calculator shows exactly which switches to flip for your target channel |
| "No glossary for non-IT terms" | 15+ terms with tooltips, searchable glossary, links to docs |
| "Config generation is manual" | Download button generates YAML; zero manual editing |
| "Network troubleshooting is silent" | Network test endpoint; clear pass/fail/latency feedback |
| "Wiring setup is fragmented" | All 6 steps in one wizard; real-time validation banner |
| "No sense of completion" | Deployment checklist at end; optional actuator test |
| "Terms undefined (I²C, GPIO, relay)" | Every term has inline help + glossary entry |
| "Must context-switch between UI pages" | Everything in one wizard flow |

---

## Roadmap: What Happens After?

### Phase 61+ Potential Enhancements

- **Video tutorials:** 5 short clips (HAT assembly, DIP switches, first boot, wiring, troubleshooting)
- **Mobile-first redesign:** Tablet + phone-optimized layout
- **Batch provisioning:** Set up 5 Pis at once
- **SSH integration:** Optional auto-copy-to-Pi (for tech-savvy users)
- **Multilingual:** French, Spanish, Mandarin (glossary first, then steps)
- **QR code:** Generate QR for API key + config (scan on phone for sharing)

---

## Files Summary

| File | Purpose | Status |
|------|---------|--------|
| [`phase_60_pi_setup_wizard_ux.plan.md`](docs/plans/phase_60_pi_setup_wizard_ux.plan.md) | Full spec + workstreams | ✅ Created |
| [`pi-setup-wizard-design-guide.md`](docs/pi-setup-wizard-design-guide.md) | UI mockups + copy + accessibility | ✅ Created |
| [`phase-60-glossary.js`](ui/src/lib/phase-60-glossary.js) | Glossary data + search functions | ✅ Created |
| [`phase-60-implementation-checklist.md`](docs/phase-60-implementation-checklist.md) | Dev guide + file structure + LOC estimate | ✅ Created |

---

## Next Steps for Implementation

1. **Week 1:** Create component skeleton files + Pinia store
2. **Week 2:** Build Step 1–3 (entry, registration, channels)
3. **Week 3:** Implement DIP calculator + network test endpoints
4. **Week 4:** Build Step 4–6 (network, download, confirm)
5. **Week 5:** Add tests (Vitest + Cypress) + refine based on feedback
6. **Week 6:** Polish, docs screenshots, OC-60 closure

**Total:** 4–6 weeks, 1 FTE.

---

## Questions & Answers

**Q: Will this break existing Pi setups?**  
A: No. Phase 60 reads from existing APIs (phases 50–51) but doesn't modify them. Existing Pis continue working unchanged.

**Q: Can farmers skip the wizard?**  
A: Yes. The wizard is at `/pi-setup-wizard`. The old reference guide remains at `/pi-setup`. Advanced users can still register devices manually.

**Q: How does this relate to phase 51 (platform sync)?**  
A: Phase 51 handles the **Pi ↔ API contract** (how Pi fetches wiring). Phase 60 handles the **UI → Pi setup experience** (how humans set it up). They're complementary, not conflicting.

**Q: What if the network test fails?**  
A: User gets a clear error (DNS failed, timeout, firewall, etc.) + links to troubleshooting. They can still proceed (might work after fixing), or go back and edit the URL.

**Q: Is this only for Sequent HATs?**  
A: Phase 60 covers the common case (Sequent HAT stack). Direct GPIO wiring (single relay on one pin) is also supported as a fallback mode in Step 3.

---

## Contact & Feedback

This phase was designed to eliminate **every gap** from the UX audit (7/10 → 9.5/10). Feedback welcome!

- Review the design guide mockups
- Suggest glossary terms we missed
- Point out unclear copy
- Contribute test cases
