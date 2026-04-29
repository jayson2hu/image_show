<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  generationId: number
}>()

const emit = defineEmits<{
  completed: [url: string]
  failed: [message: string]
  cancelled: []
  cancel: []
}>()

const messages = [
  '正在创建图片...',
  '正在整理提示词...',
  '正在生成初稿...',
  '正在设置场景...',
  '正在润色细节...',
  '即将完成...',
]
const message = ref(messages[0])
let source: EventSource | null = null
let timer = 0

onMounted(() => {
  let index = 0
  timer = window.setInterval(() => {
    index = (index + 1) % messages.length
    message.value = messages[index]
  }, 3500)

  source = new EventSource(`/api/generations/${props.generationId}/stream`)
  source.addEventListener('status', (event) => {
    const payload = JSON.parse((event as MessageEvent).data)
    if (payload.message) {
      message.value = payload.message
    }
    if (payload.status === 3) {
      emit('completed', payload.image_url)
      close()
    }
    if (payload.status === 4) {
      emit('failed', payload.error || '生成失败，请重试')
      close()
    }
    if (payload.status === 5) {
      emit('cancelled')
      close()
    }
  })
  source.onerror = () => {
    emit('failed', '连接中断，请稍后重试')
    close()
  }
})

onUnmounted(close)

function close() {
  if (timer) {
    window.clearInterval(timer)
    timer = 0
  }
  source?.close()
  source = null
}
</script>

<template>
  <div class="rounded border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-900">
    <div class="h-48 animate-pulse rounded bg-slate-100 dark:bg-slate-800"></div>
    <div class="mt-4 flex items-center justify-between gap-3">
      <p class="text-sm text-slate-700 dark:text-slate-200">{{ message }}</p>
      <button class="min-h-10 rounded border border-slate-300 px-3 py-2 text-sm dark:border-slate-600" type="button" @click="emit('cancel')">
        取消
      </button>
    </div>
  </div>
</template>
