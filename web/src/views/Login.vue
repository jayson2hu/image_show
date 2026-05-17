<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'
import { fetchSiteConfig } from '@/api/site'
import { useConversationStore } from '@/stores/conversation'
import { useUserStore } from '@/stores/user'

type EmailTab = 'login' | 'register'

const router = useRouter()
const userStore = useUserStore()
const conversationStore = useConversationStore()

const wechatCode = ref('')
const wechatQRCode = ref('')
const wechatEnabled = ref(false)
const wechatLoading = ref(false)
const wechatLoaded = ref(false)
const fallbackNotice = ref('')

const emailExpanded = ref(false)
const emailTab = ref<EmailTab>('login')
const loginEmail = ref('')
const loginPassword = ref('')
const registerEmail = ref('')
const registerPassword = ref('')
const registerCode = ref('')
const emailLoading = ref(false)
const codeSending = ref(false)
const registerEnabled = ref(true)
const error = ref('')
const info = ref('')

const wechatUnavailable = computed(() => wechatLoaded.value && !wechatEnabled.value)
const emailPanelMaxHeight = computed(() => (emailExpanded.value ? '520px' : '0px'))

onMounted(() => {
  loadSiteConfig()
  loadWechatLogin()
})

async function loadSiteConfig() {
  try {
    const response = await fetchSiteConfig()
    registerEnabled.value = response.data.register_enabled !== false
    if (!registerEnabled.value && emailTab.value === 'register') {
      emailTab.value = 'login'
    }
  } catch {
    registerEnabled.value = true
  }
}

async function loadWechatLogin() {
  error.value = ''
  fallbackNotice.value = ''
  wechatLoading.value = true
  try {
    const response = await api.get('/auth/wechat/qrcode')
    wechatEnabled.value = Boolean(response.data.enabled)
    wechatQRCode.value = response.data.qrcode_url || ''
    wechatLoaded.value = true
    if (!wechatEnabled.value) {
      fallbackToEmail()
    }
  } catch {
    wechatLoaded.value = true
    wechatEnabled.value = false
    wechatQRCode.value = ''
    fallbackToEmail()
  } finally {
    wechatLoading.value = false
  }
}

function fallbackToEmail() {
  fallbackNotice.value = registerEnabled.value ? '微信登录暂不可用，请使用邮箱登录或注册。' : '微信登录暂不可用，请使用邮箱登录。'
  emailExpanded.value = true
}

async function submitWechatCode() {
  if (!wechatCode.value || !wechatEnabled.value) {
    return
  }
  error.value = ''
  info.value = ''
  wechatLoading.value = true
  try {
    await userStore.wechatLogin(wechatCode.value.trim())
    await conversationStore.syncGuestConversation()
    await userStore.fetchUser()
    await router.push('/')
  } catch {
    error.value = '微信验证码无效或已过期'
  } finally {
    wechatLoading.value = false
  }
}

async function submitEmailLogin() {
  error.value = ''
  info.value = ''
  emailLoading.value = true
  try {
    await userStore.login(loginEmail.value.trim(), loginPassword.value)
    await conversationStore.syncGuestConversation()
    await userStore.fetchUser()
    await router.push('/')
  } catch {
    error.value = '邮箱或密码不正确'
  } finally {
    emailLoading.value = false
  }
}

async function sendRegisterCode() {
  error.value = ''
  info.value = ''
  if (!registerEnabled.value) {
    error.value = '当前暂未开放注册，请联系管理员。'
    return
  }
  if (!registerEmail.value.trim()) {
    error.value = '请先输入注册邮箱'
    return
  }
  codeSending.value = true
  try {
    await api.post('/auth/send-code', { email: registerEmail.value.trim() })
    info.value = '验证码已发送，请查看邮箱'
  } catch (err: any) {
    error.value = err.response?.data?.error || '验证码发送失败'
  } finally {
    codeSending.value = false
  }
}

