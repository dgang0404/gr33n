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

import api from '../api'
import GuardianModelSelector from '../components/GuardianModelSelector.vue'
import { useAuthStore } from '../stores/auth'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'

const ownerID = '00000000-0000-0000-0000-000000000001'

function stubApi({ members = [], farmModel = '' } = {}) {
  api.get.mockImplementation((url) => {
    if (url === '/guardian/models') {
      return Promise.resolve({
        data: {
          server_default: 'tinyllama',
            available_models: [
              { name: 'tinyllama', speed_class: 'fast', context_window: 2048, capabilities: ['completion'], loaded: false, runtime_hint: 'cold — first message may load the model' },
              { name: 'phi3:mini', speed_class: 'fast', context_window: 131072, effective_context_window: 4096, capabilities: ['completion'], loaded: false, runtime_hint: 'cold' },
            ],
        },
      })
    }
    if (url === '/farms/1') {
      return Promise.resolve({
        data: { id: 1, owner_user_id: ownerID, guardian_preferred_model: farmModel },
      })
    }
    if (url === '/farms/1/members') {
      return Promise.resolve({ data: members })
    }
    return Promise.resolve({ data: {} })
  })
}

describe('Phase 117 — GuardianModelSelector', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders model options after load', async () => {
    stubApi()
    const auth = useAuthStore()
    auth.userId = ownerID
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID, guardian_preferred_model: '' }

    const wrapper = mount(GuardianModelSelector, { props: { farmId: 1 } })
    await flushPromises()

    expect(wrapper.find('[data-test="guardian-model-selector"]').exists()).toBe(true)
    const options = wrapper.find('[data-test="guardian-session-model"]').findAll('option')
    expect(options.some((o) => o.text().includes('tinyllama'))).toBe(true)
  })

  it('shows readonly farm default for non-admin members', async () => {
    const viewerID = '00000000-0000-0000-0000-000000000099'
    stubApi({
      members: [{ user_id: viewerID, role_in_farm: 'viewer' }],
      farmModel: 'phi3:mini',
    })
    const auth = useAuthStore()
    auth.userId = viewerID

    const wrapper = mount(GuardianModelSelector, {
      props: { farmId: 1, farmContextActive: true },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="guardian-farm-model-readonly"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="guardian-pull-model-btn"]').exists()).toBe(false)
  })

  it('explains field guides are not used in quick chat mode', async () => {
    stubApi()
    const wrapper = mount(GuardianModelSelector, {
      props: { farmId: 1, farmContextActive: false },
    })
    await flushPromises()

    const banner = wrapper.find('[data-test="guardian-mode-banner"]')
    expect(banner.text()).toContain('Quick chat')
    const help = wrapper.find('[data-test="guardian-field-guides-help"]')
    expect(help.text()).toContain('only')
    expect(help.text()).toContain('off')
  })

  it('shows grounded block hint when farm context on and tinyllama selected', async () => {
    stubApi()
    const auth = useAuthStore()
    auth.userId = ownerID
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID, guardian_preferred_model: '' }

    const wrapper = mount(GuardianModelSelector, {
      props: { farmId: 1, farmContextActive: true },
    })
    await flushPromises()

    const options = wrapper.find('[data-test="guardian-session-model"]').findAll('option')
    const values = options.map((o) => o.element.value)
    expect(values).not.toContain('tinyllama')
    expect(wrapper.find('[data-test="guardian-session-model"]').element.value).toBe('phi3:mini')
    expect(wrapper.find('[data-test="guardian-grounded-block-hint"]').exists()).toBe(false)
  })

  it('disables farm save when draft is server default and server default blocks grounded', async () => {
    stubApi()
    const auth = useAuthStore()
    auth.userId = ownerID
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID, guardian_preferred_model: 'phi3:mini' }

    const wrapper = mount(GuardianModelSelector, {
      props: { farmId: 1, farmContextActive: true },
    })
    await flushPromises()

    await wrapper.find('[data-test="guardian-farm-model"]').setValue('')
    await flushPromises()

    const save = wrapper.find('[data-test="guardian-farm-model-save"]')
    expect(save.attributes('disabled')).toBeDefined()
  })
})
