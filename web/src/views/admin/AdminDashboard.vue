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
  user_id?: number
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
  status: number
  weight: number
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

const router = useRouter()
const userStore = useUserStore()
const activeTab = ref('users')
const loading = ref(false)
const message = ref('')
const userKeyword = ref('')
const users = ref<Page<User>>({ items: [], total: 0, page: 1, pageSize: 20 })
const selectedUser = ref<User | null>(null)
const userGenerations = ref<Page<Generation>>({ items: [], total: 0, page: 1, pageSize: 10 })
const creditLogs = ref<Page<CreditLog>>({ items: [], total: 0, page: 1, pageSize: 20 })
const templates = ref<PromptTemplate[]>([])
const channels = ref<Channel[]>([])
const settings = ref<Record<string, string>>({})
const monitor = ref<MonitorSummary | null>(null)
const creditForm = ref({ amount: 1, remark: '' })
const templateForm = ref<PromptTemplate>({ id: 0, category: 'default', label: '', prompt: '', sort_order: 0, status: 1 })

const tabs = [
  { id: 'users', label: '用户' },
  { id: 'credits', label: '积分' },
  { id: 'templates', label: '模板' },
  { id: 'settings', label: '设置' },
  { id: 'monitor', label: '监控' },
  { id: 'logs', label: '日志' },
  { id: 'channels', label: '渠道' },
]

const isAdmin = computed(() => (userStore.user?.role || 0) >= 10)

async function guarded<T>(fn: () => Promise<T>) {
  loading.value = true
  message.value = ''
  try {
    return await fn()
  } catch (error) {
    message.value = '操作失败，请检查权限或输入'
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
  })
}

async function updateUserRole(user: User, role: number) {
  await guarded(async () => {
    await api.put(`/admin/users/${user.id}/role`, { role })
    await loadUsers()
  })
}

async function topupCredits(user: User) {
  await guarded(async () => {
    await api.post(`/admin/users/${user.id}/credits`, creditForm.value)
    await Promise.all([loadUsers(), loadCreditLogs()])
    message.value = '充值完成'
  })
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
  templateForm.value = { id: 0, category: 'default', label: '', prompt: '', sort_order: 0, status: 1 }
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
  })
}

async function deleteTemplate(template: PromptTemplate) {
  await guarded(async () => {
    await api.delete(`/admin/prompt-templates/${template.id}`)
    await loadTemplates()
  })
}

async function loadSettings() {
  const response = await api.get('/admin/settings')
  settings.value = response.data.items
}

async function saveSettings() {
  await guarded(async () => {
    await api.put('/admin/settings', { items: settings.value })
    message.value = '设置已保存'
  })
}

async function loadChannels() {
  const response = await api.get('/admin/channels')
  channels.value = response.data.items
}

async function loadMonitor() {
  const response = await api.get('/admin/monitor/summary')
  monitor.value = response.data
}

async function checkMonitorAlert() {
  await guarded(async () => {
    const response = await api.post('/admin/monitor/check')
    message.value = response.data.sent ? '告警已发送' : '未触发或今日已发送'
    await loadMonitor()
  })
}

