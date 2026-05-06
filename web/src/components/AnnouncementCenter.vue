<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import api from '@/api'
import { useUserStore } from '@/stores/user'

interface Announcement {
  id: number
  title: string
  content: string
  status: number
  notify_mode: 'silent' | 'popup'
  starts_at?: string | null
  ends_at?: string | null
  created_at: string
  updated_at: string
  read_at?: string | null
}

const userStore = useUserStore()
const announcements = ref<Announcement[]>([])
const loading = ref(false)
const isListOpen = ref(false)
const popupAnnouncement = ref<Announcement | null>(null)
const shownPopupKey = 'announcement_popup_shown_ids'

const unreadCount = computed(() => announcements.value.filter((item) => !item.read_at).length)

onMounted(loadAnnouncements)

async function loadAnnouncements() {
  loading.value = true
  try {
    const response = await api.get('/announcements')
    announcements.value = Array.isArray(response.data.items) ? response.data.items : []
    queuePopup()
  } catch {
    announcements.value = []
  } finally {
    loading.value = false
  }
}

function queuePopup() {
  const shown = getShownPopupIds()
  const next = announcements.value.find((item) => item.notify_mode === 'popup' && !item.read_at && !shown.has(item.id))
  if (!next) {
    return
  }
  popupAnnouncement.value = next
  shown.add(next.id)
  localStorage.setItem(shownPopupKey, JSON.stringify([...shown]))
}

function getShownPopupIds() {
  try {
    const raw = localStorage.getItem(shownPopupKey)
    const ids = raw ? JSON.parse(raw) : []
    return new Set<number>(Array.isArray(ids) ? ids.map((id) => Number(id)).filter(Number.isFinite) : [])
  } catch {
    return new Set<number>()
  }
}

async function markRead(item: Announcement) {
  if (item.read_at) {
    return
  }
  item.read_at = new Date().toISOString()
  if (!userStore.user) {
    return
  }
  try {
    await api.post(`/announcements/${item.id}/read`)
  } catch {
    item.read_at = null
  }
}

async function markAllRead() {
  await Promise.all(announcements.value.filter((item) => !item.read_at).map(markRead))
}

async function dismissPopup() {
  const item = popupAnnouncement.value
  popupAnnouncement.value = null
  if (item) {
    await markRead(item)
  }
  queuePopup()
}

function openList() {
  isListOpen.value = true
  loadAnnouncements()
}

