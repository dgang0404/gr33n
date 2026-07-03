import { describe, it, expect } from 'vitest'
import { buildNavGroups } from '../lib/navGroups.js'
import { MODULE_SCHEMA } from '../lib/farmModules.js'

describe('Phase 117 — workspace nav module gating', () => {
  it('hides Animals and Aquaponics when modules are disabled', () => {
    const groups = buildNavGroups({
      modules: {
        [MODULE_SCHEMA.animals]: false,
        [MODULE_SCHEMA.aquaponics]: false,
      },
    })
    const more = groups.find((g) => g.label === 'More')
    const routes = more.items.map((i) => i.to)
    expect(routes).not.toContain('/animals')
    expect(routes).not.toContain('/aquaponics')
    expect(routes).toContain('/settings')
  })

  it('shows optional modules when enabled in farm map', () => {
    const groups = buildNavGroups({
      modules: {
        [MODULE_SCHEMA.animals]: true,
        [MODULE_SCHEMA.aquaponics]: true,
      },
    })
    const more = groups.find((g) => g.label === 'More')
    const routes = more.items.map((i) => i.to)
    expect(routes).toContain('/animals')
    expect(routes).toContain('/aquaponics')
  })
})
