import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { emptyHintConfig, EMPTY_HINT_REASONS } from '../lib/emptyStateHint.js'
import EmptyStateHint from '../components/EmptyStateHint.vue'

describe('Phase 41 WS4 — EmptyStateHint', () => {
  it('exports reason enum values', () => {
    expect(EMPTY_HINT_REASONS.no_setpoint).toBe('no_setpoint')
    expect(EMPTY_HINT_REASONS.no_telemetry).toBe('no_telemetry')
  })

  it('no_setpoint defaults to comfort-targets route', () => {
    const cfg = emptyHintConfig('no_setpoint')
    expect(cfg.actionLabel).toBe('Comfort targets')
    expect(cfg.actionTo).toBe('/comfort-targets')
  })

  it('builds config with overrides', () => {
    const cfg = emptyHintConfig('no_setpoint', {
      message: 'No humidity band yet.',
      actionTo: '/setpoints',
    })
    expect(cfg.message).toBe('No humidity band yet.')
    expect(cfg.actionTo).toBe('/setpoints')
  })

  it('renders message and action link', () => {
    const wrapper = mount(EmptyStateHint, {
      props: {
        reason: 'no_data',
        actionLabel: 'Open Tasks',
        actionTo: '/tasks',
      },
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    expect(wrapper.text()).toContain('Nothing recorded yet')
    expect(wrapper.attributes('data-test')).toBe('empty-hint-no_data')
  })

  it('emits action when no route', async () => {
    const wrapper = mount(EmptyStateHint, {
      props: {
        reason: 'no_data',
        message: 'Custom empty',
        actionLabel: 'Add',
        actionTo: null,
      },
    })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('action')).toHaveLength(1)
  })
})
