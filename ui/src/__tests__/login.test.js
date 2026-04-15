import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: { mode: 'dev' } }),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

import Login from '../views/Login.vue'

describe('Login.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('mounts without crashing', () => {
    const wrapper = mount(Login)
    expect(wrapper.exists()).toBe(true)
  })

  it('renders the sign-in heading', () => {
    const wrapper = mount(Login)
    expect(wrapper.text()).toContain('Sign in')
  })

  it('renders username and password inputs', () => {
    const wrapper = mount(Login)
    const inputs = wrapper.findAll('input')
    expect(inputs.length).toBeGreaterThanOrEqual(2)
    expect(inputs[0].attributes('type')).toBe('text')
    expect(inputs[1].attributes('type')).toBe('password')
  })

  it('renders submit button', () => {
    const wrapper = mount(Login)
    const btn = wrapper.find('button[type="submit"]')
    expect(btn.exists()).toBe(true)
    expect(btn.text()).toContain('Sign in')
  })
})
