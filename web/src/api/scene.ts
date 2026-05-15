import api from '@/api'
import type { Scene } from '@/api/types'

interface RawScene {
  id?: number
  label?: string
  name?: string
  icon?: string
  description?: string
  prompt?: string
  prompt_template?: string
  recommended_ratio?: string
  credit_cost?: number
  default_layered?: boolean
  default_layer_count?: number
  sort_order?: number
}

const fallbackScenes: Scene[] = [
  {
    id: 1,
    name: '小红书封面',
    icon: '📸',
    description: '',
    prompt_template: '小红书封面图，精致生活方式视觉，一眼能看懂主题，清晰大标题留白，明亮干净的构图',
    recommended_ratio: 'portrait_3_4',
    credit_cost: 2,
    default_layered: false,
    default_layer_count: 5,
    sort_order: 40,
  },
  {
    id: 2,
    name: '商品展示图',
    icon: '🛒',
    description: '',
    prompt_template: '电商商品展示图，主体突出，干净背景，真实材质，高级商业摄影光影',
    recommended_ratio: 'square',
    credit_cost: 1,
    default_layered: false,
    default_layer_count: 5,
    sort_order: 41,
  },
  {
    id: 3,
    name: '海报设计',
    icon: '🎨',
    description: '',
    prompt_template: '活动宣传海报视觉，主题突出，层次清晰，保留文字排版空间，适合促销活动和创意传播',
    recommended_ratio: 'portrait_3_4',
    credit_cost: 2,
    default_layered: true,
    default_layer_count: 5,
    sort_order: 43,
  },
  {
    id: 4,
    name: '社交头像',
    icon: '👤',
    description: '',
    prompt_template: '精致社交头像，主体居中，五官清晰，背景简洁，有辨识度',
    recommended_ratio: 'square',
    credit_cost: 1,
    default_layered: false,
    default_layer_count: 5,
    sort_order: 42,
  },
  {
    id: 5,
    name: '自由创作',
    icon: '✨',
    description: '',
    prompt_template: '',
    recommended_ratio: 'square',
    credit_cost: 1,
    default_layered: false,
    default_layer_count: 5,
    sort_order: 45,
  },
]

export async function listScenes(): Promise<{ items: Scene[] }> {
  const response = await api.get<{ items: RawScene[] }>('/generation/scenes')
  const items = (response.data.items || []).map((item, index) => {
    const name = item.name || item.label || '未命名场景'
    const posterScene = name.includes('海报')
    return {
      id: item.id || index + 1,
      name,
      icon: item.icon || '✨',
      description: item.description || '',
      prompt_template: item.prompt_template || item.prompt || '',
      recommended_ratio: item.recommended_ratio || 'square',
      credit_cost: item.credit_cost || 1,
      default_layered: item.default_layered ?? posterScene,
      default_layer_count: item.default_layer_count || 5,
      sort_order: item.sort_order ?? index,
    }
  })
  return { items: items.length ? items : fallbackScenes }
}

export function fallbackSceneItems() {
  return fallbackScenes
}
