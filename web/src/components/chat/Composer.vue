<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import api from '@/api'
import { useToast } from '@/composables/useToast'
import { useComposerStore } from '@/stores/composer'
import { useConversationStore } from '@/stores/conversation'

type Option = {
  value: string
  label: string
  ratio?: string
  credit_cost?: number
}

const fallbackStyles: Option[] = [
  { value: 'realistic', label: '写实' },
  { value: 'anime', label: '动漫' },
  { value: 'fantasy', label: '奇幻' },
  { value: 'cyberpunk', label: '赛博朋克' },
  { value: 'watercolor', label: '水彩' },
  { value: 'illustration', label: '插画' },
]

const fallbackRatios: Option[] = [
  { value: 'square', label: '1:1 方图', ratio: '1:1', credit_cost: 1 },
  { value: 'portrait_3_4', label: '3:4 竖图', ratio: '3:4', credit_cost: 2 },
  { value: 'story', label: '9:16 长图', ratio: '9:16', credit_cost: 2 },
  { value: 'landscape_4_3', label: '4:3 横图', ratio: '4:3', credit_cost: 2 },
  { value: 'widescreen', label: '16:9 宽屏', ratio: '16:9', credit_cost: 2 },
]

const composerStore = useComposerStore()
const conversationStore = useConversationStore()
const toast = useToast()
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const openMenu = ref<'style' | 'ratio' | null>(null)
const showUploader = ref(false)
const previewUrl = ref('')
const styles = ref<Option[]>(fallbackStyles)
const ratios = ref<Option[]>(fallbackRatios)

const styleOptions = computed<Option[]>(() => [{ value: '', label: '默认风格' }, ...styles.value])
const selectedStyle = computed(() => styleOptions.value.find((item) => item.value === composerStore.draft.style_id) || styleOptions.value[0])
const selectedRatio = computed(() => ratios.value.find((item) => item.value === composerStore.draft.size) || fallbackRatios[0])
const hasAttachment = computed(() => !!composerStore.draft.attachment)
const creditEstimate = computed(() => {
  const base = selectedRatio.value.credit_cost ?? (composerStore.draft.size === 'square' ? 1 : 2)
  return base + (composerStore.draft.layered ? composerStore.draft.layer_count : 0)
})

watch(
  () => composerStore.focusTick,
  () => {
    nextTick(() => {
      const textarea = textareaRef.value
      if (!textarea) return
      textarea.focus()
      textarea.setSelectionRange(textarea.value.length, textarea.value.length)
    })
  },
)

onMounted(() => {
  document.addEventListener('click', closeDropdown)
  loadOptions()
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown)
  revokePreview()
})

async function loadOptions() {
  try {
    const response = await api.get('/generation/options')
    if (Array.isArray(response.data.size_options) && response.data.size_options.length > 0) {
      ratios.value = response.data.size_options.map((item: Option) => ({
        value: item.value,
        label: `${item.ratio || item.label} ${item.label}`,
        ratio: item.ratio,
        credit_cost: item.credit_cost,
      }))
    }
  } catch {
    ratios.value = fallbackRatios
  }

  try {
    const response = await api.get('/prompt-templates')
    const items = Array.isArray(response.data.items) ? response.data.items : []
    const styleItems = items.filter((item: any) => item.category === 'style')
    if (styleItems.length > 0) {
      styles.value = styleItems.map((item: any) => ({
        value: String(item.id || item.label),
        label: item.label,
      }))
    }
  } catch {
    styles.value = fallbackStyles
  }
}

function closeDropdown() {
  openMenu.value = null
}

function toggleDropdown(menu: 'style' | 'ratio') {
  openMenu.value = openMenu.value === menu ? null : menu
}

function selectStyle(value: string) {
  composerStore.setDraft({ style_id: value })
  closeDropdown()
}

function selectRatio(value: string) {
  composerStore.setDraft({ size: value })
  closeDropdown()
}

function toggleLayered() {
  const next = !composerStore.draft.layered
  composerStore.setDraft({ layered: next, layer_count: composerStore.draft.layer_count || 5 })
}

function triggerFilePicker() {
  fileInputRef.value?.click()
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) setAttachment(file)
  input.value = ''
}

function handleDrop(event: DragEvent) {
  const file = event.dataTransfer?.files?.[0]
  if (file) setAttachment(file)
}

function setAttachment(file: File) {
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
    toast.error('仅支持 JPG、PNG、WebP 图片')
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    toast.error('图片不能超过 10MB')
    return
  }
  revokePreview()
  composerStore.setDraft({ attachment: file })
  previewUrl.value = URL.createObjectURL(file)
  showUploader.value = true
}

function removeAttachment() {
  revokePreview()
  composerStore.setDraft({ attachment: null })
}

function revokePreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

async function send() {
  const prompt = composerStore.draft.prompt.trim()
  if (!prompt) {
    composerStore.focusInput()
    return
  }
  try {
    await conversationStore.sendMessage({
      prompt,
      size: composerStore.draft.size,
      style_id: composerStore.draft.style_id,
      scene_id: composerStore.draft.scene_id,
      layered: composerStore.draft.layered,
      layer_count: composerStore.draft.layer_count,
      attachment: composerStore.draft.attachment,
    })
    revokePreview()
    showUploader.value = false
    composerStore.reset()
  } catch (error: any) {
    toast.error(error.response?.data?.message || error.response?.data?.error || '发送失败')
  }
}
</script>

