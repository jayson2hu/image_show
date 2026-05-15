import { defineStore } from 'pinia'

import type { Conversation, Message } from '@/api/types'

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    list: [] as Conversation[],
    currentId: null as number | null,
    messages: {} as Record<number, Message[]>,
    loading: false,
    searchQuery: '',
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
    sendLocalMessage(prompt: string) {
      if (!this.currentId) {
        this.createLocalConversation()
      }
      const conversationId = this.currentId as number
      const message: Message = {
        id: Date.now(),
        conversation_id: conversationId,
        prompt,
        task_kind: 'text2img',
        created_at: new Date().toISOString(),
      }
      this.messages[conversationId] = [...(this.messages[conversationId] || []), message]
      const conversation = this.list.find((item) => item.id === conversationId)
      if (conversation) {
        conversation.title = prompt.slice(0, 24) || '新对话'
        conversation.msg_count = this.messages[conversationId].length
        conversation.last_msg_at = new Date().toISOString()
      }
    },
  },
})
