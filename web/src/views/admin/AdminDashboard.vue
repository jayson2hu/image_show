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
  last_login_at?: string | null
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

interface FailureReasonSummary {
  category: string
  label: string
  count: number
}

interface RecentFailure {
  id: number
  user_id?: number | null
  prompt: string
  size: string
  error: string
  category: string
  label: string
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
  last_test_at?: string | null
  last_test_success?: boolean
  last_test_status?: number
  last_test_error?: string
  recent_success_count?: number
  recent_failed_count?: number
  recent_failure_rate?: number
}

interface Announcement {
  id: number
  title: string
  content: string
  status: number
  notify_mode: 'silent' | 'popup'
  target: 'all' | 'guest' | 'user' | 'admin'
  sort_order: number
  starts_at?: string | null
  ends_at?: string | null
  read_count?: number
  created_at: string
  updated_at: string
}

interface AnnouncementRead {
  user_id: number
  email: string
  username: string
  role: number
  read_at: string
}

interface MonitorSummary {
  date: string
  generation_count: number
  completed_count: number
  failed_count: number
  failure_rate: number
  credits_consumed: number
  new_users: number
  paid_order_count: number
  paid_order_amount: number
  alert_threshold: number
  alert_triggered: boolean
  failure_reasons?: FailureReasonSummary[]
  recent_failures?: RecentFailure[]
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
const isAnnouncementModalOpen = ref(false)
const isAnnouncementReadsOpen = ref(false)
const isUserRecordsOpen = ref(false)
const isCreditModalOpen = ref(false)
const loading = ref(false)
const message = ref('')
const userKeyword = ref('')
const users = ref<Page<User>>(emptyUserPage())
const selectedUser = ref<User | null>(null)
const userGenerations = ref<Page<Generation>>(emptyGenerationPage())
const creditLogs = ref<Page<CreditLog>>(emptyCreditPage())
const templates = ref<PromptTemplate[]>([])
const channels = ref<Channel[]>([])
const announcements = ref<Announcement[]>([])
const announcementReads = ref<AnnouncementRead[]>([])
const selectedAnnouncement = ref<Announcement | null>(null)
const settings = ref<Record<string, string>>({})
const originalSettings = ref<Record<string, string>>({})
const revealedSettings = ref<Record<string, boolean>>({})
const monitor = ref<MonitorSummary | null>(null)
const creditForm = ref({ amount: 1, remark: '' })
const userForm = ref({ email: '', username: '', password: '', role: 1, status: 1, credits: 0 })
const templateForm = ref<PromptTemplate>({ id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 })
const channelForm = ref<Channel>({ id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' })
const announcementForm = ref<Announcement>({ id: 0, title: '', content: '', notify_mode: 'silent', target: 'all', sort_order: 0, status: 1, starts_at: '', ends_at: '', created_at: '', updated_at: '' })
const channelTestResult = ref<Record<number, string>>({})
const settingFileInputs = ref<Record<string, HTMLInputElement | null>>({})
const refreshDiagnostics = ref<string[]>([])

const tabs = [
  { id: 'overview', label: '概览', description: '核心指标与运行状态' },
  { id: 'users', label: '用户', description: '账号、角色、状态与充值' },
  { id: 'channels', label: '渠道', description: 'Sub2API 渠道配置与测试' },
  { id: 'templates', label: '模板', description: '提示词模板管理' },
  { id: 'settings', label: '设置', description: '按场景维护系统配置' },
  { id: 'announcements', label: '公告', description: '发布前台生成页通知' },
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
  { label: '失败率', value: `${((monitor.value?.failure_rate ?? 0) * 100).toFixed(1)}%`, hint: '今日失败任务占比' },
  { label: '今日生成', value: monitor.value?.generation_count ?? 0, hint: `${monitor.value?.completed_count ?? 0} 成功 / ${monitor.value?.failed_count ?? 0} 失败` },
  { label: '新增用户', value: monitor.value?.new_users ?? 0, hint: `当前用户 ${users.value.total}` },
  { label: '积分消耗', value: monitor.value?.credits_consumed ?? 0, hint: `告警阈值 ${monitor.value?.alert_threshold ?? 0}` },
  { label: '启用渠道', value: enabledChannels.value, hint: `共 ${channels.value.length} 个渠道` },
])
const tabCounts = computed<Record<string, number | string>>(() => ({
  users: users.value.total,
  channels: channels.value.length,
  templates: templates.value.length,
  announcements: announcements.value.length,
  credits: creditLogs.value.total,
  monitor: monitor.value?.alert_triggered ? '!' : '',
}))
const creditPreviewBalance = computed(() => Number(selectedUser.value?.credits || 0) + Number(creditForm.value.amount || 0))
const pageOrigin = computed(() => window.location.origin)
const hasRefreshFailure = computed(() => refreshDiagnostics.value.some((item) => item.includes('：')))

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
    message.value = adminErrorMessage(error)
    throw error
  } finally {
    loading.value = false
  }
}

function adminErrorMessage(error: any) {
  const data = error?.response?.data
  if (data) {
    if (typeof data === 'string') {
      return data
    }
    if (typeof data.message === 'string' && data.message.trim()) {
      return data.message
    }
    if (typeof data.error === 'string' && data.error.trim()) {
      return data.error
    }
    if (typeof data.detail === 'string' && data.detail.trim()) {
      return data.detail
    }
  }
  if (error?.response?.status === 403) {
    return '当前账号没有管理员权限，请重新登录管理员账号'
  }
  if (error?.response?.status === 401) {
    return '登录已失效，请重新登录'
  }
  if (error?.message) {
    return `请求失败：${error.message}`
  }
  return '操作失败，请检查权限或输入'
}

