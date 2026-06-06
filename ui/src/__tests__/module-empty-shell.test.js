import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ModuleEmptyShell from '../components/ModuleEmptyShell.vue'
import {
  MODULE_EMPTY_SHELLS,
  moduleEmptyShellConfig,
  moduleShellZoneHint,
} from '../lib/moduleEmptyShell.js'

describe('Phase 45 WS5 — module empty shells', () => {
  it('exports animals and aquaponics shell copy with workflow doc refs', () => {
    expect(MODULE_EMPTY_SHELLS.animals.workflowDoc).toBe('docs/workflow-guide.md')
    expect(MODULE_EMPTY_SHELLS.animals.templateKey).toBe('chicken_coop_v1')
    expect(MODULE_EMPTY_SHELLS.aquaponics.workflowSection).toContain('Aquaponics')
    expect(MODULE_EMPTY_SHELLS.aquaponics.templateKey).toBe('small_aquaponics_v1')
  })

  it('suggests zones when aquaponics has fewer than two zones', () => {
    expect(moduleShellZoneHint('aquaponics', 0)?.actionTo).toBe('/zones')
    expect(moduleShellZoneHint('aquaponics', 1)?.message).toMatch(/two zones/i)
    expect(moduleShellZoneHint('aquaponics', 2)).toBeNull()
  })

  it('suggests zones for animals when farm has none', () => {
    expect(moduleShellZoneHint('animals', 0)?.actionTo).toBe('/zones')
    expect(moduleShellZoneHint('animals', 1)).toBeNull()
  })

  it('renders animals shell with doc links and primary action', async () => {
    const shell = moduleEmptyShellConfig('animals')
    const wrapper = mount(ModuleEmptyShell, {
      props: { moduleId: 'animals', zoneCount: 0 },
      global: {
        stubs: {
          RouterLink: { props: ['to'], template: '<a><slot /></a>' },
        },
      },
    })
    expect(wrapper.attributes('data-test')).toBe('module-empty-shell-animals')
    expect(wrapper.text()).toContain(shell.title)
    expect(wrapper.text()).toContain('workflow-guide.md')
    expect(wrapper.text()).toContain('chicken_coop_v1')
    expect(wrapper.find('[data-test="module-empty-shell-primary-animals"]').exists()).toBe(true)
    await wrapper.find('[data-test="module-empty-shell-primary-animals"]').trigger('click')
    expect(wrapper.emitted('primary')).toHaveLength(1)
  })

  it('renders aquaponics zone hint when only one zone exists', () => {
    const wrapper = mount(ModuleEmptyShell, {
      props: { moduleId: 'aquaponics', zoneCount: 1 },
      global: {
        stubs: {
          RouterLink: { props: ['to'], template: '<a><slot /></a>' },
        },
      },
    })
    expect(wrapper.find('[data-test="module-empty-shell-zone-hint-aquaponics"]').exists()).toBe(true)
    expect(wrapper.text()).toMatch(/two zones/i)
  })
})
