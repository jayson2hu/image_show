<script setup lang="ts">
import { computed } from 'vue'

import { useConversationStore } from '@/stores/conversation'
import { useUserStore } from '@/stores/user'

const conversationStore = useConversationStore()
const userStore = useUserStore()

const conversations = computed(() => conversationStore.list)
const displayName = computed(() => userStore.user?.username || userStore.user?.email?.split('@')[0] || '游客')
const creditText = computed(() => (userStore.user ? `${userStore.user.credits} 积分` : '体验积分 5'))
const avatarText = computed(() => displayName.value.trim().slice(0, 1).toUpperCase() || '游')

function createConversation() {
  conversationStore.createLocalConversation()
}

function firstChar(title: string) {
  return title.trim().slice(0, 1) || '新'
}
</script>

<template>
  <aside
    class="hidden h-screen shrink-0 border-r border-slate-200 bg-white/90 transition-all duration-200 lg:flex lg:flex-col"
    :class="conversationStore.sidebarCollapsed ? 'w-14 p-2' : 'w-72 p-3'"
  >
    <template v-if="conversationStore.sidebarCollapsed">
      <div class="flex flex-col items-center gap-2">
        <button
          class="flex size-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
          type="button"
          title="展开侧边栏"
          @click="conversationStore.toggleSidebar()"
        >
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 6 6 6-6 6" />
          </svg>
        </button>
        <button
          class="flex size-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
          type="button"
          title="新建创作"
          @click="createConversation"
        >
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14" />
          </svg>
        </button>
      </div>

      <div class="mt-4 flex min-h-0 flex-1 flex-col items-center gap-2 overflow-y-auto">
        <button
          v-for="conversation in conversations"
          :key="conversation.id"
          class="flex size-9 items-center justify-center rounded-full text-sm font-semibold transition"
          :class="conversationStore.currentId === conversation.id ? 'bg-mist text-ink' : 'bg-white text-slate-500 hover:bg-slate-50 hover:text-ink'"
          type="button"
          :title="conversation.title"
          @click="conversationStore.selectConversation(conversation.id)"
        >
          {{ firstChar(conversation.title) }}
        </button>
      </div>

      <div class="mt-3 flex justify-center border-t border-slate-200 pt-3">
        <div class="flex size-9 items-center justify-center rounded-full bg-ink text-sm font-semibold text-white" :title="displayName">
          {{ avatarText }}
        </div>
      </div>
    </template>

    <template v-else>
      <div class="flex items-center justify-between gap-2">
        <h2 class="text-sm font-semibold text-ink">对话列表</h2>
        <div class="flex items-center gap-2">
          <button
            class="flex size-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
            type="button"
            aria-label="新建创作"
            @click="createConversation"
          >
            <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14" />
            </svg>
          </button>
          <button
            class="flex size-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
            type="button"
            aria-label="收起侧边栏"
            @click="conversationStore.toggleSidebar()"
          >
            <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 6-6 6 6 6" />
            </svg>
          </button>
        </div>
      </div>

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

      <div class="mt-3 flex items-center gap-3 border-t border-slate-200 pt-3">
        <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-ink text-sm font-semibold text-white">
          {{ avatarText }}
        </div>
        <div class="min-w-0">
          <p class="truncate text-sm font-medium text-ink">{{ displayName }}</p>
          <p class="truncate text-xs text-slate-500">{{ creditText }}</p>
        </div>
      </div>
    </template>
  </aside>
</template>
