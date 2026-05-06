import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  config.headers['X-Fingerprint'] = getFingerprint()
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
      const loginPath = window.location.pathname.startsWith('/console/admin') ? '/console/admin/login' : '/login'
      if (window.location.pathname !== loginPath) {
        window.location.href = loginPath
      }
    }
    return Promise.reject(error)
  },
)

export default api

function getFingerprint() {
  const key = 'anonymous_fingerprint'
  const existing = localStorage.getItem(key)
  if (existing) {
    return existing
  }
  const raw = [
    navigator.userAgent,
    navigator.language,
    screen.width,
    screen.height,
    Intl.DateTimeFormat().resolvedOptions().timeZone,
    crypto.randomUUID(),
  ].join('|')
  localStorage.setItem(key, raw)
  return raw
}