function adminErrorDetail(error: any) {
  const parts = [adminErrorMessage(error).replace(/^请求失败：/, '')]
  const config = error?.config
  const response = error?.response
  if (config?.url) {
    const base = config.baseURL || ''
    parts.push(`url=${base}${config.url}`)
  }
  if (typeof response?.status === 'number') {
    parts.push(`status=${response.status}`)
  }
  if (error?.code) {
    parts.push(`code=${error.code}`)
  }
  if (typeof window !== 'undefined') {
    parts.push(`origin=${window.location.origin}`)
  }
  return parts.join('，')
}

async function loadUsers() {
  const response = await api.get('/admin/users', { params: { keyword: userKeyword.value, pageSize: 20 } })
  users.value = response.data
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
  isUserRecordsOpen.value = true
  const response = await api.get(`/admin/users/${user.id}/generations`, { params: { pageSize: 10 } })
  userGenerations.value = response.data
}

function openCreditModal(user: User) {
  selectedUser.value = user
  creditForm.value = { amount: 1, remark: '' }
  isCreditModalOpen.value = true
}

function closeUserRecordsModal() {
  isUserRecordsOpen.value = false
  userGenerations.value = emptyGenerationPage()
  if (!isCreditModalOpen.value) {
    selectedUser.value = null
  }
}

function closeCreditModal() {
  isCreditModalOpen.value = false
  creditForm.value = { amount: 1, remark: '' }
  if (!isUserRecordsOpen.value) {
    selectedUser.value = null
  }
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
    closeCreditModal()
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
  originalSettings.value = { ...response.data.items }
  if (!visibleSettingGroups.value.some((group) => group.id === activeSettingGroup.value)) {
    activeSettingGroup.value = visibleSettingGroups.value[0]?.id || 'account'
  }
}

async function loadAnnouncements() {
  const response = await api.get('/admin/announcements')
  announcements.value = response.data.items
}

function resetAnnouncement() {
  announcementForm.value = { id: 0, title: '', content: '', notify_mode: 'silent', target: 'all', sort_order: 0, status: 1, starts_at: '', ends_at: '', created_at: '', updated_at: '' }
}

function openCreateAnnouncementModal() {
  resetAnnouncement()
  isAnnouncementModalOpen.value = true
}

function editAnnouncement(item: Announcement) {
  announcementForm.value = {
    ...item,
    notify_mode: item.notify_mode || 'silent',
    target: item.target || 'all',
    starts_at: toDatetimeLocal(item.starts_at),
    ends_at: toDatetimeLocal(item.ends_at),
  }
  isAnnouncementModalOpen.value = true
}

function closeAnnouncementModal() {
  isAnnouncementModalOpen.value = false
  resetAnnouncement()
}

async function saveAnnouncement() {
  const startsAt = toRFC3339(announcementForm.value.starts_at)
  const endsAt = toRFC3339(announcementForm.value.ends_at)
  if (!startsAt && announcementForm.value.starts_at) {
    message.value = '开始时间格式不正确，请重新选择'
    return
  }
  if (!endsAt && announcementForm.value.ends_at) {
    message.value = '结束时间格式不正确，请重新选择'
    return
  }
  if (startsAt && endsAt && new Date(endsAt).getTime() <= new Date(startsAt).getTime()) {
    message.value = '结束时间必须晚于开始时间'
    return
  }
  await guarded(async () => {
    const payload = {
      ...announcementForm.value,
      starts_at: startsAt,
      ends_at: endsAt,
    }
    if (payload.id) {
      await api.put(`/admin/announcements/${payload.id}`, payload)
    } else {
      await api.post('/admin/announcements', payload)
    }
    closeAnnouncementModal()
    await loadAnnouncements()
  }, '公告已保存')
}

function toDatetimeLocal(value?: string | null) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function toRFC3339(value?: string | null) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toISOString()
}

function announcementTargetText(target?: string) {
  const map: Record<string, string> = {
    all: '全部',
    guest: '仅访客',
    user: '登录用户',
    admin: '管理员',
  }
  return map[target || 'all'] || '全部'
}

async function deleteAnnouncement(item: Announcement) {
  if (!window.confirm(`确认删除公告「${item.title}」？`)) {
    return
  }
  await guarded(async () => {
    await api.delete(`/admin/announcements/${item.id}`)
    await loadAnnouncements()
  }, '公告已删除')
}

async function openAnnouncementReads(item: Announcement) {
  selectedAnnouncement.value = item
  isAnnouncementReadsOpen.value = true
  announcementReads.value = []
  await guarded(async () => {
    const response = await api.get(`/admin/announcements/${item.id}/reads`)
    announcementReads.value = response.data.items || []
  })
}

