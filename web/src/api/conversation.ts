import api from '@/api'
import type { Conversation, Message } from '@/api/types'

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

export interface ClaimGuestConversationMessage {
  generation_id: number
  prompt: string
  task_kind: Message['task_kind']
  size?: string
  style_id?: string
  scene_id?: string
  layered?: boolean
  layer_count?: number
}

export interface ClaimGuestConversationPayload {
  title?: string
  messages: ClaimGuestConversationMessage[]
}

export interface ClaimGuestConversationResponse {
  conversation: Conversation
  messages: Message[]
  claimed: number
}

export function claimGuestConversation(payload: ClaimGuestConversationPayload) {
  return api.post<ClaimGuestConversationResponse>('/conversations/claim-guest', payload)
}
