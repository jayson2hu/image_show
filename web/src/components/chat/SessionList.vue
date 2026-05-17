<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'

import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { useComposerStore } from '@/stores/composer'
import { useConversationStore } from '@/stores/conversation'
import { useUiStore } from '@/stores/ui'
import { useUserStore } from '@/stores/user'

const composerStore = useComposerStore()
const conversationStore = useConversationStore()
const uiStore = useUiStore()
const userStore = useUserStore()
const menuOpenId = ref<number | null>(null)
const editingId = ref<number | null>(null)
const editingTitle = ref('')
const pendingDeleteId = ref<number | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

const conversations = computed(() => conversationStore.list)
const displayName = computed(() => userStore.user?.username || userStore.user?.email?.split('@')[0] || '访客')
const creditText = computed(() => (userStore.user ? `${userStore.user.credits} 点数` : '访客点数 5'))
const avatarText = computed(() => displayName.value.trim().slice(0, 1).toUpperCase() || 'U')
const deleteTarget = computed(() => conversations.value.find((item) => item.id === pendingDeleteId.value) || null)

async function createConversation() {
  closeMenu()
  if (userStore.token) {
    await conversationStore.createLocalConversation()
  } else {
    conversationStore.resetGuestConversation()
  }
  composerStore.reset()
}

function firstChar(title: string) {
  return title.trim().slice(0, 1) || '新'
}

function toggleMenu(id: number) {
  menuOpenId.value = menuOpenId.value === id ? null : id
}

function closeMenu() {
  menuOpenId.value = null
}

function startRename(id: number, title: string) {
  editingId.value = id
  editingTitle.value = title.slice(0, 128)
  closeMenu()
  nextTick(() => inputRef.value?.focus())
}

async function confirmRename() {
  if (!editingId.value || !editingTitle.value.trim()) return
  await conversationStore.updateConversationTitle(editingId.value, editingTitle.value)
  cancelRename()
}

function cancelRename() {
  editingId.value = null
  editingTitle.value = ''
}

function askDelete(id: number) {
  pendingDeleteId.value = id
  closeMenu()
}

async function confirmDelete() {
  if (pendingDeleteId.value) {
    await conversationStore.deleteLocalConversation(pendingDeleteId.value)
  }
  pendingDeleteId.value = null
}

function handleDocumentClick() {
  closeMenu()
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
})

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick)
})
</script>

<template>
  <aside
    class="hidden h-screen shrink-0 border-r border-slate-100 bg-white transition-all duration-200 lg:flex lg:flex-col"
    :class="uiStore.sidebarCollapsed ? 'w-14 p-2' : 'w-72 p-3'"
  >
    <template v-if="uiStore.sidebarCollapsed">
      <div class="flex flex-col items-center gap-2">
        <button
          class="flex size-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
          type="button"
          title="展开侧栏"
          @click="uiStore.toggleSidebar()"
        >
          <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 6 6 6-6 6" />
          </svg>
        </button>
        <button
          class="flex size-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
          type="button"
          title="新建对话"
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
          class="flex size-8 items-center justify-center rounded-lg text-sm font-semibold transition"
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
            aria-label="新建对话"
            @click="createConversation"
          >
            <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14" />
            </svg>
          </button>
          <button
            class="flex size-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition hover:border-teal hover:bg-teal/5 hover:text-teal"
            type="button"
            aria-label="收起侧栏"
            @click="uiStore.toggleSidebar()"
          >
            <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m15 6-6 6 6 6" />
            </svg>
          </button>
        </div>
      </div>

      <div class="mt-4 min-h-0 flex-1 space-y-1 overflow-y-auto">
        <div
          v-for="conversation in conversations"
          :key="conversation.id"
          class="group relative rounded-lg transition"
          :class="conversationStore.currentId === conversation.id ? 'bg-mist text-ink' : 'text-slate-600 hover:bg-slate-50 hover:text-ink'"
          @click.stop
        >
          <div v-if="editingId === conversation.id" class="px-2 py-2">
            <input
              ref="inputRef"
              v-model="editingTitle"
              class="w-full rounded-lg border-2 border-teal bg-white px-2 py-1.5 text-sm text-ink outline-none"
              maxlength="128"
              @keydown.enter.prevent="confirmRename"
              @keydown.esc.prevent="cancelRename"
            />
            <div class="mt-2 flex gap-2">
              <button
                class="rounded-md bg-teal px-2 py-1 text-xs font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
                type="button"
                :disabled="!editingTitle.trim()"
                @click="confirmRename"
              >
                保存
              </button>
              <button class="rounded-md border border-slate-200 px-2 py-1 text-xs font-medium text-slate-500 hover:bg-white" type="button" @click="cancelRename">取消</button>
            </div>
          </div>

          <button v-else class="w-full px-3 py-2 pr-9 text-left text-sm" type="button" @click="conversationStore.selectConversation(conversation.id)">
            <span class="block truncate font-medium">{{ conversation.title }}</span>
            <span class="mt-0.5 block text-xs text-slate-400">{{ conversation.msg_count }} 条消息</span>
          </button>

          <button
            v-if="editingId !== conversation.id"
            class="absolute right-2 top-2 flex size-7 items-center justify-center rounded-md text-slate-400 opacity-0 transition hover:bg-white hover:text-ink group-hover:opacity-100"
            :class="menuOpenId === conversation.id || conversationStore.currentId === conversation.id ? 'opacity-100' : ''"
            type="button"
            aria-label="更多"
            @click.stop="toggleMenu(conversation.id)"
          >
            <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 12h.01M12 12h.01M18 12h.01" />
            </svg>
          </button>

          <div
            v-if="menuOpenId === conversation.id"
            class="absolute right-2 top-9 z-20 w-28 rounded-lg border border-slate-200 bg-white py-1 text-sm shadow-lg"
          >
            <button class="block w-full px-3 py-2 text-left text-slate-700 hover:bg-slate-50" type="button" @click="startRename(conversation.id, conversation.title)">重命名</button>
            <button class="block w-full px-3 py-2 text-left text-red-600 hover:bg-red-50" type="button" @click="askDelete(conversation.id)">删除</button>
          </div>
        </div>
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

  <ConfirmDialog
    :open="pendingDeleteId !== null"
    title="删除对话"
    :message="`确定删除「${deleteTarget?.title || '未命名对话'}」吗？删除后不会再显示该会话。`"
    confirm-text="删除"
    confirm-color="red"
    @cancel="pendingDeleteId = null"
    @confirm="confirmDelete"
  />
</template>
