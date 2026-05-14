<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function submitAdminLogin() {
  error.value = ''
  loading.value = true
  try {
    await userStore.login(email.value.trim(), password.value)
    if ((userStore.user?.role || 0) < 10) {
      userStore.logout()
      error.value = '当前账号不是管理员，请使用管理员账号登录。'
      return
    }
    await router.push('/console/admin')
  } catch {
    error.value = '管理员邮箱或密码不正确。'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="mx-auto flex min-h-[calc(100vh-160px)] max-w-4xl items-center justify-center px-4 py-8 text-slate-900 dark:text-slate-100">
    <div class="grid w-full overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm md:grid-cols-[0.85fr_1.15fr] dark:border-slate-700 dark:bg-slate-900">
      <div class="hidden border-r border-slate-200 bg-slate-950 p-8 text-white md:flex md:flex-col md:justify-between dark:border-slate-700">
        <div>
          <RouterLink class="inline-flex text-sm font-medium text-slate-300 transition hover:text-white" to="/">
            返回首页
          </RouterLink>
          <p class="mt-10 text-sm font-semibold uppercase tracking-wide text-teal">Admin Console</p>
          <h1 class="mt-3 text-3xl font-semibold leading-tight">管理员后台登录</h1>
          <p class="mt-4 text-sm leading-6 text-slate-300">后台入口仅用于管理员账号。普通用户请从用户登录入口进入创作页面。</p>
        </div>
        <RouterLink class="inline-flex w-fit min-h-10 items-center rounded-full border border-slate-600 px-5 text-sm font-medium text-slate-200 transition hover:bg-slate-900" to="/login">
          用户登录入口
        </RouterLink>
      </div>

      <div class="p-6 sm:p-8">
        <div class="mx-auto max-w-md">
          <div class="flex items-start justify-between gap-4 md:hidden">
            <RouterLink class="text-sm font-medium text-slate-500 transition hover:text-slate-900 dark:text-slate-400 dark:hover:text-white" to="/">
              返回首页
            </RouterLink>
            <RouterLink class="text-sm font-medium text-teal transition hover:text-teal/80" to="/login">
              用户登录
            </RouterLink>
          </div>

          <div class="mt-6 md:mt-0">
            <p class="text-sm font-medium text-teal">管理员登录</p>
            <h2 class="mt-2 text-2xl font-semibold text-slate-950 dark:text-white">进入后台控制台</h2>
            <p class="mt-2 text-sm leading-6 text-slate-500 dark:text-slate-400">请输入管理员邮箱和密码。普通账号不会进入后台。</p>
          </div>

          <form class="mt-6 space-y-4 rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950" @submit.prevent="submitAdminLogin">
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">
              管理员邮箱
              <input
                v-model="email"
                class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
                type="email"
                autocomplete="email"
                placeholder="admin@example.com"
                required
              />
            </label>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">
              管理员密码
              <input
                v-model="password"
                class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
                type="password"
                autocomplete="current-password"
                placeholder="输入密码"
                required
              />
            </label>
            <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
            <button class="min-h-11 w-full rounded-lg bg-slate-950 px-4 text-sm font-medium text-white transition hover:bg-slate-800 disabled:opacity-60 dark:bg-teal dark:hover:bg-teal/90" type="submit" :disabled="loading">
              {{ loading ? '登录中...' : '登录后台' }}
            </button>
          </form>
        </div>
      </div>
    </div>
  </section>
</template>
