<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { useUserStore } from './stores/user'

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()
const theme = ref(localStorage.getItem('theme') || defaultTheme())
const isHome = computed(() => route.name === 'home' || route.path === '/')
const isAdmin = computed(() => (userStore.user?.role || 0) >= 10)
const roleLabel = computed(() => (isAdmin.value ? '管理员已登录' : userStore.user ? '普通用户已登录' : '未登录'))

function defaultTheme() {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyTheme(value: string) {
  document.documentElement.classList.toggle('dark', value === 'dark')
  localStorage.setItem('theme', value)
}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

function handleUnauthorized() {
  userStore.logout()
}

async function logout() {
  userStore.logout()
  await router.push('/login')
}

watch(theme, applyTheme, { immediate: true })

onMounted(() => {
  userStore.fetchUser()
  window.addEventListener('auth:unauthorized', handleUnauthorized)
})

onUnmounted(() => {
  window.removeEventListener('auth:unauthorized', handleUnauthorized)
})
</script>

<template>
  <div class="min-h-screen bg-mist text-ink dark:bg-slate-950 dark:text-slate-100">
    <header class="border-b border-gray-200 bg-white dark:border-slate-800 dark:bg-slate-900">
      <nav class="flex min-h-16 items-center justify-between gap-3 px-4 py-3 sm:px-6">
        <RouterLink to="/" class="flex items-center gap-3">
          <span class="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-violet-600 to-blue-600 text-white">
            <svg class="size-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
            </svg>
          </span>
          <span>
            <span class="block text-xl font-medium tracking-tight text-gray-900 dark:text-white">来看看巴</span>
            <span class="hidden text-xs text-gray-500 sm:block">来看看巴</span>
          </span>
        </RouterLink>

        <div class="flex flex-wrap items-center justify-end gap-2 text-sm">
          <RouterLink
            v-if="isAdmin"
            class="min-h-10 rounded border border-slate-300 px-3 py-2 hover:bg-slate-100 dark:border-slate-600 dark:hover:bg-slate-800"
            to="/admin"
          >
            管理后台
          </RouterLink>
          <RouterLink
            v-if="userStore.user && !isAdmin"
            class="min-h-10 rounded border border-slate-300 px-3 py-2 hover:bg-slate-100 dark:border-slate-600 dark:hover:bg-slate-800"
            to="/history"
          >
            历史
          </RouterLink>
          <span
            v-if="userStore.user"
            class="rounded-xl px-3 py-2 text-sm"
            :class="isAdmin ? 'bg-amber-50 text-amber-700' : 'bg-violet-50 text-violet-900'"
          >
            {{ roleLabel }}<template v-if="!isAdmin"> · {{ userStore.user.credits }} 积分</template>
          </span>
          <button class="min-h-10 rounded border border-slate-300 px-3 py-1.5 dark:border-slate-600" type="button" @click="toggleTheme">
            {{ theme === 'dark' ? '浅色' : '深色' }}
          </button>
          <button v-if="userStore.user" class="min-h-10 rounded border border-slate-300 px-3 py-1.5 dark:border-slate-600" type="button" @click="logout">
            退出
          </button>
          <RouterLink
            v-else
            class="min-h-10 rounded bg-gradient-to-r from-violet-600 to-blue-600 px-4 py-2.5 text-white shadow-lg shadow-violet-500/20 hover:from-violet-700 hover:to-blue-700"
            to="/login"
          >
            登录 / 注册
          </RouterLink>
        </div>
      </nav>
    </header>

    <main :class="isHome ? 'mx-auto max-w-none p-0' : 'mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8'">
      <RouterView />
    </main>
  </div>
</template>
