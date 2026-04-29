<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'

import { useUserStore } from './stores/user'

const userStore = useUserStore()
const router = useRouter()
const theme = ref(localStorage.getItem('theme') || defaultTheme())

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
    <header class="border-b border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
      <nav class="mx-auto flex min-h-16 max-w-6xl flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
        <RouterLink to="/" class="text-lg font-semibold text-ink dark:text-white">Image Show</RouterLink>
        <div class="flex flex-wrap items-center gap-2 text-sm">
          <button class="min-h-10 rounded border border-slate-300 px-3 py-1.5 dark:border-slate-600" type="button" @click="toggleTheme">
            {{ theme === 'dark' ? '浅色' : '深色' }}
          </button>
          <template v-if="userStore.user">
            <span class="hidden text-slate-600 dark:text-slate-300 sm:inline">{{ userStore.user.email }}</span>
            <span class="rounded bg-teal px-2 py-1 text-white">{{ userStore.user.credits }} 积分</span>
            <RouterLink class="min-h-10 rounded border border-slate-300 px-3 py-2 hover:bg-slate-100 dark:border-slate-600 dark:hover:bg-slate-800" to="/history">
              历史
            </RouterLink>
            <RouterLink
              v-if="userStore.user.role >= 10"
              class="min-h-10 rounded border border-slate-300 px-3 py-2 hover:bg-slate-100 dark:border-slate-600 dark:hover:bg-slate-800"
              to="/admin"
            >
              管理
            </RouterLink>
            <button class="min-h-10 rounded border border-slate-300 px-3 py-1.5 dark:border-slate-600" type="button" @click="logout">
              退出
            </button>
          </template>
          <template v-else>
            <RouterLink class="min-h-10 rounded px-3 py-2 text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800" to="/login">
              登录
            </RouterLink>
            <RouterLink class="min-h-10 rounded bg-coral px-3 py-2 text-white hover:bg-red-500" to="/register">
              注册
            </RouterLink>
          </template>
        </div>
      </nav>
    </header>

    <main class="mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8">
      <RouterView />
    </main>
  </div>
</template>