function fmtTime(value: string) {
  return value ? new Date(value).toLocaleString() : '-'
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
  <section class="space-y-5">
    <div class="flex flex-col gap-3 border-b border-slate-200 pb-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-ink">管理员后台</h1>
        <p class="mt-1 text-sm text-slate-600">用户、积分、模板、设置和运行记录管理。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="rounded border px-3 py-1.5 text-sm"
          :class="activeTab === tab.id ? 'border-teal bg-teal text-white' : 'border-slate-300 bg-white text-slate-700'"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <p v-if="message" class="rounded border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
      {{ message }}
    </p>

    <div v-if="activeTab === 'users'" class="space-y-4">
      <div class="flex gap-2">
        <input v-model="userKeyword" class="min-w-0 flex-1 rounded border border-slate-300 px-3 py-2" placeholder="搜索邮箱或用户名" />
        <button class="rounded bg-teal px-4 py-2 text-white" type="button" @click="loadUsers">搜索</button>
      </div>
      <div class="overflow-x-auto rounded border border-slate-200 bg-white">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 text-left text-slate-600">
            <tr>
              <th class="px-3 py-2">ID</th>
              <th class="px-3 py-2">邮箱</th>
              <th class="px-3 py-2">角色</th>
              <th class="px-3 py-2">状态</th>
              <th class="px-3 py-2">积分</th>
              <th class="px-3 py-2">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users.items" :key="user.id" class="border-t border-slate-100">
              <td class="px-3 py-2">{{ user.id }}</td>
              <td class="px-3 py-2">{{ user.email }}</td>
              <td class="px-3 py-2">{{ user.role }}</td>
              <td class="px-3 py-2">{{ user.status === 1 ? '正常' : '封禁' }}</td>
              <td class="px-3 py-2">{{ user.credits }}</td>
              <td class="flex flex-wrap gap-2 px-3 py-2">
                <button class="rounded border px-2 py-1" type="button" @click="loadUserGenerations(user)">图片</button>
                <button class="rounded border px-2 py-1" type="button" @click="updateUserStatus(user, user.status === 1 ? 2 : 1)">
                  {{ user.status === 1 ? '封禁' : '解封' }}
                </button>
                <button class="rounded border px-2 py-1" type="button" @click="updateUserRole(user, user.role >= 10 ? 1 : 10)">
                  {{ user.role >= 10 ? '设为用户' : '设为管理员' }}
                </button>
                <button class="rounded border px-2 py-1" type="button" @click="selectedUser = user">充值</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="selectedUser" class="grid gap-4 lg:grid-cols-2">
        <form class="rounded border border-slate-200 bg-white p-4" @submit.prevent="topupCredits(selectedUser)">
          <h2 class="text-base font-semibold">给 {{ selectedUser.email }} 充值</h2>
          <div class="mt-3 grid gap-3 sm:grid-cols-2">
            <input v-model.number="creditForm.amount" class="rounded border border-slate-300 px-3 py-2" min="0.01" step="0.01" type="number" />
            <input v-model="creditForm.remark" class="rounded border border-slate-300 px-3 py-2" placeholder="备注" />
          </div>
          <button class="mt-3 rounded bg-coral px-4 py-2 text-white" type="submit">确认充值</button>
        </form>
        <div class="rounded border border-slate-200 bg-white p-4">
          <h2 class="text-base font-semibold">最近图片</h2>
          <div v-for="item in userGenerations.items" :key="item.id" class="mt-3 border-t border-slate-100 pt-3 text-sm">
            <div class="font-medium">#{{ item.id }} · {{ item.quality }} · {{ item.status }}</div>
            <p class="mt-1 line-clamp-2 text-slate-600">{{ item.prompt }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'credits'" class="overflow-x-auto rounded border border-slate-200 bg-white">
      <table class="min-w-full text-sm">
        <thead class="bg-slate-50 text-left text-slate-600">
          <tr>
            <th class="px-3 py-2">用户</th>
            <th class="px-3 py-2">类型</th>
            <th class="px-3 py-2">金额</th>
            <th class="px-3 py-2">余额</th>
            <th class="px-3 py-2">备注</th>
            <th class="px-3 py-2">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in creditLogs.items" :key="log.id" class="border-t border-slate-100">
            <td class="px-3 py-2">{{ log.user_id }}</td>
            <td class="px-3 py-2">{{ log.type }}</td>
            <td class="px-3 py-2">{{ log.amount }}</td>
            <td class="px-3 py-2">{{ log.balance }}</td>
            <td class="px-3 py-2">{{ log.remark }}</td>
            <td class="px-3 py-2">{{ fmtTime(log.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="activeTab === 'templates'" class="grid gap-4 lg:grid-cols-[1fr_360px]">
      <div class="rounded border border-slate-200 bg-white">
        <div v-for="template in templates" :key="template.id" class="flex gap-3 border-b border-slate-100 p-3 text-sm">
          <div class="min-w-0 flex-1">
            <div class="font-medium">{{ template.label }} · {{ template.category }}</div>
            <p class="mt-1 text-slate-600">{{ template.prompt }}</p>
          </div>
          <button class="rounded border px-2 py-1" type="button" @click="editTemplate(template)">编辑</button>
          <button class="rounded border px-2 py-1" type="button" @click="deleteTemplate(template)">删除</button>
        </div>
      </div>
      <form class="rounded border border-slate-200 bg-white p-4" @submit.prevent="saveTemplate">
        <h2 class="text-base font-semibold">{{ templateForm.id ? '编辑模板' : '新增模板' }}</h2>
        <input v-model="templateForm.label" class="mt-3 w-full rounded border border-slate-300 px-3 py-2" placeholder="名称" />
        <select v-model="templateForm.category" class="mt-3 w-full rounded border border-slate-300 px-3 py-2">
          <option value="default">default</option>
          <option value="repair">repair</option>
          <option value="style">style</option>
        </select>
        <textarea v-model="templateForm.prompt" class="mt-3 min-h-28 w-full rounded border border-slate-300 px-3 py-2" placeholder="Prompt" />
        <div class="mt-3 grid grid-cols-2 gap-3">
          <input v-model.number="templateForm.sort_order" class="rounded border border-slate-300 px-3 py-2" type="number" />
          <select v-model.number="templateForm.status" class="rounded border border-slate-300 px-3 py-2">
            <option :value="1">启用</option>
            <option :value="2">禁用</option>
          </select>
        </div>
        <div class="mt-3 flex gap-2">
          <button class="rounded bg-teal px-4 py-2 text-white" type="submit">保存</button>
          <button class="rounded border border-slate-300 px-4 py-2" type="button" @click="resetTemplate">清空</button>
        </div>
      </form>
    </div>

    <form v-if="activeTab === 'settings'" class="max-w-2xl rounded border border-slate-200 bg-white p-4" @submit.prevent="saveSettings">
      <label v-for="(_, key) in settings" :key="key" class="mb-3 block text-sm">
        <span class="font-medium text-slate-700">{{ key }}</span>
        <input v-model="settings[key]" class="mt-1 w-full rounded border border-slate-300 px-3 py-2" />
      </label>
      <button class="rounded bg-teal px-4 py-2 text-white" type="submit">保存设置</button>
    </form>

    <div v-if="activeTab === 'monitor' && monitor" class="space-y-4">
      <div class="grid gap-3 md:grid-cols-4">
        <div class="rounded border border-slate-200 bg-white p-4">
          <div class="text-sm text-slate-500">今日生成</div>
          <div class="mt-2 text-2xl font-semibold">{{ monitor.generation_count }}</div>
        </div>
        <div class="rounded border border-slate-200 bg-white p-4">
          <div class="text-sm text-slate-500">成功 / 失败</div>
          <div class="mt-2 text-2xl font-semibold">{{ monitor.completed_count }} / {{ monitor.failed_count }}</div>
        </div>
        <div class="rounded border border-slate-200 bg-white p-4">
          <div class="text-sm text-slate-500">积分消耗</div>
          <div class="mt-2 text-2xl font-semibold">{{ monitor.credits_consumed }}</div>
        </div>
        <div class="rounded border border-slate-200 bg-white p-4">
          <div class="text-sm text-slate-500">支付金额</div>
          <div class="mt-2 text-2xl font-semibold">¥{{ monitor.paid_order_amount }}</div>
        </div>
      </div>
      <div class="rounded border border-slate-200 bg-white p-4 text-sm">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div class="font-medium">每日积分阈值 {{ monitor.alert_threshold }}</div>
            <div class="mt-1 text-slate-600">{{ monitor.alert_triggered ? '当前已触发告警条件' : '当前未触发告警条件' }}</div>
          </div>
          <button class="rounded bg-coral px-4 py-2 text-white" type="button" @click="checkMonitorAlert">检查告警</button>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'logs'" class="rounded border border-slate-200 bg-white p-4 text-sm">
      <p class="text-slate-700">生成日志和登录日志接口已接入后端；当前页面先保留入口，详细筛选在后续日志管理页扩展。</p>
    </div>

    <div v-if="activeTab === 'channels'" class="rounded border border-slate-200 bg-white">
      <div v-for="channel in channels" :key="channel.id" class="flex items-center justify-between border-b border-slate-100 p-3 text-sm">
        <div>
          <div class="font-medium">{{ channel.name }}</div>
          <div class="text-slate-600">{{ channel.base_url }}</div>
        </div>
        <div class="text-slate-700">权重 {{ channel.weight }} · {{ channel.status === 1 ? '启用' : '禁用' }}</div>
      </div>
    </div>

    <div v-if="loading" class="fixed bottom-4 right-4 rounded bg-ink px-3 py-2 text-sm text-white">处理中</div>
  </section>
</template>
