---
name: Phase 60 — Pi Setup Wizard & Onboarding Clarity
overview: >
  Close every gap in the Pi → UI → Pi loop. Add an interactive setup wizard with contextual
  tooltips, network diagnostics, auto-generated config download, and step-by-step guidance.
  Farmer with zero IT background can wire a Pi from the UI alone—no SSH, no YAML editing,
  no guessing about DIP switches or channel numbers. Standalone effort; does not depend on
  or disturb Phases 50–59.
todos:
  - id: ws1-wizard-component
    content: "WS1: Create PiSetupWizard.vue component with 6-step flow: Hardware checklist → Register Pi → DIP & wiring → Network test → Config preview → Confirmation"
    status: not-started
  - id: ws2-comprehensive-tooltips
    content: "WS2: Add HelpTip / InfoPanel components throughout; cover: I²C vs GPIO, DIP addressing, channel math, relay vs pin, IP discovery, API key security"
    status: not-started
  - id: ws3-network-diagnostics
    content: "WS3: POST /devices/{id}/network-test — Pi → API connectivity check; diagnose firewall, DNS, timeout, response time; return human-readable feedback (✓ OK, ✗ unreachable, ⚠ slow)"
    status: not-started
  - id: ws4-config-download
    content: "WS4: Generate minimal config.yaml on-the-fly in the UI (no manual editing); download button produces YAML with api.base_url + api_key + device_uid + farm_id; includes live comments"
    status: not-started
  - id: ws5-dip-calculator
    content: "WS5: Interactive DIP switch calculator — select desired channel or stack number; live visual feedback (ON/OFF for ID0/ID1/ID2); matches physical board diagram"
    status: not-started
  - id: ws6-channel-assign-flow
    content: "WS6: Streamlined actuator/sensor wiring — single panel to assign all actuators at once; visual grid showing current assignments; one-click reassign"
    status: not-started
  - id: ws7-glossary-panel
    content: "WS7: Inline glossary — hover or click terms (channel, relay, GPIO, I²C, BCM pin) → quick definition in context"
    status: not-started
  - id: ws8-validation-feedback
    content: "WS8: Real-time validation — warn if channel is out of range, device unassigned, wiring incomplete; show green ✓ only when ready to deploy"
    status: not-started
  - id: ws9-video-or-animations
    content: "WS9: Optional—embed mini-videos or animated GIFs showing physical stacking, DIP switch positioning, wire routing (can be added post-ship)"
    status: not-started
  - id: ws10-docs-and-tests
    content: "WS10: Update pi-setup guide with embedded wizard screenshots; add Vitest + E2E tests for wizard flow; OC-60 closure"
    status: not-started
isProject: false
---

# Phase 60 — Pi Setup Wizard & Onboarding Clarity

## Status

**Not started.** Scoped as a standalone sprint; does not depend on or modify Phases 50–59 beyond read-only API calls.

---

## Problem

From earlier UX audit:

| Step | Current blocker | Farmer impact |
|------|-----------------|---------------|
| **Hardware assembly** | DIP switch table is a reference, not guided | Must flip switches by hand; easy to mis-set address |
| **Register Pi in UI** | Simple but isolated; doesn't tell them *why* | Rote checkbox, no mental model |
| **Assign wiring in UI** | Click each actuator individually; no overview | No sense of "am I done?"; unclear what channel numbers mean |
| **Generate config** | Not in the UI; must SSH and edit YAML by hand | Non-IT user hits YAML indentation errors; doesn't understand fields |
| **Network setup** | Hidden in a config file | No feedback until Pi fails to connect; hard to debug |
| **Troubleshoot** | Cryptic error messages | "API unreachable" — is it my network? My Pi? Firewall? |

**Net:** User can follow docs and *eventually* get it working, but the experience is **7/10 clear**. Gaps:
- No wizard flow (must context-switch between Settings, Hardware, Controls)
- No glossary for non-IT terms (I²C, GPIO, channel, stack, DIP)
- Config generation is manual + error-prone
- Network diagnostics are silent
- No feedback when wiring is incomplete

**Target:** **9.5/10 clear** — farmer never leaves the wizard, understands every choice, and sees live validation.

---

## Solution

### Part A: Six-step Wizard

**Entry point:** `/pi-setup-wizard` (new route) or button in Settings / Hardware.

