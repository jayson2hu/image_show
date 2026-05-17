import { onUnmounted, ref, watch, type Ref } from 'vue'

import api from '@/api'
import type { Generation } from '@/api/types'

export function useGenerationPoll(generationId: Ref<number | null | undefined>) {
  const generation = ref<Generation | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  let eventSource: EventSource | null = null
  let fallbackTimer: number | null = null

  function close() {
    eventSource?.close()
    eventSource = null
    if (fallbackTimer) {
      window.clearInterval(fallbackTimer)
      fallbackTimer = null
    }
  }

  function isTerminal(status: number) {
    return status === 3 || status === 4 || status === 5
  }

  async function fetchDetail(id: number) {
    const response = await api.get(`/generations/${id}`)
    const item = response.data.item || response.data
    if (!item.image_url && item.id) {
      item.image_url = `/api/generations/${item.id}/image`
    }
    generation.value = item
    if (isTerminal(item.status)) {
      loading.value = false
      if (item.status === 4) {
        error.value = item.error_msg || item.error || 'generation failed'
      }
      close()
    }
  }

  function startFallback(id: number) {
    if (fallbackTimer) return
    fetchDetail(id).catch(() => undefined)
    fallbackTimer = window.setInterval(() => {
      fetchDetail(id).catch(() => undefined)
    }, 2000)
  }

  function connect(id: number) {
    close()
    loading.value = true
    error.value = null
    eventSource = new EventSource(`/api/generations/${id}/stream`)
    eventSource.addEventListener('status', (event) => {
      const payload = JSON.parse((event as MessageEvent).data)
      if (payload.status === 3 && !payload.image_url) {
        payload.image_url = `/api/generations/${id}/image`
      }
      generation.value = payload
      if (isTerminal(payload.status)) {
        loading.value = false
        if (payload.status === 4) {
          error.value = payload.error || 'generation failed'
        }
        close()
      }
    })
    eventSource.onerror = () => {
      eventSource?.close()
      eventSource = null
      startFallback(id)
    }
  }

  watch(
    generationId,
    (id) => {
      if (id) {
        connect(id)
      } else {
        close()
        generation.value = null
        loading.value = false
        error.value = null
      }
    },
    { immediate: true },
  )

  onUnmounted(close)

  return { generation, loading, error }
}
