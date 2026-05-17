<script setup lang="ts">
import { computed, nextTick, onMounted, watch } from 'vue'

import { useConversationStore } from '@/stores/conversation'

import ImageReply from './ImageReply.vue'

const props = defineProps<{
  scrollContainer?: HTMLElement | null
}>()

const conversationStore = useConversationStore()
const messages = computed(() => conversationStore.currentMessages)

function scrollToBottom() {
  nextTick(() => {
    const container = props.scrollContainer
    if (!container) return
    container.scrollTop = container.scrollHeight
  })
}

onMounted(scrollToBottom)

watch(
  () => messages.value.map((message) => `${message.id}:${message.generation_id || ''}:${message._sending ? 'sending' : 'done'}`).join('|'),
  scrollToBottom,
)
</script>

<template>
  <div class="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
    <template v-for="message in messages" :key="message.id">
      <article class="ml-auto max-w-[80%] rounded-2xl bg-ink px-4 py-3 text-sm leading-6 text-white shadow-sm">
        <img v-if="message.attachment_url" class="mb-2 max-h-40 rounded-xl object-cover" :src="message.attachment_url" alt="attachment" />
        <p>{{ message.prompt }}</p>
        <div v-if="message._error" class="mt-2 text-[11px] text-red-200">{{ message._error }}</div>
      </article>
      <ImageReply :message="message" @image-load="scrollToBottom" />
    </template>
  </div>
</template>
