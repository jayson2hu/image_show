import api from '@/api'
import type { Conversation } from '@/api/types'

export interface ConversationListResponse {
  items: Conversation[]
  next_cursor: string
}

export function listConversations(params?: { q?: string; cursor?: string; limit?: number }) {
  return api.get<ConversationListResponse>('/conversations', { params })
}

export function createConversation(title?: string) {
  return api.post<Conversation>('/conversations', { title })
}

export function renameConversation(id: number, title: string) {
  return api.patch<Conversation>(`/conversations/${id}`, { title })
}

export function deleteConversation(id: number) {
  return api.delete(`/conversations/${id}`)
}
