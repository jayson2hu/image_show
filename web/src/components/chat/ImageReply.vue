<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

import type { Message } from '@/api/types'
import { useGenerationPoll } from '@/composables/useGenerationPoll'
import { useProtectedImageURL } from '@/composables/useProtectedImageURL'
import { useToast } from '@/composables/useToast'
import { useComposerStore } from '@/stores/composer'

const props = defineProps<{ message: Message }>()
const emit = defineEmits<{ imageLoad: [] }>()

const toast = useToast()
const composerStore = useComposerStore()
const showMore = ref(false)
const previewOpen = ref(false)
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
const sizeLabel = computed(() => formatSizeLabel(props.message.size))
const imageSizeClass = computed(() => {
  switch (props.message.size) {
    case 'square':
    case '1024x1024':
      return 'max-w-[300px] max-h-[300px]'
    case 'portrait_3_4':
    case 'portrait':
    case '1152x1536':
    case 'story':
    case '1008x1792':
      return 'max-h-[400px] max-w-full'
    case 'landscape_4_3':
    case 'landscape':
    case '1536x1152':
    case 'widescreen':
    case '1792x1008':
      return 'max-w-[500px] max-h-full'
    default:
      return 'max-w-[400px] max-h-[400px]'
  }
})
const layeredLabel = computed(() => (props.message.layered ? `${props.message.layer_count || 0} 层` : '关闭'))
const taskId = computed(() => props.message.generation_id || props.message.id)

function formatSizeLabel(size?: string) {
  const labels: Record<string, string> = {
    square: '1:1 方图',
    portrait: '3:4 竖图',
    portrait_3_4: '3:4 竖图',
    story: '9:16 长图',
    landscape: '4:3 横图',
    landscape_4_3: '4:3 横图',
    widescreen: '16:9 宽屏',
    '1024x1024': '1:1 方图',
    '1152x1536': '3:4 竖图',
    '1008x1792': '9:16 长图',
    '1536x1152': '4:3 横图',
    '1792x1008': '16:9 宽屏',
  }
  if (!size) return '图片'
  return labels[size] || size.replace('x', ' x ')
}

function download() {
  const url = imageUrl.value || rawImageUrl.value
  if (!url) return
  const link = document.createElement('a')
  link.href = url
  link.download = `image-${taskId.value}.png`
  link.click()
}

function editImage() {
  composerStore.setDraft({
    prompt: props.message.prompt,
    size: props.message.size || composerStore.draft.size,
    style_id: props.message.style_id || '',
    scene_id: props.message.scene_id || '',
    layered: Boolean(props.message.layered),
    layer_count: props.message.layer_count || 5,
  })
  composerStore.focusInput()
  toast.info('已填入提示词')
}

async function copyPrompt() {
  await navigator.clipboard?.writeText(props.message.prompt)
  toast.success('提示词已复制')
}

async function copyImageUrl() {
  const url = imageUrl.value || rawImageUrl.value
  if (!url) return
  await navigator.clipboard?.writeText(url)
  toast.success('图片链接已复制')
  showMore.value = false
}

function saveImage() {
  toast.info('收藏功能暂未开放')
  showMore.value = false
}

function closeMenus() {
  showMore.value = false
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    previewOpen.value = false
    closeMenus()
  }
}

onMounted(() => {
  document.addEventListener('click', closeMenus)
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', closeMenus)
  document.removeEventListener('keydown', handleKeydown)
})
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
        <img
          :src="imageUrl"
          class="cursor-pointer rounded-2xl object-contain"
          :class="imageSizeClass"
          alt="generated image"
          @click="previewOpen = true"
          @load="emit('imageLoad')"
        />
      </div>
      <div v-else class="aspect-square w-72 max-w-full rounded-2xl bg-slate-100"></div>

      <div v-if="imageUrl" class="mt-2 flex items-center gap-0.5 text-slate-500">
        <button class="grid size-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" type="button" title="编辑" @click="editImage">
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L8.582 18.07 4 19l.93-4.582L16.862 4.487Z" />
          </svg>
        </button>
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
        <div class="relative" @click.stop>
          <button class="grid size-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" type="button" title="更多" @click="showMore = !showMore">
            <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 12h.01M12 12h.01M18 12h.01" />
            </svg>
          </button>
          <div v-if="showMore" class="absolute bottom-9 left-0 z-20 w-44 rounded-xl border border-slate-200 bg-white py-1 text-xs shadow-lg">
            <button class="flex w-full items-center gap-2 px-3 py-2 text-left text-slate-700 hover:bg-slate-50" type="button" @click="saveImage">
              <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m5 5 7-2 7 2v16l-7-3-7 3V5Z" />
              </svg>
              <span>收藏</span>
            </button>
            <button class="flex w-full items-center gap-2 px-3 py-2 text-left text-slate-700 hover:bg-slate-50" type="button" @click="copyImageUrl">
              <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 6H15a3 3 0 0 1 0 6h-1.5M10.5 18H9a3 3 0 0 1 0-6h1.5M8 12h8" />
              </svg>
              <span>复制图片链接</span>
            </button>
            <div class="my-1 border-t border-slate-100"></div>
            <div class="px-3 py-1.5 text-[11px] leading-5 text-slate-400">
              <p>尺寸：{{ sizeLabel }}</p>
              <p>分层：{{ layeredLabel }}</p>
              <p class="truncate">任务 ID：{{ taskId }}</p>
            </div>
          </div>
        </div>
        <span class="ml-2 text-[11px] text-slate-400">{{ sizeLabel }}</span>
      </div>
    </div>
  </div>

  <Teleport to="body">
    <div v-if="previewOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4" @click.self="previewOpen = false">
      <button
        class="absolute right-4 top-4 grid size-10 place-items-center rounded-full bg-white/15 text-white backdrop-blur transition hover:bg-white/25"
        type="button"
        aria-label="关闭预览"
        @click="previewOpen = false"
      >
        <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 6l12 12M18 6 6 18" />
        </svg>
      </button>
      <img :src="imageUrl" class="max-h-[92vh] max-w-[92vw] object-contain" alt="generated image preview" />
    </div>
  </Teleport>
</template>
