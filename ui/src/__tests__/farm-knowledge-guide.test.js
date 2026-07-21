import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import HelpFarmKnowledgeGuide from '../components/HelpFarmKnowledgeGuide.vue'
import {
  FARM_KNOWLEDGE_ACTIONS,
  FARM_KNOWLEDGE_BOUNDARIES,
  INSERT_COMMONS_SUMMARY,
} from '../lib/farmKnowledgeGuide.js'

describe('farm knowledge guide', () => {
  it('exports action and boundary tables for Help glossary', () => {
    expect(FARM_KNOWLEDGE_ACTIONS.length).toBeGreaterThanOrEqual(5)
    expect(FARM_KNOWLEDGE_BOUNDARIES.some((r) => r.category.includes('Insert Commons'))).toBe(true)
    expect(INSERT_COMMONS_SUMMARY.includes.some((s) => s.includes('pseudonym'))).toBe(true)
  })

  it('renders tables at bottom of Help glossary section', () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/operator-guide', component: { template: '<div/>' } }],
    })
    const wrapper = mount(HelpFarmKnowledgeGuide, {
      global: { plugins: [router], stubs: { RouterLink: true } },
    })
    expect(wrapper.find('[data-test="help-farm-knowledge-guide"]').exists()).toBe(true)
    expect(wrapper.text()).toMatch(/If you want X, do Y/)
    expect(wrapper.text()).toMatch(/Insert Commons/)
    expect(wrapper.text()).toMatch(/coarse stats/)
  })
})
