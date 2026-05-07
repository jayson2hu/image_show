<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useToast } from '@/composables/useToast'
import { useUserStore } from '@/stores/user'

import AdminSidebar from './AdminSidebar.vue'
import ChannelsTab from './ChannelsTab.vue'
import OverviewTab from './OverviewTab.vue'
import TemplatesTab from './TemplatesTab.vue'
import UsersTab from './UsersTab.vue'

const router = useRouter()
const userStore = useUserStore()
const toast = useToast()
const activeTab = ref('overview')
const ready = ref(false)

const isAdmin = computed(() => (userStore.user?.role || 0) >= 10)

const tabTitles: Record<string, string> = {
  overview: '概览',
  users: '用户',
  channels: '渠道',
  templates: '模板',
  settings: '设置',
  announcements: '公告',
  credits: '积分',
  monitor: '监控',
}

onMounted(async () => {
  await userStore.fetchUser()
  if (!isAdmin.value) {
    toast.error('当前账号没有管理员权限')
    await router.push('/')
    return
  }
  ready.value = true
})
</script>

<template>
  <section v-if="ready" class="min-h-[calc(100vh-65px)] bg-slate-50 text-slate-950">
    <div class="grid min-h-[calc(100vh-65px)] lg:grid-cols-[260px_1fr]">
      <AdminSidebar v-model:active-tab="activeTab" />
      <main class="min-w-0 p-4 sm:p-6 lg:p-8">
        <OverviewTab v-if="activeTab === 'overview'" @change-tab="activeTab = $event" />
        <UsersTab v-else-if="activeTab === 'users'" />
        <ChannelsTab v-else-if="activeTab === 'channels'" />
        <TemplatesTab v-else-if="activeTab === 'templates'" />
        <div v-else class="rounded-3xl border border-dashed border-slate-300 bg-white p-10 text-center shadow-sm">
          <p class="text-sm font-medium text-teal">Admin redesign preview</p>
          <h2 class="mt-2 text-2xl font-semibold text-slate-950">{{ tabTitles[activeTab] || '概览' }}</h2>
          <p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-slate-500">
            新后台布局基础已就绪。当前阶段只验证布局、侧边栏和权限守卫，具体 Tab 功能将在阶段 C 逐步迁移。
          </p>
        </div>
      </main>
    </div>
  </section>
</template>
