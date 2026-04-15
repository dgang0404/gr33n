import { defineStore } from 'pinia'
import api from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('gr33n_token') ?? null,
    username: localStorage.getItem('gr33n_user') ?? null,
    userId: localStorage.getItem('gr33n_user_id') ?? null,
    authMode: null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isDevMode: (state) => state.authMode === 'dev',
    isAuthTestMode: (state) => state.authMode === 'auth_test',
  },

  actions: {
    async fetchAuthMode() {
      try {
        const res = await api.get('/auth/mode')
        this.authMode = res.data?.mode ?? 'production'
      } catch {
        this.authMode = 'production'
      }
    },

    async login(username, password) {
      const res = await api.post('/auth/login', { username, password })
      this.token = res.data.token
      this.username = username
      localStorage.setItem('gr33n_token', this.token)
      localStorage.setItem('gr33n_user', username)
      if (res.data.user_id) {
        this.userId = res.data.user_id
        localStorage.setItem('gr33n_user_id', res.data.user_id)
      }
    },

    logout() {
      this.token = null
      this.username = null
      this.userId = null
      localStorage.removeItem('gr33n_token')
      localStorage.removeItem('gr33n_user')
      localStorage.removeItem('gr33n_user_id')
    },
  },
})
