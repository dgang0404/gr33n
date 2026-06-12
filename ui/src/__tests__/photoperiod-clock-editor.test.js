import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PhotoperiodClockEditor from '../components/PhotoperiodClockEditor.vue'

describe('PhotoperiodClockEditor — Phase 35 WS8', () => {
  it('derives OFF time from start + duration (18/6 at 06:00 → midnight)', () => {
    const wrapper = mount(PhotoperiodClockEditor, {
      props: {
        modelLightsOnAt: '06:00',
        modelOnHours: 18,
        timezone: 'America/New_York',
      },
    })

    const offInput = wrapper.findAll('input[type="time"]')[1]
    expect(offInput.element.value).toBe('00:00')
    expect(wrapper.text()).toContain('18h ON')
    expect(wrapper.text()).toContain('6h OFF')
  })

  it('updates OFF time when duration changes to 12h', async () => {
    const wrapper = mount(PhotoperiodClockEditor, {
      props: {
        modelLightsOnAt: '06:00',
        modelOnHours: 18,
      },
    })

    const durationInput = wrapper.find('input[type="number"]')
    await durationInput.setValue(12)
    await durationInput.trigger('change')

    const offInput = wrapper.findAll('input[type="time"]')[1]
    expect(offInput.element.value).toBe('18:00')
    expect(wrapper.emitted('update:modelOnHours')?.slice(-1)[0]).toEqual([12])
  })

  it('12/12 preset chip sets duration to 12 hours', async () => {
    const wrapper = mount(PhotoperiodClockEditor, {
      props: {
        modelLightsOnAt: '06:00',
        modelOnHours: 18,
        presets: [{ key: 'flower_12_12', label: '12/12', onHours: 12 }],
      },
    })

    const chip = wrapper.findAll('button.chip').find((b) => b.text() === '12/12')
    expect(chip).toBeTruthy()
    await chip.trigger('click')

    expect(wrapper.emitted('update:modelOnHours')?.slice(-1)[0]).toEqual([12])
    expect(wrapper.text()).toContain('12h ON')
    expect(wrapper.text()).toContain('12h OFF')
  })
})