async function saveSettings() {
  const keysToSave = activeSettingKeys.value
  const payload = keysToSave.reduce<Record<string, string>>((items, key) => {
    items[key] = settings.value[key] || ''
    return items
  }, {})
  const changedSecretKeys = keysToSave.filter((key) => isSecretSetting(key) && payload[key] !== originalSettings.value[key])
  if (changedSecretKeys.length > 0) {
    const labels = changedSecretKeys.map((key) => settingLabel(key)).join('、')
    const confirmed = window.confirm(`你正在修改敏感配置：${labels}。\n这些配置可能影响微信登录、图片存储、人机验证或上游访问，请确认已经核对无误。是否继续保存？`)
    if (!confirmed) {
      return
    }
  }
  await guarded(async () => {
    await api.put('/admin/settings', { items: payload })
    originalSettings.value = { ...originalSettings.value, ...payload }
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
  await loadChannels()
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
  loading.value = true
  message.value = ''
  refreshDiagnostics.value = []
  const tasks = [
    { label: '用户', run: loadUsers },
    { label: '积分流水', run: loadCreditLogs },
    { label: '提示词模板', run: loadTemplates },
    { label: '系统设置', run: loadSettings },
    { label: '渠道', run: loadChannels },
    { label: '公告', run: loadAnnouncements },
    { label: '监控', run: loadMonitor },
  ]
  const results = await Promise.allSettled(tasks.map((task) => task.run()))
  const failed = results
    .map((result, index) => ({ result, task: tasks[index] }))
    .filter((item): item is { result: PromiseRejectedResult; task: typeof tasks[number] } => item.result.status === 'rejected')
  if (failed.length) {
    const details = failed.map((item) => {
      const detail = adminErrorDetail(item.result.reason)
      refreshDiagnostics.value.push(`${item.task.label}：${detail}`)
      return `${item.task.label}：${detail}`
    }).join('；')
    message.value = `部分数据刷新失败，${details}`
  } else {
    refreshDiagnostics.value = ['所有后台数据接口刷新成功']
    message.value = '数据已刷新'
  }
  loading.value = false
}

async function initialLoadDashboard() {
  loading.value = true
  message.value = ''
  refreshDiagnostics.value = []
  const tasks = [
    { label: '用户', run: loadUsers },
    { label: '积分流水', run: loadCreditLogs },
    { label: '提示词模板', run: loadTemplates },
    { label: '系统设置', run: loadSettings },
    { label: '渠道', run: loadChannels },
    { label: '公告', run: loadAnnouncements },
    { label: '监控', run: loadMonitor },
  ]
  const results = await Promise.allSettled(tasks.map((task) => task.run()))
  const failed = results
    .map((result, index) => ({ result, task: tasks[index] }))
    .filter((item): item is { result: PromiseRejectedResult; task: typeof tasks[number] } => item.result.status === 'rejected')
  if (failed.length) {
    const details = failed.map((item) => {
      const detail = adminErrorDetail(item.result.reason)
      refreshDiagnostics.value.push(`${item.task.label}：${detail}`)
      return `${item.task.label}：${detail}`
    }).join('；')
    message.value = `部分数据加载失败，${details}`
  } else {
    refreshDiagnostics.value = ['所有后台数据接口加载成功']
  }
  loading.value = false
}

function fmtTime(value?: string | null) {
  return value ? new Date(value).toLocaleString() : '-'
}

function fmtNumber(value: number | undefined) {
  return Number(value ?? 0).toLocaleString()
}

function userInitial(user: User) {
  return (user.email || user.username || 'U').charAt(0).toUpperCase()
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
  isUserRecordsOpen.value = false
  isCreditModalOpen.value = false
  userGenerations.value = emptyGenerationPage()
  creditForm.value = { amount: 1, remark: '' }
}

function templateCategoryText(category: string) {
  const map: Record<string, string> = {
    default: '默认标签',
    repair: '修复标签',
    style: '首页风格预设',
    sample: '首页推荐样例',
    scenario: '首页场景入口',
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
    register_gift_credits: '新用户注册赠送积分',
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
    register_gift_credits: '控制新账号注册成功后的初始额度。示例：10 表示注册即送 10 积分；0 表示不赠送。',
    credit_exhausted_message: '用户免费额度或积分用完时展示的温馨提示文案。',
    credit_exhausted_wechat_qrcode_url: '额度用完提示展示的联系二维码。可填写图片 URL，也可直接上传本地图片。',
    credit_exhausted_qq: '可填写 QQ 号码或 QQ 群号，额度用完提示会展示联系方式。',
    ip_blacklist: '拦截不允许访问的来源 IP。每行一个 IP 或 CIDR，例如 192.168.1.10、10.0.0.0/24；留空表示不拦截。',
    monitor_daily_credit_threshold: '每日积分消耗告警阈值。示例：500 表示当天消耗达到 500 积分后触发告警检查；0 表示不告警。',
    monitor_alert_last_date: '系统记录的最近告警日期，格式示例 2026-05-07。通常无需手动填写，清空后当天可重新触发告警。',
  }
  return map[key] || ''
}

function settingPlaceholder(key: string) {
  const map: Record<string, string> = {
    ip_blacklist: '192.168.1.10\n10.0.0.0/24\n203.0.113.5',
    enabled_image_sizes: 'square,portrait_3_4,story,landscape_4_3,widescreen',
    register_gift_credits: '10',
    monitor_daily_credit_threshold: '500',
    monitor_alert_last_date: '2026-05-07',
  }
  return map[key] || ''
}

function isSecretSetting(key: string) {
  return key.includes('secret') || key.includes('password') || key.includes('access_key') || key.endsWith('_token')
}

function toggleSettingReveal(key: string) {
  revealedSettings.value[key] = !revealedSettings.value[key]
}

function settingInputType(key: string) {
  if (key.includes('url')) {
    return 'url'
  }
  if (key.includes('credits')) {
    return 'number'
  }
  return isSecretSetting(key) && !revealedSettings.value[key] ? 'password' : 'text'
}

function isTextareaSetting(key: string) {
  return key.includes('message') || key === 'ip_blacklist'
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
  await initialLoadDashboard()
})
</script>

<template>
  <section class="admin-shell min-h-screen bg-gray-50 text-slate-950">
    <div class="pointer-events-none fixed inset-0 admin-bg-mesh"></div>
    <div class="relative grid min-h-screen lg:grid-cols-[256px_1fr]">
      <aside class="admin-sidebar">
        <div class="admin-sidebar-header">
          <div class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-teal to-blue-500 font-bold text-white shadow-lg shadow-teal/20">
            来
          </div>
          <div class="min-w-0">
            <h1 class="truncate text-lg font-bold text-slate-950">来看看巴后台</h1>
            <p class="mt-0.5 text-xs text-slate-500">Console</p>
          </div>
        </div>
        <nav class="admin-sidebar-nav">
          <div class="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-slate-400">管理菜单</div>
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            class="admin-sidebar-link"
            :class="activeTab === tab.id ? 'admin-sidebar-link-active' : ''"
            @click="activeTab = tab.id"
          >
            <span class="min-w-0 truncate">{{ tab.label }}</span>
            <span v-if="tabCounts[tab.id]" class="rounded-full px-2 py-0.5 text-xs" :class="activeTab === tab.id ? 'bg-teal/10 text-teal' : 'bg-slate-100 text-slate-500'">
              {{ tabCounts[tab.id] }}
            </span>
          </button>
        </nav>
        <div class="admin-sidebar-footer">
          <div class="truncate text-sm font-medium text-slate-700">{{ userStore.user?.email }}</div>
          <p class="mt-1 text-xs text-slate-400">管理员账号</p>
        </div>
      </aside>

      <main class="min-w-0 p-4 md:p-6 lg:p-8">
        <header class="admin-topbar mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
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
        <div
          v-if="refreshDiagnostics.length"
          class="mb-4 rounded-xl border px-4 py-3 text-sm shadow-sm"
          :class="hasRefreshFailure ? 'border-amber-200 bg-amber-50 text-amber-900' : 'border-emerald-100 bg-emerald-50 text-emerald-900'"
        >
          <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div class="flex min-w-0 gap-3">
              <span class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full" :class="hasRefreshFailure ? 'bg-amber-100 text-amber-700' : 'bg-emerald-100 text-emerald-700'">
                {{ hasRefreshFailure ? '!' : '✓' }}
              </span>
              <div class="min-w-0">
                <div class="font-semibold">{{ hasRefreshFailure ? '部分模块需要检查' : '后台数据已同步' }}</div>
                <div class="mt-1 text-xs opacity-80">页面来源：{{ pageOrigin }}</div>
              </div>
            </div>
            <button v-if="hasRefreshFailure" class="shrink-0 rounded-lg border border-amber-200 bg-white/70 px-3 py-1.5 text-xs font-medium text-amber-800 transition hover:bg-white" type="button" @click="refreshDiagnostics = []">
              收起
            </button>
          </div>
          <ul v-if="hasRefreshFailure" class="mt-3 space-y-1.5 border-t border-amber-200/70 pt-3 text-xs">
            <li v-for="item in refreshDiagnostics" :key="item" class="break-all rounded-lg bg-white/60 px-3 py-2">{{ item }}</li>
          </ul>
        </div>

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
          <div class="admin-toolbar">
            <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
              <div class="flex min-w-0 flex-1 flex-col gap-2 sm:flex-row sm:items-center">
                <div class="relative min-w-0 flex-1 sm:max-w-sm">
                  <svg class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 21-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
                  </svg>
                  <input v-model="userKeyword" class="admin-input w-full pl-9" placeholder="搜索邮箱或用户名" @keydown.enter.prevent="loadUsers" />
                </div>
                <button class="admin-btn" type="button" @click="guarded(loadUsers, '用户列表已刷新')">刷新 / 搜索</button>
              </div>
              <button class="admin-primary" type="button" @click="openCreateUserModal">新建用户</button>
            </div>
          </div>

          <div class="admin-panel overflow-hidden">
            <div class="border-b border-slate-100 px-5 py-4">
              <div class="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
                <div>
                  <h3 class="text-base font-semibold text-slate-950">用户列表</h3>
                  <p class="mt-1 text-sm text-slate-500">管理账号状态、角色、积分和最近使用情况。</p>
                </div>
                <p class="text-sm text-slate-500">共 {{ users.total }} 个用户</p>
              </div>
            </div>

            <div class="admin-table-scroll">
              <table class="admin-table">
                <thead>
                  <tr>
                    <th>用户</th>
                    <th>角色</th>
                    <th>状态</th>
                    <th>积分</th>
                    <th>最近使用</th>
                    <th>创建时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="user in users.items" :key="user.id" :class="selectedUser?.id === user.id ? 'bg-teal/5' : ''">
                    <td>
                      <div class="flex items-center gap-3">
                        <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-slate-100 text-sm font-semibold text-slate-700">
                          {{ userInitial(user) }}
                        </div>
                        <div class="min-w-0">
                          <div class="truncate font-medium text-slate-900">{{ user.email }}</div>
                          <div class="truncate text-xs text-slate-500">ID {{ user.id }} · {{ user.username || '未设置用户名' }}</div>
                        </div>
                      </div>
                    </td>
                    <td><span class="admin-badge" :class="user.role >= 10 ? 'admin-badge-info' : 'admin-badge-muted'">{{ user.role >= 10 ? '管理员' : '用户' }}</span></td>
                    <td><span class="admin-badge" :class="user.status === 1 ? 'admin-badge-ok' : 'admin-badge-danger'">{{ user.status === 1 ? '正常' : '封禁' }}</span></td>
                    <td class="font-medium">{{ user.credits }}</td>
                    <td class="text-slate-500">{{ fmtTime(user.last_login_at) }}</td>
                    <td class="text-slate-500">{{ fmtTime(user.created_at) }}</td>
                    <td>
                      <div class="flex flex-nowrap gap-1.5">
                        <button class="admin-btn" type="button" @click="loadUserGenerations(user)">记录</button>
                        <button class="admin-btn" type="button" @click="openCreditModal(user)">充值</button>
                        <button class="admin-btn" type="button" @click="updateUserStatus(user, user.status === 1 ? 2 : 1)">{{ user.status === 1 ? '封禁' : '解封' }}</button>
                        <button class="admin-btn" type="button" @click="updateUserRole(user, user.role >= 10 ? 1 : 10)">{{ user.role >= 10 ? '设为用户' : '设为管理员' }}</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="!users.items.length">
                    <td class="py-12 text-center text-slate-500" colspan="7">没有匹配的用户</td>
                  </tr>
                </tbody>
              </table>
            </div>
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
                <span v-if="channel.last_test_at" class="rounded bg-slate-100 px-2 py-1">最近测试 {{ fmtTime(channel.last_test_at) }}</span>
                <span v-if="channel.last_test_at" class="rounded px-2 py-1" :class="channel.last_test_success ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'">{{ channel.last_test_success ? '测试可用' : (channel.last_test_error || `测试失败 ${channel.last_test_status || ''}`) }}</span>
                <span v-if="channelTestResult[channel.id]" class="rounded bg-slate-100 px-2 py-1">{{ channelTestResult[channel.id] }}</span>
              </div>
              <div class="mt-3 grid gap-2 text-xs sm:grid-cols-3">
                <div class="rounded-lg border border-slate-100 bg-slate-50 px-3 py-2">
                  <div class="text-slate-500">近 24 小时成功</div>
                  <div class="mt-1 text-base font-semibold text-slate-900">{{ channel.recent_success_count ?? 0 }}</div>
                </div>
                <div class="rounded-lg border border-slate-100 bg-slate-50 px-3 py-2">
                  <div class="text-slate-500">近 24 小时失败</div>
                  <div class="mt-1 text-base font-semibold text-slate-900">{{ channel.recent_failed_count ?? 0 }}</div>
                </div>
                <div class="rounded-lg border px-3 py-2" :class="(channel.recent_failure_rate ?? 0) > 0 ? 'border-red-100 bg-red-50' : 'border-slate-100 bg-slate-50'">
                  <div :class="(channel.recent_failure_rate ?? 0) > 0 ? 'text-red-600' : 'text-slate-500'">近 24 小时失败率</div>
                  <div class="mt-1 text-base font-semibold" :class="(channel.recent_failure_rate ?? 0) > 0 ? 'text-red-700' : 'text-slate-900'">{{ ((channel.recent_failure_rate ?? 0) * 100).toFixed(1) }}%</div>
                </div>
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
                :class="activeSettingGroup === group.id ? 'bg-teal/10 text-teal shadow-sm' : 'text-slate-600 hover:bg-gray-50 hover:text-slate-950'"
                @click="activeSettingGroup = group.id"
              >
                <span class="flex items-center justify-between gap-3">
                  <span class="font-medium">{{ group.title }}</span>
                  <span class="rounded-full px-2 py-0.5 text-xs" :class="activeSettingGroup === group.id ? 'bg-teal/10 text-teal' : 'bg-gray-100 text-slate-500'">{{ group.keys.length }}</span>
                </span>
                <span class="mt-1 block text-xs leading-5" :class="activeSettingGroup === group.id ? 'text-teal/75' : 'text-slate-400'">{{ group.description }}</span>
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
                    <textarea v-if="isTextareaSetting(key)" v-model="settings[key]" class="admin-textarea w-full" :placeholder="settingPlaceholder(key)" />
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
                    <span v-else-if="isSecretSetting(key)" class="flex min-w-0 gap-2">
                      <input v-model="settings[key]" :type="settingInputType(key)" class="admin-input min-w-0 flex-1" :placeholder="settingPlaceholder(key)" />
                      <button class="admin-icon-btn" type="button" :aria-label="revealedSettings[key] ? '隐藏敏感内容' : '显示敏感内容'" @click="toggleSettingReveal(key)">
                        <svg v-if="!revealedSettings[key]" class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 12s3.75-6.75 9.75-6.75S21.75 12 21.75 12s-3.75 6.75-9.75 6.75S2.25 12 2.25 12Z" />
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15.25A3.25 3.25 0 1 0 12 8.75a3.25 3.25 0 0 0 0 6.5Z" />
                        </svg>
                        <svg v-else class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m3 3 18 18M10.7 5.5A8.3 8.3 0 0 1 12 5.4c6 0 9.75 6.6 9.75 6.6a18 18 0 0 1-2.7 3.45M6.1 6.8C3.7 8.6 2.25 12 2.25 12s3.75 6.75 9.75 6.75c1.7 0 3.2-.53 4.48-1.28" />
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.9 9.9A3.25 3.25 0 0 0 14.1 14.1" />
                        </svg>
                      </button>
                    </span>
                    <input v-else v-model="settings[key]" :type="settingInputType(key)" class="admin-input w-full" :placeholder="settingPlaceholder(key)" />
                  </span>
                </label>
              </div>
            </section>
          </div>
        </form>

        <div v-if="activeTab === 'announcements'" class="space-y-4">
          <div class="admin-toolbar">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-sm text-slate-500">发布给生成图片页面的通知。启用状态下，前台会展示排序最靠前且最近更新的一条公告。</p>
              <button class="admin-primary" type="button" @click="openCreateAnnouncementModal">发布公告</button>
            </div>
          </div>
          <div class="admin-list">
            <div v-for="item in announcements" :key="item.id" class="admin-list-row">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="font-semibold text-slate-900">{{ item.title }}</h3>
                    <span class="admin-badge" :class="item.status === 1 ? 'admin-badge-ok' : 'admin-badge-muted'">{{ statusText(item.status) }}</span>
                    <span class="admin-badge" :class="item.notify_mode === 'popup' ? 'admin-badge-warning' : 'admin-badge-muted'">{{ item.notify_mode === 'popup' ? '弹窗提醒' : '仅公告中心' }}</span>
                    <span class="admin-badge admin-badge-info">{{ announcementTargetText(item.target) }}</span>
                    <span class="admin-badge admin-badge-muted">已读 {{ item.read_count || 0 }}</span>
                    <span class="admin-badge admin-badge-muted">排序 {{ item.sort_order }}</span>
                  </div>
                  <p class="mt-2 whitespace-pre-line text-sm leading-6 text-slate-600">{{ item.content }}</p>
                  <p class="mt-2 text-xs text-slate-400">展示时间：{{ fmtTime(item.starts_at) }} - {{ fmtTime(item.ends_at) }} · 更新时间：{{ fmtTime(item.updated_at) }}</p>
                </div>
                <div class="flex shrink-0 gap-1.5">
                  <button class="admin-btn" type="button" @click="openAnnouncementReads(item)">阅读</button>
                  <button class="admin-btn" type="button" @click="editAnnouncement(item)">编辑</button>
                  <button class="admin-btn-danger" type="button" @click="deleteAnnouncement(item)">删除</button>
                </div>
              </div>
            </div>
            <p v-if="!announcements.length" class="admin-empty">暂无公告。发布后会在生成图片页面展示给用户。</p>
          </div>
        </div>

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
              <div class="mt-2 text-2xl font-semibold">{{ typeof card.value === 'number' ? fmtNumber(card.value) : card.value }}</div>
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

        <div v-if="activeTab === 'monitor' && monitor" class="mt-4 grid gap-4 xl:grid-cols-[360px_1fr]">
          <div class="admin-panel p-4">
            <div class="font-medium text-slate-900">失败原因</div>
            <div class="mt-3 space-y-2">
              <div v-for="reason in monitor.failure_reasons || []" :key="reason.category" class="flex items-center justify-between gap-3 rounded-lg bg-slate-50 px-3 py-2 text-sm">
                <span class="text-slate-700">{{ reason.label }}</span>
                <span class="font-semibold text-slate-950">{{ reason.count }}</span>
              </div>
              <p v-if="!(monitor.failure_reasons || []).length" class="rounded-lg bg-slate-50 px-3 py-6 text-center text-sm text-slate-500">今日暂无失败任务</p>
            </div>
          </div>
          <div class="admin-panel overflow-hidden">
            <div class="border-b border-slate-200/70 px-4 py-3 font-medium text-slate-900">最近失败任务</div>
            <div class="divide-y divide-slate-200/70">
              <div v-for="failure in monitor.recent_failures || []" :key="failure.id" class="grid gap-3 px-4 py-3 text-sm lg:grid-cols-[120px_120px_1fr]">
                <div class="text-slate-500">{{ fmtTime(failure.created_at) }}</div>
                <div><span class="admin-badge admin-badge-danger">{{ failure.label }}</span></div>
                <div class="min-w-0">
                  <div class="truncate font-medium text-slate-900">#{{ failure.id }} · {{ failure.size }}</div>
                  <p class="mt-1 line-clamp-2 text-slate-500">{{ failure.error || failure.prompt || '暂无错误摘要' }}</p>
                </div>
              </div>
              <p v-if="!(monitor.recent_failures || []).length" class="px-4 py-8 text-center text-sm text-slate-500">暂无失败记录</p>
            </div>
          </div>
        </div>

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

        <div v-if="isUserRecordsOpen && selectedUser" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="user-records-title" @click.self="closeUserRecordsModal">
          <section class="flex max-h-[86vh] w-full max-w-3xl flex-col rounded-2xl bg-white shadow-2xl shadow-slate-950/20">
            <div class="flex items-start justify-between gap-4 border-b border-slate-100 p-5">
              <div class="min-w-0">
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">Records</p>
                <h3 id="user-records-title" class="mt-1 text-xl font-semibold text-slate-950">生成记录</h3>
                <p class="mt-1 break-all text-sm text-slate-500">{{ selectedUser.email }}</p>
              </div>
              <button class="admin-btn" type="button" @click="closeUserRecordsModal">关闭</button>
            </div>
            <div class="min-h-0 overflow-y-auto px-5 py-2">
              <div v-for="item in userGenerations.items" :key="item.id" class="border-b border-slate-100 py-4 text-sm last:border-b-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-semibold text-slate-900">#{{ item.id }}</span>
                  <span class="admin-badge admin-badge-muted">{{ item.quality }}</span>
                  <span class="admin-badge" :class="item.status === 3 ? 'admin-badge-ok' : item.status === 4 ? 'admin-badge-danger' : 'admin-badge-info'">{{ generationStatus(item.status) }}</span>
                  <span class="text-xs text-slate-400">{{ fmtTime(item.created_at) }}</span>
                </div>
                <p class="mt-2 line-clamp-3 leading-6 text-slate-600">{{ item.prompt }}</p>
                <a v-if="item.image_url" class="mt-2 inline-flex text-sm font-medium text-blue-700 hover:text-blue-900" :href="item.image_url" target="_blank" rel="noreferrer">查看图片</a>
              </div>
              <p v-if="!userGenerations.items.length" class="py-12 text-center text-sm text-slate-500">暂无生成记录</p>
            </div>
          </section>
        </div>

        <div v-if="isCreditModalOpen && selectedUser" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="credit-modal-title" @click.self="closeCreditModal">
          <form class="w-full max-w-md rounded-2xl bg-white p-5 shadow-2xl shadow-slate-950/20" @submit.prevent="topupCredits(selectedUser)">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">Credits</p>
                <h3 id="credit-modal-title" class="mt-1 text-xl font-semibold text-slate-950">用户充值</h3>
                <p class="mt-1 break-all text-sm text-slate-500">{{ selectedUser.email }}</p>
              </div>
              <button class="admin-btn" type="button" @click="closeCreditModal">关闭</button>
            </div>

            <div class="mt-5 rounded-2xl bg-slate-50 px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-slate-200 text-base font-semibold text-slate-700">{{ userInitial(selectedUser) }}</div>
                <div class="min-w-0 flex-1">
                  <p class="truncate font-medium text-slate-900">{{ selectedUser.email }}</p>
                  <p class="text-sm text-slate-500">当前积分：{{ selectedUser.credits }}</p>
                </div>
              </div>
            </div>

            <label class="block mt-5">
              <span class="admin-label">充值积分</span>
              <input v-model.number="creditForm.amount" class="admin-input mt-2 w-full" min="0.01" step="0.01" type="number" />
            </label>
            <label class="block mt-4">
              <span class="admin-label">备注</span>
              <input v-model="creditForm.remark" class="admin-input mt-2 w-full" placeholder="例如：后台手动补偿" />
            </label>

            <div v-if="creditForm.amount > 0" class="mt-4 rounded-xl border border-blue-100 bg-blue-50 px-4 py-3 text-sm">
              <div class="flex items-center justify-between gap-3">
                <span class="text-slate-600">充值后积分</span>
                <span class="font-semibold text-slate-950">{{ creditPreviewBalance }}</span>
              </div>
            </div>

            <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button class="admin-btn" type="button" @click="closeCreditModal">取消</button>
              <button class="admin-primary" type="submit" :disabled="loading">确认充值</button>
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
                  <option value="scenario">首页场景入口</option>
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

        <div v-if="isAnnouncementModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="announcement-modal-title" @click.self="closeAnnouncementModal">
          <form class="w-full max-w-2xl rounded-2xl bg-white p-5 shadow-2xl shadow-slate-950/20" @submit.prevent="saveAnnouncement">
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">Notice</p>
                <h3 id="announcement-modal-title" class="mt-1 text-xl font-semibold text-slate-950">{{ announcementForm.id ? '编辑公告' : '发布公告' }}</h3>
                <p class="mt-1 text-sm text-slate-500">用于生成图片页面的公告通知，适合维护、活动、额度说明等短消息。</p>
              </div>
              <button class="admin-btn" type="button" @click="closeAnnouncementModal">关闭</button>
            </div>

            <div class="mt-5 grid gap-4 sm:grid-cols-2">
              <label class="block sm:col-span-2">
                <span class="admin-label">标题</span>
                <input v-model="announcementForm.title" class="admin-input mt-2 w-full" placeholder="例如：系统维护通知" required />
              </label>
              <label class="block">
                <span class="admin-label">排序</span>
                <input v-model.number="announcementForm.sort_order" class="admin-input mt-2 w-full" type="number" />
              </label>
              <label class="block">
                <span class="admin-label">状态</span>
                <select v-model.number="announcementForm.status" class="admin-input mt-2 w-full">
                  <option :value="1">启用</option>
                  <option :value="2">禁用</option>
                </select>
              </label>
              <label class="block">
                <span class="admin-label">通知方式</span>
                <select v-model="announcementForm.notify_mode" class="admin-input mt-2 w-full">
                  <option value="silent">仅公告中心</option>
                  <option value="popup">未读弹窗提醒</option>
                </select>
              </label>
              <label class="block">
                <span class="admin-label">投放范围</span>
                <select v-model="announcementForm.target" class="admin-input mt-2 w-full">
                  <option value="all">全部</option>
                  <option value="guest">仅访客</option>
                  <option value="user">登录用户</option>
                  <option value="admin">管理员</option>
                </select>
                <span class="mt-1 block text-xs text-slate-500">第二阶段先支持按登录状态和角色投放。</span>
              </label>
              <label class="block">
                <span class="admin-label">开始时间</span>
                <input v-model="announcementForm.starts_at" class="admin-input mt-2 w-full" type="datetime-local" />
                <span class="mt-1 block text-xs text-slate-500">留空表示立即展示。</span>
              </label>
              <label class="block">
                <span class="admin-label">结束时间</span>
                <input v-model="announcementForm.ends_at" class="admin-input mt-2 w-full" type="datetime-local" />
                <span class="mt-1 block text-xs text-slate-500">留空表示不过期。</span>
              </label>
              <label class="block sm:col-span-2">
                <span class="admin-label">内容</span>
                <textarea v-model="announcementForm.content" class="admin-textarea mt-2 w-full" placeholder="写给前台用户看的公告内容" required />
              </label>
            </div>

            <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button class="admin-btn" type="button" @click="closeAnnouncementModal">取消</button>
              <button class="admin-primary" type="submit" :disabled="loading">保存公告</button>
            </div>
          </form>
        </div>

        <div v-if="isAnnouncementReadsOpen && selectedAnnouncement" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4" role="dialog" aria-modal="true" aria-labelledby="announcement-reads-title" @click.self="isAnnouncementReadsOpen = false">
          <section class="flex max-h-[86vh] w-full max-w-2xl flex-col rounded-2xl bg-white shadow-2xl shadow-slate-950/20">
            <div class="flex items-start justify-between gap-4 border-b border-slate-100 p-5">
              <div class="min-w-0">
                <p class="text-xs font-semibold uppercase tracking-wide text-teal">Reads</p>
                <h3 id="announcement-reads-title" class="mt-1 text-xl font-semibold text-slate-950">阅读明细</h3>
                <p class="mt-1 truncate text-sm text-slate-500">{{ selectedAnnouncement.title }} · {{ announcementTargetText(selectedAnnouncement.target) }}</p>
              </div>
              <button class="admin-btn" type="button" @click="isAnnouncementReadsOpen = false">关闭</button>
            </div>
            <div class="min-h-0 overflow-y-auto">
              <table class="admin-table">
                <thead>
                  <tr>
                    <th>用户</th>
                    <th>角色</th>
                    <th>阅读时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in announcementReads" :key="`${item.user_id}-${item.read_at}`">
                    <td>
                      <div class="font-medium text-slate-900">{{ item.email || '-' }}</div>
                      <div class="text-xs text-slate-500">ID {{ item.user_id }} · {{ item.username || '未设置用户名' }}</div>
                    </td>
                    <td><span class="admin-badge" :class="item.role >= 10 ? 'admin-badge-info' : 'admin-badge-muted'">{{ item.role >= 10 ? '管理员' : '用户' }}</span></td>
                    <td class="text-slate-500">{{ fmtTime(item.read_at) }}</td>
                  </tr>
                  <tr v-if="!announcementReads.length">
                    <td class="py-12 text-center text-slate-500" colspan="3">暂无阅读记录</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>
      </main>
    </div>
  </section>
</template>

<style scoped>
.admin-bg-mesh {
  background:
    radial-gradient(circle at 18% 18%, rgb(23 126 137 / 0.08), transparent 28%),
    radial-gradient(circle at 82% 4%, rgb(59 130 246 / 0.08), transparent 24%),
    radial-gradient(circle at 65% 88%, rgb(229 111 90 / 0.06), transparent 26%);
}

.admin-sidebar {
  @apply flex flex-col border-b border-gray-200 bg-white/95 shadow-sm backdrop-blur lg:sticky lg:top-0 lg:h-screen lg:border-b-0 lg:border-r;
}

.admin-sidebar-header {
  @apply flex h-16 items-center gap-3 border-b border-gray-100 px-6;
}

.admin-sidebar-nav {
  @apply flex gap-1 overflow-x-auto px-3 py-4 lg:block lg:flex-1 lg:space-y-1 lg:overflow-y-auto;
}

.admin-sidebar-link {
  @apply flex min-w-24 items-center justify-between gap-3 rounded-xl px-4 py-2.5 text-left text-sm font-medium text-slate-600 transition-all hover:bg-gray-100 hover:text-slate-950 lg:w-full;
}

.admin-sidebar-link-active {
  @apply bg-teal/10 text-teal shadow-sm hover:bg-teal/15 hover:text-teal;
}

.admin-sidebar-footer {
  @apply hidden border-t border-gray-100 p-4 lg:block;
}

.admin-panel {
  @apply rounded-2xl border border-gray-100 bg-white shadow-sm shadow-slate-900/[0.04];
}

.admin-topbar {
  @apply rounded-2xl border border-gray-100 bg-white/85 px-5 py-4 shadow-sm shadow-slate-900/[0.03] backdrop-blur;
}

.admin-hero {
  @apply flex flex-col gap-4 rounded-2xl border border-gray-100 bg-white px-6 py-5 shadow-sm shadow-slate-900/[0.04] sm:flex-row sm:items-center sm:justify-between;
}

.admin-metric-grid {
  @apply grid overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-sm shadow-slate-900/[0.04] md:grid-cols-2 xl:grid-cols-4;
}

.admin-metric {
  @apply border-b border-gray-100 p-6 md:border-r xl:border-b-0;
}

.admin-toolbar {
  @apply rounded-2xl border border-gray-100 bg-white px-4 py-4 shadow-sm shadow-slate-900/[0.04];
}

.admin-list {
  @apply overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-sm shadow-slate-900/[0.04];
}

.admin-list-row {
  @apply border-t border-gray-100 p-5 first:border-t-0 hover:bg-gray-50/70;
}

.admin-empty {
  @apply border-t border-gray-100 p-10 text-center text-sm text-slate-500;
}

.admin-settings-nav {
  @apply overflow-hidden rounded-2xl border border-gray-100 bg-white p-2 shadow-sm shadow-slate-900/[0.04];
}

.admin-settings-content {
  @apply overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-sm shadow-slate-900/[0.04];
}

.admin-section-title {
  @apply text-sm font-semibold text-slate-950;
}

.admin-input {
  @apply min-h-10 rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-slate-950 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:opacity-60;
}

.admin-textarea {
  @apply min-h-28 rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-slate-950 outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:opacity-60;
}

.admin-label {
  @apply block text-xs font-medium uppercase tracking-wide text-slate-500;
}

.admin-primary {
  @apply inline-flex min-h-10 items-center justify-center rounded-xl bg-gradient-to-r from-teal to-blue-600 px-4 py-2 text-sm font-medium text-white shadow-md shadow-teal/20 transition hover:shadow-lg hover:shadow-teal/25 disabled:opacity-60;
}

.admin-btn {
  @apply inline-flex min-h-9 items-center justify-center rounded-xl border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 shadow-sm transition hover:border-gray-300 hover:bg-gray-50 disabled:opacity-60;
}

.admin-btn-danger {
  @apply inline-flex min-h-9 items-center justify-center rounded-xl border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-600 shadow-sm transition hover:bg-red-50 disabled:opacity-60;
}

.admin-icon-btn {
  @apply inline-flex min-h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-gray-200 bg-white text-slate-500 transition hover:border-gray-300 hover:bg-gray-50 hover:text-slate-900;
}

.admin-badge {
  @apply inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium;
}

.admin-badge-ok {
  @apply bg-emerald-100 text-emerald-700;
}

.admin-badge-danger {
  @apply bg-red-100 text-red-700;
}

.admin-badge-muted {
  @apply bg-gray-100 text-slate-600;
}

.admin-badge-info {
  @apply bg-blue-100 text-blue-700;
}

.admin-badge-warning {
  @apply bg-amber-100 text-amber-700;
}

.admin-table {
  @apply min-w-full text-sm;
}

.admin-table-scroll {
  @apply overflow-x-auto;
}

.admin-table thead {
  @apply bg-gray-50/80 text-left text-xs font-medium uppercase tracking-wide text-slate-500;
}

.admin-table th {
  @apply whitespace-nowrap px-4 py-3;
}

.admin-table td {
  @apply whitespace-nowrap border-t border-gray-100 px-4 py-4 align-middle text-slate-700;
}

.admin-table tbody tr {
  @apply transition hover:bg-gray-50;
}
</style>