```
Step 1: Welcome & Hardware Checklist
  ├─ What is a Relay HAT? (with inline video/GIF placeholder)
  ├─ Parts list (interactive; can check off ✓)
  └─ "Next: Register your Pi"

Step 2: Register Pi in Dashboard
  ├─ Device name (e.g., "Flower Room Pi")
  ├─ Device UID (auto-suggest from hostname or let them type)
  ├─ Generate API key (show `gdev_*` secret once, warn "save now")
  └─ "Next: Hardware Wiring"

Step 3: Assign Relay Channels (Visual + Interactive)
  ├─ Visual grid: 8 cards (ch0–7 for first stack)
  ├─ For each channel:
  │   ├─ Dropdown: "What actuator goes here?" (e.g., "Main Pump")
  │   ├─ If relay_hat: show DIP calculator inline
  │   └─ Instant validation: ✓ OK or ⚠ Missing actuator def
  ├─ "Preview" button → shows all assignments
  └─ "Next: Configure Network"

Step 4: Network & Bootstrap Config
  ├─ API base URL (auto-fill from current origin; let them override)
  ├─ Test connectivity: "Check if your Pi can reach the API"
  │   └─ Button runs POST /devices/{id}/network-test
  │   └─ Shows: ✓ Connected in 120ms | ⚠ Timeout | ✗ Refused | ⚠ No route
  ├─ Live config preview (read-only YAML box)
  └─ "Next: Download & Deploy"

Step 5: Download Config
  ├─ Show generated config.yaml (read-only syntax-highlighted)
  ├─ Copy-to-clipboard button
  ├─ Download button (config.yaml file)
  ├─ Show scp/rsync one-liner if user provides Pi hostname/IP
  └─ "Next: Verify"

Step 6: Confirm & Verify
  ├─ Checklist:
  │   ├─ [ ] config.yaml copied to Pi
  │   ├─ [ ] Pi service restarted
  │   ├─ [ ] Actuator wiring is physically correct
  │   └─ [ ] All sensors/relays assigned
  ├─ "Test actuator" — single relay pulse to verify
  └─ "Finish" → dashboard / Hardware board with this Pi pre-selected
```

---

### Part B: Comprehensive Tooltips & Glossary

Add a `TermGlossary.vue` component (bottom-right) with click-to-expand entries:

