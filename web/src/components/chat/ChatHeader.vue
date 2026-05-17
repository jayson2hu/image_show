<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { useConversationStore } from '@/stores/conversation'

const conversationStore = useConversationStore()
const title = computed(() => conversationStore.currentConversation?.title || '新对话')
const subtitle = computed(() => (conversationStore.currentConversation ? `${conversationStore.currentMessages.length} 条消息` : '未选择会话'))
</script>

<template>
  <header class="flex h-16 shrink-0 items-center justify-between border-b border-slate-200 bg-white/90 px-4 backdrop-blur">
    <div class="min-w-0">
      <h1 class="truncate text-base font-semibold text-ink">{{ title }}</h1>
      <p class="text-xs text-slate-500">{{ subtitle }}</p>
    </div>
    <RouterLink class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700 hover:border-teal hover:text-teal" to="/classic">
      经典模式
    </RouterLink>
  </header>
</template>
