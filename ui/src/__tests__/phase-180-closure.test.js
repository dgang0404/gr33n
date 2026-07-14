/**
 * Phase 180 WS1 — Help knowledge surfaces map.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import HelpKnowledgeSurfacesMap from '../components/HelpKnowledgeSurfacesMap.vue'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/operator-guide', component: { template: '<div/>' } },
    { path: '/symptom-guide', component: { template: '<div/>' } },
  ],
})

describe('Phase 180 WS1 — help knowledge surfaces map', () => {
  it('renders four surface cards with correct links', async () => {
    const wrapper = mount(HelpKnowledgeSurfacesMap, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('[data-test="help-knowledge-surfaces-map"]').exists()).toBe(true)

    const guide = wrapper.find('[data-test="help-surface-card-guide"]')
    const knowledge = wrapper.find('[data-test="help-surface-card-knowledge"]')
    const catalog = wrapper.find('[data-test="help-surface-card-catalog"]')
    const symptoms = wrapper.find('[data-test="help-surface-card-symptoms"]')

    expect(guide.exists()).toBe(true)
    expect(knowledge.exists()).toBe(true)
    expect(catalog.exists()).toBe(true)
    expect(symptoms.exists()).toBe(true)

    expect(guide.attributes('href')).toContain('/operator-guide')
    expect(guide.attributes('href')).toContain('tab=guide')
    expect(knowledge.attributes('href')).toContain('tab=knowledge')
    expect(catalog.attributes('href')).toContain('tab=catalog')
    expect(symptoms.attributes('href')).toContain('/symptom-guide')

    expect(guide.text()).toMatch(/Guide/i)
    expect(knowledge.text()).toMatch(/semantic search/i)
    expect(catalog.text()).toMatch(/Commons/i)
    expect(symptoms.text()).toMatch(/Symptom/i)
  })
})
