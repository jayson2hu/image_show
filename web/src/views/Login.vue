<script setup lang="ts">
import { computed, ref } from 'vue'
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
const qrDialogOpen = ref(false)
const showEmailLogin = ref(false)
const titleText = computed(() => (wechatLoaded.value ? '微信登录 / 注册' : '登录 / 注册'))
const subtitleText = computed(() => (wechatLoaded.value ? '关注公众号获取验证码，首次使用会自动创建账号。' : '点击获取验证码后扫码关注公众号。'))

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

async function openWechatQRCode() {
  qrDialogOpen.value = true
  if (!wechatLoaded.value) {
    await loadWechatLogin()
  }
}

function closeWechatQRCode() {
  qrDialogOpen.value = false
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
</script>

<template>
  <section class="mx-auto flex min-h-[calc(100vh-160px)] max-w-5xl items-center justify-center px-4 py-8 text-slate-900 dark:text-slate-100">
    <div class="grid w-full overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm md:grid-cols-[0.82fr_1.18fr] dark:border-slate-700 dark:bg-slate-900">
      <div class="hidden border-r border-slate-200 bg-slate-50 p-8 md:flex md:flex-col md:justify-between dark:border-slate-700 dark:bg-slate-950">
        <div>
          <RouterLink class="inline-flex text-sm font-medium text-slate-500 transition hover:text-slate-900 dark:text-slate-400 dark:hover:text-white" to="/">
            返回首页
          </RouterLink>
          <h1 class="mt-10 text-3xl font-semibold leading-tight text-slate-950 dark:text-white">一个入口完成登录和注册</h1>
          <p class="mt-4 text-sm leading-6 text-slate-600 dark:text-slate-300">微信验证码可直接登录。首次使用会自动创建账号并获得注册积分。</p>
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
            <p class="text-sm font-medium text-teal">用户入口</p>
            <h2 class="mt-2 text-2xl font-semibold text-slate-950 dark:text-white">{{ titleText }}</h2>
            <p class="mt-2 text-sm leading-6 text-slate-500 dark:text-slate-400">{{ subtitleText }}</p>
          </div>

          <div class="mt-6 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-950">
            <div class="space-y-3">
              <button class="min-h-11 w-full rounded-lg border border-teal/30 bg-teal/5 px-4 text-sm font-medium text-teal transition hover:bg-teal/10 disabled:opacity-60" type="button" :disabled="wechatLoading" @click="openWechatQRCode">
                {{ wechatLoading ? '正在加载公众号二维码...' : '获取验证码' }}
              </button>
              <input
                v-model="wechatCode"
                class="min-h-11 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:opacity-60 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100"
                placeholder="输入公众号返回的验证码"
                :disabled="!wechatLoaded || !wechatEnabled"
                @keydown.enter.prevent="submitWechatCode"
              />
              <button class="min-h-11 w-full rounded-lg bg-teal px-4 text-sm font-medium text-white transition hover:bg-teal/90 disabled:opacity-60" type="button" :disabled="!wechatLoaded || !wechatEnabled || wechatLoading || !wechatCode" @click="submitWechatCode">
                {{ wechatLoading ? '处理中...' : '登录 / 注册' }}
              </button>
            </div>
            <p v-if="error" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
          </div>

          <div class="mt-5 rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950">
            <button class="flex w-full items-center justify-between text-left text-sm font-medium text-slate-700 dark:text-slate-200" type="button" @click="showEmailLogin = !showEmailLogin">
              <span>已有邮箱账号</span>
              <span class="text-slate-400">{{ showEmailLogin ? '收起' : '展开' }}</span>
            </button>
            <form v-if="showEmailLogin" class="mt-4 space-y-3" @submit.prevent="submitEmailLogin">
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
          </div>
        </div>
      </div>
    </div>

    <div v-if="qrDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/55 px-4" role="dialog" aria-modal="true" aria-labelledby="wechat-qrcode-title" @click.self="closeWechatQRCode">
      <div class="w-full max-w-sm rounded-2xl bg-white p-5 text-slate-900 shadow-xl dark:bg-slate-900 dark:text-slate-100">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 id="wechat-qrcode-title" class="text-lg font-semibold">关注公众号获取验证码</h3>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">扫码关注后，公众号会返回验证码。</p>
          </div>
          <button class="rounded-full border border-slate-200 px-3 py-1 text-sm text-slate-500 transition hover:bg-slate-50 dark:border-slate-700 dark:hover:bg-slate-800" type="button" @click="closeWechatQRCode">关闭</button>
        </div>
        <div class="mt-5 flex min-h-64 items-center justify-center rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950">
          <img v-if="wechatQRCode" :src="wechatQRCode" class="size-56 object-contain" alt="微信公众号二维码" />
          <p v-else class="text-sm text-slate-500">{{ wechatLoading ? '二维码加载中...' : '未配置公众号二维码' }}</p>
        </div>
        <div class="mt-4 grid gap-2 sm:grid-cols-2">
          <button class="min-h-10 rounded-lg border border-slate-300 px-4 text-sm font-medium text-slate-700 transition hover:bg-slate-50 disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800" type="button" :disabled="wechatLoading" @click="loadWechatLogin">刷新二维码</button>
          <button class="min-h-10 rounded-lg bg-teal px-4 text-sm font-medium text-white transition hover:bg-teal/90" type="button" @click="closeWechatQRCode">我已拿到验证码</button>
        </div>
      </div>
    </div>
  </section>
</template>
