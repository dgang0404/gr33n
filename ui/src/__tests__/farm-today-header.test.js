import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import FarmTodayHeader from '../components/FarmTodayHeader.vue'
import {
  buildFarmTodayRollup,
  todayHeaderSubtitle,
  todayTimeGreeting,
} from '../lib/farmTodayHeader.js'

const zones = [
  { id: 1, name: 'Veg Room', zone_type: 'indoor' },
  { id: 2, name: 'Flower Room', zone_type: 'indoor' },
  { id: 3, name: 'Outdoor Garden', zone_type: 'outdoor' },
]

function statusFor(zone) {
  if (zone.id === 2) {
    return { health: 'warn', attention: [{ label: 'Humidity high' }], plants: { state: 'growing' } }
  }
  if (zone.id === 3) {
    return { health: 'unconfigured', plants: { state: 'empty' } }
  }
  return { health: 'ok', plants: { state: 'growing' } }
}

describe('Phase 174 WS2 — farmTodayHeader lib', () => {
  it('rolls up healthy and attention zone counts', () => {
    const rollup = buildFarmTodayRollup({
      zones,
      getStatus: statusFor,
      tasksTodayCount: 5,
      unreadAlerts: 2,
      overdueTaskCount: 1,
    })
    expect(rollup.healthy).toBe(1)
    expect(rollup.attention).toBe(1)
    expect(rollup.tasksTodayCount).toBe(5)
    expect(rollup.unreadAlerts).toBe(2)
    expect(rollup.overdueTaskCount).toBe(1)
  })

  it('returns time-of-day greeting', () => {
    expect(todayTimeGreeting(new Date('2026-07-12T09:00:00'))).toBe('Good morning')
    expect(todayTimeGreeting(new Date('2026-07-12T14:00:00'))).toBe('Good afternoon')
    expect(todayTimeGreeting(new Date('2026-07-12T19:00:00'))).toBe('Good evening')
  })

  it('extends subtitle when site weather has solar times', () => {
    const plain = todayHeaderSubtitle(null)
    expect(plain).toMatch(/Good (morning|afternoon|evening)/)
    const withSun = todayHeaderSubtitle({
      solar: { sunrise_at: '2026-07-12T06:00:00Z', sunset_at: '2026-07-12T20:00:00Z' },
    })
    expect(withSun).toContain('daylight on your farm today')
  })
})

describe('Phase 174 WS2 — FarmTodayHeader', () => {
  it('renders farm name and health pills with links', () => {
    const wrapper = mount(FarmTodayHeader, {
      props: {
        farmName: 'gr33n Demo Farm',
        zones,
        getStatus: statusFor,
        tasksTodayCount: 3,
        unreadAlerts: 2,
        overdueTaskCount: 0,
        tasksLink: '/zones/1?tab=tasks',
        alertsLink: '/zones/1?tab=alerts',
        siteWeather: null,
      },
      global: { stubs: { HelpTip: true, RouterLink: true } },
    })
    expect(wrapper.find('[data-test="farm-today-header"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('gr33n Demo Farm')
    expect(wrapper.find('[data-test="farm-today-pill-healthy"]').text()).toContain('1 healthy')
    expect(wrapper.find('[data-test="farm-today-pill-attention"]').text()).toContain('1 need attention')
    expect(wrapper.find('[data-test="farm-today-pill-tasks"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="farm-today-pill-alerts"]').exists()).toBe(true)
  })

  it('emits filter-attention when attention pill clicked', async () => {
    const wrapper = mount(FarmTodayHeader, {
      props: {
        farmName: 'Demo',
        zones,
        getStatus: statusFor,
        tasksTodayCount: 0,
        unreadAlerts: 0,
        overdueTaskCount: 0,
        tasksLink: '/zones',
        alertsLink: '/zones',
      },
      global: { stubs: { HelpTip: true, RouterLink: true } },
    })
    await wrapper.find('[data-test="farm-today-pill-attention"]').trigger('click')
    expect(wrapper.emitted('filter-attention')).toHaveLength(1)
  })

  it('emits refresh when refresh clicked', async () => {
    const wrapper = mount(FarmTodayHeader, {
      props: {
        farmName: 'Demo',
        zones: [],
        getStatus: () => ({ health: 'ok' }),
        tasksTodayCount: 0,
        unreadAlerts: 0,
        overdueTaskCount: 0,
        tasksLink: '/zones',
        alertsLink: '/zones',
      },
      global: { stubs: { HelpTip: true, RouterLink: true } },
    })
    await wrapper.find('[data-test="farm-today-refresh"]').trigger('click')
    expect(wrapper.emitted('refresh')).toHaveLength(1)
  })
})
