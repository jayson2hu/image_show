<script setup lang="ts">
import { computed } from 'vue'

import type { Message } from '@/api/types'
import { useGenerationPoll } from '@/composables/useGenerationPoll'
import { useProtectedImageURL } from '@/composables/useProtectedImageURL'

const props = defineProps<{ message: Message }>()
const generationId = computed(() => props.message.generation_id)
const { generation, loading, error } = useGenerationPoll(generationId)

const rawImageUrl = computed(() => generation.value?.image_url || '')
const imageUrl = useProtectedImageURL(rawImageUrl)
const progress = computed(() => {
  if (props.message._sending) return 8
  switch (generation.value?.status) {
    case 0:
      return 12
    case 1:
      return 45
    case 2:
      return 78
    case 3:
      return 100
    case 4:
    case 5:
      return 100
    default:
      return loading.value ? 24 : 0
  }
})
const statusText = computed(() => {
  if (props.message._sending) return '提交中...'
  if (error.value || props.message._error) return error.value || props.message._error
  if (loading.value) return generation.value?.message || '生成中...'
  if (imageUrl.value) return '已完成'
  return '等待生成'
})

function download() {
  const url = imageUrl.value || rawImageUrl.value
  if (!url) return
  const link = document.createElement('a')
  link.href = url
  link.download = `image-${props.message.generation_id || props.message.id}.png`
  link.click()
}

async function copyPrompt() {
  await navigator.clipboard?.writeText(props.message.prompt)
}
</script>

<template>
  <div class="mr-auto flex max-w-[86%] items-start gap-3">
    <div class="mt-1 grid size-7 shrink-0 place-items-center rounded-full bg-white text-xs font-semibold text-teal ring-1 ring-slate-200">
      AI
    </div>
    <div class="min-w-0">
      <div class="mb-2 inline-flex items-center gap-1 text-[12.5px] text-slate-500">
        <span class="font-medium">ImageShow</span>
        <span>{{ statusText }}</span>
      </div>

      <div v-if="loading || message._sending" class="w-72 max-w-full rounded-2xl bg-slate-100 p-4">
        <div class="aspect-square animate-pulse rounded-xl bg-white/70"></div>
        <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-white">
          <div class="h-full rounded-full bg-teal transition-all duration-500" :style="{ width: `${progress}%` }"></div>
        </div>
      </div>
      <div v-else-if="error || message._error" class="rounded-2xl bg-red-50 px-4 py-3 text-[13px] text-red-700">
        {{ error || message._error }}
      </div>
      <div v-else-if="imageUrl" class="group relative w-fit">
        <img :src="imageUrl" class="max-h-[520px] max-w-full rounded-2xl object-contain" alt="generated image" />
        <button
          class="absolute bottom-3 right-3 grid size-9 place-items-center rounded-full bg-black/55 text-white opacity-90 backdrop-blur transition hover:bg-black/75"
          type="button"
          title="下载"
          @click="download"
        >
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v12m0 0 4-4m-4 4-4-4M4 21h16" />
          </svg>
        </button>
      </div>
      <div v-else class="aspect-square w-72 max-w-full rounded-2xl bg-slate-100"></div>

      <div v-if="imageUrl" class="mt-2 flex items-center gap-0.5 text-slate-500">
        <button class="grid size-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" type="button" title="复制提示词" @click="copyPrompt">
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 8h10v10H8zM6 16H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
          </svg>
        </button>
        <button class="grid size-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" type="button" title="下载" @click="download">
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v12m0 0 4-4m-4 4-4-4M4 21h16" />
          </svg>
        </button>
        <button class="grid size-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" type="button" title="收藏">
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m5 5 7-2 7 2v16l-7-3-7 3V5Z" />
          </svg>
        </button>
        <button class="grid size-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" type="button" title="更多">
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 12h.01M12 12h.01M18 12h.01" />
          </svg>
        </button>
        <span class="ml-2 text-[11px] text-slate-400">{{ message.size || 'image' }}</span>
      </div>
    </div>
  </div>
</template>
