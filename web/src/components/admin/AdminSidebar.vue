<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { useUserStore } from '@/stores/user'

const props = defineProps<{
  activeTab: string
}>()

const emit = defineEmits<{
  'update:activeTab': [tab: string]
}>()

const userStore = useUserStore()

const tabs = [
  { id: 'overview', label: '概览', description: '核心指标与运行状态', icon: 'M3 13.5h6v7H3v-7Zm12-10h6v17h-6v-17ZM9 8.5h6v12H9v-12Z' },
  { id: 'users', label: '用户', description: '账号、角色与充值', icon: 'M16 11a4 4 0 1 0-8 0m8 0a4 4 0 1 1-8 0m8 0c2.5.8 4 2.2 4 4v1H4v-1c0-1.8 1.5-3.2 4-4' },
  { id: 'channels', label: '渠道', description: 'API 渠道配置与测试', icon: 'M4 6h16M4 12h16M4 18h16M7 6v.01M7 12v.01M7 18v.01' },
  { id: 'templates', label: '模板', description: '提示词模板管理', icon: 'M7 3h7l5 5v13H7V3Zm7 0v5h5M10 13h6M10 17h6' },
  { id: 'settings', label: '设置', description: '系统开关和配置', icon: 'M12 8a4 4 0 1 1 0 8 4 4 0 0 1 0-8Zm0-5v3m0 12v3M4.6 4.6l2.1 2.1m10.6 10.6 2.1 2.1M3 12h3m12 0h3M4.6 19.4l2.1-2.1M17.3 6.7l2.1-2.1' },
  { id: 'announcements', label: '公告', description: '前台通知和弹窗', icon: 'M5 8h14v9H8l-3 3V8Zm3-3h8' },
  { id: 'credits', label: '积分', description: '积分流水审计', icon: 'M4 7h16v10H4V7Zm0 3h16M8 15h3' },
  { id: 'monitor', label: '监控', description: '每日指标和告警', icon: 'M15 17h5l-1.4-1.4A2 2 0 0 1 18 14.2V11a6 6 0 1 0-12 0v3.2a2 2 0 0 1-.6 1.4L4 17h5m6 0a3 3 0 0 1-6 0' },
]

const adminEmail = computed(() => userStore.user?.email || '管理员')

function selectTab(tab: string) {
  emit('update:activeTab', tab)
}
</script>

<template>
  <aside class="border-b border-slate-200 bg-white lg:min-h-[calc(100vh-65px)] lg:border-b-0 lg:border-r">
    <div class="hidden h-full flex-col lg:flex">
      <div class="border-b border-slate-200 p-5">
        <p class="text-xs font-semibold uppercase tracking-wide text-teal">Console</p>
        <h1 class="mt-2 text-xl font-semibold text-slate-950">管理后台</h1>
        <p class="mt-1 text-sm text-slate-500">运营、配置和监控工作台</p>
      </div>
      <nav class="flex-1 space-y-1 p-3">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="flex w-full items-center gap-3 rounded-2xl px-3 py-3 text-left transition"
          :class="props.activeTab === tab.id ? 'bg-slate-950 text-white shadow-lg shadow-slate-950/15' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-950'"
          type="button"
          @click="selectTab(tab.id)"
        >
          <svg class="size-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" :d="tab.icon" />
          </svg>
          <span class="min-w-0">
            <span class="block text-sm font-semibold">{{ tab.label }}</span>
            <span class="mt-0.5 block truncate text-xs opacity-70">{{ tab.description }}</span>
          </span>
        </button>
      </nav>
      <div class="border-t border-slate-200 p-4">
        <p class="truncate text-sm font-medium text-slate-800">{{ adminEmail }}</p>
        <RouterLink class="mt-3 inline-flex w-full justify-center rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-50" to="/">返回前台</RouterLink>
      </div>
    </div>

    <div class="lg:hidden">
      <div class="flex items-center justify-between gap-3 border-b border-slate-200 px-4 py-3">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-teal">Console</p>
          <h1 class="text-lg font-semibold text-slate-950">管理后台</h1>
        </div>
        <RouterLink class="rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600" to="/">前台</RouterLink>
      </div>
      <nav class="flex gap-2 overflow-x-auto px-3 py-2">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="flex min-w-20 flex-col items-center gap-1 rounded-2xl px-3 py-2 text-xs transition"
          :class="props.activeTab === tab.id ? 'bg-slate-950 text-white' : 'text-slate-600 hover:bg-slate-50'"
          type="button"
          @click="selectTab(tab.id)"
        >
          <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" :d="tab.icon" />
          </svg>
          <span>{{ tab.label }}</span>
        </button>
      </nav>
    </div>
  </aside>
</template>
