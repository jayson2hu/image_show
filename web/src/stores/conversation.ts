import { defineStore } from 'pinia'

import type { Conversation, Message } from '@/api/types'

const SIDEBAR_COLLAPSED_KEY = 'sidebar_collapsed'

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    list: [] as Conversation[],
    currentId: null as number | null,
    messages: {} as Record<number, Message[]>,
    loading: false,
    searchQuery: '',
    sidebarCollapsed: localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true',
  }),
  getters: {
    currentConversation: (state) => state.list.find((item) => item.id === state.currentId) || null,
    currentMessages: (state) => (state.currentId ? state.messages[state.currentId] || [] : []),
  },
  actions: {
    createLocalConversation() {
      const id = Date.now()
      const conversation: Conversation = {
        id,
        title: '新对话',
        msg_count: 0,
        last_msg_at: new Date().toISOString(),
        is_layered: false,
        total_cost: 0,
      }
      this.list.unshift(conversation)
      this.currentId = id
      this.messages[id] = []
    },
    selectConversation(id: number) {
      this.currentId = id
      if (!this.messages[id]) {
        this.messages[id] = []
      }
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(this.sidebarCollapsed))
    },
    sendLocalMessage(prompt: string) {
      const normalizedPrompt = prompt.trim()
      if (!normalizedPrompt) return

      if (!this.currentId) {
        this.createLocalConversation()
      }
      const conversationId = this.currentId as number
      const message: Message = {
        id: Date.now(),
        conversation_id: conversationId,
        prompt: normalizedPrompt,
        task_kind: 'text2img',
        created_at: new Date().toISOString(),
      }
      this.messages[conversationId] = [...(this.messages[conversationId] || []), message]
      const conversation = this.list.find((item) => item.id === conversationId)
      if (conversation) {
        const previousCount = conversation.msg_count
        conversation.msg_count = this.messages[conversationId].length
        conversation.last_msg_at = new Date().toISOString()
        if (previousCount === 0) {
          conversation.title = normalizedPrompt.length > 12 ? `${normalizedPrompt.slice(0, 12)}...` : normalizedPrompt
        }
      }
    },
    updateConversationTitle(id: number, title: string) {
      const nextTitle = title.trim().slice(0, 128)
      if (!nextTitle) return
      const conversation = this.list.find((item) => item.id === id)
      if (conversation) {
        conversation.title = nextTitle
        conversation.last_msg_at = new Date().toISOString()
      }
    },
    deleteLocalConversation(id: number) {
      this.list = this.list.filter((item) => item.id !== id)
      delete this.messages[id]

      if (this.currentId !== id) return

      if (this.list.length > 0) {
        this.selectConversation(this.list[0].id)
      } else {
        this.createLocalConversation()
      }
    },
  },
})