<template>
  <div class="border-t border-slate-200 bg-white/90 p-4">
    <div class="mx-auto mb-2 flex max-w-3xl items-center gap-2 overflow-x-auto pb-1 text-xs text-slate-500">
      <span class="rounded-full bg-mist px-3 py-1">文字生成</span>
      <span class="rounded-full bg-mist px-3 py-1">图片参考</span>
    </div>

    <div class="mx-auto max-w-3xl rounded-2xl border border-slate-200 bg-white p-2 shadow-sm">
      <textarea
        ref="textareaRef"
        v-model="composerStore.draft.prompt"
        class="min-h-24 w-full resize-none rounded-xl border-0 bg-transparent px-3 py-2 text-sm leading-6 text-ink outline-none placeholder:text-slate-400"
        placeholder="描述你想生成或修改的图片..."
        @keydown.enter.exact.prevent="send"
      ></textarea>

      <div
        v-if="showUploader"
        class="mx-2 mb-2 rounded-xl border border-dashed border-slate-300 bg-slate-50 p-3"
        @dragover.prevent
        @drop.prevent="handleDrop"
      >
        <input ref="fileInputRef" class="sr-only" type="file" accept="image/jpeg,image/png,image/webp" @change="handleFileChange" />
        <div v-if="previewUrl" class="relative size-20">
          <img class="size-20 rounded-lg object-cover" :src="previewUrl" alt="参考图预览" />
          <button class="absolute -right-2 -top-2 flex size-6 items-center justify-center rounded-full bg-ink text-xs text-white" type="button" @click="removeAttachment">x</button>
        </div>
        <button v-else class="flex w-full flex-col items-center justify-center rounded-lg px-4 py-6 text-center text-sm text-slate-500 hover:bg-white" type="button" @click="triggerFilePicker">
          <svg class="mb-2 size-6 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 16V4m0 0 4 4m-4-4-4 4M4 20h16" />
          </svg>
          点击上传参考图片
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2 px-2 pb-1" @click.stop>
        <div class="relative">
          <button class="rounded-full bg-mist px-3 py-1.5 text-xs font-medium text-ink" type="button" @click="toggleDropdown('style')">
            风格: {{ selectedStyle.label }}
          </button>
          <div v-if="openMenu === 'style'" class="absolute bottom-full left-0 z-20 mb-2 w-44 rounded-lg border border-slate-200 bg-white py-1 text-sm shadow-lg">
            <button
              v-for="option in styleOptions"
              :key="option.value || 'none'"
              class="flex w-full items-center justify-between px-3 py-2 text-left text-slate-700 hover:bg-slate-50"
              type="button"
              @click="selectStyle(option.value)"
            >
              <span>{{ option.label }}</span>
              <span v-if="composerStore.draft.style_id === option.value">✓</span>
            </button>
          </div>
        </div>

        <div class="relative">
          <button class="rounded-full bg-mist px-3 py-1.5 text-xs font-medium text-ink" type="button" @click="toggleDropdown('ratio')">
            {{ selectedRatio.label }}
          </button>
          <div v-if="openMenu === 'ratio'" class="absolute bottom-full left-0 z-20 mb-2 w-40 rounded-lg border border-slate-200 bg-white py-1 text-sm shadow-lg">
            <button
              v-for="option in ratios"
              :key="option.value"
              class="flex w-full items-center justify-between px-3 py-2 text-left text-slate-700 hover:bg-slate-50"
              type="button"
              @click="selectRatio(option.value)"
            >
              <span>{{ option.label }}</span>
              <span v-if="composerStore.draft.size === option.value">✓</span>
            </button>
          </div>
        </div>

        <button
          class="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-medium transition"
          :class="composerStore.draft.layered ? 'bg-coral text-white' : 'bg-mist text-ink'"
          type="button"
          @click="toggleLayered"
        >
          <span>分层 {{ composerStore.draft.layered ? '开' : '关' }}</span>
          <span class="flex h-4 w-7 items-center rounded-full bg-white/60 p-0.5">
            <span class="size-3 rounded-full bg-current transition" :class="composerStore.draft.layered ? 'translate-x-3' : 'translate-x-0'"></span>
          </span>
        </button>

        <button
          class="rounded-full px-3 py-1.5 text-xs font-medium transition"
          :class="hasAttachment ? 'border border-teal/30 bg-teal/10 text-teal' : 'bg-mist text-ink'"
          type="button"
          @click="showUploader = !showUploader"
        >
          参考图{{ hasAttachment ? ' 已选' : '' }}
        </button>
        <span class="ml-auto text-xs font-medium text-slate-500">预计 {{ creditEstimate }} 点数</span>
        <button class="flex size-9 items-center justify-center rounded-full bg-teal text-white transition hover:bg-ink disabled:cursor-not-allowed disabled:bg-slate-300" type="button" :disabled="!composerStore.draft.prompt.trim() || conversationStore.sending" @click="send">
          <svg v-if="conversationStore.sending" class="size-4 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 0 1 8-8v4a4 4 0 0 0-4 4H4Z" />
          </svg>
          <span v-else>发送</span>
        </button>
      </div>
    </div>

    <div class="mx-auto mt-2 max-w-3xl text-center text-xs text-slate-400">
      Enter 发送 · Shift+Enter 换行
    </div>
  </div>
</template>
