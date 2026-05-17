import api from '@/api'
import type { Message } from '@/api/types'

export interface MessageListResponse {
  items: Message[]
}

export interface CreateMessageResponse {
  message: Message
  generation_id: number | null
}

export interface CreateMessagePayload {
  prompt: string
  size: string
  style_id: string
  scene_id: string
  layered: boolean
  layer_count: number
  attachment?: File | null
}

export function listMessages(conversationId: number) {
  return api.get<MessageListResponse>(`/conversations/${conversationId}/messages`)
}

export function createMessage(conversationId: number, payload: CreateMessagePayload) {
  if (payload.attachment) {
    const form = new FormData()
    form.append('prompt', payload.prompt)
    form.append('size', payload.size)
    form.append('style_id', payload.style_id)
    form.append('scene_id', payload.scene_id)
    form.append('layered', String(payload.layered))
    form.append('layer_count', String(payload.layer_count))
    form.append('image', payload.attachment)
    return api.post<CreateMessageResponse>(`/conversations/${conversationId}/messages`, form)
  }
  return api.post<CreateMessageResponse>(`/conversations/${conversationId}/messages`, {
    prompt: payload.prompt,
    size: payload.size,
    style_id: payload.style_id,
    scene_id: payload.scene_id,
    layered: payload.layered,
    layer_count: payload.layer_count,
  })
}
