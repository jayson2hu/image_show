<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'

import { useComposerStore } from '@/stores/composer'
import { useConversationStore } from '@/stores/conversation'

const props = withDefaults(defineProps<{ compact?: boolean }>(), {
  compact: false,
})

const composerStore = useComposerStore()
const conversationStore = useConversationStore()
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const ratioLabel = computed(() => {
  const map: Record<string, string> = {
    square: '1:1 方图',
    portrait_3_4: '3:4 竖版',
    story: '9:16 竖屏',
    landscape_4_3: '4:3 横版',
    widescreen_16_9: '16:9 宽屏',
  }
  return map[composerStore.draft.size] || composerStore.draft.size
})
const creditEstimate = computed(() => {
  const base = composerStore.draft.size === 'square' ? 1 : 2
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

function send() {
  const prompt = composerStore.draft.prompt.trim()
  if (!prompt) {
    composerStore.focusInput()
    return
  }
  conversationStore.sendLocalMessage(prompt)
  composerStore.reset()
}
</script>

<template>
  <div :class="props.compact ? 'w-full' : 'border-t border-slate-200 bg-white/90 p-4'">
    <div v-if="!props.compact" class="mx-auto mb-2 flex max-w-3xl items-center gap-2 overflow-x-auto pb-1 text-xs text-slate-500">
      <span class="rounded-full bg-mist px-3 py-1">场景</span>
      <span class="rounded-full bg-mist px-3 py-1">推荐样例</span>
    </div>

    <div class="mx-auto max-w-3xl rounded-2xl border border-slate-200 bg-white p-2 shadow-sm">
      <textarea
        ref="textareaRef"
        v-model="composerStore.draft.prompt"
        class="min-h-24 w-full resize-none rounded-xl border-0 bg-transparent px-3 py-2 text-sm leading-6 text-ink outline-none placeholder:text-slate-400"
        placeholder="描述你想生成的画面..."
        @keydown.enter.exact.prevent="send"
      ></textarea>

      <div class="flex flex-wrap items-center gap-2 px-2 pb-1">
        <button class="rounded-full bg-mist px-3 py-1.5 text-xs font-medium text-ink" type="button">风格</button>
        <button class="rounded-full bg-mist px-3 py-1.5 text-xs font-medium text-ink" type="button">{{ ratioLabel }}</button>
        <button
          class="rounded-full px-3 py-1.5 text-xs font-medium"
          :class="composerStore.draft.layered ? 'bg-coral text-white' : 'bg-mist text-ink'"
          type="button"
          @click="composerStore.setDraft({ layered: !composerStore.draft.layered })"
        >
          分层{{ composerStore.draft.layered ? ` · ${composerStore.draft.layer_count} 层` : '' }}
        </button>
        <button class="rounded-full bg-mist px-3 py-1.5 text-xs font-medium text-ink" type="button">质量</button>
        <button class="rounded-full bg-mist px-3 py-1.5 text-xs font-medium text-ink" type="button">附图</button>
        <span class="ml-auto text-xs font-medium text-slate-500">本次 {{ creditEstimate }} 积分</span>
        <button class="flex size-9 items-center justify-center rounded-full bg-teal text-white transition hover:bg-ink disabled:cursor-not-allowed disabled:bg-slate-300" type="button" :disabled="!composerStore.draft.prompt.trim()" @click="send">
          ↑
        </button>
      </div>
    </div>

    <div v-if="!props.compact" class="mx-auto mt-2 max-w-3xl text-center text-xs text-slate-400">
      Enter 发送 · Shift+Enter 换行
    </div>
  </div>
</template>
