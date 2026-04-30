<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

interface Page<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

interface User {
  id: number
  username: string
  email: string
  role: number
  status: number
  credits: number
  created_at: string
}

interface CreditLog {
  id: number
  user_id: number
  type: number
  amount: number
  balance: number
  remark: string
  created_at: string
}

interface Generation {
  id: number
  prompt: string
  quality: string
  size: string
  status: number
  image_url: string
  created_at: string
}

interface PromptTemplate {
  id: number
  category: string
  label: string
  prompt: string
  sort_order: number
  status: number
}

interface Channel {
  id: number
  name: string
  base_url: string
  api_key?: string
  headers?: string
  status: number
  weight: number
  remark?: string
}

interface MonitorSummary {
  date: string
  generation_count: number
  completed_count: number
  failed_count: number
  credits_consumed: number
  new_users: number
  paid_order_count: number
  paid_order_amount: number
  alert_threshold: number
  alert_triggered: boolean
}

const emptyUserPage = (): Page<User> => ({ items: [], total: 0, page: 1, pageSize: 20 })
const emptyGenerationPage = (): Page<Generation> => ({ items: [], total: 0, page: 1, pageSize: 10 })
const emptyCreditPage = (): Page<CreditLog> => ({ items: [], total: 0, page: 1, pageSize: 20 })

