<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const wechatLoading = ref(false)
const wechatCode = ref('')
const wechatQRCode = ref('')
const wechatEnabled = ref(false)
const wechatLoaded = ref(false)

async function submitEmailLogin() {
  error.value = ''
  loading.value = true
  try {
    await userStore.login(email.value, password.value)
    await router.push('/')
  } catch {
    error.value = '邮箱或密码不正确'
  } finally {
    loading.value = false
  }
}

async function loadWechatLogin() {
  error.value = ''
  wechatLoading.value = true
  try {
    const response = await api.get('/auth/wechat/qrcode')
    wechatEnabled.value = response.data.enabled
    wechatQRCode.value = response.data.qrcode_url
    wechatLoaded.value = true
    if (!response.data.enabled) {
      error.value = '微信登录未开启'
    }
  } catch {
    error.value = '微信登录配置读取失败'
  } finally {
    wechatLoading.value = false
  }
}

async function submitWechatCode() {
  if (!wechatCode.value || !wechatEnabled.value) {
    return
  }
  error.value = ''
  wechatLoading.value = true
  try {
    await userStore.wechatLogin(wechatCode.value)
    await router.push('/')
  } catch {
    error.value = '微信验证码无效或已过期'
  } finally {
    wechatLoading.value = false
  }
}

onMounted(() => {
  loadWechatLogin()
})
</script>

<template>
  <section class="mx-auto flex min-h-[calc(100vh-160px)] max-w-5xl items-center justify-center px-4 py-8 text-slate-900 dark:text-slate-100">
    <div class="grid w-full overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm md:grid-cols-[0.92fr_1.08fr] dark:border-slate-700 dark:bg-slate-900">
      <div class="hidden border-r border-slate-200 bg-slate-50 p-8 md:flex md:flex-col md:justify-between dark:border-slate-700 dark:bg-slate-950">
        <div>
          <RouterLink class="inline-flex text-sm font-medium text-slate-500 transition hover:text-slate-900 dark:text-slate-400 dark:hover:text-white" to="/">
            返回首页
          </RouterLink>
          <h1 class="mt-10 text-3xl font-semibold leading-tight text-slate-950 dark:text-white">登录后保存历史记录和积分余额</h1>
          <p class="mt-4 text-sm leading-6 text-slate-600 dark:text-slate-300">游客仍可免费体验一次。新账号使用微信注册，已有邮箱账号可继续登录。</p>
        </div>
        <RouterLink class="inline-flex w-fit min-h-10 items-center rounded-full border border-slate-300 px-5 text-sm font-medium text-slate-700 transition hover:bg-white dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-900" to="/">
          游客继续体验
        </RouterLink>
      </div>

      <div class="p-6 sm:p-8">
        <div class="mx-auto max-w-md">
          <div class="flex items-start justify-between gap-4 md:hidden">
            <RouterLink class="text-sm font-medium text-slate-500 transition hover:text-slate-900 dark:text-slate-400 dark:hover:text-white" to="/">
              返回首页
            </RouterLink>
            <RouterLink class="text-sm font-medium text-teal transition hover:text-teal/80" to="/">
              游客体验
            </RouterLink>
          </div>

          <div class="mt-6 md:mt-0">
            <p class="text-sm font-medium text-teal">登录 / 注册</p>
            <h2 class="mt-2 text-2xl font-semibold text-slate-950 dark:text-white">微信验证码登录</h2>
            <p class="mt-2 text-sm leading-6 text-slate-500 dark:text-slate-400">扫码获取验证码并提交。首次使用会自动创建微信账号。</p>
          </div>

          <div class="mt-6 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-950">
            <div class="grid gap-4 sm:grid-cols-[168px_1fr] sm:items-center">
              <div class="mx-auto flex size-40 items-center justify-center rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-900">
                <img v-if="wechatQRCode" :src="wechatQRCode" class="size-32 object-contain" alt="微信二维码" />
                <div v-else class="text-center text-slate-400">
                  <svg class="mx-auto size-10" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                    <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1 .023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.03zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z" />
                  </svg>
                  <p class="mt-2 text-xs">{{ wechatLoading ? '加载中' : '未加载' }}</p>
                </div>
              </div>

              <div class="space-y-3">
                <input
                  v-model="wechatCode"
                  class="min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:opacity-60 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
                  placeholder="输入微信验证码"
                  :disabled="!wechatEnabled"
                  @keydown.enter.prevent="submitWechatCode"
                />
                <button class="min-h-11 w-full rounded-lg bg-teal px-4 text-sm font-medium text-white transition hover:bg-teal/90 disabled:opacity-60" type="button" :disabled="!wechatEnabled || wechatLoading || !wechatCode" @click="submitWechatCode">
                  {{ wechatLoading ? '处理中...' : '微信登录 / 注册' }}
                </button>
                <button class="min-h-10 w-full rounded-lg border border-slate-300 px-4 text-sm font-medium text-slate-700 transition hover:bg-slate-50 disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-900" type="button" :disabled="wechatLoading" @click="loadWechatLogin">
                  {{ wechatLoaded ? '刷新二维码' : '加载二维码' }}
                </button>
              </div>
            </div>
            <p v-if="error" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
          </div>

          <details class="mt-5 rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950">
            <summary class="cursor-pointer text-sm font-medium text-slate-700 dark:text-slate-200">已有邮箱账号登录</summary>
            <form class="mt-4 space-y-3" @submit.prevent="submitEmailLogin">
              <input
                v-model="email"
                class="min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
                type="email"
                autocomplete="email"
                placeholder="邮箱地址"
                required
              />
              <input
                v-model="password"
                class="min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
                type="password"
                autocomplete="current-password"
                placeholder="密码"
                required
              />
              <button class="min-h-11 w-full rounded-lg border border-slate-300 bg-white px-4 text-sm font-medium text-slate-700 transition hover:bg-slate-50 disabled:opacity-60 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800" type="submit" :disabled="loading">
                {{ loading ? '登录中...' : '邮箱登录' }}
              </button>
            </form>
          </details>

          <p class="mt-5 text-center text-sm text-slate-500 dark:text-slate-400">
            新用户请使用微信注册。
            <RouterLink class="font-medium text-teal transition hover:text-teal/80" to="/register">查看注册页</RouterLink>
          </p>
        </div>
      </div>
    </div>
  </section>
</template>