async function submitEmailRegister() {
  error.value = ''
  info.value = ''
  if (!registerEnabled.value) {
    error.value = '当前暂未开放注册，请联系管理员。'
    return
  }
  emailLoading.value = true
  try {
    await userStore.register(registerEmail.value.trim(), registerPassword.value, registerCode.value.trim())
    await conversationStore.syncGuestConversation()
    await userStore.fetchUser()
    await router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || '注册失败，请检查邮箱、密码和验证码'
  } finally {
    emailLoading.value = false
  }
}

function toggleEmailPanel() {
  emailExpanded.value = !emailExpanded.value
}

function switchEmailTab(tab: EmailTab) {
  if (tab === 'register' && !registerEnabled.value) {
    error.value = '当前暂未开放注册，请联系管理员。'
    info.value = ''
    return
  }
  emailTab.value = tab
  emailExpanded.value = true
  error.value = ''
  info.value = ''
}
</script>

<template>
  <section class="mx-auto flex min-h-[calc(100vh-130px)] max-w-5xl items-center justify-center px-4 py-8 text-slate-900">
    <div class="w-full overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl shadow-slate-900/8">
      <div class="bg-gradient-to-br from-violet-600 via-blue-600 to-cyan-500 px-6 py-8 text-white sm:px-10">
        <div class="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <RouterLink class="text-sm font-medium text-white/75 transition hover:text-white" to="/">返回首页</RouterLink>
            <h1 class="mt-6 text-4xl font-semibold tracking-normal">Image Show</h1>
            <p class="mt-3 max-w-xl text-sm leading-6 text-white/82">微信验证码优先登录；邮箱账号入口保留给已有账号和管理员使用。</p>
          </div>
          <RouterLink class="inline-flex min-h-10 w-fit items-center rounded-full border border-white/30 px-4 text-sm font-medium text-white transition hover:bg-white/12" to="/">
            游客体验
          </RouterLink>
        </div>
      </div>

      <div class="px-5 py-6 sm:px-8 sm:py-8">
        <div v-if="fallbackNotice" class="mb-5 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          {{ fallbackNotice }}
        </div>

        <div class="mx-auto max-w-md rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <div class="text-center">
            <p class="text-sm font-semibold text-violet-700">微信登录</p>
            <h2 class="mt-1 text-2xl font-semibold text-slate-950">扫码获取验证码</h2>
            <p class="mt-2 text-sm leading-6 text-slate-500">关注公众号获取验证码，首次使用会自动创建账号。</p>
          </div>

          <div class="mt-5 flex min-h-64 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <div v-if="wechatLoading" class="size-52 animate-pulse rounded-2xl bg-slate-200"></div>
            <img v-else-if="wechatEnabled && wechatQRCode" :src="wechatQRCode" class="size-52 object-contain" alt="微信公众号二维码" />
            <div v-else class="text-center text-sm text-slate-500">
              <div class="mx-auto mb-3 size-16 rounded-2xl bg-slate-200"></div>
              <p>{{ wechatUnavailable ? '微信登录暂不可用' : '未配置公众号二维码' }}</p>
            </div>
          </div>

          <div class="mt-5 space-y-3">
            <input
              v-model="wechatCode"
              class="auth-input"
              placeholder="输入公众号返回的验证码"
              :disabled="!wechatEnabled"
              @keydown.enter.prevent="submitWechatCode"
            />
            <button class="auth-primary" type="button" :disabled="!wechatEnabled || wechatLoading || !wechatCode" @click="submitWechatCode">
              {{ wechatLoading ? '处理中...' : '微信登录 / 注册' }}
            </button>
            <button class="auth-secondary" type="button" :disabled="wechatLoading" @click="loadWechatLogin">
              刷新二维码
            </button>
          </div>

          <p v-if="error" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</p>
          <p v-if="info" class="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{{ info }}</p>
        </div>

        <div class="mx-auto mt-5 max-w-md rounded-2xl border border-slate-200 bg-slate-50">
          <button class="flex w-full items-center justify-between px-5 py-4 text-left text-sm font-semibold text-slate-700 transition hover:text-violet-700" type="button" @click="toggleEmailPanel">
            <span>已有邮箱账号？邮箱登录</span>
            <span class="transition-transform" :class="emailExpanded ? 'rotate-180' : ''">⌄</span>
          </button>

          <div class="overflow-hidden transition-[max-height] duration-300 ease-in-out" :style="{ maxHeight: emailPanelMaxHeight }">
            <div class="border-t border-slate-200 p-5">
              <div class="grid grid-cols-2 rounded-xl bg-white p-1 shadow-sm">
                <button class="rounded-lg px-3 py-2 text-sm font-semibold transition" :class="emailTab === 'login' ? 'bg-violet-600 text-white shadow-sm' : 'text-slate-600 hover:text-violet-700'" type="button" @click="switchEmailTab('login')">
                  邮箱登录
                </button>
                <button v-if="registerEnabled" class="rounded-lg px-3 py-2 text-sm font-semibold transition" :class="emailTab === 'register' ? 'bg-violet-600 text-white shadow-sm' : 'text-slate-600 hover:text-violet-700'" type="button" @click="switchEmailTab('register')">
                  邮箱注册
                </button>
                <button v-else class="rounded-lg px-3 py-2 text-sm font-semibold text-slate-400" type="button" disabled>
                  注册已关闭
                </button>
              </div>

              <form v-if="emailTab === 'login'" class="mt-4 space-y-3" @submit.prevent="submitEmailLogin">
                <input v-model="loginEmail" class="auth-input" type="email" autocomplete="email" placeholder="邮箱地址" required />
                <input v-model="loginPassword" class="auth-input" type="password" autocomplete="current-password" placeholder="密码" required />
                <button class="auth-primary" type="submit" :disabled="emailLoading">
                  {{ emailLoading ? '登录中...' : '邮箱登录' }}
                </button>
              </form>

              <form v-else class="mt-4 space-y-3" @submit.prevent="submitEmailRegister">
                <input v-model="registerEmail" class="auth-input" type="email" autocomplete="email" placeholder="邮箱地址" required />
                <input v-model="registerPassword" class="auth-input" type="password" autocomplete="new-password" placeholder="至少 8 位密码" required minlength="8" />
                <div class="flex gap-2">
                  <input v-model="registerCode" class="auth-input" inputmode="numeric" placeholder="邮箱验证码" required maxlength="6" />
                  <button class="shrink-0 rounded-xl border border-violet-200 bg-white px-4 text-sm font-semibold text-violet-700 transition hover:bg-violet-50 disabled:opacity-60" type="button" :disabled="codeSending" @click="sendRegisterCode">
                    {{ codeSending ? '发送中' : '发送验证码' }}
                  </button>
                </div>
                <button class="auth-primary" type="submit" :disabled="emailLoading">
                  {{ emailLoading ? '注册中...' : '邮箱注册并登录' }}
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.auth-input {
  min-height: 2.75rem;
  width: 100%;
  border-radius: 0.75rem;
  border: 1px solid rgb(203 213 225);
  background: white;
  padding: 0.625rem 0.875rem;
  color: rgb(15 23 42);
  outline: none;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.auth-input:focus {
  border-color: rgb(124 58 237);
  box-shadow: 0 0 0 3px rgb(124 58 237 / 0.18);
}

.auth-input:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.auth-primary {
  min-height: 2.75rem;
  width: 100%;
  border-radius: 0.75rem;
  background: linear-gradient(90deg, rgb(124 58 237), rgb(37 99 235));
  padding: 0 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: white;
  box-shadow: 0 12px 24px rgb(79 70 229 / 0.22);
  transition: transform 0.18s ease, filter 0.18s ease, opacity 0.18s ease;
}

.auth-primary:hover:not(:disabled) {
  filter: brightness(0.95);
  transform: translateY(-1px);
}

.auth-primary:active:not(:disabled) {
  transform: translateY(0);
}

.auth-primary:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.auth-secondary {
  min-height: 2.75rem;
  width: 100%;
  border-radius: 0.75rem;
  border: 1px solid rgb(221 214 254);
  background: white;
  padding: 0 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: rgb(109 40 217);
  transition: background-color 0.18s ease, border-color 0.18s ease;
}

.auth-secondary:hover:not(:disabled) {
  border-color: rgb(167 139 250);
  background: rgb(245 243 255);
}

.auth-secondary:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
</style>
