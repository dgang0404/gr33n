# Phase 60 — Implementation Checklist

> **Developer's quick reference** for building Phase 60 Pi Setup Wizard & Onboarding Clarity.

---

## File Structure

```
ui/src/
├── views/
│   └── PiSetupWizard.vue                    ← Main 6-step wizard
├── components/
│   ├── PiWizard/
│   │   ├── Welcome.vue                      ← Step 1
│   │   ├── RegisterPi.vue                   ← Step 2
│   │   ├── AssignChannels.vue               ← Step 3
│   │   ├── NetworkTest.vue                  ← Step 4
│   │   ├── ConfigDownload.vue               ← Step 5
│   │   ├── Confirm.vue                      ← Step 6
│   │   ├── DipSwitchCalculator.vue          ← Reusable in Step 3
│   │   ├── ValidationBanner.vue             ← Real-time status
│   │   └── GlossaryPanel.vue                ← Inline help
│   └── TermBadge.vue                        ← Reusable glossary term
├── lib/
│   ├── phase-60-glossary.js                 ← Glossary data + search
│   ├── phase-60-validation.js               ← Form validation logic
│   └── phase-60-config-generator.js         ← YAML generation
├── stores/
│   └── piWizardStore.js                     ← Pinia state (steps, form data)
└── __tests__/
    └── PiSetupWizard.spec.js                ← Unit + E2E tests

cmd/api/
├── devices.go                               ← New routes:
│                                             - POST /devices/{id}/network-test
│                                             - GET /devices/{id}/network-test-status
│                                             - GET /devices/{id}/config/download
├── routes.go                                ← Register new routes
└── openapi.yaml (update)                    ← Document new endpoints
```

---

## Component Breakdown

### 1. PiSetupWizard.vue (Main Container)

**Responsibilities:**
- Render the 6-step flow (current step, content, buttons)
- Progress bar at top
- Back/Next/Cancel button logic
- Route to sub-components
- Persist state across steps (via Pinia store)

**Key data:**
```vue
const currentStep = ref(1)
const formData = reactive({
  device: { name: '', uid: '', apiKey: '' },
  channels: [...],
  networkConfig: { baseUrl: '', tested: false },
  ...
})
const validation = reactive({
  step1: { complete: true, ... },
  step2: { complete: false, errors: [...] },
  ...
})
```

**Tests:**
- [ ] Can navigate forward with valid data
- [ ] Cannot proceed if validation fails
- [ ] Back button works
- [ ] Cancel confirms before exiting

---

### 2. Welcome.vue (Step 1)

**Renders:**
- Title + intro text
- Checklist of parts (interactive checkboxes)
- Video placeholder (optional)
- Help panel (right side)

**Props / Events:**
- `:onNext` callback (to advance to step 2)

**Tests:**
- [ ] Checklist items render
- [ ] Checkbox state persists
- [ ] Help panel content displays

---

### 3. RegisterPi.vue (Step 2)

**Renders:**
- Form: Device name, Device UID, Zone (optional)
- API key generation button
- API key display (one-time, highlighted warning)
- Copy-to-clipboard button
- Glossary tooltips for "UID" and "API Key"

**Props / Events:**
- `:devices` (list of existing devices)
- `:onNext` (advance with device ID)

**Async logic:**
- Generate API key (backend call)
- Show success toast

**Tests:**
- [ ] Form fields render
- [ ] API key generation button works
- [ ] Copy-to-clipboard button works
- [ ] Validation blocks Next if required fields empty

---

### 4. AssignChannels.vue (Step 3)

**Renders:**
- Title + intro
- Visual grid of channels (8 per stack, expandable to stacks 1+)
- For each channel: dropdown selector (actuators), status badge
- DIP calculator modal (on-click)
- Validation banner
- Help panel (glossary, DIP info)

**Props / Events:**
- `:actuators` (list of available actuators)
- `:devices` (to auto-select current device)
- `:onNext` (advance with channel assignments)

**Sub-components:**
- `DipSwitchCalculator.vue` (modal or inline)
- `ValidationBanner.vue` (real-time status)

**Reactive state:**
```vue
const channelAssignments = reactive({
  0: null,  // channel 0 → actuator ID or null
  1: null,
  2: null,
  ...
})
```

**Tests:**
- [ ] Dropdown selects actuator
- [ ] Status badge updates (✓ assigned, ⚠ empty)
- [ ] DIP calculator opens + computes correctly
- [ ] Validation banner shows blockers

---

### 5. NetworkTest.vue (Step 4)

