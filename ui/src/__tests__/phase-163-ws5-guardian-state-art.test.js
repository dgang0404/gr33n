import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import GuardianStateArt from '../components/GuardianStateArt.vue'
import {
  guardianStateArtUrl,
  guardianStateArtAlt,
  guardianStateHasArt,
  normalizeManifest,
  resetGuardianStateArtManifestCache,
  GUARDIAN_STATE_ART_STATES,
} from '../lib/guardianStateArt'

describe('Phase 163 WS5 — guardianStateArt helpers', () => {
  it('lists all awakening states', () => {
    expect(GUARDIAN_STATE_ART_STATES).toEqual([
      'sleeping',
      'dormant',
      'stirring',
      'ready',
      'busy',
      'unavailable',
    ])
  })

  it('builds URLs only for manifest-registered files', () => {
    const manifest = normalizeManifest({
      version: 1,
      files: { dormant: 'dormant.webp', stirring: 'stirring.png' },
    })
    expect(guardianStateArtUrl('dormant', manifest)).toBe('/assets/guardian/druid/dormant.webp')
    expect(guardianStateArtUrl('ready', manifest)).toBeNull()
    expect(guardianStateHasArt('stirring', manifest)).toBe(true)
  })

  it('rejects unsafe manifest paths', () => {
    const manifest = normalizeManifest({
      version: 1,
      files: { dormant: '../secret.png', ready: 'ok.webp' },
    })
    expect(guardianStateArtUrl('dormant', manifest)).toBeNull()
    expect(guardianStateArtUrl('ready', manifest)).toBe('/assets/guardian/druid/ok.webp')
  })

  it('provides accessible alt text per state', () => {
    expect(guardianStateArtAlt('dormant')).toContain('resting')
    expect(guardianStateArtAlt('stirring')).toContain('awakening')
  })
})

describe('Phase 163 WS5 — GuardianStateArt', () => {
  beforeEach(() => {
    resetGuardianStateArtManifestCache()
    vi.stubGlobal(
      'fetch',
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ version: 1, files: {} }),
        }),
      ),
    )
  })

  it('stays hidden when manifest has no art for the state', async () => {
    const wrapper = mount(GuardianStateArt, { props: { state: 'dormant' } })
    await flushPromises()
    expect(wrapper.find('[data-test="guardian-state-art"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows art when manifest registers a file and image loads', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ version: 1, files: { dormant: 'dormant.webp' } }),
        }),
      ),
    )
    const wrapper = mount(GuardianStateArt, { props: { state: 'dormant', size: 'md' } })
    await flushPromises()
    const img = wrapper.find('img')
    expect(img.exists()).toBe(true)
    expect(img.attributes('src')).toBe('/assets/guardian/druid/dormant.webp')
    expect(wrapper.get('[data-test="guardian-state-art"]').classes()).toContain('hidden')
    await img.trigger('load')
    await nextTick()
    expect(wrapper.get('[data-test="guardian-state-art"]').classes()).not.toContain('hidden')
    wrapper.unmount()
  })

  it('hides art when image fails to load', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ version: 1, files: { stirring: 'missing.webp' } }),
        }),
      ),
    )
    const wrapper = mount(GuardianStateArt, { props: { state: 'stirring' } })
    await flushPromises()
    const img = wrapper.find('img')
    await img.trigger('error')
    await nextTick()
    expect(wrapper.get('[data-test="guardian-state-art"]').classes()).toContain('hidden')
    wrapper.unmount()
  })
})
