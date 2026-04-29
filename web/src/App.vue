<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useRouter } from 'vue-router'

import { useUserStore } from './stores/user'

const userStore = useUserStore()
const router = useRouter()

function handleUnauthorized() {
  userStore.logout()
}

async function logout() {
  userStore.logout()
  await router.push('/login')
}

onMounted(() => {
  userStore.fetchUser()
  window.addEventListener('auth:unauthorized', handleUnauthorized)
})

onUnmounted(() => {
  window.removeEventListener('auth:unauthorized', handleUnauthorized)
})
</script>

<template>
  <div class="min-h-screen bg-mist text-ink">
    <header class="border-b border-slate-200 bg-white">
      <nav class="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
        <RouterLink to="/" class="text-lg font-semibold text-ink">Image Show</RouterLink>
        <div class="flex items-center gap-2 text-sm">
          <template v-if="userStore.user">
            <span class="hidden text-slate-600 sm:inline">{{ userStore.user.email }}</span>
            <span class="rounded bg-teal px-2 py-1 text-white">{{ userStore.user.credits }} 积分</span>
            <RouterLink
              v-if="userStore.user.role >= 10"
              class="rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-100"
              to="/admin"
            >
              管理
            </RouterLink>
            <button class="rounded border border-slate-300 px-3 py-1.5" type="button" @click="logout">
              退出
            </button>
          </template>
          <template v-else>
            <RouterLink class="rounded px-3 py-1.5 text-slate-700 hover:bg-slate-100" to="/login">登录</RouterLink>
            <RouterLink class="rounded bg-coral px-3 py-1.5 text-white hover:bg-red-500" to="/register">
              注册
            </RouterLink>
          </template>
        </div>
      </nav>
    </header>

    <main class="mx-auto max-w-6xl px-4 py-8 sm:px-6">
      <RouterView />
    </main>
  </div>
</template>