**Renders:**
- Input field for API base URL (pre-filled with current origin)
- Edit button
- "Test connectivity" button
- Spinner while testing (10s countdown)
- Result (✓ OK + latency, ⚠ slow, ✗ failed, etc.)
- Help panel (what is API, troubleshooting)

**Props / Events:**
- `:deviceId` (which device to test)
- `:onNext` (advance regardless of test result)

**Async logic:**
- POST `/devices/{id}/network-test` → get `ping_url`
- Countdown (10s)
- Poll `/devices/{id}/network-test-status`
- Display result

**Tests:**
- [ ] Form pre-fills with current origin
- [ ] Test button starts countdown
- [ ] Result displays pass/fail
- [ ] Can edit URL and retry
- [ ] Timeout after 10s with friendly message

---

### 6. ConfigDownload.vue (Step 5)

**Renders:**
- Title + intro
- Live preview of config.yaml (syntax-highlighted, read-only)
- "Download config.yaml" button (triggers file download)
- "Copy to clipboard" button
- Instructions (scp one-liner, nano steps, optional)
- Help panel (what is YAML, SSH basics)

**Props / Events:**
- `:deviceId` (fetch config from API)
- `:onNext` (advance)

**Async logic:**
- GET `/devices/{id}/config/download` → YAML file
- Render in code block
- Handle download

**Tests:**
- [ ] Config preview renders
- [ ] Download button produces valid YAML file
- [ ] Copy button works
- [ ] Instructions are clear

---

### 7. Confirm.vue (Step 6)

**Renders:**
- Title + celebration emoji
- Deployment checklist (4 items, checkboxes)
- Optional "Test actuator" button (pulse Main Pump for 1s)
- Help panel (next steps, links to other docs)
- "Finish" button → exit wizard, go to Hardware board

**Props / Events:**
- `:deviceId`
- `:onFinish` (callback to redirect)

**Async logic:**
- Optional: POST test pulse command
- Display result

**Tests:**
- [ ] Checklist items render + toggle
- [ ] Test button works (safe pulse)
- [ ] Finish button exits wizard

---

### 8. DipSwitchCalculator.vue (Reusable)

**Renders:**
- Input: target channel (0–63)
- Output: DIP state (ON/OFF for ID0, ID1, ID2)
- Stack level + I²C address
- Optional visual diagram (placeholder image)

**Logic:**
```js
const stackLevel = computed(() => Math.floor(channel / 8))
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
```

**Tests:**
- [ ] Computes DIP state correctly for channels 0–63
- [ ] I²C address matches spec
- [ ] Visual updates in real-time

---

### 9. ValidationBanner.vue (Reusable)

**Renders:**
- Title: "STEP N VALIDATION"
- List of checks (✓, ⚠, ✗, —)
- Blocker message (red) if validation fails

**Props:**
```vue
const validationState = {
  checks: [
    { icon: '✓', message: 'Device registered', level: 'ok' },
    { icon: '⚠', message: 'Network test not run', level: 'warning' },
    { icon: '✗', message: 'Actuators must be assigned', level: 'error', blocker: true },
  ],
}
```

**Tests:**
- [ ] Renders all checks
- [ ] Color-codes correctly
- [ ] Blocker state communicated to parent

---

### 10. GlossaryPanel.vue (Reusable)

**Renders:**
- Search input
- List of terms (filtered)
- Each term: title + short + long + links

**Props:**
- `:context` (optional, to filter terms by step)
- `:compact` (toggle panel size)

**Logic:**
- Search on keystroke
- Import from `phase-60-glossary.js`

**Tests:**
- [ ] Search filters terms
- [ ] Click term → expand/collapse
- [ ] Links work

---

### 11. TermBadge.vue (Reusable)

**Renders:**
- Inline term badge: "I²C" with `?` icon
- Hover → tooltip (short definition)
- Click → open full glossary entry

**Props:**
```vue
const term = ref('i2c')  // ID from glossary
```

**Usage in other steps:**
```vue
<p>
  Each <TermBadge term="relay" /> controls one load via
  <TermBadge term="channel" />.
</p>
```

**Tests:**
- [ ] Renders badge
- [ ] Tooltip appears on hover
- [ ] Link to glossary works

---

## Pinia Store: piWizardStore.js

