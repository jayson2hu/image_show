import api from '@/api'

export interface CreateGenerationResponse {
  id: number
  status: number
  anonymous_id?: string
}

export interface CreateGenerationPayload {
  prompt: string
  size: string
  captcha_token?: string
}

export interface CreateImageEditPayload extends CreateGenerationPayload {
  image: File
}

export function createGeneration(payload: CreateGenerationPayload) {
  return api.post<CreateGenerationResponse>('/generations', {
    prompt: payload.prompt,
    size: payload.size,
    captcha_token: payload.captcha_token || '',
  })
}

export function createImageEdit(payload: CreateImageEditPayload) {
  const form = new FormData()
  form.append('prompt', payload.prompt)
  form.append('size', payload.size)
  form.append('captcha_token', payload.captcha_token || '')
  form.append('image', payload.image)
  return api.post<CreateGenerationResponse>('/generations/edit', form)
}
