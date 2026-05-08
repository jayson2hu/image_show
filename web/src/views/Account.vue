<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

interface AccountUser {
  id: number
  username: string
  email: string
  avatar_url: string
  role: number
  status: number
  credits: number
  credits_expiry?: string | null
  created_at: string
  updated_at: string
  last_login_at?: string | null
  last_login_ip: string
}

interface CreditLog {
  id: number
  type: number
  amount: number
  balance: number
  remark: string
  created_at: string
}

interface RecentGeneration {
  id: number
  prompt: string
  status: number
  image_url: string
}

interface RecentAnnouncement {
  id: number
  title: string
  read_at?: string | null
}

interface AccountOverview {
  user: AccountUser
  credits: {
    recent_logs: CreditLog[]
  }
  creations: {
    total: number
    completed: number
    failed: number
    latest_at?: string | null
    recent_items: RecentGeneration[]
  }
  announcements: {
    unread_count: number
    recent_items: RecentAnnouncement[]
  }
  security?: {
    latest_login?: {
      method: string
      ip: string
      created_at: string
    } | null
  }
}

const router = useRouter()
const userStore = useUserStore()
const overview = ref<AccountOverview | null>(null)
const loading = ref(false)
const savingProfile = ref(false)
const uploadingAvatar = ref(false)
const error = ref('')
const notice = ref('')
const profileForm = ref({ username: '', avatar_url: '' })
const avatarPreviewFailed = ref(false)

const user = computed(() => overview.value?.user || userStore.user)
const displayName = computed(() => {
  const name = user.value?.username?.trim()
  return name || user.value?.email?.split('@')[0] || '用户'
})
const initials = computed(() => displayName.value.slice(0, 1).toUpperCase())
const roleText = computed(() => ((user.value?.role || 0) >= 10 ? '管理员' : '普通用户'))
const statusText = computed(() => (user.value?.status === 1 ? '正常' : '已停用'))
const expiryText = computed(() => formatDate(user.value?.credits_expiry, '暂无到期时间'))
const latestLogin = computed(() => overview.value?.security?.latest_login || null)
const recentLogs = computed(() => overview.value?.credits.recent_logs || [])
const recentItems = computed(() => overview.value?.creations.recent_items || [])
const recentAnnouncements = computed(() => overview.value?.announcements.recent_items || [])

onMounted(async () => {
  if (!userStore.token) {
    await router.push('/login')
    return
  }
  await loadOverview()
})

