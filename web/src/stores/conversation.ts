import { defineStore } from 'pinia'

import { claimGuestConversation, createConversation, deleteConversation, listConversations, renameConversation } from '@/api/conversation'
import { createGeneration, createImageEdit } from '@/api/generation'
import { createMessage, listMessages, type CreateMessagePayload } from '@/api/message'
import type { Conversation, Message } from '@/api/types'
import { useUserStore } from '@/stores/user'

const GUEST_CONVERSATION_ID = -1

function autoTitleFromPrompt(prompt: string) {
  return prompt.slice(0, 12) + (prompt.length > 12 ? '...' : '')
}

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    list: [] as Conversation[],
    currentId: null as number | null,
    messages: {} as Record<number, Message[]>,
    loading: false,
    messageLoading: false,
    sending: false,
    searchQuery: '',
  }),
  getters: {
    currentConversation: (state) => state.list.find((item) => item.id === state.currentId) || null,
    currentMessages: (state) => (state.currentId ? state.messages[state.currentId] || [] : []),
  },
  actions: {
    async loadConversations() {
      const userStore = useUserStore()
      if (!userStore.token) {
        this.list = []
        this.currentId = null
        this.messages = {}
        return
      }
      this.loading = true
      try {
        const response = await listConversations({ q: this.searchQuery || undefined })
        this.list = response.data.items
        if (this.currentId !== null && this.currentId < 0) {
          this.currentId = null
        }
        delete this.messages[GUEST_CONVERSATION_ID]
        if (!this.currentId && this.list.length > 0) {
          await this.selectConversation(this.list[0].id)
        }
      } finally {
        this.loading = false
      }
    },
    hasClaimableGuestConversation() {
      return (this.messages[GUEST_CONVERSATION_ID] || []).some((message) => Number(message.generation_id) > 0)
    },
    clearGuestConversation() {
      this.list = this.list.filter((item) => item.id !== GUEST_CONVERSATION_ID)
      delete this.messages[GUEST_CONVERSATION_ID]
      if (this.currentId === GUEST_CONVERSATION_ID) {
        this.currentId = null
      }
    },
    async syncGuestConversation() {
      const userStore = useUserStore()
      if (!userStore.token || !this.hasClaimableGuestConversation()) {
        return
      }

      const guestConversation = this.list.find((item) => item.id === GUEST_CONVERSATION_ID)
      const guestMessages = this.messages[GUEST_CONVERSATION_ID] || []
      const claimableMessages = guestMessages
        .filter((message) => Number(message.generation_id) > 0)
        .map((message) => ({
          generation_id: Number(message.generation_id),
          prompt: message.prompt,
          task_kind: message.task_kind,
          size: message.size,
          style_id: message.style_id,
          scene_id: message.scene_id,
          layered: Boolean(message.layered),
          layer_count: message.layer_count || 0,
        }))

      if (claimableMessages.length === 0) {
        this.clearGuestConversation()
        return
      }

      try {
        const response = await claimGuestConversation({
          title: guestConversation?.title,
          messages: claimableMessages,
        })
        if (response.data.claimed > 0 && response.data.conversation?.id) {
          const conversation = response.data.conversation
          this.list = [conversation, ...this.list.filter((item) => item.id !== GUEST_CONVERSATION_ID && item.id !== conversation.id)]
          this.messages[conversation.id] = response.data.messages || []
          this.currentId = conversation.id
          delete this.messages[GUEST_CONVERSATION_ID]
          return
        }
      } catch (error) {
        console.warn('sync guest conversation failed', error)
        return
      }

      this.clearGuestConversation()
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
      const userStore = useUserStore()
      if (!userStore.token) {
        const existing = this.list.find((item) => item.id === GUEST_CONVERSATION_ID)
        if (!existing) {
          this.list = [
            {
              id: GUEST_CONVERSATION_ID,
              title: '新对话',
              msg_count: 0,
              last_msg_at: new Date().toISOString(),
              is_layered: false,
              total_cost: 0,
            },
            ...this.list,
          ]
          this.messages[GUEST_CONVERSATION_ID] = []
        }
        this.currentId = GUEST_CONVERSATION_ID
        return GUEST_CONVERSATION_ID
      }
      if (this.currentId !== null && this.currentId < 0) {
        this.currentId = null
        this.list = this.list.filter((item) => item.id >= 0)
        delete this.messages[GUEST_CONVERSATION_ID]
      }
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
      if (id < 0) {
        this.messages[id] = this.messages[id] || []
        return
      }
      this.messageLoading = true
      try {
        const response = await listMessages(id)
        this.messages[id] = response.data.items
      } finally {
        this.messageLoading = false
      }
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
        const userStore = useUserStore()
        if (!userStore.token) {
          const response = payload.attachment
            ? await createImageEdit({ prompt: normalizedPrompt, size: payload.size, image: payload.attachment })
            : await createGeneration({ prompt: normalizedPrompt, size: payload.size })
          const message: Message = {
            ...pendingMessage,
            id: response.data.id,
            generation_id: response.data.id,
            anonymous_id: response.data.anonymous_id,
            _sending: false,
          }
          this.messages[conversationId] = (this.messages[conversationId] || []).map((item) => (item.id === tempId ? message : item))
          this.list = this.list.map((item) =>
            item.id === conversationId
              ? {
                  ...item,
                  title: item.msg_count === 0 ? autoTitleFromPrompt(normalizedPrompt) : item.title,
                  msg_count: item.msg_count + 1,
                  last_msg_at: message.created_at || new Date().toISOString(),
                  is_layered: Boolean(payload.layered),
                }
              : item,
          )
          return
        }

        const response = await createMessage(conversationId, { ...payload, prompt: normalizedPrompt })
        const message = response.data.message
        this.messages[conversationId] = (this.messages[conversationId] || []).map((item) => (item.id === tempId ? message : item))
        const current = this.list.find((item) => item.id === conversationId)
        await this.loadConversations()
        if (current && current.msg_count === 0) {
          const nextTitle = autoTitleFromPrompt(normalizedPrompt)
          await this.updateConversationTitle(conversationId, nextTitle)
        }
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
      if (id < 0) {
        this.list = this.list.filter((item) => item.id !== id)
        delete this.messages[id]
        if (this.currentId === id) {
          this.currentId = null
          await this.ensureConversation()
        }
        return
      }
      await deleteConversation(id)
      this.list = this.list.filter((item) => item.id !== id)
      delete this.messages[id]

      if (this.currentId !== id) return

      if (this.list.length > 0) {
        await this.selectConversation(this.list[0].id)
      } else {
        this.currentId = null
        await this.ensureConversation()
      }
    },
  },
})