```js
import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'

export const usePiWizardStore = defineStore('piWizard', () => {
  const currentStep = ref(1)
  const formData = reactive({
    // Step 2
    device: {
      name: '',
      uid: '',
      apiKey: '', // generated, not user input
    },
    // Step 3
    channelAssignments: {}, // channel → actuator_id
    // Step 4
    network: {
      apiBaseUrl: '',
      testResult: null, // { success, latency, error }
    },
    // Step 5+6
    configYaml: '', // generated
  })

  const validation = reactive({
    step1: { complete: true },
    step2: { complete: false, errors: [] },
    step3: { complete: false, errors: [] },
    step4: { complete: false, errors: [] },
    step5: { complete: false, errors: [] },
    step6: { complete: false, errors: [] },
  })

  function canAdvance(step) {
    return !validation[`step${step}`].errors.length
  }

  function setStep(n) {
    currentStep.value = n
  }

  function updateFormData(path, value) {
    // deep update nested object
  }

  function resetWizard() {
    currentStep.value = 1
    formData.device = { name: '', uid: '', apiKey: '' }
    // ... reset all
  }

  return {
    currentStep,
    formData,
    validation,
    canAdvance,
    setStep,
    updateFormData,
    resetWizard,
  }
})
```

---

## Utility Functions

### phase-60-validation.js

```js
/**
 * Validation rules for each step
 */

export function validateStep2(data) {
  const errors = []
  if (!data.device.name?.trim()) errors.push('Device name required')
  if (!data.device.uid?.trim()) errors.push('Device UID required')
  if (!data.device.apiKey) errors.push('API key not generated')
  return errors
}

export function validateStep3(data) {
  const errors = []
  if (!Object.keys(data.channelAssignments).length) {
    errors.push('At least one actuator must be assigned')
  }
  return errors
}

// ... etc
```

### phase-60-config-generator.js

```js
/**
 * Generate config.yaml from form data
 */

export function generateConfigYaml(formData) {
  const config = {
    api: {
      base_url: formData.network.apiBaseUrl,
      timeout_seconds: 5,
      api_key: formData.device.apiKey,
    },
    device: {
      uid: formData.device.uid,
    },
    farm: {
      farm_id: formData.device.farmId,
    },
    schedule_poll_interval_seconds: 30,
    offline_queue_path: '/var/lib/gr33n/queue.db',
  }

  // Convert to YAML string
  return toYaml(config)
}

export function downloadYaml(yaml, filename = 'config.yaml') {
  const blob = new Blob([yaml], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function copyToClipboard(text) {
  return navigator.clipboard.writeText(text)
}
```

---

## API Endpoints to Create

### cmd/api/devices.go

```go
// POST /devices/{id}/network-test
// Start a network test; return ping URL
func (s *Server) handleNetworkTest(w http.ResponseWriter, r *http.Request) {
  deviceID := r.PathValue("id")
  
  // Generate test token
  testToken := generateTestToken(deviceID)
  testURL := fmt.Sprintf(
    "%s/devices/%s/network-test-ping?token=%s",
    s.baseURL, deviceID, testToken,
  )
  
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]interface{}{
    "ping_url": testURL,
    "ttl_seconds": 10,
  })
}

// GET /devices/{id}/network-test-status
// Poll status of ongoing test
func (s *Server) handleNetworkTestStatus(w http.ResponseWriter, r *http.Request) {
  deviceID := r.PathValue("id")
  
  // Check if token was pinged back
  status := checkTestStatus(deviceID)
  
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(status)
  // { status: "ok", latency_ms: 120 } or
  // { status: "waiting" } or
  // { status: "timeout" }
}

// GET /devices/{id}/config/download
// Generate config.yaml
func (s *Server) handleConfigDownload(w http.ResponseWriter, r *http.Request) {
  deviceID := r.PathValue("id")
  
  device, _ := s.db.GetDeviceByID(r.Context(), deviceID)
  
  yaml := buildBootstrapYaml(device, r.Host)
  
  w.Header().Set("Content-Type", "application/yaml")
  w.Header().Set("Content-Disposition", 
    fmt.Sprintf("attachment; filename=config-%s.yaml", device.DeviceUID))
  w.Write([]byte(yaml))
}
```

---

## Tests to Write

### ui/src/components/__tests__/PiSetupWizard.spec.js

```js
import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import PiSetupWizard from '../PiSetupWizard.vue'

describe('PiSetupWizard', () => {
  let wrapper

  beforeEach(() => {
    wrapper = mount(PiSetupWizard)
  })

  it('renders step 1', () => {
    expect(wrapper.text()).toContain('Welcome to the Pi Setup Wizard')
  })

  it('shows next button on step 1', () => {
    const nextBtn = wrapper.find('[data-test="next-button"]')
    expect(nextBtn.exists()).toBe(true)
  })

  it('advances to step 2 when next clicked', async () => {
    const nextBtn = wrapper.find('[data-test="next-button"]')
    await nextBtn.trigger('click')
    expect(wrapper.vm.currentStep).toBe(2)
  })

  // ... more tests for each step, validation, etc.
})
```

