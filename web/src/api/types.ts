export interface Conversation {
  id: number
  title: string
  msg_count: number
  last_msg_at: string
  is_layered: boolean
  total_cost: number
}

export interface Message {
  id: number
  conversation_id: number
  user_id?: number
  anonymous_id?: string
  prompt: string
  attachment_url?: string
  attachment_name?: string
  attachment_size?: number
  task_kind: 'text2img' | 'img_restore' | 'img_style_transfer' | 'img2img_generic'
  size?: string
  style_id?: string
  scene_id?: string
  layered?: boolean
  layer_count?: number
  generation_id?: number | null
  created_at?: string
  _sending?: boolean
  _error?: string
}

export interface Generation {
  id?: number
  status: number
  message?: string
  image_url?: string
  error?: string
  error_msg?: string
  size?: string
  credits_cost?: number
}

export interface Scene {
  id: number
  name: string
  icon: string
  description: string
  prompt_template: string
  recommended_ratio: string
  credit_cost: number
  default_layered: boolean
  default_layer_count: number
  sort_order: number
}
