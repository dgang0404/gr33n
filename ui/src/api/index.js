import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
  timeout: 8000,
  headers: { 'Content-Type': 'application/json' },
})

/** Set from `main.js` so Pinia stays in sync with localStorage (avoids stale UI after 401). */
let onUnauthorized = null
export function setUnauthorizedHandler(fn) {
  onUnauthorized = fn
}

// Attach JWT token to every request (if present)
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('gr33n_token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  // Default axios Content-Type is application/json — that breaks multipart uploads
  // (layout background, zone photos, receipts). Let the browser set the boundary.
  if (typeof FormData !== 'undefined' && config.data instanceof FormData) {
    if (config.headers && typeof config.headers.delete === 'function') {
      config.headers.delete('Content-Type')
    } else if (config.headers) {
      delete config.headers['Content-Type']
      delete config.headers['content-type']
    }
  }
  return config
})

/** Noisy failures callers already handle or that happen during normal navigation / API restarts. */
function shouldLogApiError(err) {
  const url = err.config?.url || ''
  if (err.response?.status === 404 && url.includes('/readings/latest')) return false
  if (err.response?.status === 404 && url.includes('/layout-background')) return false
  if (url.includes('/v1/chat/sessions/') && url.endsWith('/close')) return false
  if (err.code === 'ERR_CANCELED' || axios.isCancel?.(err)) return false
  if (/abort|canceled/i.test(String(err.message || ''))) return false
  return true
}

// On 401 — clear session and redirect to login
api.interceptors.response.use(
  r => r,
  err => {
    if (err.response?.status === 401) {
      if (onUnauthorized) onUnauthorized()
      else {
        localStorage.removeItem('gr33n_token')
        localStorage.removeItem('gr33n_user')
        localStorage.removeItem('gr33n_user_id')
      }
      // Only redirect if not already on /login to avoid loops
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    if (shouldLogApiError(err)) {
      console.error('[gr33n api]', err.config?.url || '', err.message)
    }
    return Promise.reject(err)
  }
)

export default api
