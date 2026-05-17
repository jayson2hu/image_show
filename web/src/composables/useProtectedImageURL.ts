import { onUnmounted, ref, watch, type Ref } from 'vue'

import api from '@/api'

export function useProtectedImageURL(source: Ref<string>) {
  const imageURL = ref('')
  let objectURL = ''

  function revoke() {
    if (objectURL) {
      URL.revokeObjectURL(objectURL)
      objectURL = ''
    }
  }

  watch(
    source,
    async (next) => {
      revoke()
      imageURL.value = ''
      if (!next) return

      if (!next.startsWith('/api/')) {
        imageURL.value = next
        return
      }

      try {
        const response = await api.get(next.replace(/^\/api/, ''), { responseType: 'blob' })
        objectURL = URL.createObjectURL(response.data)
        imageURL.value = objectURL
      } catch {
        imageURL.value = ''
      }
    },
    { immediate: true },
  )

  onUnmounted(revoke)

  return imageURL
}
