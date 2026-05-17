<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import ChatEmptyState from '@/components/chat/ChatEmptyState.vue'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import Composer from '@/components/chat/Composer.vue'
import MessageList from '@/components/chat/MessageList.vue'
import SessionList from '@/components/chat/SessionList.vue'
import { useConversationStore } from '@/stores/conversation'

const conversationStore = useConversationStore()
const scrollContainerRef = ref<HTMLElement | null>(null)

const isEmpty = computed(() => {
  const conversation = conversationStore.currentConversation
  if (!conversation) return true
  const messages = conversationStore.messages[conversation.id]
  return !messages || messages.length === 0
})
const showEmptyState = computed(() => !conversationStore.loading && !conversationStore.messageLoading && isEmpty.value)

onMounted(() => {
  conversationStore.loadConversations()
})
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden bg-mist">
    <SessionList />
    <main class="flex min-w-0 flex-1 flex-col transition-all duration-200">
      <ChatHeader />
      <div ref="scrollContainerRef" class="flex-1 overflow-y-auto">
        <ChatEmptyState v-if="showEmptyState" />
        <div v-else-if="conversationStore.loading || conversationStore.messageLoading" class="flex h-full items-center justify-center text-sm text-slate-500">
          加载中...
        </div>
        <MessageList v-else :scroll-container="scrollContainerRef" />
      </div>
      <Composer class="shrink-0" />
    </main>
  </div>
</template>
