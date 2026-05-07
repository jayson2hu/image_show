import { inject, provide, reactive } from 'vue'

export type ToastType = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
}

export interface ToastContext {
  items: ToastItem[]
  success: (message: string) => void
  error: (message: string) => void
  info: (message: string) => void
  remove: (id: number) => void
}

const toastKey = Symbol('toast')
let nextToastId = 0

export function createToastContext(): ToastContext {
  const items = reactive<ToastItem[]>([])

  function remove(id: number) {
    const index = items.findIndex((item) => item.id === id)
    if (index !== -1) {
      items.splice(index, 1)
    }
  }

  function add(type: ToastType, message: string) {
    const id = ++nextToastId
    items.push({ id, type, message })
    if (items.length > 5) {
      items.shift()
    }
    window.setTimeout(() => remove(id), type === 'error' ? 5000 : 3000)
  }

  return {
    items,
    success: (message) => add('success', message),
    error: (message) => add('error', message),
    info: (message) => add('info', message),
    remove,
  }
}

export function provideToast() {
  const context = createToastContext()
  provide(toastKey, context)
  return context
}

export function useToast() {
  const context = inject<ToastContext>(toastKey)
  if (!context) {
    throw new Error('useToast() requires provideToast() in an ancestor component')
  }
  return context
}
