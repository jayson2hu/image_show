import { defineStore } from 'pinia'

import { createConversation, deleteConversation, listConversations, renameConversation } from '@/api/conversation'
import { createMessage, listMessages, type CreateMessagePayload } from '@/api/message'
import type { Conversation, Message } from '@/api/types'

const SIDEBAR_COLLAPSED_KEY = 'sidebar_collapsed'

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    list: [] as Conversation[],
    currentId: null as number | null,
    messages: {} as Record<number, Message[]>,
    loading: false,
    messageLoading: false,
    sending: false,
    searchQuery: '',
    sidebarCollapsed: localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true',
  }),
  getters: {
    currentConversation: (state) => state.list.find((item) => item.id === state.currentId) || null,
    currentMessages: (state) => (state.currentId ? state.messages[state.currentId] || [] : []),
  },
  actions: {
    async loadConversations() {
      this.loading = true
      try {
        const response = await listConversations({ q: this.searchQuery || undefined })
        this.list = response.data.items
        if (!this.currentId && this.list.length > 0) {
          await this.selectConversation(this.list[0].id)
        }
      } finally {
        this.loading = false
      }
    },
    async createLocalConversation() {
      const response = await createConversation()
      const conversation = response.data
      this.list = [conversation, ...this.list.filter((item) => item.id !== conversation.id)]
      this.currentId = conversation.id
      this.messages[conversation.id] = []
      return conversation
    },
    async ensureConversation() {
      if (this.currentId) return this.currentId
      if (this.list.length > 0) {
        await this.selectConversation(this.list[0].id)
        return this.currentId as number
      }
      const conversation = await this.createLocalConversation()
      return conversation.id
    },
    async selectConversation(id: number) {
      this.currentId = id
      if (!this.messages[id]) {
        await this.loadMessages(id)
      }
    },
    async loadMessages(id: number) {
      this.messageLoading = true
      try {
        const response = await listMessages(id)
        this.messages[id] = response.data.items
      } finally {
        this.messageLoading = false
      }
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(this.sidebarCollapsed))
    },
    async sendMessage(payload: CreateMessagePayload) {
      const normalizedPrompt = payload.prompt.trim()
      if (!normalizedPrompt) return

      const conversationId = await this.ensureConversation()
      const tempId = -Date.now()
      const pendingMessage: Message = {
        id: tempId,
        conversation_id: conversationId,
        prompt: normalizedPrompt,
        task_kind: payload.attachment ? 'img2img_generic' : 'text2img',
        size: payload.size,
        style_id: payload.style_id,
        scene_id: payload.scene_id,
        layered: payload.layered,
        layer_count: payload.layer_count,
        created_at: new Date().toISOString(),
        _sending: true,
      }
      this.messages[conversationId] = [...(this.messages[conversationId] || []), pendingMessage]
      this.sending = true

      try {
        const response = await createMessage(conversationId, { ...payload, prompt: normalizedPrompt })
        const message = response.data.message
        this.messages[conversationId] = (this.messages[conversationId] || []).map((item) => (item.id === tempId ? message : item))
        await this.loadConversations()
        this.currentId = conversationId
      } catch (error: any) {
        const message = error.response?.data?.message || error.response?.data?.error || 'send failed'
        this.messages[conversationId] = (this.messages[conversationId] || []).map((item) => (item.id === tempId ? { ...item, _sending: false, _error: message } : item))
        throw error
      } finally {
        this.sending = false
      }
    },
    async updateConversationTitle(id: number, title: string) {
      const nextTitle = title.trim().slice(0, 128)
      if (!nextTitle) return
      const response = await renameConversation(id, nextTitle)
      const index = this.list.findIndex((item) => item.id === id)
      if (index >= 0) {
        this.list[index] = response.data
      }
    },
    async deleteLocalConversation(id: number) {
      await deleteConversation(id)
      this.list = this.list.filter((item) => item.id !== id)
      delete this.messages[id]

      if (this.currentId !== id) return

      if (this.list.length > 0) {
        await this.selectConversation(this.list[0].id)
      } else {
        this.currentId = null
      }
    },
  },
})
