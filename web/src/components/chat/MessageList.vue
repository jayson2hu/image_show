<script setup lang="ts">
import { computed } from 'vue'

import { useConversationStore } from '@/stores/conversation'

const conversationStore = useConversationStore()
const messages = computed(() => conversationStore.currentMessages)
</script>

<template>
  <div class="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
    <article v-for="message in messages" :key="message.id" class="ml-auto max-w-[80%] rounded-2xl bg-ink px-4 py-3 text-sm leading-6 text-white shadow-sm">
      {{ message.prompt }}
    </article>
    <article v-if="messages.length" class="mr-auto max-w-[80%] rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-600 shadow-sm">
      正在准备生成任务...
    </article>
  </div>
</template>
