import api from '@/api'

export interface CreateGenerationResponse {
  id: number
  status: number
  anonymous_id?: string
}

export interface CreateGenerationPayload {
  prompt: string
  size: string
  style_id?: string
  layered?: boolean
  layer_count?: number
  captcha_token?: string
}

export interface CreateImageEditPayload extends CreateGenerationPayload {
  image: File
}

export function createGeneration(payload: CreateGenerationPayload) {
  return api.post<CreateGenerationResponse>('/generations', {
    prompt: payload.prompt,
    size: payload.size,
    style_id: payload.style_id || '',
    layered: Boolean(payload.layered),
    layer_count: payload.layer_count || 0,
    captcha_token: payload.captcha_token || '',
  })
}

export function createImageEdit(payload: CreateImageEditPayload) {
  const form = new FormData()
  form.append('prompt', payload.prompt)
  form.append('size', payload.size)
  form.append('style_id', payload.style_id || '')
  form.append('layered', String(Boolean(payload.layered)))
  form.append('layer_count', String(payload.layer_count || 0))
  form.append('captcha_token', payload.captcha_token || '')
  form.append('image', payload.image)
  return api.post<CreateGenerationResponse>('/generations/edit', form)
}
