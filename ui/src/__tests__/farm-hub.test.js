import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import FarmMorningStrip from '../components/FarmMorningStrip.vue'
import { parseZoneIdQuery, filterAlertsForZone, programAppliesToZone } from '../lib/zoneContext.js'

describe('Phase 41 — farm hub coherence', () => {
  it('parses zone_id query param', () => {
    expect(parseZoneIdQuery('3')).toBe(3)
    expect(parseZoneIdQuery(['7'])).toBe(7)
    expect(parseZoneIdQuery(undefined)).toBeNull()
    expect(parseZoneIdQuery('abc')).toBeNull()
  })

  it('filters alerts to zone sensors', () => {
    const sensors = [{ id: 10, zone_id: 3, sensor_type: 'humidity' }]
    const alerts = [
      {
        id: 1,
        triggering_event_source_type: 'sensor',
        triggering_event_source_id: 10,
        subject_rendered: 'High humidity',
      },
      { id: 2, subject_rendered: 'Farm stock low' },
    ]
    expect(filterAlertsForZone(alerts, 3, 'Flower Room', sensors)).toHaveLength(1)
  })

  it('matches programs by target zone', () => {
    expect(programAppliesToZone({ id: 1, target_zone_id: 3 }, 3)).toBe(true)
    expect(programAppliesToZone({ id: 1, target_zone_id: 99 }, 3)).toBe(false)
  })

  it('renders fertigation zone banner with back to Water', () => {
    const wrapper = mount(ZoneContextBanner, {
      props: {
        zoneId: 3,
        zoneName: 'Flower Room',
        pageLabel: 'Fertigation',
        variant: 'fertigation',
        backToZoneTab: 'water',
        clearRoute: { name: 'fertigation', query: { tab: 'events' } },
      },
      global: {
        stubs: {
          RouterLink: { props: ['to'], template: '<a :href="JSON.stringify(to)"><slot /></a>' },
        },
      },
    })
    expect(wrapper.text()).toContain('Flower Room')
    expect(wrapper.text()).toContain('Back to zone Water')
    expect(wrapper.attributes('data-test')).toBe('zone-context-banner')
  })

  it('renders clickable morning chips', () => {
    const wrapper = mount(FarmMorningStrip, {
      props: {
        chips: [{
          id: 'tasks-due',
          icon: '✅',
          label: 'Tasks due today',
          value: '2',
          to: { path: '/tasks' },
        }],
      },
      global: {
        stubs: {
          RouterLink: { props: ['to'], template: '<a class="chip-link"><slot /></a>' },
        },
      },
    })
    expect(wrapper.find('[data-test="farm-morning-chip-tasks-due"]').exists()).toBe(true)
  })
})
