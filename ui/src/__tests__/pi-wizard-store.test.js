import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePiWizardStore } from '../stores/piWizardStore.js'

describe('Phase 117 — Pi wizard store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts on step 1 with step1 validation complete', () => {
    const store = usePiWizardStore()
    expect(store.currentStep).toBe(1)
    expect(store.validation.step1.complete).toBe(true)
    expect(store.canAdvance).toBe(true)
  })

  it('advances and retreats between steps when validation passes', () => {
    const store = usePiWizardStore()
    store.setStep(2)
    store.updateValidation(2, [])
    store.nextStep()
    expect(store.currentStep).toBe(3)
    store.prevStep()
    expect(store.currentStep).toBe(2)
  })

  it('blocks advance when step has validation errors', () => {
    const store = usePiWizardStore()
    store.setStep(2)
    store.updateValidation(2, ['Device name required'])
    store.nextStep()
    expect(store.currentStep).toBe(2)
    expect(store.canAdvance).toBe(false)
  })

  it('tracks channel assignments and device fields', () => {
    const store = usePiWizardStore()
    store.updateDevice({ name: 'Relay Pi', uid: 'pi-demo-01', farmId: 1 })
    store.updateChannelAssignment('CH1', 42)
    expect(store.formData.device.name).toBe('Relay Pi')
    expect(store.channelCount).toBe(1)
    store.clearChannelAssignments()
    expect(store.channelCount).toBe(0)
  })

  it('reset returns to initial wizard state', () => {
    const store = usePiWizardStore()
    store.setStep(4)
    store.setApiKey('secret')
    store.reset()
    expect(store.currentStep).toBe(1)
    expect(store.formData.device.apiKey).toBe('')
  })
})
