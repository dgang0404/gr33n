import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import { useAuthStore } from '../stores/auth'
import api from '../api'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('starts logged out', () => {
    const auth = useAuthStore()
    expect(auth.isLoggedIn).toBe(false)
    expect(auth.token).toBeNull()
  })

  it('login() stores token and username', async () => {
    api.post.mockResolvedValue({ data: { token: 'jwt-abc' } })
    const auth = useAuthStore()

    await auth.login('admin', 'secret')

    expect(auth.token).toBe('jwt-abc')
    expect(auth.username).toBe('admin')
    expect(auth.isLoggedIn).toBe(true)
    expect(localStorage.getItem('gr33n_token')).toBe('jwt-abc')
    expect(localStorage.getItem('gr33n_user')).toBe('admin')
  })

  it('logout() clears token and username', async () => {
    api.post.mockResolvedValue({ data: { token: 'jwt-abc' } })
    const auth = useAuthStore()
    await auth.login('admin', 'secret')

    auth.logout()

    expect(auth.token).toBeNull()
    expect(auth.username).toBeNull()
    expect(auth.isLoggedIn).toBe(false)
    expect(localStorage.getItem('gr33n_token')).toBeNull()
  })

  it('fetchAuthMode() defaults to dev on error', async () => {
    api.get.mockRejectedValue(new Error('network'))
    const auth = useAuthStore()

    await auth.fetchAuthMode()

    expect(auth.authMode).toBe('dev')
    expect(auth.isDevMode).toBe(true)
  })
})
