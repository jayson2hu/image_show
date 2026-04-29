import { defineStore } from 'pinia'

import api from '@/api'

interface User {
  id: number
  email: string
  credits: number
  role: number
  status: number
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
