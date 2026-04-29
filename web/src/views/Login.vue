<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
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
    await router.push('/')
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
    await router.push('/')
  } catch {
    error.value = '微信验证码无效或已过期'
  } finally {
    wechatLoading.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-md rounded border border-slate-200 bg-white p-6">
    <h1 class="text-xl font-semibold">登录</h1>
    <form class="mt-6 space-y-4" @submit.prevent="submit">
      <label class="block text-sm font-medium">
        邮箱
        <input v-model="email" class="mt-1 w-full rounded border border-slate-300 px-3 py-2" type="email" autocomplete="email" required />
      </label>
      <label class="block text-sm font-medium">
        密码
        <input
          v-model="password"
          class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
          type="password"
          autocomplete="current-password"
          required
        />
      </label>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="w-full rounded bg-teal px-4 py-2 text-white disabled:opacity-60" type="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>
    <div class="mt-4 border-t border-slate-200 pt-4">
      <button class="w-full rounded border border-slate-300 px-4 py-2 text-sm disabled:opacity-60" type="button" :disabled="wechatLoading" @click="openWechatLogin">
        微信登录
      </button>
      <div v-if="wechatOpen" class="mt-4 space-y-3 rounded border border-slate-200 p-3">
        <img v-if="wechatQRCode" :src="wechatQRCode" class="mx-auto h-40 w-40 object-contain" alt="微信二维码" />
        <p class="text-sm text-slate-600">扫码关注公众号，输入验证码后提交。</p>
        <input v-model="wechatCode" class="w-full rounded border border-slate-300 px-3 py-2" placeholder="微信验证码" :disabled="!wechatEnabled" />
        <button class="w-full rounded bg-coral px-4 py-2 text-white disabled:opacity-60" type="button" :disabled="!wechatEnabled || wechatLoading" @click="submitWechatCode">
          提交验证码
        </button>
      </div>
    </div>
  </section>
</template>