function fmtTime(value?: string | null) {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

<template>
  <div class="relative">
    <button
      class="relative flex size-10 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 hover:text-blue-700"
      type="button"
      aria-label="公告"
      @click="openList"
    >
      <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.4-1.4A2 2 0 0 1 18 14.2V11a6 6 0 1 0-12 0v3.2c0 .5-.2 1-.6 1.4L4 17h11Zm0 0a3 3 0 0 1-6 0" />
      </svg>
      <span v-if="unreadCount" class="absolute right-1 top-1 flex size-2">
        <span class="absolute inline-flex size-full animate-ping rounded-full bg-red-500 opacity-75"></span>
        <span class="relative inline-flex size-2 rounded-full bg-red-500"></span>
      </span>
    </button>

    <Teleport to="body">
      <div v-if="isListOpen" class="fixed inset-0 z-[100] flex items-start justify-center overflow-y-auto bg-black/55 p-4 pt-[8vh] backdrop-blur-sm" @click="isListOpen = false">
        <section class="w-full max-w-2xl overflow-hidden rounded-3xl bg-white shadow-2xl" @click.stop>
          <header class="relative overflow-hidden border-b border-blue-100 bg-gradient-to-br from-blue-50 to-indigo-50 px-6 py-5">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="flex items-center gap-2">
                  <span class="flex size-9 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/25">
                    <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.4-1.4A2 2 0 0 1 18 14.2V11a6 6 0 1 0-12 0v3.2c0 .5-.2 1-.6 1.4L4 17h11Z" /></svg>
                  </span>
                  <h2 class="text-lg font-semibold text-slate-950">公告中心</h2>
                </div>
                <p class="mt-2 text-sm text-slate-600">当前有 {{ unreadCount }} 条未读公告</p>
              </div>
              <div class="flex gap-2">
                <button v-if="unreadCount" class="rounded-xl bg-blue-600 px-3 py-2 text-xs font-medium text-white transition hover:bg-blue-700" type="button" @click="markAllRead">全部已读</button>
                <button class="admin-close-btn" type="button" @click="isListOpen = false">关闭</button>
              </div>
            </div>
          </header>
          <div class="max-h-[64vh] overflow-y-auto">
            <div v-if="loading" class="p-10 text-center text-sm text-slate-500">加载中...</div>
            <div v-else-if="announcements.length">
              <article
                v-for="item in announcements"
                :key="item.id"
                class="group relative border-b border-slate-100 px-6 py-4 transition hover:bg-slate-50"
                :class="!item.read_at ? 'bg-blue-50/40' : ''"
              >
                <div class="flex items-start gap-4">
                  <div class="mt-1 flex size-10 shrink-0 items-center justify-center rounded-xl" :class="item.read_at ? 'bg-slate-100 text-slate-400' : 'bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/20'">
                    <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path v-if="item.read_at" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m5 13 4 4L19 7" /><path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z" /></svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <h3 class="font-semibold text-slate-950">{{ item.title }}</h3>
                      <span v-if="!item.read_at" class="rounded-md bg-blue-100 px-1.5 py-0.5 text-xs font-medium text-blue-700">未读</span>
                      <span v-if="item.notify_mode === 'popup'" class="rounded-md bg-amber-100 px-1.5 py-0.5 text-xs font-medium text-amber-700">弹窗</span>
                    </div>
                    <p class="mt-1 text-xs text-slate-400">{{ fmtTime(item.created_at) }}</p>
                    <p class="mt-2 whitespace-pre-line text-sm leading-6 text-slate-600">{{ item.content }}</p>
                    <button v-if="!item.read_at" class="mt-3 text-sm font-medium text-blue-700 hover:text-blue-900" type="button" @click="markRead(item)">标记已读</button>
                  </div>
                </div>
              </article>
            </div>
            <div v-else class="p-12 text-center text-sm text-slate-500">暂无公告</div>
          </div>
        </section>
      </div>

      <div v-if="popupAnnouncement" class="fixed inset-0 z-[110] flex items-start justify-center overflow-y-auto bg-black/60 p-4 pt-[8vh] backdrop-blur-sm">
        <section class="w-full max-w-2xl overflow-hidden rounded-3xl bg-white shadow-2xl">
          <header class="relative overflow-hidden border-b border-amber-100 bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50 px-7 py-6">
            <div class="flex items-center gap-3">
              <span class="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-amber-500 to-orange-600 text-white shadow-lg shadow-amber-500/25">
                <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.4-1.4A2 2 0 0 1 18 14.2V11a6 6 0 1 0-12 0v3.2c0 .5-.2 1-.6 1.4L4 17h11Z" /></svg>
              </span>
              <div>
                <span class="rounded-lg bg-amber-100 px-2 py-1 text-xs font-medium text-amber-700">新公告</span>
                <h2 class="mt-2 text-2xl font-bold text-slate-950">{{ popupAnnouncement.title }}</h2>
                <p class="mt-1 text-sm text-slate-500">{{ fmtTime(popupAnnouncement.created_at) }}</p>
              </div>
            </div>
          </header>
          <div class="max-h-[50vh] overflow-y-auto px-7 py-6">
            <div class="border-l-4 border-amber-500 pl-5">
              <p class="whitespace-pre-line text-sm leading-7 text-slate-700">{{ popupAnnouncement.content }}</p>
            </div>
          </div>
          <footer class="border-t border-slate-100 bg-slate-50 px-7 py-4 text-right">
            <button class="rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 px-5 py-2.5 text-sm font-medium text-white shadow-lg shadow-amber-500/25 transition hover:shadow-xl" type="button" @click="dismissPopup">
              我知道了
            </button>
          </footer>
        </section>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.admin-close-btn {
  @apply rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-medium text-slate-600 transition hover:bg-slate-50;
}
</style>
