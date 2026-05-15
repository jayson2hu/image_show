<script setup lang="ts">
import { computed } from 'vue'

import { useConversationStore } from '@/stores/conversation'

const conversationStore = useConversationStore()
const conversations = computed(() => conversationStore.list)

function createConversation() {
  conversationStore.createLocalConversation()
}
</script>

<template>
  <aside class="hidden h-screen w-72 shrink-0 border-r border-slate-200 bg-white/90 p-3 lg:flex lg:flex-col">
    <button
      class="rounded-lg bg-ink px-3 py-2 text-left text-sm font-semibold text-white transition hover:bg-teal"
      type="button"
      @click="createConversation"
    >
      新建创作
    </button>

    <div class="mt-4 min-h-0 flex-1 space-y-1 overflow-y-auto">
      <button
        v-for="conversation in conversations"
        :key="conversation.id"
        class="w-full rounded-lg px-3 py-2 text-left text-sm transition"
        :class="conversationStore.currentId === conversation.id ? 'bg-mist text-ink' : 'text-slate-600 hover:bg-slate-50 hover:text-ink'"
        type="button"
        @click="conversationStore.selectConversation(conversation.id)"
      >
        <span class="block truncate font-medium">{{ conversation.title }}</span>
        <span class="mt-0.5 block text-xs text-slate-400">{{ conversation.msg_count }} 条消息</span>
      </button>
    </div>
  </aside>
</template>