async function loadOverview() {
  loading.value = true
  error.value = ''
  try {
    const response = await api.get('/account/overview')
    overview.value = response.data
    userStore.user = response.data.user
    profileForm.value = {
      username: response.data.user?.username || '',
      avatar_url: response.data.user?.avatar_url || '',
    }
    avatarPreviewFailed.value = false
  } catch (err: any) {
    error.value = err.response?.data?.error || '个人中心加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function saveProfile() {
  savingProfile.value = true
  error.value = ''
  notice.value = ''
  try {
    const response = await api.put('/account/profile', {
      username: profileForm.value.username,
      avatar_url: profileForm.value.avatar_url,
    })
    applyUser(response.data.user)
    notice.value = '个人资料已更新'
  } catch (err: any) {
    const message = err.response?.data?.error || '个人资料保存失败'
    if (message.includes('avatar_url')) {
      error.value = '头像地址必须以 http://、https:// 或本地上传地址开头'
    } else if (message.includes('username')) {
      error.value = '昵称太长，请控制在 64 个字符以内'
    } else {
      error.value = message
    }
  } finally {
    savingProfile.value = false
  }
}

async function handleAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  uploadingAvatar.value = true
  error.value = ''
  notice.value = ''
  try {
    const formData = new FormData()
    formData.append('avatar', file)
    const response = await api.post('/account/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    applyUser(response.data.user)
    profileForm.value.avatar_url = response.data.avatar_url || response.data.user?.avatar_url || ''
    notice.value = '头像已更新'
  } catch (err: any) {
    const message = err.response?.data?.error || '头像上传失败'
    if (message.includes('less than')) {
      error.value = '头像文件过大，请按后台配置的大小限制上传'
    } else if (message.includes('type')) {
      error.value = '头像格式不支持，请上传 jpg、jpeg、png 或 webp'
    } else {
      error.value = message
    }
  } finally {
    uploadingAvatar.value = false
  }
}

function applyUser(nextUser?: AccountUser) {
  if (!nextUser) return
  if (overview.value) {
    overview.value.user = nextUser
  }
  userStore.user = nextUser
  profileForm.value = {
    username: nextUser.username || '',
    avatar_url: nextUser.avatar_url || '',
  }
  avatarPreviewFailed.value = false
}

function handleAvatarPreviewError() {
  avatarPreviewFailed.value = true
}

function formatDate(value?: string | null, fallback = '-') {
  return value ? new Date(value).toLocaleString() : fallback
}

function loginMethodText(method?: string) {
  const map: Record<string, string> = { email: '邮箱登录', wechat: '微信验证码' }
  return method ? map[method] || method : '暂无登录方式'
}
</script>

<template>
  <section class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
    <div class="flex flex-col gap-4 rounded-3xl border border-slate-200 bg-white p-5 shadow-sm sm:p-6 lg:flex-row lg:items-center lg:justify-between">
      <div class="flex min-w-0 items-center gap-4">
        <img v-if="user?.avatar_url" class="size-16 shrink-0 rounded-2xl object-cover ring-1 ring-slate-200" :src="user.avatar_url" alt="用户头像" />
        <div v-else class="flex size-16 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-2xl font-semibold text-white">
          {{ initials }}
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="truncate text-2xl font-semibold text-slate-950">{{ displayName }}</h1>
            <span class="rounded-full bg-teal/10 px-2.5 py-1 text-xs font-medium text-teal">{{ roleText }}</span>
            <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="user?.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'">{{ statusText }}</span>
          </div>
          <p class="mt-1 truncate text-sm text-slate-500">{{ user?.email || '未登录' }}</p>
          <p class="mt-2 text-xs text-slate-400">注册时间：{{ formatDate(user?.created_at) }}</p>
        </div>
      </div>

      <div class="grid gap-3 sm:grid-cols-3 lg:min-w-[520px]">
        <div class="rounded-2xl bg-slate-50 px-4 py-3">
          <p class="text-xs text-slate-500">当前积分</p>
          <p class="mt-1 text-2xl font-semibold text-slate-950">{{ user?.credits ?? 0 }}</p>
        </div>
        <div class="rounded-2xl bg-slate-50 px-4 py-3">
          <p class="text-xs text-slate-500">积分有效期</p>
          <p class="mt-1 truncate text-sm font-medium text-slate-900">{{ expiryText }}</p>
        </div>
        <div class="rounded-2xl bg-slate-50 px-4 py-3">
          <p class="text-xs text-slate-500">最近登录</p>
          <p class="mt-1 truncate text-sm font-medium text-slate-900">{{ formatDate(user?.last_login_at, '暂无登录记录') }}</p>
          <p class="mt-1 truncate text-xs text-slate-400">{{ user?.last_login_ip || '暂无 IP' }}</p>
        </div>
      </div>
    </div>

    <p v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</p>
    <p v-if="notice" class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{{ notice }}</p>

    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <RouterLink class="account-action" to="/">开始生成</RouterLink>
      <RouterLink class="account-action" to="/history">图片历史</RouterLink>
      <RouterLink class="account-action" to="/credits">积分流水</RouterLink>
      <RouterLink class="account-action account-action-primary" to="/packages">购买积分</RouterLink>
    </div>

    <div v-if="loading && !overview" class="rounded-3xl border border-slate-200 bg-white p-10 text-center text-sm text-slate-500 shadow-sm">
      正在加载个人中心...
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
      <section class="space-y-6">
        <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 class="text-lg font-semibold text-slate-950">个人资料</h2>
          <p class="mt-1 text-sm text-slate-500">维护展示昵称和头像，头像会上传到当前配置的本地存储。</p>

          <form class="mt-5 space-y-4" @submit.prevent="saveProfile">
            <div class="flex items-center gap-4 rounded-2xl bg-slate-50 p-4">
              <img v-if="profileForm.avatar_url && !avatarPreviewFailed" class="size-14 shrink-0 rounded-2xl object-cover ring-1 ring-slate-200" :src="profileForm.avatar_url" alt="头像预览" @error="handleAvatarPreviewError" />
              <div v-else class="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-xl font-semibold text-white">{{ initials }}</div>
              <div class="min-w-0 text-sm text-slate-500">
                <p class="font-medium text-slate-900">{{ profileForm.username.trim() || displayName }}</p>
                <p class="mt-1 truncate">{{ profileForm.avatar_url || '未设置头像' }}</p>
              </div>
            </div>

            <label class="block">
              <span class="text-sm font-medium text-slate-700">昵称</span>
              <input v-model="profileForm.username" class="mt-2 min-h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20" maxlength="64" placeholder="输入昵称" />
            </label>

            <label class="block">
              <span class="text-sm font-medium text-slate-700">上传头像</span>
              <input class="mt-2 min-h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 py-2 text-sm outline-none transition file:mr-4 file:rounded-xl file:border-0 file:bg-slate-950 file:px-3 file:py-2 file:text-sm file:font-semibold file:text-white hover:file:bg-slate-800 focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:opacity-60" type="file" accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp" :disabled="uploadingAvatar" @change="handleAvatarFileChange" />
              <span class="mt-2 block text-xs leading-5 text-slate-500">支持 jpg、jpeg、png、webp。上传成功后会自动同步头像。</span>
            </label>

            <button class="inline-flex min-h-11 items-center justify-center rounded-2xl bg-slate-950 px-5 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60" type="submit" :disabled="savingProfile || uploadingAvatar">
              {{ savingProfile ? '保存中...' : uploadingAvatar ? '头像上传中...' : '保存资料' }}
            </button>
          </form>
        </div>

        <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 class="text-lg font-semibold text-slate-950">安全与通知</h2>
          <div class="mt-4 grid gap-3">
            <div class="rounded-2xl bg-slate-50 px-4 py-3">
              <p class="text-xs text-slate-500">登录信息</p>
              <p class="mt-1 text-sm text-slate-800">{{ latestLogin ? formatDate(latestLogin.created_at) : formatDate(user?.last_login_at, '暂无登录记录') }}</p>
              <p class="mt-1 text-xs text-slate-500">{{ loginMethodText(latestLogin?.method) }} / {{ latestLogin?.ip || user?.last_login_ip || '暂无 IP' }}</p>
            </div>
            <div class="rounded-2xl bg-slate-50 px-4 py-3">
              <p class="text-xs text-slate-500">公告通知</p>
              <p class="mt-1 text-sm text-slate-800">未读 {{ overview?.announcements.unread_count || 0 }} 条</p>
              <p v-if="recentAnnouncements.length" class="mt-1 truncate text-xs text-slate-500">{{ recentAnnouncements[0].title }}</p>
              <p v-else class="mt-1 text-xs text-slate-500">暂无新的通知。</p>
            </div>
          </div>
        </div>
      </section>

      <section class="space-y-6">
        <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 class="text-lg font-semibold text-slate-950">账号摘要</h2>
          <p class="mt-1 text-sm text-slate-500">默认页只展示摘要信息，历史、流水和购买中心从入口进入。</p>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="rounded-2xl bg-slate-50 p-4">
              <p class="text-xs text-slate-500">累计生成</p>
              <p class="mt-1 text-2xl font-semibold text-slate-950">{{ overview?.creations.total || 0 }}</p>
              <p class="mt-1 text-xs text-slate-500">完成 {{ overview?.creations.completed || 0 }}，失败 {{ overview?.creations.failed || 0 }}</p>
            </div>
            <div class="rounded-2xl bg-slate-50 p-4">
              <p class="text-xs text-slate-500">最近作品</p>
              <p class="mt-1 text-2xl font-semibold text-slate-950">{{ recentItems.length }}</p>
              <p class="mt-1 text-xs text-slate-500">点击图片历史查看全部</p>
            </div>
            <div class="rounded-2xl bg-slate-50 p-4">
              <p class="text-xs text-slate-500">最近流水</p>
              <p class="mt-1 text-2xl font-semibold text-slate-950">{{ recentLogs.length }}</p>
            </div>
            <div class="rounded-2xl bg-slate-50 p-4">
              <p class="text-xs text-slate-500">未读通知</p>
              <p class="mt-1 text-2xl font-semibold text-slate-950">{{ overview?.announcements.unread_count || 0 }}</p>
            </div>
          </div>
        </div>

        <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 class="text-lg font-semibold text-slate-950">常用入口</h2>
          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <RouterLink class="rounded-2xl border border-slate-200 p-4 text-sm font-semibold text-slate-700 transition hover:border-teal/40 hover:bg-slate-50" to="/history">查看图片历史</RouterLink>
            <RouterLink class="rounded-2xl border border-slate-200 p-4 text-sm font-semibold text-slate-700 transition hover:border-teal/40 hover:bg-slate-50" to="/credits">查看积分流水</RouterLink>
            <RouterLink class="rounded-2xl border border-slate-200 p-4 text-sm font-semibold text-slate-700 transition hover:border-teal/40 hover:bg-slate-50" to="/packages">购买积分</RouterLink>
            <RouterLink class="rounded-2xl border border-slate-200 p-4 text-sm font-semibold text-slate-700 transition hover:border-teal/40 hover:bg-slate-50" to="/">开始生成</RouterLink>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.account-action {
  display: inline-flex;
  min-height: 3rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid rgb(226 232 240);
  background: white;
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(51 65 85);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
  transition: all 0.2s ease;
}

.account-action:hover {
  border-color: rgb(20 184 166 / 0.35);
  color: rgb(15 23 42);
  transform: translateY(-1px);
}

.account-action-primary {
  border-color: transparent;
  background: rgb(15 23 42);
  color: white;
}

.account-action-primary:hover {
  background: rgb(30 41 59);
  color: white;
}
</style>