### E2E Test (Cypress)

```js
// e2e/pi-setup-wizard.cy.js

describe('Pi Setup Wizard E2E', () => {
  it('completes full wizard flow', () => {
    cy.visit('/pi-setup-wizard')
    
    // Step 1: Welcome
    cy.contains('Welcome to the Pi Setup Wizard').should('be.visible')
    cy.get('[data-test="next-button"]').click()
    
    // Step 2: Register
    cy.get('input[name="deviceName"]').type('Demo Pi')
    cy.get('input[name="deviceUid"]').type('demo-01')
    cy.get('[data-test="generate-api-key"]').click()
    cy.get('[data-test="api-key-display"]').should('contain', 'gdev_')
    cy.get('[data-test="next-button"]').click()
    
    // Step 3: Channels
    cy.get('[data-test="channel-0-select"]').select('Main Pump')
    cy.get('[data-test="dip-calculator"]').click()
    cy.get('input[name="targetChannel"]').type('0')
    // ... verify DIP state
    cy.get('[data-test="next-button"]').click()
    
    // ... more steps
    
    // Step 6: Confirm
    cy.get('[data-test="finish-button"]').click()
    cy.url().should('include', '/hardware')
  })
})
```

---

## Routes to Add

### ui/src/router/index.js

```js
import PiSetupWizard from '../views/PiSetupWizard.vue'

export const routes = [
  // ...existing...
  {
    path: '/pi-setup-wizard',
    component: PiSetupWizard,
    name: 'pi-setup-wizard',
    meta: { title: 'Pi Setup Wizard' },
  },
]
```

---

## OpenAPI Spec Updates

Add to `openapi.yaml`:

```yaml
/devices/{id}/network-test:
  post:
    summary: Start network connectivity test
    tags: [Devices]
    parameters:
      - in: path
        name: id
        required: true
        schema: { type: integer }
    responses:
      200:
        description: Test started
        content:
          application/json:
            schema:
              type: object
              properties:
                ping_url: { type: string }
                ttl_seconds: { type: integer }

/devices/{id}/network-test-status:
  get:
    summary: Check network test status
    tags: [Devices]
    parameters:
      - in: path
        name: id
        required: true
        schema: { type: integer }
    responses:
      200:
        description: Test status
        content:
          application/json:
            schema:
              type: object
              properties:
                status: { enum: [ok, waiting, timeout] }
                latency_ms: { type: integer, nullable: true }

/devices/{id}/config/download:
  get:
    summary: Download config.yaml
    tags: [Devices]
    parameters:
      - in: path
        name: id
        required: true
        schema: { type: integer }
    responses:
      200:
        description: YAML file
        content:
          application/yaml:
            schema: { type: string }
```

---

## Summary: Lines of Code Estimate

| Component / File | Est. LOC | Priority |
|------------------|----------|----------|
| PiSetupWizard.vue | 300 | P0 |
| Step 1–6 (6 files) | 200 ea × 6 = 1200 | P0 |
| DipSwitchCalculator.vue | 150 | P0 |
| GlossaryPanel.vue | 200 | P1 |
| piWizardStore.js | 150 | P0 |
| Validation + Config libs | 200 | P0 |
| API endpoints (Go) | 300 | P0 |
| Tests (Vitest + Cypress) | 800 | P0 |
| Docs updates | 200 | P1 |
| **Total** | **~4300** | — |

**Estimate:** 4–6 weeks, 1 FTE (design + frontend + backend).

---

## Next Steps

1. [ ] Create component skeleton files
2. [ ] Build Step 1–2 (entry, registration)
3. [ ] Implement DIP calculator (Step 3 foundation)
4. [ ] Build Step 3 (channel assignment)
5. [ ] Add network test endpoints (Step 4)
6. [ ] Build Step 5–6 (config download + confirm)
7. [ ] Add comprehensive tests
8. [ ] Documentation & screenshots

---

## Acceptance Criteria Checklist

- [ ] All 6 steps render without errors
- [ ] Navigation (back/next) works
- [ ] Validation prevents incomplete submission
- [ ] Tooltips appear on hover (10+ terms covered)
- [ ] Glossary search works (5+ term results)
- [ ] DIP calculator computes correctly (all 8 stacks)
- [ ] Config download produces valid YAML
- [ ] Network test endpoint responds
- [ ] Mobile layout is readable
- [ ] E2E test covers full wizard flow
- [ ] Zero console errors
- [ ] Accessibility: keyboard nav + aria labels
- [ ] Docs updated with wizard screenshots
- [ ] OC-60 closed with test evidence
