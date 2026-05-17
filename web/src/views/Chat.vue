<script setup lang="ts">
import { computed, onMounted } from 'vue'

import ChatEmptyState from '@/components/chat/ChatEmptyState.vue'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import Composer from '@/components/chat/Composer.vue'
import MessageList from '@/components/chat/MessageList.vue'
import SessionList from '@/components/chat/SessionList.vue'
import { useConversationStore } from '@/stores/conversation'

const conversationStore = useConversationStore()

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
      <ChatEmptyState v-if="showEmptyState" class="flex-1" />
      <div v-else-if="conversationStore.loading || conversationStore.messageLoading" class="flex flex-1 items-center justify-center text-sm text-slate-500">
        加载中...
      </div>
      <template v-else>
        <div class="flex-1 overflow-y-auto">
          <MessageList />
        </div>
        <Composer />
      </template>
    </main>
  </div>
</template>