const router = useRouter()
const userStore = useUserStore()
const activeTab = ref('overview')
const loading = ref(false)
const message = ref('')
const userKeyword = ref('')
const users = ref<Page<User>>(emptyUserPage())
const selectedUser = ref<User | null>(null)
const userGenerations = ref<Page<Generation>>(emptyGenerationPage())
const creditLogs = ref<Page<CreditLog>>(emptyCreditPage())
const templates = ref<PromptTemplate[]>([])
const channels = ref<Channel[]>([])
const settings = ref<Record<string, string>>({})
const monitor = ref<MonitorSummary | null>(null)
const creditForm = ref({ amount: 1, remark: '' })
const templateForm = ref<PromptTemplate>({ id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 })
const channelForm = ref<Channel>({ id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' })
const channelTestResult = ref<Record<number, string>>({})

const tabs = [
  { id: 'overview', label: '概览', description: '核心指标与运行状态' },
  { id: 'users', label: '用户', description: '账号、角色、状态与充值' },
  { id: 'channels', label: '渠道', description: 'Sub2API 渠道配置与测试' },
  { id: 'templates', label: '模板', description: '提示词模板管理' },
  { id: 'settings', label: '设置', description: '系统开关和第三方配置' },
  { id: 'credits', label: '积分', description: '积分流水审计' },
  { id: 'monitor', label: '监控', description: '每日指标和告警检查' },
]

const isAdmin = computed(() => (userStore.user?.role || 0) >= 10)
const enabledChannels = computed(() => channels.value.filter((item) => item.status === 1).length)
const currentTab = computed(() => tabs.find((item) => item.id === activeTab.value) || tabs[0])
const settingEntries = computed(() => Object.keys(settings.value).sort())
const overviewCards = computed(() => [
  { label: '今日生成', value: monitor.value?.generation_count ?? 0, hint: `${monitor.value?.completed_count ?? 0} 成功 / ${monitor.value?.failed_count ?? 0} 失败` },
  { label: '新增用户', value: monitor.value?.new_users ?? 0, hint: `当前用户 ${users.value.total}` },
  { label: '积分消耗', value: monitor.value?.credits_consumed ?? 0, hint: `告警阈值 ${monitor.value?.alert_threshold ?? 0}` },
  { label: '启用渠道', value: enabledChannels.value, hint: `共 ${channels.value.length} 个渠道` },
])

async function guarded<T>(fn: () => Promise<T>, successMessage = '') {
  loading.value = true
  message.value = ''
  try {
    const result = await fn()
    if (successMessage) {
      message.value = successMessage
    }
    return result
  } catch (error: any) {
    message.value = error.response?.data?.error || '操作失败，请检查权限或输入'
    throw error
  } finally {
    loading.value = false
  }
}

async function loadUsers() {
  await guarded(async () => {
    const response = await api.get('/admin/users', { params: { keyword: userKeyword.value, pageSize: 20 } })
    users.value = response.data
  })
}

async function loadUserGenerations(user: User) {
  selectedUser.value = user
  const response = await api.get(`/admin/users/${user.id}/generations`, { params: { pageSize: 10 } })
  userGenerations.value = response.data
}

async function updateUserStatus(user: User, status: number) {
  await guarded(async () => {
    await api.put(`/admin/users/${user.id}/status`, { status })
    await loadUsers()
  }, status === 1 ? '用户已解封' : '用户已封禁')
}

async function updateUserRole(user: User, role: number) {
  await guarded(async () => {
    await api.put(`/admin/users/${user.id}/role`, { role })
    await loadUsers()
  }, role >= 10 ? '已设为管理员' : '已设为普通用户')
}

async function topupCredits(user: User) {
  await guarded(async () => {
    await api.post(`/admin/users/${user.id}/credits`, creditForm.value)
    creditForm.value = { amount: 1, remark: '' }
    await Promise.all([loadUsers(), loadCreditLogs()])
  }, '充值完成')
}

async function loadCreditLogs() {
  const response = await api.get('/admin/credits/logs', { params: { pageSize: 20 } })
  creditLogs.value = response.data
}

async function loadTemplates() {
  const response = await api.get('/admin/prompt-templates')
  templates.value = response.data.items
}

function editTemplate(template: PromptTemplate) {
  templateForm.value = { ...template }
}

function resetTemplate() {
  templateForm.value = { id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 }
}

async function saveTemplate() {
  await guarded(async () => {
    const payload = { ...templateForm.value }
    if (payload.id) {
      await api.put(`/admin/prompt-templates/${payload.id}`, payload)
    } else {
      await api.post('/admin/prompt-templates', payload)
    }
    resetTemplate()
    await loadTemplates()
  }, '模板已保存')
}

async function deleteTemplate(template: PromptTemplate) {
  await guarded(async () => {
    await api.delete(`/admin/prompt-templates/${template.id}`)
    await loadTemplates()
  }, '模板已删除')
}

async function loadSettings() {
  const response = await api.get('/admin/settings')
  settings.value = response.data.items
}

async function saveSettings() {
  await guarded(async () => {
    await api.put('/admin/settings', { items: settings.value })
  }, '设置已保存')
}

async function loadChannels() {
  const response = await api.get('/admin/channels')
  channels.value = response.data.items
}

function editChannel(channel: Channel) {
  channelForm.value = { ...channel }
}

function resetChannel() {
  channelForm.value = { id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' }
}

async function saveChannel() {
  await guarded(async () => {
    const payload = { ...channelForm.value }
    if (payload.id) {
      await api.put(`/admin/channels/${payload.id}`, payload)
    } else {
      await api.post('/admin/channels', payload)
    }
    resetChannel()
    await loadChannels()
  }, '渠道已保存')
}

async function deleteChannel(channel: Channel) {
  await guarded(async () => {
    await api.delete(`/admin/channels/${channel.id}`)
    await loadChannels()
  }, '渠道已删除')
}

async function testChannel(channel: Channel) {
  channelTestResult.value[channel.id] = '测试中...'
  const response = await api.post(`/admin/channels/${channel.id}/test`)
  channelTestResult.value[channel.id] = response.data.ok ? `可用，状态码 ${response.data.status}` : response.data.error || `不可用，状态码 ${response.data.status}`
}

async function loadMonitor() {
  const response = await api.get('/admin/monitor/summary')
  monitor.value = response.data
}

async function checkMonitorAlert() {
  await guarded(async () => {
    const response = await api.post('/admin/monitor/check')
    await loadMonitor()
    message.value = response.data.sent ? '告警已发送' : '未触发或今日已发送'
  })
}

function fmtTime(value: string) {
  return value ? new Date(value).toLocaleString() : '-'
}

function statusText(status: number) {
  return status === 1 ? '启用' : '禁用'
}

function templateCategoryText(category: string) {
  const map: Record<string, string> = {
    default: '默认标签',
    repair: '修复标签',
    style: '首页风格预设',
    sample: '首页推荐样例',
  }
  return map[category] || category
}

function generationStatus(status: number) {
  const map: Record<number, string> = { 0: '待处理', 1: '生成中', 2: '上传中', 3: '完成', 4: '失败', 5: '取消' }
  return map[status] || `状态 ${status}`
}

function settingLabel(key: string) {
  const map: Record<string, string> = {
    r2_endpoint: 'Cloudflare R2 Endpoint',
    r2_access_key: 'Cloudflare R2 Access Key',
    r2_secret_key: 'Cloudflare R2 Secret Key',
    r2_bucket: 'Cloudflare R2 Bucket',
    r2_public_url: 'Cloudflare R2 Public URL / CDN',
    image_model: '生成模型',
    enabled_image_sizes: '可用图片尺寸',
    captcha_enabled: 'Turnstile 验证开关',
    turnstile_site_key: 'Turnstile Site Key',
    turnstile_secret: 'Turnstile Secret',
    register_enabled: '注册开关',
    ip_blacklist: 'IP 黑名单',
  }
  return map[key] || key
}

function settingHelp(key: string) {
  const map: Record<string, string> = {
    r2_endpoint: '形如 https://<account_id>.r2.cloudflarestorage.com',
    r2_access_key: 'Cloudflare R2 API Token 的 Access Key ID',
    r2_secret_key: 'Cloudflare R2 API Token 的 Secret Access Key',
    r2_bucket: 'R2 bucket 名称，例如 image-show',
    r2_public_url: '可选。绑定自定义域名或 CDN 后填写，例如 https://cdn.example.com；为空时使用 1 小时签名链接。',
    image_model: '默认 gpt-image-1。OpenAI 官方当前没有 gpt-image-2；如果 sub2api 提供自定义模型名，可在这里切换。',
    enabled_image_sizes: '逗号分隔，例如 512x512,768x768,1024x1024,1024x1536,1536x1024。未登录用户只开放宽高都不超过 1024 的尺寸。',
  }
  return map[key] || ''
}

function settingInputType(key: string) {
  return key.includes('secret') || key.includes('password') || key.includes('access_key') ? 'password' : 'text'
}

onMounted(async () => {
  await userStore.fetchUser()
  if (!isAdmin.value) {
    await router.push('/')
    return
  }
  await Promise.all([loadUsers(), loadCreditLogs(), loadTemplates(), loadSettings(), loadChannels(), loadMonitor()])
})
</script>

<template>
  <section class="min-h-[calc(100vh-65px)] bg-slate-50 text-slate-950">
    <div class="grid gap-0 lg:grid-cols-[260px_1fr]">
      <aside class="border-b border-slate-200 bg-white lg:min-h-[calc(100vh-65px)] lg:border-b-0 lg:border-r">
        <div class="p-5">
          <p class="text-xs font-medium uppercase tracking-wide text-slate-400">Admin Console</p>
          <h1 class="mt-2 text-xl font-semibold text-slate-950">管理后台</h1>
          <p class="mt-1 text-sm text-slate-500">用户、渠道、模板和系统运行配置。</p>
        </div>
        <nav class="flex gap-2 overflow-x-auto px-4 pb-4 lg:block lg:space-y-1 lg:overflow-visible">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            class="min-w-28 rounded-lg px-3 py-2 text-left text-sm transition lg:w-full"
            :class="activeTab === tab.id ? 'bg-slate-950 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-100 hover:text-slate-950'"
            @click="activeTab = tab.id"
          >
            <span class="block font-medium">{{ tab.label }}</span>
            <span class="hidden text-xs opacity-70 lg:block">{{ tab.description }}</span>
          </button>
        </nav>
      </aside>

      <main class="min-w-0 p-4 sm:p-6 lg:p-8">
        <header class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 class="text-2xl font-semibold text-slate-950">{{ currentTab.label }}</h2>
            <p class="mt-1 text-sm text-slate-500">{{ currentTab.description }}</p>
          </div>
          <div class="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-600">
            {{ userStore.user?.email }}
          </div>
        </header>

        <p v-if="message" class="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          {{ message }}
        </p>

        <div v-if="activeTab === 'overview'" class="space-y-6">
          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <div v-for="card in overviewCards" :key="card.label" class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <div class="text-sm text-slate-500">{{ card.label }}</div>
              <div class="mt-3 text-3xl font-semibold text-slate-950">{{ card.value }}</div>
              <div class="mt-2 text-sm text-slate-500">{{ card.hint }}</div>
            </div>
          </div>
          <div class="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
            <div class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <div class="flex items-center justify-between">
                <h3 class="font-semibold">渠道状态</h3>
                <button class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm" type="button" @click="activeTab = 'channels'">管理渠道</button>
              </div>
              <div class="mt-4 space-y-3">
                <div v-for="channel in channels.slice(0, 4)" :key="channel.id" class="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-sm">
                  <div class="min-w-0">
                    <div class="truncate font-medium">{{ channel.name }}</div>
                    <div class="truncate text-slate-500">{{ channel.base_url }}</div>
                  </div>
                  <span class="rounded-full px-2 py-1 text-xs" :class="channel.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-200 text-slate-600'">
                    {{ statusText(channel.status) }}
                  </span>
                </div>
                <p v-if="!channels.length" class="text-sm text-slate-500">还没有配置渠道，生成时会使用环境变量兜底。</p>
              </div>
            </div>
            <div class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <h3 class="font-semibold">今日监控</h3>
              <dl class="mt-4 space-y-3 text-sm">
                <div class="flex justify-between"><dt class="text-slate-500">支付订单</dt><dd>{{ monitor?.paid_order_count ?? 0 }}</dd></div>
                <div class="flex justify-between"><dt class="text-slate-500">支付金额</dt><dd>¥{{ monitor?.paid_order_amount ?? 0 }}</dd></div>
                <div class="flex justify-between"><dt class="text-slate-500">告警状态</dt><dd>{{ monitor?.alert_triggered ? '已触发' : '正常' }}</dd></div>
              </dl>
              <button class="mt-5 w-full rounded-lg bg-slate-950 px-4 py-2 text-sm text-white" type="button" @click="checkMonitorAlert">检查告警</button>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'users'" class="space-y-4">
          <div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div class="flex flex-col gap-3 sm:flex-row">
              <input v-model="userKeyword" class="min-h-11 min-w-0 flex-1 rounded-lg border border-slate-300 px-3 py-2" placeholder="搜索邮箱或用户名" />
              <button class="rounded-lg bg-slate-950 px-4 py-2 text-white" type="button" @click="loadUsers">搜索</button>
            </div>
          </div>
          <div class="overflow-x-auto rounded-xl border border-slate-200 bg-white shadow-sm">
            <table class="min-w-full text-sm">
              <thead class="bg-slate-50 text-left text-slate-500">
                <tr>
                  <th class="px-4 py-3">用户</th>
                  <th class="px-4 py-3">角色</th>
                  <th class="px-4 py-3">状态</th>
                  <th class="px-4 py-3">积分</th>
                  <th class="px-4 py-3">创建时间</th>
                  <th class="px-4 py-3">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in users.items" :key="user.id" class="border-t border-slate-100">
                  <td class="px-4 py-3">
                    <div class="font-medium">{{ user.email }}</div>
                    <div class="text-xs text-slate-500">ID {{ user.id }} · {{ user.username || '未设置用户名' }}</div>
                  </td>
                  <td class="px-4 py-3">{{ user.role >= 10 ? '管理员' : '用户' }}</td>
                  <td class="px-4 py-3">
                    <span class="rounded-full px-2 py-1 text-xs" :class="user.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'">
                      {{ user.status === 1 ? '正常' : '封禁' }}
                    </span>
                  </td>
                  <td class="px-4 py-3">{{ user.credits }}</td>
                  <td class="px-4 py-3 text-slate-500">{{ fmtTime(user.created_at) }}</td>
                  <td class="px-4 py-3">
                    <div class="flex flex-wrap gap-2">
                      <button class="rounded-lg border border-slate-200 px-2 py-1" type="button" @click="loadUserGenerations(user)">记录</button>
                      <button class="rounded-lg border border-slate-200 px-2 py-1" type="button" @click="selectedUser = user">充值</button>
                      <button class="rounded-lg border border-slate-200 px-2 py-1" type="button" @click="updateUserStatus(user, user.status === 1 ? 2 : 1)">
                        {{ user.status === 1 ? '封禁' : '解封' }}
                      </button>
                      <button class="rounded-lg border border-slate-200 px-2 py-1" type="button" @click="updateUserRole(user, user.role >= 10 ? 1 : 10)">
                        {{ user.role >= 10 ? '设为用户' : '设为管理员' }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="selectedUser" class="grid gap-4 xl:grid-cols-[380px_1fr]">
            <form class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm" @submit.prevent="topupCredits(selectedUser)">
              <h3 class="font-semibold">给 {{ selectedUser.email }} 充值</h3>
              <input v-model.number="creditForm.amount" class="mt-4 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" min="0.01" step="0.01" type="number" />
              <input v-model="creditForm.remark" class="mt-3 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" placeholder="备注" />
              <button class="mt-4 w-full rounded-lg bg-gradient-to-r from-violet-600 to-blue-600 px-4 py-2 text-white" type="submit">确认充值</button>
            </form>
            <div class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <h3 class="font-semibold">最近生成记录</h3>
              <div v-for="item in userGenerations.items" :key="item.id" class="mt-3 border-t border-slate-100 pt-3 text-sm">
                <div class="font-medium">#{{ item.id }} · {{ item.quality }} · {{ generationStatus(item.status) }}</div>
                <p class="mt-1 line-clamp-2 text-slate-500">{{ item.prompt }}</p>
              </div>
              <p v-if="!userGenerations.items.length" class="mt-3 text-sm text-slate-500">暂无记录。</p>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'channels'" class="grid gap-4 xl:grid-cols-[1fr_420px]">
          <div class="space-y-3">
            <div v-for="channel in channels" :key="channel.id" class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <h3 class="font-semibold">{{ channel.name }}</h3>
                    <span class="rounded-full px-2 py-1 text-xs" :class="channel.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-200 text-slate-600'">{{ statusText(channel.status) }}</span>
                  </div>
                  <p class="mt-1 break-all text-sm text-slate-500">{{ channel.base_url }}</p>
                  <p v-if="channel.remark" class="mt-2 text-sm text-slate-500">{{ channel.remark }}</p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <button class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm" type="button" @click="testChannel(channel)">测试</button>
                  <button class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm" type="button" @click="editChannel(channel)">编辑</button>
                  <button class="rounded-lg border border-red-200 px-3 py-1.5 text-sm text-red-600" type="button" @click="deleteChannel(channel)">删除</button>
                </div>
              </div>
              <div class="mt-3 flex flex-wrap gap-3 text-xs text-slate-500">
                <span>权重 {{ channel.weight }}</span>
                <span>API Key {{ channel.api_key ? '已配置' : '未配置' }}</span>
                <span v-if="channelTestResult[channel.id]">{{ channelTestResult[channel.id] }}</span>
              </div>
            </div>
            <p v-if="!channels.length" class="rounded-xl border border-dashed border-slate-300 bg-white p-8 text-center text-sm text-slate-500">
              暂无渠道。没有启用渠道时，后端会使用 `SUB2API_BASE_URL` 环境变量作为兜底。
            </p>
          </div>

          <form class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm" @submit.prevent="saveChannel">
            <h3 class="font-semibold">{{ channelForm.id ? '编辑渠道' : '新增渠道' }}</h3>
            <label class="mt-4 block text-sm font-medium">
              名称
              <input v-model="channelForm.name" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" required />
            </label>
            <label class="mt-3 block text-sm font-medium">
              Base URL
              <input v-model="channelForm.base_url" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" placeholder="http://sub2api:8080" required />
            </label>
            <label class="mt-3 block text-sm font-medium">
              API Key
              <input v-model="channelForm.api_key" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" />
            </label>
            <label class="mt-3 block text-sm font-medium">
              Headers JSON
              <textarea v-model="channelForm.headers" class="mt-2 min-h-24 w-full rounded-lg border border-slate-300 px-3 py-2" placeholder='{"X-Custom":"value"}' />
            </label>
            <div class="mt-3 grid grid-cols-2 gap-3">
              <label class="text-sm font-medium">
                权重
                <input v-model.number="channelForm.weight" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" min="1" type="number" />
              </label>
              <label class="text-sm font-medium">
                状态
                <select v-model.number="channelForm.status" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2">
                  <option :value="1">启用</option>
                  <option :value="2">禁用</option>
                </select>
              </label>
            </div>
            <label class="mt-3 block text-sm font-medium">
              备注
              <input v-model="channelForm.remark" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" />
            </label>
            <div class="mt-4 flex gap-2">
              <button class="flex-1 rounded-lg bg-slate-950 px-4 py-2 text-white" type="submit">保存渠道</button>
              <button class="rounded-lg border border-slate-200 px-4 py-2" type="button" @click="resetChannel">清空</button>
            </div>
          </form>
        </div>

        <div v-if="activeTab === 'templates'" class="grid gap-4 xl:grid-cols-[1fr_420px]">
          <div class="rounded-xl border border-slate-200 bg-white shadow-sm">
            <div class="border-b border-slate-100 p-4 text-sm text-slate-500">
              启用状态会展示到前台；分类为“首页风格预设”时展示在风格按钮，分类为“首页推荐样例”时展示在推荐样例。
            </div>
            <div v-for="template in templates" :key="template.id" class="flex gap-3 border-b border-slate-100 p-4 text-sm last:border-b-0">
              <div class="min-w-0 flex-1">
                <div class="font-medium">{{ template.label }} · {{ templateCategoryText(template.category) }}</div>
                <div class="mt-1 text-xs" :class="template.status === 1 ? 'text-emerald-700' : 'text-slate-400'">{{ statusText(template.status) }}</div>
                <p class="mt-1 line-clamp-2 text-slate-500">{{ template.prompt }}</p>
              </div>
              <button class="rounded-lg border border-slate-200 px-3 py-1.5" type="button" @click="editTemplate(template)">编辑</button>
              <button class="rounded-lg border border-red-200 px-3 py-1.5 text-red-600" type="button" @click="deleteTemplate(template)">删除</button>
            </div>
          </div>
          <form class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm" @submit.prevent="saveTemplate">
            <h3 class="font-semibold">{{ templateForm.id ? '编辑模板' : '新增模板' }}</h3>
            <input v-model="templateForm.label" class="mt-4 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" placeholder="名称" />
            <select v-model="templateForm.category" class="mt-3 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2">
              <option value="style">首页风格预设</option>
              <option value="sample">首页推荐样例</option>
              <option value="default">默认标签</option>
              <option value="repair">修复标签</option>
            </select>
            <textarea v-model="templateForm.prompt" class="mt-3 min-h-32 w-full rounded-lg border border-slate-300 px-3 py-2" placeholder="Prompt" />
            <div class="mt-3 grid grid-cols-2 gap-3">
              <input v-model.number="templateForm.sort_order" class="rounded-lg border border-slate-300 px-3 py-2" type="number" />
              <select v-model.number="templateForm.status" class="rounded-lg border border-slate-300 px-3 py-2">
                <option :value="1">启用</option>
                <option :value="2">禁用</option>
              </select>
            </div>
            <div class="mt-4 flex gap-2">
              <button class="flex-1 rounded-lg bg-slate-950 px-4 py-2 text-white" type="submit">保存模板</button>
              <button class="rounded-lg border border-slate-200 px-4 py-2" type="button" @click="resetTemplate">清空</button>
            </div>
          </form>
        </div>

        <form v-if="activeTab === 'settings'" class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm" @submit.prevent="saveSettings">
          <div class="grid gap-4 md:grid-cols-2">
            <label v-for="key in settingEntries" :key="key" class="block text-sm">
              <span class="font-medium text-slate-700">{{ settingLabel(key) }}</span>
              <input v-model="settings[key]" :type="settingInputType(key)" class="mt-2 min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2" />
              <span v-if="settingHelp(key)" class="mt-1 block text-xs text-slate-500">{{ settingHelp(key) }}</span>
            </label>
          </div>
          <button class="mt-5 rounded-lg bg-slate-950 px-4 py-2 text-white" type="submit">保存设置</button>
        </form>

        <div v-if="activeTab === 'credits'" class="overflow-x-auto rounded-xl border border-slate-200 bg-white shadow-sm">
          <table class="min-w-full text-sm">
            <thead class="bg-slate-50 text-left text-slate-500">
              <tr>
                <th class="px-4 py-3">用户</th>
                <th class="px-4 py-3">类型</th>
                <th class="px-4 py-3">金额</th>
                <th class="px-4 py-3">余额</th>
                <th class="px-4 py-3">备注</th>
                <th class="px-4 py-3">时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in creditLogs.items" :key="log.id" class="border-t border-slate-100">
                <td class="px-4 py-3">{{ log.user_id }}</td>
                <td class="px-4 py-3">{{ log.type }}</td>
                <td class="px-4 py-3">{{ log.amount }}</td>
                <td class="px-4 py-3">{{ log.balance }}</td>
                <td class="px-4 py-3">{{ log.remark }}</td>
                <td class="px-4 py-3 text-slate-500">{{ fmtTime(log.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="activeTab === 'monitor' && monitor" class="space-y-4">
          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <div v-for="card in overviewCards" :key="card.label" class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <div class="text-sm text-slate-500">{{ card.label }}</div>
              <div class="mt-3 text-3xl font-semibold">{{ card.value }}</div>
              <div class="mt-2 text-sm text-slate-500">{{ card.hint }}</div>
            </div>
          </div>
          <div class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <div class="font-medium">每日积分阈值 {{ monitor.alert_threshold }}</div>
                <div class="mt-1 text-sm text-slate-500">{{ monitor.alert_triggered ? '当前已触发告警条件' : '当前未触发告警条件' }}</div>
              </div>
              <button class="rounded-lg bg-slate-950 px-4 py-2 text-white" type="button" @click="checkMonitorAlert">检查告警</button>
            </div>
          </div>
        </div>

        <div v-if="loading" class="fixed bottom-4 right-4 rounded-lg bg-slate-950 px-4 py-3 text-sm text-white shadow-lg">处理中...</div>
      </main>
    </div>
  </section>
</template>
