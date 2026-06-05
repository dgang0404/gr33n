import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import ZoneAlertsPanel from '../components/ZoneAlertsPanel.vue'
import { useFarmStore } from '../stores/farm'

describe('Phase 40 WS4 — ZoneAlertsPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows zone-matched alerts and acknowledges inline', async () => {
    const store = useFarmStore()
    store.markAlertAcknowledged = vi.fn().mockResolvedValue({
      id: 5,
      is_acknowledged: true,
      is_read: true,
    })

    const wrapper = mount(ZoneAlertsPanel, {
      props: {
        zoneId: 3,
        zoneName: 'Flower Room',
        sensors: [{ id: 10, sensor_type: 'humidity' }],
        alerts: [
          {
            id: 5,
            is_read: false,
            is_acknowledged: false,
            severity: 'high',
            subject_rendered: 'Humidity high — Flower Room',
            message_text_rendered: 'Zone: Flower Room. 72% RH',
            triggering_event_source_type: 'sensor',
            triggering_event_source_id: 10,
            created_at: new Date().toISOString(),
          },
          {
            id: 6,
            is_read: false,
            is_acknowledged: false,
            severity: 'low',
            subject_rendered: 'Farm stock',
            message_text_rendered: 'OHN batch low',
            triggering_event_source_type: 'input_batch',
            triggering_event_source_id: 1,
            created_at: new Date().toISOString(),
          },
        ],
      },
      global: {
        stubs: {
          RouterLink: true,
          AskGuardianButton: true,
        },
      },
    })

    expect(wrapper.findAll('[data-test^="zone-alert-row-"]')).toHaveLength(1)
    expect(wrapper.find('[data-test="zone-alerts-unread-badge"]').text()).toContain('1 open')

    await wrapper.find('[data-test="zone-alert-ack"]').trigger('click')
    await flushPromises()

    expect(store.markAlertAcknowledged).toHaveBeenCalledWith(5)
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })
})
