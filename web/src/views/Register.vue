<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const error = ref('')
const wechatLoading = ref(false)
const wechatCode = ref('')
const wechatQRCode = ref('')
const wechatEnabled = ref(false)
const wechatLoaded = ref(false)

async function loadWechatRegister() {
  error.value = ''
  wechatLoading.value = true
  try {
    const response = await api.get('/auth/wechat/qrcode')
    wechatEnabled.value = response.data.enabled
    wechatQRCode.value = response.data.qrcode_url
    wechatLoaded.value = true
    if (!response.data.enabled) {
      error.value = '微信注册未开启'
    }
  } catch {
    error.value = '微信注册配置读取失败'
  } finally {
    wechatLoading.value = false
  }
}

async function submitWechatRegister() {
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
  loadWechatRegister()
})
</script>

<template>
  <section class="mx-auto max-w-3xl overflow-hidden rounded-2xl border border-slate-200 bg-white text-slate-900 shadow-sm dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100">
    <div class="border-b border-slate-200 bg-slate-50 px-6 py-7 sm:px-8 dark:border-slate-700 dark:bg-slate-950">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="text-sm font-medium text-teal">微信注册</p>
          <h1 class="mt-2 text-3xl font-semibold text-slate-950 dark:text-white">新账号仅支持微信注册</h1>
          <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-600 dark:text-slate-300">扫码获取验证码后提交，系统会自动创建账号并登录。邮箱注册入口已关闭，已有邮箱账号仍可从登录页进入。</p>
        </div>
        <RouterLink class="shrink-0 rounded-full border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-white dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-900" to="/login">
          去登录
        </RouterLink>
        <RouterLink class="shrink-0 rounded-full border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-white dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-900" to="/">
          游客体验
        </RouterLink>
      </div>
    </div>

    <div class="grid gap-6 p-6 sm:p-8 md:grid-cols-[260px_1fr]">
      <div class="flex flex-col items-center rounded-2xl border border-slate-200 bg-slate-50 p-5 dark:border-slate-700 dark:bg-slate-950">
        <div class="flex size-52 items-center justify-center rounded-2xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900">
          <img v-if="wechatQRCode" :src="wechatQRCode" class="size-44 object-contain" alt="微信二维码" />
          <div v-else class="text-center text-slate-400">
            <svg class="mx-auto size-14" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1 .023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.03zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z" />
            </svg>
            <p class="mt-3 text-sm">{{ wechatLoading ? '加载中...' : '等待加载' }}</p>
          </div>
        </div>
        <button class="mt-4 min-h-10 rounded-full border border-slate-300 px-5 text-sm font-medium text-slate-700 transition hover:bg-white disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-900" type="button" :disabled="wechatLoading" @click="loadWechatRegister">
          {{ wechatLoaded ? '刷新二维码' : '加载二维码' }}
        </button>
      </div>

      <div class="flex flex-col justify-center">
        <div class="space-y-4">
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">
            微信验证码
            <input
              v-model="wechatCode"
              class="mt-2 w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:opacity-60 dark:border-slate-600 dark:bg-slate-950 dark:text-slate-100"
              placeholder="输入微信验证码"
              :disabled="!wechatEnabled"
              @keydown.enter.prevent="submitWechatRegister"
            />
          </label>
          <button class="w-full rounded-xl bg-teal px-4 py-3 font-medium text-white shadow-sm transition hover:bg-teal/90 disabled:opacity-60" type="button" :disabled="!wechatEnabled || wechatLoading || !wechatCode" @click="submitWechatRegister">
            {{ wechatLoading ? '注册中...' : '微信注册并登录' }}
          </button>
          <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        </div>

        <div class="mt-7 rounded-2xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-300">
          <p class="font-medium text-slate-800 dark:text-slate-100">当前注册规则</p>
          <p class="mt-2 leading-6">新用户必须通过微信验证码注册。游客仍可返回首页免费体验一次，注册后可获得积分并保存历史记录。</p>
        </div>
      </div>
    </div>
  </section>
</template>
