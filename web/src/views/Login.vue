<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

type LoginMethod = 'email' | 'wechat'

const router = useRouter()
const userStore = useUserStore()
const loginMethod = ref<LoginMethod>('email')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const wechatOpen = ref(false)
const wechatLoading = ref(false)
const wechatCode = ref('')
const wechatQRCode = ref('')
const wechatEnabled = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await userStore.login(email.value, password.value)
    await router.push((userStore.user?.role || 0) >= 10 ? '/admin' : '/')
  } catch {
    error.value = '邮箱或密码不正确'
  } finally {
    loading.value = false
  }
}

async function openWechatLogin() {
  error.value = ''
  wechatLoading.value = true
  try {
    const response = await api.get('/auth/wechat/qrcode')
    wechatEnabled.value = response.data.enabled
    wechatQRCode.value = response.data.qrcode_url
    wechatOpen.value = true
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
  if (!wechatCode.value) {
    return
  }
  error.value = ''
  wechatLoading.value = true
  try {
    await userStore.wechatLogin(wechatCode.value)
    await router.push((userStore.user?.role || 0) >= 10 ? '/admin' : '/')
  } catch {
    error.value = '微信验证码无效或已过期'
  } finally {
    wechatLoading.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-md rounded-2xl border border-slate-200 bg-white p-8 text-slate-900 shadow-2xl dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100">
    <h1 class="mb-6 text-center text-2xl font-medium text-gray-900 dark:text-white">欢迎登录</h1>

    <div class="mb-6 flex gap-2 rounded-xl bg-gray-100 p-1 dark:bg-slate-800">
      <button
        class="flex-1 rounded-lg py-2.5 transition"
        :class="loginMethod === 'email' ? 'bg-white text-gray-900 shadow-sm dark:bg-slate-950 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-slate-300 dark:hover:text-white'"
        type="button"
        @click="loginMethod = 'email'"
      >
        <span class="inline-flex items-center justify-center gap-2">
          <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          邮箱
        </span>
      </button>
      <button
        class="flex-1 rounded-lg py-2.5 transition"
        :class="loginMethod === 'wechat' ? 'bg-white text-gray-900 shadow-sm dark:bg-slate-950 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-slate-300 dark:hover:text-white'"
        type="button"
        @click="loginMethod = 'wechat'"
      >
        <span class="inline-flex items-center justify-center gap-2">
          <svg class="size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1 .023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.03zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z" />
          </svg>
          微信
        </span>
      </button>
    </div>

    <form v-if="loginMethod === 'email'" class="space-y-4" @submit.prevent="submit">
      <label class="block text-sm font-medium text-gray-700 dark:text-slate-200">
        邮箱
        <input
          v-model="email"
          class="mt-2 w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-slate-900 outline-none focus:border-transparent focus:ring-2 focus:ring-violet-500 dark:border-slate-600 dark:bg-slate-950 dark:text-slate-100"
          type="email"
          autocomplete="email"
          placeholder="请输入邮箱地址"
          required
        />
      </label>
      <label class="block text-sm font-medium text-gray-700 dark:text-slate-200">
        密码
        <input
          v-model="password"
          class="mt-2 w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-slate-900 outline-none focus:border-transparent focus:ring-2 focus:ring-violet-500 dark:border-slate-600 dark:bg-slate-950 dark:text-slate-100"
          type="password"
          autocomplete="current-password"
          placeholder="请输入密码"
          required
        />
      </label>
      <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
      <button class="w-full rounded-xl bg-gradient-to-r from-violet-600 to-blue-600 py-3 text-white shadow-lg shadow-violet-500/30 transition hover:from-violet-700 hover:to-blue-700 disabled:opacity-60" type="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>

    <div v-else class="py-2">
      <div class="flex flex-col items-center">
        <div class="mb-4 flex size-48 items-center justify-center rounded-2xl border-2 border-dashed border-gray-300 bg-gray-100 dark:border-slate-600 dark:bg-slate-950">
          <img v-if="wechatQRCode" :src="wechatQRCode" class="size-40 object-contain" alt="微信二维码" />
          <div v-else class="text-center">
            <svg class="mx-auto mb-2 size-16 text-gray-400" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1 .023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.03zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z" />
            </svg>
            <p class="text-sm text-gray-500 dark:text-slate-400">{{ wechatLoading ? '二维码加载中...' : '点击下方加载二维码' }}</p>
          </div>
        </div>
        <p class="mb-2 text-gray-700 dark:text-slate-200">打开微信扫一扫</p>
        <p class="text-sm text-gray-500 dark:text-slate-400">扫描二维码登录，或输入公众号验证码</p>
        <button class="mt-4 rounded-lg border border-gray-300 px-6 py-2 text-sm text-gray-600 transition hover:bg-gray-50 disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800" type="button" :disabled="wechatLoading" @click="openWechatLogin">
          {{ wechatOpen ? '刷新二维码' : '加载二维码' }}
        </button>
        <div v-if="wechatOpen" class="mt-4 w-full space-y-3">
          <input
            v-model="wechatCode"
            class="w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-slate-900 outline-none focus:border-transparent focus:ring-2 focus:ring-violet-500 dark:border-slate-600 dark:bg-slate-950 dark:text-slate-100"
            placeholder="微信验证码"
            :disabled="!wechatEnabled"
          />
          <button class="w-full rounded-xl bg-gradient-to-r from-violet-600 to-blue-600 py-3 text-white disabled:opacity-60" type="button" :disabled="!wechatEnabled || wechatLoading" @click="submitWechatCode">
            提交验证码
          </button>
        </div>
        <p v-if="error" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
      </div>
    </div>

    <div v-if="loginMethod === 'email'" class="mt-6 text-center">
      <RouterLink class="text-violet-600 transition hover:text-violet-700" to="/register">没有账号？立即注册</RouterLink>
    </div>
  </section>
</template>
