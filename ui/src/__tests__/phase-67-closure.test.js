/**
 * Phase 67 WS7 / OC-67 — hands-free field assistant closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { loadGuardianFieldPrefs, saveGuardianFieldPrefs } from '../lib/guardianFieldPrefs.js'
import { speechRecognitionSupported, speechSynthesisSupported } from '../lib/guardianFieldVoice.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 67 WS7 / OC-67 — field assistant closure', () => {
  it('documents field assistant and plan shipped', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_67_guardian_field_assistant.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
    expect(tour).toContain('Field assistant')
    expect(arch).toContain('Phase 67')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/vision_field.go'))).toBe(true)
  })

  it('field prefs persist read-aloud and STT provider', () => {
    saveGuardianFieldPrefs({ readAloud: true, sttProvider: 'browser' })
    const prefs = loadGuardianFieldPrefs()
    expect(prefs.readAloud).toBe(true)
    expect(prefs.sttProvider).toBe('browser')
    saveGuardianFieldPrefs({ readAloud: false })
  })

  it('Guardian panel settings and STT route ship field assistant controls', () => {
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(settings).toContain('settings-field-read-aloud')
    expect(routes).toContain('POST /v1/chat/stt')
  })

  it('zone detail links photos to Guardian', () => {
    const zone = readFileSync(join(process.cwd(), 'src/views/ZoneDetail.vue'), 'utf8')
    expect(zone).toContain('zone-photo-ask-guardian')
    expect(zone).toContain('askGuardianAboutPhoto')
  })

  it('voice helpers expose browser capability probes', () => {
    expect(typeof speechRecognitionSupported()).toBe('boolean')
    expect(typeof speechSynthesisSupported()).toBe('boolean')
  })
})
