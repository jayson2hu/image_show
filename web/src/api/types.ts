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
  prompt: string
  task_kind: 'text2img' | 'img_restore' | 'img_style_transfer' | 'img2img_generic'
  created_at?: string
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
