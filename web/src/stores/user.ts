import { defineStore } from 'pinia'

import api from '@/api'

interface User {
  id: number
  username?: string
  email: string
  avatar_url?: string
  credits: number
  credits_expiry?: string | null
  role: number
  status: number
  created_at?: string
  updated_at?: string
  last_login_at?: string | null
  last_login_ip?: string
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null as User | null,
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    },
    async login(email: string, password: string) {
      const response = await api.post('/auth/login', { email, password })
      this.setToken(response.data.token)
      this.user = response.data.user
    },
    async wechatLogin(code: string) {
      const response = await api.get('/auth/wechat/callback', { params: { code } })
      this.setToken(response.data.token)
      this.user = response.data.user
    },
    async register(email: string, password: string, code: string) {
      const response = await api.post('/auth/register', { email, password, code })
      this.setToken(response.data.token)
      this.user = response.data.user
    },
    async fetchUser() {
      if (!this.token) {
        return
      }
      try {
        const response = await api.get('/auth/me')
        this.user = response.data.user
      } catch {
        this.logout()
      }
    },
  },
})
