<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'
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
const activeSettingGroup = ref('account')
const isCreateUserOpen = ref(false)
const isChannelModalOpen = ref(false)
const isTemplateModalOpen = ref(false)
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
const userForm = ref({ email: '', username: '', password: '', role: 1, status: 1, credits: 0 })
const templateForm = ref<PromptTemplate>({ id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 })
const channelForm = ref<Channel>({ id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' })
const channelTestResult = ref<Record<number, string>>({})
const settingFileInputs = ref<Record<string, HTMLInputElement | null>>({})

const tabs = [
  { id: 'overview', label: '概览', description: '核心指标与运行状态' },
  { id: 'users', label: '用户', description: '账号、角色、状态与充值' },
  { id: 'channels', label: '渠道', description: 'Sub2API 渠道配置与测试' },
  { id: 'templates', label: '模板', description: '提示词模板管理' },
  { id: 'settings', label: '设置', description: '按场景维护系统配置' },
  { id: 'credits', label: '积分', description: '积分流水审计' },
  { id: 'monitor', label: '监控', description: '每日指标和告警检查' },
]

const settingGroups = [
  {
    id: 'account',
    title: '账号与额度',
    description: '注册开关、新用户赠送积分和额度用完后的联系提示。',
    keys: ['register_enabled', 'register_gift_credits', 'credit_exhausted_message', 'credit_exhausted_wechat_qrcode_url', 'credit_exhausted_qq'],
  },
  {
    id: 'wechat',
    title: '微信登录',
    description: '公众号二维码、验证码服务地址和访问凭证。敏感项只在后台展示。',
    keys: ['wechat_auth_enabled', 'wechat_qrcode_url', 'wechat_server_address', 'wechat_server_token'],
  },
  {
    id: 'generation',
    title: '图像生成',
    description: '模型名称和前台可选尺寸比例。',
    keys: ['image_model', 'enabled_image_sizes'],
  },
  {
    id: 'storage',
    title: '图片存储',
    description: 'Cloudflare R2 上传和公开访问地址。',
    keys: ['r2_endpoint', 'r2_access_key', 'r2_secret_key', 'r2_bucket', 'r2_public_url'],
  },
  {
    id: 'captcha',
    title: '人机验证',
    description: 'Cloudflare Turnstile 验证开关和密钥。',
    keys: ['captcha_enabled', 'turnstile_site_key', 'turnstile_secret'],
  },
  {
    id: 'security',
    title: '安全与监控',
    description: 'IP 黑名单和每日消耗告警。',
    keys: ['ip_blacklist', 'monitor_daily_credit_threshold', 'monitor_alert_last_date'],
  },
]

const isAdmin = computed(() => (userStore.user?.role || 0) >= 10)
const enabledChannels = computed(() => channels.value.filter((item) => item.status === 1).length)
const currentTab = computed(() => tabs.find((item) => item.id === activeTab.value) || tabs[0])
const knownSettingKeys = computed(() => new Set(settingGroups.flatMap((group) => group.keys)))
const visibleSettingGroups = computed(() => {
  const groups = settingGroups
    .map((group) => ({
      ...group,
      keys: group.keys.filter((key) => Object.prototype.hasOwnProperty.call(settings.value, key)),
    }))
    .filter((group) => group.keys.length > 0)
  const otherKeys = Object.keys(settings.value)
    .filter((key) => !knownSettingKeys.value.has(key))
    .sort()
  if (otherKeys.length > 0) {
    groups.push({
      id: 'other',
      title: '其他配置',
      description: '暂未归类的系统配置项。',
      keys: otherKeys,
    })
  }
  return groups
})
const activeSettingGroupInfo = computed(() => visibleSettingGroups.value.find((group) => group.id === activeSettingGroup.value) || visibleSettingGroups.value[0])
const activeSettingKeys = computed(() => activeSettingGroupInfo.value?.keys || [])
const overviewCards = computed(() => [
  { label: '今日生成', value: monitor.value?.generation_count ?? 0, hint: `${monitor.value?.completed_count ?? 0} 成功 / ${monitor.value?.failed_count ?? 0} 失败` },
  { label: '新增用户', value: monitor.value?.new_users ?? 0, hint: `当前用户 ${users.value.total}` },
  { label: '积分消耗', value: monitor.value?.credits_consumed ?? 0, hint: `告警阈值 ${monitor.value?.alert_threshold ?? 0}` },
  { label: '启用渠道', value: enabledChannels.value, hint: `共 ${channels.value.length} 个渠道` },
])
const tabCounts = computed<Record<string, number | string>>(() => ({
  users: users.value.total,
  channels: channels.value.length,
  templates: templates.value.length,
  credits: creditLogs.value.total,
  monitor: monitor.value?.alert_triggered ? '!' : '',
}))

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

function resetUserForm() {
  userForm.value = { email: '', username: '', password: '', role: 1, status: 1, credits: 0 }
}

function openCreateUserModal() {
  resetUserForm()
  isCreateUserOpen.value = true
}

function closeCreateUserModal() {
  isCreateUserOpen.value = false
  resetUserForm()
}

async function createUser() {
  await guarded(async () => {
    await api.post('/admin/users', userForm.value)
    closeCreateUserModal()
    await loadUsers()
  }, '用户已创建')
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
  isTemplateModalOpen.value = true
}

function resetTemplate() {
  templateForm.value = { id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 }
}

function openCreateTemplateModal() {
  resetTemplate()
  isTemplateModalOpen.value = true
}

function closeTemplateModal() {
  isTemplateModalOpen.value = false
  resetTemplate()
}

async function saveTemplate() {
  await guarded(async () => {
    const payload = { ...templateForm.value }
    if (payload.id) {
      await api.put(`/admin/prompt-templates/${payload.id}`, payload)
    } else {
      await api.post('/admin/prompt-templates', payload)
    }
    closeTemplateModal()
    await loadTemplates()
  }, '模板已保存')
}

async function deleteTemplate(template: PromptTemplate) {
  if (!window.confirm(`确认删除模板「${template.label}」？`)) {
    return
  }
  await guarded(async () => {
    await api.delete(`/admin/prompt-templates/${template.id}`)
    await loadTemplates()
  }, '模板已删除')
}

async function loadSettings() {
  const response = await api.get('/admin/settings')
  settings.value = response.data.items
  if (!visibleSettingGroups.value.some((group) => group.id === activeSettingGroup.value)) {
    activeSettingGroup.value = visibleSettingGroups.value[0]?.id || 'account'
  }
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
  isChannelModalOpen.value = true
}

function resetChannel() {
  channelForm.value = { id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' }
}

function openCreateChannelModal() {
  resetChannel()
  isChannelModalOpen.value = true
}

function closeChannelModal() {
  isChannelModalOpen.value = false
  resetChannel()
}

async function saveChannel() {
  await guarded(async () => {
    const payload = { ...channelForm.value }
    if (payload.id) {
      await api.put(`/admin/channels/${payload.id}`, payload)
    } else {
      await api.post('/admin/channels', payload)
    }
    closeChannelModal()
    await loadChannels()
  }, '渠道已保存')
}

async function deleteChannel(channel: Channel) {
  if (!window.confirm(`确认删除渠道「${channel.name}」？`)) {
    return
  }
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

async function refreshDashboard() {
  await guarded(async () => {
    await Promise.all([loadUsers(), loadCreditLogs(), loadTemplates(), loadSettings(), loadChannels(), loadMonitor()])
  }, '数据已刷新')
}

function fmtTime(value: string) {
  return value ? new Date(value).toLocaleString() : '-'
}

function fmtNumber(value: number | undefined) {
  return Number(value ?? 0).toLocaleString()
}

function statusText(status: number) {
  return status === 1 ? '启用' : '禁用'
}

function creditTypeText(type: number) {
  const map: Record<number, string> = { 1: '充值', 2: '消费', 3: '退款', 4: '赠送', 5: '注册赠送' }
  return map[type] || `类型 ${type}`
}

function clearSelectedUser() {
  selectedUser.value = null
  userGenerations.value = emptyGenerationPage()
  creditForm.value = { amount: 1, remark: '' }
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
    wechat_auth_enabled: '微信登录开关',
    wechat_qrcode_url: '微信登录二维码',
    wechat_server_address: '微信服务地址',
    wechat_server_token: '微信服务 Token',
    register_gift_credits: '注册赠送积分',
    credit_exhausted_message: '额度用完提示文案',
    credit_exhausted_wechat_qrcode_url: '额度用完微信二维码',
    credit_exhausted_qq: '额度用完 QQ 联系',
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
    image_model: '默认 gpt-image-2。如果 sub2api 要求其他模型名，可在这里切换。',
    enabled_image_sizes: '逗号分隔，默认开放 5 个比例：square 方形 1:1、portrait_3_4 竖版 3:4、story 故事版 9:16、landscape_4_3 横版 4:3、widescreen 宽屏 16:9；后端会映射到 GPT Image 2 合规像素尺寸。',
    wechat_auth_enabled: '控制前台微信登录/注册是否启用。启用后需同时配置二维码、服务地址和 Token。',
    wechat_qrcode_url: '前台登录/注册页展示的微信二维码图片。可填写图片 URL，也可直接上传本地图片。',
    wechat_server_address: '微信验证码服务地址，例如 https://wechat.example.com；后端会请求 /api/wechat/user?code=xxx。',
    wechat_server_token: '请求微信验证码服务时写入 Authorization 头的 Token。',
    register_gift_credits: '新用户注册成功后赠送的积分，默认 10；设置为 0 表示不赠送。',
    credit_exhausted_message: '用户免费额度或积分用完时展示的温馨提示文案。',
    credit_exhausted_wechat_qrcode_url: '额度用完提示展示的联系二维码。可填写图片 URL，也可直接上传本地图片。',
    credit_exhausted_qq: '可填写 QQ 号码或 QQ 群号，额度用完提示会展示联系方式。',
  }
  return map[key] || ''
}

function settingInputType(key: string) {
  if (key.includes('url')) {
    return 'url'
  }
  if (key.includes('credits')) {
    return 'number'
  }
  return key.includes('secret') || key.includes('password') || key.includes('access_key') ? 'password' : 'text'
}

function isTextareaSetting(key: string) {
  return key.includes('message')
}

function isBooleanSetting(key: string) {
  return key.endsWith('_enabled') || key === 'wechat_auth_enabled'
}

function isImageSetting(key: string) {
  return key === 'wechat_qrcode_url' || key === 'credit_exhausted_wechat_qrcode_url'
}

function setSettingFileInput(key: string, el: Element | ComponentPublicInstance | null) {
  settingFileInputs.value[key] = el instanceof HTMLInputElement ? el : null
}

function chooseSettingImage(key: string) {
  settingFileInputs.value[key]?.click()
}

function clearSettingImage(key: string) {
  settings.value[key] = ''
  if (settingFileInputs.value[key]) {
    settingFileInputs.value[key]!.value = ''
  }
}

function handleSettingImageChange(key: string, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }
  if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
    message.value = '二维码图片仅支持 PNG、JPG、WebP'
    input.value = ''
    return
  }
  if (file.size > 512 * 1024) {
    message.value = '二维码图片不能超过 512KB'
    input.value = ''
    return
  }
  const reader = new FileReader()
  reader.onload = () => {
    settings.value[key] = String(reader.result || '')
    message.value = '二维码图片已读取，请点击保存设置'
  }
  reader.onerror = () => {
    message.value = '二维码图片读取失败'
  }
  reader.readAsDataURL(file)
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
  <section class="admin-shell min-h-[calc(100vh-65px)] bg-[#f6f7f9] text-slate-950">
    <div class="grid min-h-[calc(100vh-65px)] lg:grid-cols-[248px_1fr]">
      <aside class="bg-slate-950 text-white">
        <div class="flex items-center justify-between gap-3 px-6 py-6 lg:block">
          <div>
            <p class="text-xs font-semibold uppercase tracking-wide text-teal-300">Console</p>
            <h1 class="mt-1 text-lg font-semibold text-white">来看看巴后台</h1>
          </div>
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-slate-300 lg:mt-4 lg:inline-block">{{ userStore.user?.email }}</span>
        </div>
        <nav class="flex gap-1 overflow-x-auto px-3 pb-4 lg:block lg:space-y-1 lg:overflow-visible">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            class="min-w-24 rounded-md px-3 py-2 text-left text-sm transition lg:w-full"
            :class="activeTab === tab.id ? 'bg-white text-slate-950 shadow-sm' : 'text-slate-300 hover:bg-white/8 hover:text-white'"
            @click="activeTab = tab.id"
          >
            <span class="flex items-center justify-between gap-3">
              <span class="font-medium">{{ tab.label }}</span>
              <span v-if="tabCounts[tab.id]" class="rounded-full px-2 py-0.5 text-xs" :class="activeTab === tab.id ? 'bg-slate-100 text-slate-600' : 'bg-white/10 text-slate-300'">
                {{ tabCounts[tab.id] }}
              </span>
            </span>
          </button>
        </nav>
      </aside>

      <main class="min-w-0 px-4 py-5 sm:px-7 lg:px-10">
        <header class="admin-topbar mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-2xl font-semibold tracking-tight text-slate-950">{{ currentTab.label }}</h2>
            <p class="mt-1 text-sm text-slate-500">{{ currentTab.description }}</p>
          </div>
          <div class="flex items-center gap-2">
            <span v-if="loading" class="text-sm text-slate-500">处理中...</span>
            <button class="admin-btn" type="button" :disabled="loading" @click="refreshDashboard">刷新数据</button>
          </div>
        </header>

        <p v-if="message" class="mb-4 rounded-md border px-4 py-3 text-sm" :class="message.includes('失败') ? 'border-red-200 bg-red-50 text-red-700' : 'border-emerald-200 bg-emerald-50 text-emerald-700'">
          {{ message }}
        </p>

        <div v-if="activeTab === 'overview'" class="space-y-5">
          <div class="admin-metric-grid">
            <div v-for="card in overviewCards" :key="card.label" class="admin-metric">
              <div class="text-xs font-medium uppercase tracking-wide text-slate-400">{{ card.label }}</div>
              <div class="mt-2 text-2xl font-semibold text-slate-950">{{ fmtNumber(Number(card.value)) }}</div>
              <div class="mt-1 text-sm text-slate-500">{{ card.hint }}</div>
            </div>
          </div>
          <div class="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
            <div class="admin-panel p-4">
              <div class="flex items-center justify-between gap-3">
                <h3 class="admin-section-title">渠道状态</h3>
                <button class="admin-btn" type="button" @click="activeTab = 'channels'">管理渠道</button>
              </div>
              <div class="mt-3 divide-y divide-slate-100">
                <div v-for="channel in channels.slice(0, 5)" :key="channel.id" class="flex items-center justify-between gap-3 py-3 text-sm">
                  <div class="min-w-0">
                    <div class="truncate font-medium text-slate-800">{{ channel.name }}</div>
                    <div class="truncate text-xs text-slate-500">{{ channel.base_url }}</div>
                  </div>
                  <span class="admin-badge" :class="channel.status === 1 ? 'admin-badge-ok' : 'admin-badge-muted'">{{ statusText(channel.status) }}</span>
                </div>
                <p v-if="!channels.length" class="py-8 text-center text-sm text-slate-500">暂无渠道配置</p>
              </div>
            </div>
            <div class="admin-panel p-4">
              <div class="flex items-center justify-between gap-3">
                <h3 class="admin-section-title">今日监控</h3>
                <span class="admin-badge" :class="monitor?.alert_triggered ? 'admin-badge-danger' : 'admin-badge-ok'">{{ monitor?.alert_triggered ? '已触发' : '正常' }}</span>
              </div>
              <dl class="mt-4 space-y-3 text-sm">
                <div class="flex justify-between gap-3"><dt class="text-slate-500">支付订单</dt><dd class="font-medium">{{ monitor?.paid_order_count ?? 0 }}</dd></div>
                <div class="flex justify-between gap-3"><dt class="text-slate-500">支付金额</dt><dd class="font-medium">¥{{ monitor?.paid_order_amount ?? 0 }}</dd></div>
                <div class="flex justify-between gap-3"><dt class="text-slate-500">告警阈值</dt><dd class="font-medium">{{ monitor?.alert_threshold ?? 0 }}</dd></div>
              </dl>
              <button class="admin-primary mt-5 w-full" type="button" @click="checkMonitorAlert">检查告警</button>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'users'" class="space-y-4">
          <div class="space-y-4">
          <div class="admin-toolbar">
            <div class="flex flex-col gap-2 sm:flex-row">
              <input v-model="userKeyword" class="admin-input min-w-0 flex-1" placeholder="搜索邮箱或用户名" @keydown.enter.prevent="loadUsers" />
              <button class="admin-primary" type="button" @click="loadUsers">搜索</button>
                <button class="admin-btn" type="button" @click="openCreateUserModal">新建用户</button>
              </div>
            </div>

            <div class="admin-panel overflow-x-auto">
              <table class="admin-table">
                <thead>
                  <tr>
                    <th>用户</th>
                    <th>角色</th>
                    <th>状态</th>
                    <th>积分</th>
                    <th>创建时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="user in users.items" :key="user.id" :class="selectedUser?.id === user.id ? 'bg-teal/5' : ''">
                    <td>
                      <div class="font-medium text-slate-900">{{ user.email }}</div>
                      <div class="text-xs text-slate-500">ID {{ user.id }} · {{ user.username || '未设置用户名' }}</div>
                    </td>
                    <td><span class="admin-badge" :class="user.role >= 10 ? 'admin-badge-info' : 'admin-badge-muted'">{{ user.role >= 10 ? '管理员' : '用户' }}</span></td>
                    <td><span class="admin-badge" :class="user.status === 1 ? 'admin-badge-ok' : 'admin-badge-danger'">{{ user.status === 1 ? '正常' : '封禁' }}</span></td>
                    <td class="font-medium">{{ user.credits }}</td>
                    <td class="text-slate-500">{{ fmtTime(user.created_at) }}</td>
                    <td>
                      <div class="flex flex-wrap gap-1.5">
                        <button class="admin-btn" type="button" @click="loadUserGenerations(user)">记录</button>
                        <button class="admin-btn" type="button" @click="selectedUser = user">充值</button>
                        <button class="admin-btn" type="button" @click="updateUserStatus(user, user.status === 1 ? 2 : 1)">{{ user.status === 1 ? '封禁' : '解封' }}</button>
                        <button class="admin-btn" type="button" @click="updateUserRole(user, user.role >= 10 ? 1 : 10)">{{ user.role >= 10 ? '设为用户' : '设为管理员' }}</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="!users.items.length">
                    <td class="py-12 text-center text-slate-500" colspan="6">没有匹配的用户</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="selectedUser" class="admin-panel p-4">
              <h3 class="admin-section-title">最近生成记录</h3>
              <div v-for="item in userGenerations.items" :key="item.id" class="border-t border-slate-100 py-3 text-sm first:mt-3">
                <div class="font-medium">#{{ item.id }} · {{ item.quality }} · {{ generationStatus(item.status) }}</div>
                <p class="mt-1 line-clamp-2 text-slate-500">{{ item.prompt }}</p>
              </div>
              <p v-if="!userGenerations.items.length" class="py-8 text-center text-sm text-slate-500">暂无记录</p>
            </div>
          </div>

          <div v-if="selectedUser" class="grid gap-4 xl:grid-cols-[360px_1fr]">
            <form class="admin-panel p-4" @submit.prevent="topupCredits(selectedUser)">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h3 class="admin-section-title">用户充值</h3>
                  <p class="mt-1 break-all text-sm text-slate-500">{{ selectedUser.email }}</p>
                </div>
                <button class="admin-btn" type="button" @click="clearSelectedUser">关闭</button>
              </div>
              <label class="admin-label mt-4">金额</label>
              <input v-model.number="creditForm.amount" class="admin-input mt-2 w-full" min="0.01" step="0.01" type="number" />
              <label class="admin-label mt-3">备注</label>
              <input v-model="creditForm.remark" class="admin-input mt-2 w-full" placeholder="备注" />
              <button class="admin-primary mt-4 w-full" type="submit">确认充值</button>
            </form>
          </div>
        </div>

        <div v-if="activeTab === 'channels'" class="space-y-4">
          <div class="admin-toolbar">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-sm text-slate-500">管理 Sub2API 渠道，支持权重、状态和连通性测试。</p>
              <button class="admin-primary" type="button" @click="openCreateChannelModal">新增渠道</button>
            </div>
          </div>
          <div class="admin-list">
            <div v-for="channel in channels" :key="channel.id" class="admin-list-row">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <h3 class="font-semibold text-slate-900">{{ channel.name }}</h3>
                    <span class="admin-badge" :class="channel.status === 1 ? 'admin-badge-ok' : 'admin-badge-muted'">{{ statusText(channel.status) }}</span>
                  </div>
                  <p class="mt-1 break-all text-sm text-slate-500">{{ channel.base_url }}</p>
                  <p v-if="channel.remark" class="mt-2 text-sm text-slate-500">{{ channel.remark }}</p>
                </div>
                <div class="flex flex-wrap gap-1.5">
                  <button class="admin-btn" type="button" @click="testChannel(channel)">测试</button>
                  <button class="admin-btn" type="button" @click="editChannel(channel)">编辑</button>
                  <button class="admin-btn-danger" type="button" @click="deleteChannel(channel)">删除</button>
                </div>
              </div>
              <div class="mt-3 flex flex-wrap gap-2 text-xs text-slate-500">
                <span class="rounded bg-slate-100 px-2 py-1">权重 {{ channel.weight }}</span>
                <span class="rounded bg-slate-100 px-2 py-1">API Key {{ channel.api_key ? '已配置' : '未配置' }}</span>
                <span v-if="channelTestResult[channel.id]" class="rounded bg-slate-100 px-2 py-1">{{ channelTestResult[channel.id] }}</span>
              </div>
            </div>
            <p v-if="!channels.length" class="admin-empty">暂无渠道。没有启用渠道时，后端会使用 SUB2API_BASE_URL 环境变量作为兜底。</p>
          </div>
        </div>

        <div v-if="activeTab === 'templates'" class="space-y-4">
          <div class="admin-toolbar">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-sm text-slate-500">维护首页风格预设和推荐样例，启用后会展示到前台。</p>
              <button class="admin-primary" type="button" @click="openCreateTemplateModal">新增模板</button>
            </div>
          </div>
          <div class="admin-list">
            <div class="px-5 py-4 text-sm text-slate-500">启用状态会展示到前台，排序越小越靠前。</div>
            <div v-for="template in templates" :key="template.id" class="admin-list-row text-sm">
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium text-slate-900">{{ template.label }}</span>
                  <span class="admin-badge admin-badge-info">{{ templateCategoryText(template.category) }}</span>
                  <span class="admin-badge" :class="template.status === 1 ? 'admin-badge-ok' : 'admin-badge-muted'">{{ statusText(template.status) }}</span>
                </div>
                <p class="mt-2 line-clamp-2 text-slate-500">{{ template.prompt }}</p>
              </div>
              <div class="flex shrink-0 gap-1.5">
                <button class="admin-btn" type="button" @click="editTemplate(template)">编辑</button>
                <button class="admin-btn-danger" type="button" @click="deleteTemplate(template)">删除</button>
              </div>
            </div>
            <p v-if="!templates.length" class="admin-empty">暂无提示词模板</p>
          </div>
        </div>

        <form v-if="activeTab === 'settings'" class="space-y-4" @submit.prevent="saveSettings">
          <div class="admin-hero">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wide text-teal">Settings</p>
              <h3 class="mt-2 text-xl font-semibold text-slate-950">配置按使用场景整理</h3>
              <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">左侧选择配置类型，右侧只显示相关字段。保存按钮会一次性保存全部设置，敏感凭证只在管理员后台可见。</p>
            </div>
            <button class="admin-primary shrink-0" type="submit" :disabled="loading">保存全部设置</button>
          </div>

          <div class="grid gap-4 xl:grid-cols-[260px_1fr]">
            <aside class="admin-settings-nav xl:sticky xl:top-4 xl:self-start">
              <button
                v-for="group in visibleSettingGroups"
                :key="group.id"
                type="button"
                class="w-full rounded-md px-3 py-3 text-left transition"
                :class="activeSettingGroup === group.id ? 'bg-slate-950 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-950'"
                @click="activeSettingGroup = group.id"
              >
                <span class="flex items-center justify-between gap-3">
                  <span class="font-medium">{{ group.title }}</span>
                  <span class="rounded-full px-2 py-0.5 text-xs" :class="activeSettingGroup === group.id ? 'bg-white/15 text-white' : 'bg-slate-100 text-slate-500'">{{ group.keys.length }}</span>
                </span>
                <span class="mt-1 block text-xs leading-5" :class="activeSettingGroup === group.id ? 'text-white/65' : 'text-slate-400'">{{ group.description }}</span>
              </button>
            </aside>

            <section class="admin-settings-content">
              <div class="border-b border-slate-200/70 px-6 py-5">
                <h3 class="text-base font-semibold text-slate-950">{{ activeSettingGroupInfo?.title }}</h3>
                <p class="mt-1 text-sm text-slate-500">{{ activeSettingGroupInfo?.description }}</p>
              </div>
              <div class="grid gap-0 divide-y divide-slate-200/70">
                <label v-for="key in activeSettingKeys" :key="key" class="grid gap-3 px-6 py-5 text-sm lg:grid-cols-[240px_1fr] lg:items-start">
                  <span>
                    <span class="block font-medium text-slate-900">{{ settingLabel(key) }}</span>
                    <span v-if="settingHelp(key)" class="mt-1 block text-xs leading-5 text-slate-500">{{ settingHelp(key) }}</span>
                  </span>
                  <span class="min-w-0">
                    <textarea v-if="isTextareaSetting(key)" v-model="settings[key]" class="admin-textarea w-full" />
                    <select v-else-if="isBooleanSetting(key)" v-model="settings[key]" class="admin-input w-full">
                      <option value="true">开启</option>
                      <option value="false">关闭</option>
                    </select>
                    <span v-else-if="isImageSetting(key)" class="block space-y-2">
                      <input v-model="settings[key]" type="text" class="admin-input w-full" placeholder="图片 URL，或点击下方选择本地图片" />
                      <span class="flex flex-wrap gap-2">
                        <button class="admin-btn" type="button" @click="chooseSettingImage(key)">选择图片</button>
                        <button v-if="settings[key]" class="admin-btn-danger" type="button" @click="clearSettingImage(key)">清除图片</button>
                      </span>
                      <input :ref="(el) => setSettingFileInput(key, el)" class="hidden" type="file" accept="image/png,image/jpeg,image/webp" @change="handleSettingImageChange(key, $event)" />
                      <img v-if="settings[key]" class="size-28 rounded-lg border border-slate-200 bg-white object-contain p-1" :src="settings[key]" alt="二维码预览" />
                    </span>
                    <input v-else v-model="settings[key]" :type="settingInputType(key)" class="admin-input w-full" />
                  </span>
                </label>
              </div>
            </section>
          </div>
        </form>

        <div v-if="activeTab === 'credits'" class="admin-panel overflow-x-auto">
          <table class="admin-table">
            <thead>
              <tr>
                <th>用户</th>
                <th>类型</th>
                <th>金额</th>
                <th>余额</th>
                <th>备注</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in creditLogs.items" :key="log.id">
                <td>{{ log.user_id }}</td>
                <td><span class="admin-badge admin-badge-muted">{{ creditTypeText(log.type) }}</span></td>
                <td class="font-medium" :class="log.amount >= 0 ? 'text-emerald-700' : 'text-red-700'">{{ log.amount }}</td>
                <td>{{ log.balance }}</td>
                <td class="max-w-md truncate">{{ log.remark || '-' }}</td>
                <td class="text-slate-500">{{ fmtTime(log.created_at) }}</td>
              </tr>
              <tr v-if="!creditLogs.items.length">
                <td class="py-12 text-center text-slate-500" colspan="6">暂无积分流水</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="activeTab === 'monitor' && monitor" class="space-y-4">
          <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <div v-for="card in overviewCards" :key="card.label" class="admin-panel p-4">
              <div class="text-xs font-medium uppercase tracking-wide text-slate-400">{{ card.label }}</div>
              <div class="mt-2 text-2xl font-semibold">{{ fmtNumber(Number(card.value)) }}</div>
              <div class="mt-1 text-sm text-slate-500">{{ card.hint }}</div>
            </div>
          </div>
          <div class="admin-panel p-4">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <div class="font-medium text-slate-900">每日积分阈值 {{ monitor.alert_threshold }}</div>
                <div class="mt-1 text-sm text-slate-500">{{ monitor.alert_triggered ? '当前已触发告警条件' : '当前未触发告警条件' }}</div>
              </div>
              <button class="admin-primary" type="button" @click="checkMonitorAlert">检查告警</button>
            </div>
          </div>
        </div>
        <p v-else-if="activeTab === 'monitor'" class="admin-panel p-8 text-center text-sm text-slate-500">监控数据加载中</p>

        <div v-if="loading" class="fixed bottom-4 right-4 rounded-md bg-slate-950 px-4 py-3 text-sm text-white shadow-lg">处理中...</div>

        <div v-if="isCreateUserOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="create-user-title" @click.self="closeCreateUserModal">
          <form class="w-full max-w-lg rounded-2xl bg-white p-5 shadow-2xl shadow-slate-950/20" @submit.prevent="createUser">
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">User</p>
                <h3 id="create-user-title" class="mt-1 text-xl font-semibold text-slate-950">新建用户</h3>
                <p class="mt-1 text-sm text-slate-500">用于后台手动开通邮箱账号，创建后可继续充值或调整角色。</p>
              </div>
              <button class="admin-btn" type="button" @click="closeCreateUserModal">关闭</button>
            </div>

            <div class="mt-5 grid gap-4 sm:grid-cols-2">
              <label class="block sm:col-span-2">
                <span class="admin-label">邮箱</span>
                <input v-model="userForm.email" class="admin-input mt-2 w-full" type="email" autocomplete="email" required />
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">用户名</span>
                <input v-model="userForm.username" class="admin-input mt-2 w-full" autocomplete="username" placeholder="可选" />
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">初始密码</span>
                <input v-model="userForm.password" class="admin-input mt-2 w-full" minlength="8" type="password" autocomplete="new-password" required />
              </label>
              <label class="block">
                <span class="admin-label">角色</span>
                <select v-model.number="userForm.role" class="admin-input mt-2 w-full">
                  <option :value="1">用户</option>
                  <option :value="10">管理员</option>
                </select>
              </label>
              <label class="block">
                <span class="admin-label">状态</span>
                <select v-model.number="userForm.status" class="admin-input mt-2 w-full">
                  <option :value="1">正常</option>
                  <option :value="2">封禁</option>
                </select>
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">初始积分</span>
                <input v-model.number="userForm.credits" class="admin-input mt-2 w-full" min="0" step="0.01" type="number" />
              </label>
            </div>

            <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button class="admin-btn" type="button" @click="closeCreateUserModal">取消</button>
              <button class="admin-primary" type="submit" :disabled="loading">创建用户</button>
            </div>
          </form>
        </div>

        <div v-if="isChannelModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="channel-modal-title" @click.self="closeChannelModal">
          <form class="w-full max-w-2xl rounded-2xl bg-white p-5 shadow-2xl shadow-slate-950/20" @submit.prevent="saveChannel">
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">Channel</p>
                <h3 id="channel-modal-title" class="mt-1 text-xl font-semibold text-slate-950">{{ channelForm.id ? '编辑渠道' : '新增渠道' }}</h3>
                <p class="mt-1 text-sm text-slate-500">配置 Sub2API 渠道地址、密钥、请求头和调度权重。</p>
              </div>
              <button class="admin-btn" type="button" @click="closeChannelModal">关闭</button>
            </div>

            <div class="mt-5 grid gap-4 sm:grid-cols-2">
              <label class="block">
                <span class="admin-label">名称</span>
                <input v-model="channelForm.name" class="admin-input mt-2 w-full" required />
              </label>
              <label class="block">
                <span class="admin-label">权重</span>
                <input v-model.number="channelForm.weight" class="admin-input mt-2 w-full" min="1" type="number" />
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">Base URL</span>
                <input v-model="channelForm.base_url" class="admin-input mt-2 w-full" placeholder="http://sub2api:8080" required />
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">API Key</span>
                <input v-model="channelForm.api_key" class="admin-input mt-2 w-full" />
              </label>
              <label class="block">
                <span class="admin-label">状态</span>
                <select v-model.number="channelForm.status" class="admin-input mt-2 w-full">
                  <option :value="1">启用</option>
                  <option :value="2">禁用</option>
                </select>
              </label>
              <label class="block">
                <span class="admin-label">备注</span>
                <input v-model="channelForm.remark" class="admin-input mt-2 w-full" />
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">Headers JSON</span>
                <textarea v-model="channelForm.headers" class="admin-textarea mt-2 w-full" placeholder='{"X-Custom":"value"}' />
              </label>
            </div>

            <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button class="admin-btn" type="button" @click="closeChannelModal">取消</button>
              <button class="admin-primary" type="submit" :disabled="loading">保存渠道</button>
            </div>
          </form>
        </div>

        <div v-if="isTemplateModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="template-modal-title" @click.self="closeTemplateModal">
          <form class="w-full max-w-2xl rounded-2xl bg-white p-5 shadow-2xl shadow-slate-950/20" @submit.prevent="saveTemplate">
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">Prompt</p>
                <h3 id="template-modal-title" class="mt-1 text-xl font-semibold text-slate-950">{{ templateForm.id ? '编辑模板' : '新增模板' }}</h3>
                <p class="mt-1 text-sm text-slate-500">用于前台风格预设、推荐样例和默认标签。</p>
              </div>
              <button class="admin-btn" type="button" @click="closeTemplateModal">关闭</button>
            </div>

            <div class="mt-5 grid gap-4 sm:grid-cols-2">
              <label class="block">
                <span class="admin-label">名称</span>
                <input v-model="templateForm.label" class="admin-input mt-2 w-full" placeholder="名称" required />
              </label>
              <label class="block">
                <span class="admin-label">分类</span>
                <select v-model="templateForm.category" class="admin-input mt-2 w-full">
                  <option value="style">首页风格预设</option>
                  <option value="sample">首页推荐样例</option>
                  <option value="default">默认标签</option>
                  <option value="repair">修复标签</option>
                </select>
              </label>
              <label class="block">
                <span class="admin-label">排序</span>
                <input v-model.number="templateForm.sort_order" class="admin-input mt-2 w-full" type="number" />
              </label>
              <label class="block">
                <span class="admin-label">状态</span>
                <select v-model.number="templateForm.status" class="admin-input mt-2 w-full">
                  <option :value="1">启用</option>
                  <option :value="2">禁用</option>
                </select>
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">Prompt</span>
                <textarea v-model="templateForm.prompt" class="admin-textarea mt-2 w-full" placeholder="Prompt" required />
              </label>
            </div>

            <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button class="admin-btn" type="button" @click="closeTemplateModal">取消</button>
              <button class="admin-primary" type="submit" :disabled="loading">保存模板</button>
            </div>
          </form>
        </div>
      </main>
    </div>
  </section>
</template>

<style scoped>
.admin-panel {
  @apply rounded-2xl border border-slate-200/80 bg-white shadow-sm shadow-slate-900/[0.03];
}

.admin-topbar {
  @apply border-b border-slate-200/80 pb-6;
}

.admin-hero {
  @apply flex flex-col gap-4 rounded-3xl bg-white px-7 py-6 shadow-sm shadow-slate-900/[0.04] ring-1 ring-slate-200/70 sm:flex-row sm:items-center sm:justify-between;
}

.admin-metric-grid {
  @apply grid overflow-hidden rounded-3xl bg-white shadow-sm shadow-slate-900/[0.04] ring-1 ring-slate-200/70 md:grid-cols-2 xl:grid-cols-4;
}

.admin-metric {
  @apply border-b border-slate-200/70 p-6 md:border-r xl:border-b-0;
}

.admin-toolbar {
  @apply rounded-2xl bg-white px-4 py-4 shadow-sm shadow-slate-900/[0.03] ring-1 ring-slate-200/70;
}

.admin-list {
  @apply overflow-hidden rounded-3xl bg-white shadow-sm shadow-slate-900/[0.04] ring-1 ring-slate-200/70;
}

.admin-list-row {
  @apply border-t border-slate-200/70 p-5 first:border-t-0;
}

.admin-empty {
  @apply border-t border-slate-200/70 p-10 text-center text-sm text-slate-500;
}

.admin-settings-nav {
  @apply overflow-hidden rounded-3xl bg-white p-2 shadow-sm shadow-slate-900/[0.04] ring-1 ring-slate-200/70;
}

.admin-settings-content {
  @apply overflow-hidden rounded-3xl bg-white shadow-sm shadow-slate-900/[0.04] ring-1 ring-slate-200/70;
}

.admin-section-title {
  @apply text-sm font-semibold text-slate-950;
}

.admin-input {
  @apply min-h-10 rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-950 outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:opacity-60;
}

.admin-textarea {
  @apply min-h-28 rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-950 outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:opacity-60;
}

.admin-label {
  @apply block text-xs font-medium uppercase tracking-wide text-slate-500;
}

.admin-primary {
  @apply min-h-10 rounded-md bg-slate-950 px-4 py-2 text-sm font-medium text-white shadow-sm shadow-slate-900/10 transition hover:bg-slate-800 disabled:opacity-60;
}

.admin-btn {
  @apply min-h-9 rounded-md border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 disabled:opacity-60;
}

.admin-btn-danger {
  @apply min-h-9 rounded-md border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-600 transition hover:bg-red-50 disabled:opacity-60;
}

.admin-badge {
  @apply inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium;
}

.admin-badge-ok {
  @apply bg-emerald-50 text-emerald-700;
}

.admin-badge-danger {
  @apply bg-red-50 text-red-700;
}

.admin-badge-muted {
  @apply bg-slate-100 text-slate-600;
}

.admin-badge-info {
  @apply bg-blue-50 text-blue-700;
}

.admin-table {
  @apply min-w-full text-sm;
}

.admin-table thead {
  @apply bg-slate-50 text-left text-xs font-medium uppercase tracking-wide text-slate-500;
}

.admin-table th {
  @apply whitespace-nowrap px-4 py-3;
}

.admin-table td {
  @apply whitespace-nowrap border-t border-slate-100 px-4 py-3 align-middle;
}

.admin-table tbody tr {
  @apply transition hover:bg-slate-50;
}
</style>
