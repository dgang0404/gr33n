/**
 * Phase 167 — Mobile stack + quick actions closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 167 WS1 — responsive stack', () => {
  it('ships FarmZoneStack with mobile-only stack layout', () => {
    const stack = readFileSync(join(repoRoot, 'ui/src/components/FarmZoneStack.vue'), 'utf8')
    expect(stack).toContain('md:hidden')
    expect(stack).toContain('sortZonesForStack')
  })
})

describe('Phase 167 WS2–WS5 — quick actions', () => {
  it('ships ZoneQuickActions sheet with water, light, triage, Guardian', () => {
    const sheet = readFileSync(join(repoRoot, 'ui/src/components/ZoneQuickActions.vue'), 'utf8')
    expect(sheet).toContain('runFertigationProgramNow')
    expect(sheet).toContain('markAlertAcknowledged')
    expect(sheet).toContain('updateTaskStatus')
    expect(sheet).toContain('buildZoneQuickStarters')
    expect(sheet).toContain('useDialogFocusTrap')
    expect(sheet).toContain('min-h-[44px]')
  })

  it('ships zoneQuickActions lib and actuator composable', () => {
    const lib = readFileSync(join(repoRoot, 'ui/src/lib/zoneQuickActions.js'), 'utf8')
    expect(lib).toContain('resolveWaterNowAction')
    expect(lib).toContain('sortZonesForStack')

    const composable = readFileSync(join(repoRoot, 'ui/src/composables/useActuatorCommands.js'), 'utf8')
    expect(composable).toContain('enqueueActuatorCommand')

    const starters = readFileSync(join(repoRoot, 'ui/src/lib/guardianStarters.js'), 'utf8')
    expect(starters).toContain('buildZoneQuickStarters')
  })

  it('FarmCanvas emits select-zone instead of hard navigation', () => {
    const canvas = readFileSync(join(repoRoot, 'ui/src/components/FarmCanvas.vue'), 'utf8')
    expect(canvas).toContain("emit('select-zone'")
    expect(canvas).not.toContain('router.push')
  })
})

describe('Phase 167 WS6 — tests', () => {
  it('ships phase 167 test bundle', () => {
    expect(readFileSync(join(repoRoot, 'ui/src/__tests__/zone-quick-actions.test.js'), 'utf8')).toContain('Herb Room Gravity Drip')
    expect(readFileSync(join(repoRoot, 'ui/src/__tests__/farm-zone-stack.test.js'), 'utf8')).toContain('FarmZoneStack')
  })
})
