import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import { buildNavGroups, mobileBottomNav } from '../lib/navGroups.js'
import { buildSunsetWorkspaceRedirects, redirectSunsetWorkspace } from '../lib/workspaces.js'
import router from '../router/index.js'
import ZoneAlertsPanel from '../components/ZoneAlertsPanel.vue'
import ZoneHardwarePanel from '../components/ZoneHardwarePanel.vue'
import { useFarmStore } from '../stores/farm'

describe('Phase 78 — zone-first hardware & feed consolidation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('removes Feed & water and Hardware from sidebar grow group', () => {
    const grow = buildNavGroups().find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.to === '/feed-water')).toBe(false)
    expect(grow.items.some((i) => i.to === '/hardware')).toBe(false)
    expect(grow.items.some((i) => i.to === '/zones')).toBe(true)
    expect(grow.items.some((i) => i.to === '/virtual-pi')).toBe(true)
  })

  it('mobile nav uses Money instead of Feed workspace', () => {
    expect(mobileBottomNav.find((i) => i.to === '/feed-water')).toBeUndefined()
    expect(mobileBottomNav.find((i) => i.to === '/money')?.label).toBe('Money')
  })

  it('zone-scoped feed-water visits redirect to zone water tab', () => {
    const result = redirectSunsetWorkspace({
      path: '/feed-water',
      query: { zone_id: '4', tab: 'daily' },
    })
    expect(result.path).toBe('/zones/4')
    expect(result.query.tab).toBe('water')
    expect(result.query.zone_id).toBeUndefined()
  })

  it('registers hardware and feed-water workspace routes', () => {
    expect(router.resolve('/hardware').name).toBe('hardware')
    expect(router.resolve('/feed-water').name).toBe('feed-water')
    expect(buildSunsetWorkspaceRedirects()).toEqual([])
  })

  it('ZoneAlertsPanel shows GPIO line on sensor alerts', () => {
    const wrapper = mount(ZoneAlertsPanel, {
      props: {
        zoneId: 3,
        zoneName: 'Flower Room',
        sensors: [{
          id: 10,
          name: 'RH sensor',
          sensor_type: 'humidity',
          wiring: { source: 'dht22', gpio_pin: 4 },
        }],
        actuators: [],
        alerts: [{
          id: 5,
          is_read: false,
          is_acknowledged: false,
          severity: 'high',
          subject_rendered: 'Humidity high',
          message_text_rendered: '72% RH',
          triggering_event_source_type: 'sensor',
          triggering_event_source_id: 10,
          created_at: new Date().toISOString(),
        }],
      },
      global: { stubs: { RouterLink: true, AskGuardianButton: true } },
    })

    expect(wrapper.find('[data-test="alert-hardware-line"]').text()).toContain('BCM GPIO 4')
  })

  it('ZoneHardwarePanel lists all zone sensors regardless of need tab', () => {
    const wrapper = mount(ZoneHardwarePanel, {
      props: {
        zoneId: 2,
        sensors: [
          { id: 1, name: 'Moisture', sensor_type: 'soil_moisture' },
          { id: 2, name: 'Lux', sensor_type: 'lux' },
        ],
        actuators: [{ id: 9, name: 'Pump', actuator_type: 'pump', current_state_text: 'offline' }],
      },
      global: {
        stubs: {
          RouterLink: true,
          SensorTile: true,
          HardwareWiringBadge: true,
          HardwareWiringPanel: true,
          ActuatorWiringPanel: true,
          ActuatorPulseControl: true,
          EmptyStateHint: true,
        },
      },
    })

    expect(wrapper.find('[data-test="zone-hardware-sensor-1"]')).toBeTruthy()
    expect(wrapper.find('[data-test="zone-hardware-sensor-2"]')).toBeTruthy()
    expect(wrapper.find('[data-test="zone-hardware-actuator-9"]')).toBeTruthy()
  })

  it('ZoneDetail overview embeds hardware panel (not on Ops or other tabs)', () => {
    const src = readFileSync(join(process.cwd(), 'src/views/ZoneDetail.vue'), 'utf8')
    expect(src).toContain('ZoneHardwarePanel')
    expect(src).toContain('id="zone-hardware"')
    expect(src).not.toContain('jump to GPIO')
  })

  it('ZoneAlertsPanel acknowledges with actuators prop wired', async () => {
    const store = useFarmStore()
    store.markAlertAcknowledged = vi.fn().mockResolvedValue({ id: 5, is_acknowledged: true })

    const wrapper = mount(ZoneAlertsPanel, {
      props: {
        zoneId: 3,
        zoneName: 'Flower Room',
        sensors: [{ id: 10, sensor_type: 'humidity' }],
        actuators: [],
        alerts: [{
          id: 5,
          is_read: false,
          is_acknowledged: false,
          severity: 'high',
          subject_rendered: 'Humidity high',
          message_text_rendered: '72% RH',
          triggering_event_source_type: 'sensor',
          triggering_event_source_id: 10,
          created_at: new Date().toISOString(),
        }],
      },
      global: { stubs: { RouterLink: true, AskGuardianButton: true } },
    })

    await wrapper.find('[data-test="zone-alert-ack"]').trigger('click')
    await flushPromises()
    expect(store.markAlertAcknowledged).toHaveBeenCalledWith(5)
  })
})
