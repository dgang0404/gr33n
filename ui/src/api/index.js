import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
  timeout: 8000,
  headers: { 'Content-Type': 'application/json' },
})

// Attach JWT token to every request (if present)
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('gr33n_token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

// On 401 — clear token and redirect to login
api.interceptors.response.use(
  r => r,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('gr33n_token')
      localStorage.removeItem('gr33n_user')
      // Only redirect if not already on /login to avoid loops
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    console.error('[gr33n api]', err.config?.url, err.message)
    return Promise.reject(err)
  }
)

export default api
