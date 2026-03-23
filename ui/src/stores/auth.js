import { defineStore } from 'pinia'
import api from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('gr33n_token') ?? null,
    username: localStorage.getItem('gr33n_user') ?? null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    async login(username, password) {
      const res = await api.post('/auth/login', { username, password })
      this.token = res.data.token
      this.username = username
      localStorage.setItem('gr33n_token', this.token)
      localStorage.setItem('gr33n_user', username)
    },

    logout() {
      this.token = null
      this.username = null
      localStorage.removeItem('gr33n_token')
      localStorage.removeItem('gr33n_user')
    },
  },
})