| Term | Tooltip (1–2 sentences) | Learn more link |
|------|-------------------------|-----------------|
| **I²C** | Serial protocol that daisy-chains devices on 2 wires (clock + data). All Sequent relay cards stack on one I²C bus. | [`pi-sequent-hat-setup.md`](../pi-sequent-hat-setup.md#i2c) |
| **GPIO** | General Purpose I/O pins on the Pi; direct digital on/off. Used for single relays or sensors. | — |
| **BCM pin** | Broadcom pin numbering (GPIO 17, 27, etc.); the *numbers* on the Pi's pinout diagram. | — |
| **Relay** | Electronic switch; closes a circuit when powered. Sequent HAT has 8 per card. | — |
| **Channel** | Numbering for relay cards: 0–7 (card 1), 8–15 (card 2), etc. | [`pi-sequent-hat-setup.md`](../pi-sequent-hat-setup.md#channel-numbering) |
| **Stack level** | Physical position of relay card on the Pi (0 = closest to Pi, 1, 2, …). | [`pi-sequent-hat-setup.md`](../pi-sequent-hat-setup.md#stack-diagram) |
| **DIP switch** | Three tiny switches (ID0, ID1, ID2) that assign a unique I²C address to each relay card. | [`pi-sequent-hat-setup.md`](../pi-sequent-hat-setup.md#dip-switch-address-table) |
| **API key** | Secret token (e.g., `gdev_123_abc`) that lets your Pi prove it's trusted. Keep it private. | [`pi-integration-guide.md`](../pi-integration-guide.md#api-auth) |
| **Offline queue** | If your Pi loses network, readings are saved locally and sync when reconnected. | [`pi-integration-guide.md`](../pi-integration-guide.md#offline-resilience) |

**UX:** On any wizard step with unfamiliar terms, add a **?** icon next to each term. Click → tooltip + link to glossary or docs.

---

### Part C: Network Diagnostics API

**New endpoint:** `POST /devices/{id}/network-test`

```go
// cmd/api/devices.go
func (s *Server) handleNetworkTest(w http.ResponseWriter, r *http.Request) {
  deviceID := r.PathValue("id")
  device, err := s.db.GetDeviceByID(r.Context(), deviceID)
  if err != nil {
    http.Error(w, "device not found", http.StatusNotFound)
    return
  }

  // Return a "ping back" URL that the Pi can call to prove connectivity
  testToken := generateTestToken(device.ID)
  testURL := fmt.Sprintf("%s/devices/%s/network-test-ping?token=%s", s.baseURL, device.ID, testToken)

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]string{
    "ping_url": testURL,
    "ttl_seconds": "10",
  })
}

func (s *Server) handleNetworkTestPing(w http.ResponseWriter, r *http.Request) {
  // Pi hits this; if it succeeds, the Pi can reach us
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]interface{}{
    "status": "ok",
    "server_time": time.Now().Unix(),
    "ping_latency_ms": /* calculated from token timestamp */,
  })
}
```

**UI flow:**
1. User clicks "Test connectivity" in step 4 of wizard
2. Frontend POSTs to `/devices/{id}/network-test` → gets `ping_url`
3. Display countdown: "Waiting for your Pi to call in… (10s)"
4. Pi's background task (or a manual curl) calls the ping_url
5. Frontend polls `/devices/{id}/network-test-status` → shows result:
   - ✓ **Connected in 120ms** (green, latency displayed)
   - ⚠ **Slow (3.2s)** — may timeout during actual use (yellow)
   - ✗ **No ping received** — Pi never called back (red)
   - ⚠ **Timeout during test** — network is unreachable

---

### Part D: DIP Switch Calculator (Interactive)

**Component:** `DipSwitchCalculator.vue`

```vue
<template>
  <div class="dip-calculator">
    <div class="flex gap-4">
      <!-- Input: Channel or Stack -->
      <label>
        <span>Target channel (0–63)</span>
        <input v-model.number="channel" type="number" min="0" max="63" />
      </label>

      <!-- Calculated output -->
      <div class="dip-state">
        <div class="text-sm font-semibold">Set DIP switches:</div>
        <div class="flex gap-2">
          <DipBit label="ID0" :on="dipState.id0" />
          <DipBit label="ID1" :on="dipState.id1" />
          <DipBit label="ID2" :on="dipState.id2" />
        </div>
        <div class="text-xs text-zinc-500 mt-1">
          Stack {{ stackLevel }}, I²C addr: {{ i2cAddress }}
        </div>
      </div>

      <!-- Visual feedback -->
      <img v-if="showDiagram" :src="dipDiagram" alt="DIP switch physical position" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const channel = ref(0)

const stackLevel = computed(() => Math.floor(channel.value / 8))
const relayInStack = computed(() => (channel.value % 8) + 1)

const dipState = computed(() => {
  const level = stackLevel.value
  return {
    id0: !!(level & 1),
    id1: !!(level & 2),
    id2: !!(level & 4),
  }
})

const i2cAddress = computed(() => {
  const addr = 0x27 - stackLevel.value
  return `0x${addr.toString(16).toUpperCase()}`
})
</script>
```

---

### Part E: Config Download (No Manual Editing)

**Route:** `GET /devices/{id}/config/download`

```go
func (s *Server) handleConfigDownload(w http.ResponseWriter, r *http.Request) {
  deviceID := r.PathValue("id")
  device, err := s.db.GetDeviceByID(r.Context(), deviceID)
  if err != nil {
    http.Error(w, "device not found", http.StatusNotFound)
    return
  }

  // Build minimal bootstrap YAML
  bootstrap := map[string]interface{}{
    "api": map[string]interface{}{
      "base_url": r.Header.Get("X-Origin-URL"), // or ask UI to provide
      "timeout_seconds": 5,
      "api_key": "GR33N_DEVICE_API_KEY", // placeholder; user fills in from UI display
    },
    "device": map[string]interface{}{
      "uid": device.DeviceUID,
    },
    "farm": map[string]interface{}{
      "farm_id": device.FarmID,
    },
    "schedule_poll_interval_seconds": 30,
    "offline_queue_path": "/var/lib/gr33n/queue.db",
  }

  w.Header().Set("Content-Type", "application/yaml")
  w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=config.yaml"))
  
  encoder := yaml.NewEncoder(w)
  encoder.Encode(bootstrap)
}
```

**UI display (step 5):**
```yaml
# Paste this into your Pi at: /home/pi/gr33n-platform/pi_client/config.yaml
# Edit only the api_key (copy from above)

api:
  base_url: "http://192.168.1.100:8080"  # ← replace with your API server
  timeout_seconds: 5
  api_key: "gdev_123_secrettoken"         # ← paste the key shown above
device:
  uid: "demo-veg-relay-01"
farm:
  farm_id: 1
schedule_poll_interval_seconds: 30
offline_queue_path: "/var/lib/gr33n/queue.db"
```

---

### Part F: Validation & Completion Checklist

Add real-time validation banner at the top of wizard:

```
Step 3: Wiring
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Device registered
✓ All 8 channels assigned
⚠ 2 actuators still undefined (create in Controls first)
✗ No network test yet (will run in step 4)

[Can't proceed] until: All actuators assigned & device UID set
```

**Color coding:**
- ✓ **Green** — complete, no action needed
- ⚠ **Yellow** — complete but warnings (e.g., slow network)
- ✗ **Red** — incomplete; blocks next step
- — **Gray** — not yet reached

---

## Workstreams

### WS1: Create PiSetupWizard.vue

**Files:**
- `ui/src/views/PiSetupWizard.vue` — main 6-step component
- `ui/src/components/PiWizard/` — substeps:
  - `Welcome.vue`
  - `RegisterPi.vue`
  - `AssignChannels.vue`
  - `NetworkTest.vue`
  - `ConfigDownload.vue`
  - `Confirm.vue`

**Routes:**
- `/pi-setup` → embed current `PiSetupGuide.vue` (reference)
- `/pi-setup-wizard` → NEW interactive wizard

**Acceptance:**
- [ ] User can step through all 6 stages without leaving the page
- [ ] "Back" and "Next" buttons work correctly
- [ ] Validation prevents premature advance (red checks block)
- [ ] At end, user has downloaded config.yaml (or copied to clipboard)

---

### WS2: Comprehensive Tooltips & Glossary

**Files:**
- `ui/src/components/TermGlossary.vue` — bottom-right panel; searchable
- `ui/src/lib/glossary.js` — term database (text + links)
- Add `HelpTip` callouts in each wizard step

**Acceptance:**
- [ ] Hover/click each unfamiliar term → tooltip appears
- [ ] Tooltips reference actual docs (pi-sequent-hat-setup, pi-integration-guide)
- [ ] Glossary search works (e.g., search "I2C" → returns I²C definition + GPIO + BCM)
- [ ] At least 10 terms covered

---

### WS3: Network Diagnostics API

**Files:**
- `cmd/api/devices.go` — `POST /devices/{id}/network-test` + `GET /devices/{id}/network-test-status`
- `cmd/api/routes.go` — register routes
- `pi_client/gr33n_client.py` — add network test client call (ping_url callback)

**Acceptance:**
- [ ] Endpoint returns a `ping_url` and `ttl_seconds`
- [ ] Pi can call ping_url from background task or curl
- [ ] Frontend polls status endpoint; shows latency + pass/fail
- [ ] Timeout after 10s with user-friendly message

---

### WS4: Config Download

**Files:**
- `cmd/api/devices.go` — `GET /devices/{id}/config/download`
- `ui/src/components/PiWizard/ConfigDownload.vue` — display + download button

**Acceptance:**
- [ ] Downloaded YAML is valid (can parse in Python)
- [ ] API key is shown separately (user must copy it in)
- [ ] File includes comments explaining each field
- [ ] Download works; scp one-liner is optional (nice-to-have)

---

### WS5: DIP Switch Calculator

**Files:**
- `ui/src/components/DipSwitchCalculator.vue`

**Acceptance:**
- [ ] User enters channel 0–63 → shows DIP state (ON/OFF for each switch)
- [ ] Shows stack level + I²C address
- [ ] Visual diagram (can be placeholder image for now)

---

### WS6: Channel Assignment Flow

**Files:**
- Update `ui/src/components/PiWizard/AssignChannels.vue`
- Optionally refactor `ActuatorWiringPanel.vue` to reuse logic

**Acceptance:**
- [ ] User sees 8-card grid for first stack
- [ ] Each card has a dropdown: "Select actuator" → list of all actuators
- [ ] Visual feedback: ✓ OK, ⚠ no actuator assigned yet
- [ ] Option to add more cards (for stacks 1, 2, …)

---

### WS7: Inline Glossary

**Files:**
- `ui/src/lib/glossary.js` — centralized term + definition database
- `ui/src/components/GlossaryTerm.vue` — reusable badge with hover

**Acceptance:**
- [ ] At least 10 terms defined with brief + link
- [ ] Hover on badge → tooltip; click → full entry (or link to docs)
- [ ] Glossary accessible from any wizard step

---

### WS8: Validation Feedback

**Files:**
- `ui/src/stores/piWizardValidation.js` — Pinia store tracking state
- Update each wizard step component to report validation state

**Acceptance:**
- [ ] Banner at top of wizard shows ✓, ⚠, ✗ for each condition
- [ ] "Next" button disabled until all ✗ are resolved
- [ ] Messages are clear ("Add at least one actuator before proceeding")

---

### WS9: Video/Animation (Optional, post-ship)

**Files:**
- `docs/pi-setup-videos/` — animated GIFs or embedded YouTube
- `ui/src/components/PiWizard/Welcome.vue` — embed video placeholder

**Acceptance:**
- [ ] Placeholder for 3–5 short clips:
  - "What is a relay HAT?" (15s)
  - "Setting DIP switches" (30s)
  - "Wiring the power supply" (30s)
  - "First boot on the Pi" (30s)
  - (Can be added in Phase 61 or later)

---

### WS10: Docs & Tests

**Files:**
- Update `docs/pi-sequent-hat-setup.md` — add "Guided wizard" section with screenshots
- Add Vitest tests: `ui/src/components/__tests__/PiSetupWizard.spec.js`
- Add E2E tests: `e2e/pi-setup-wizard.cy.js` (Cypress or Playwright)
- Update `phase-60-closure.test.js` (test file proving all requirements met)

**Acceptance:**
- [ ] All 6 wizard steps have unit tests (mount, input, validation)
- [ ] E2E test walks through full wizard once
- [ ] Screenshots in docs show before/after
- [ ] OC-60 closed with all test evidence

---

## Success Criteria

1. **Completeness:** Non-IT farmer never leaves `/pi-setup-wizard`; never opens a terminal or text editor.
2. **Clarity:** Every field has a tooltip or inline help text; terms are glossary-linked.
3. **Validation:** User cannot accidentally submit incomplete setup; banner warns of blockers.
4. **Confidence:** User sees live network test results before deploying; knows if config is correct.
5. **Speed:** Wizard walk-through takes **≤5 min** for experienced user; **≤20 min** for first-timer.

---

## Phase contract (OC-60)

When Phase 60 ships:
- ✅ User can register a Pi, assign wiring, generate config, and test network without leaving the UI
- ✅ Every unfamiliar term (DIP, I²C, channel, GPIO) has a tooltip + glossary entry
- ✅ Config download produces valid YAML requiring **zero manual editing**
- ✅ Network test gives pass/fail + latency feedback
- ✅ Validation banner prevents incomplete submission
- ✅ Wizard has E2E test coverage (full flow tested)
- ✅ Docs updated with wizard screenshots + guidance
- ✅ Zero dependency on Phases 50–59 (reads their APIs, doesn't modify)

**Estimated effort:** 4–6 weeks, 1 full-time engineer (design + frontend + API endpoints).

---

## Related docs

- [`pi-sequent-hat-setup.md`](../pi-sequent-hat-setup.md) — technical reference (not changed)
- [`pi-integration-guide.md`](../pi-integration-guide.md) — API contracts (not changed)
- [`local-operator-bootstrap.md`](../local-operator-bootstrap.md) — quickstart (add wizard link)
- `phase_50_hardware_wiring_visibility.plan.md` — prerequisite (wiring model)
- `phase_51_pi_config_sync.plan.md` — prerequisite (platform sync)

---

## Notes

- **Isolation:** Phase 60 reads from phases 50–51 but does not modify their logic or routes.
- **Future phases:** Phase 61+ could add video tutorials, mobile UI for Pi setup, or batch device provisioning.
- **Iteration:** Start with steps 1–5 (registration + download), then add step 6 (confirm) + network test in a follow-up if needed.
