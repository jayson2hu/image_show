# 管理后台前端界面重设计 — 开发计划

> 状态：待开发 | 创建日期：2026-05-07 | 最后更新：2026-05-07

## 一、现状分析

当前管理后台 `AdminDashboard.vue` 是一个 **659 行的单文件组件**，包含 7 个 Tab（概览、用户、渠道、模板、设置、积分、监控），所有逻辑、接口、类型定义、模板全部耦合在一个文件中。

### 现有问题

| 类别 | 问题描述 |
|------|---------|
| 架构 | 单文件 659 行，所有 Tab 混在一起，维护困难 |
| 架构 | 所有接口定义（User/Channel/Generation/CreditLog/PromptTemplate/MonitorSummary）写在组件内，无法复用 |
| 架构 | 所有 API 调用散落在组件方法中，无统一的 admin API 模块 |
| 交互 | 删除操作使用 `window.confirm`，无统一的确认弹窗组件 |
| 交互 | 全局 loading/message 只有一个 ref，多操作并发时互相覆盖 |
| 交互 | 分页只有 pageSize 无翻页器，用户列表/积分流水无法翻页 |
| 交互 | 表单编辑和列表在同一视图，编辑渠道/模板时需要上下滚动 |
| 交互 | 设置页保存无二次确认，敏感配置一键覆盖 |
| UI | 概览卡片只有数字，无趋势对比（昨日、环比） |
| UI | 用户表格在移动端严重溢出，操作按钮挤在一行 |
| UI | 积分流水的 `type` 字段显示原始数字，不是可读文案 |
| UI | 没有空状态插画，空列表只有文字 |
| UI | 侧边栏 Tab 在移动端是横向滚动，容易忽略后面的 Tab |
| UI | 渠道测试结果只有文字，无颜色状态标识 |
| UI | 消息提示是固定在内容区的黄色条，容易被忽略，缺少自动消失 |

---

## 二、设计目标

1. **组件拆分**：将单文件拆成 Layout + 7 个独立子页面 + 共享组件
2. **类型/API 分离**：抽取 `types/admin.ts` 和 `api/admin.ts`
3. **交互升级**：统一的 Toast 通知、确认弹窗、分页器、抽屉式表单编辑
4. **响应式优化**：移动端表格改为卡片列表，侧边栏改为底部 Sheet
5. **视觉提升**：概览仪表盘数据卡片、状态颜色系统、空状态插画、骨架屏加载

---

## 三、文件结构规划

```
web/src/
├── types/
│   └── admin.ts                    # Admin 相关类型定义
├── api/
│   └── admin.ts                    # Admin API 封装
├── composables/
│   └── useToast.ts                 # Toast 通知 composable
├── components/
│   ├── ui/
│   │   ├── AppToast.vue            # Toast 通知组件
│   │   ├── ConfirmDialog.vue       # 确认弹窗组件
│   │   ├── Pagination.vue          # 通用分页器
│   │   ├── EmptyState.vue          # 空状态占位组件
│   │   └── SkeletonCard.vue        # 骨架屏加载组件
│   └── admin/
│       ├── AdminLayout.vue         # 后台整体布局（侧边栏 + 内容区）
│       ├── AdminSidebar.vue        # 侧边栏导航
│       ├── OverviewTab.vue         # 概览 Tab
│       ├── UsersTab.vue            # 用户管理 Tab
│       ├── ChannelsTab.vue         # 渠道管理 Tab
│       ├── TemplatesTab.vue        # 模板管理 Tab
│       ├── SettingsTab.vue         # 系统设置 Tab
│       ├── CreditsTab.vue          # 积分流水 Tab
│       └── MonitorTab.vue          # 监控告警 Tab
└── views/
    └── admin/
        └── AdminDashboard.vue      # 入口页（仅引入 AdminLayout）
```

---

## 四、开发阶段与任务拆分

### 阶段 A：基础设施层（共享类型、API、通用组件）

本阶段搭建所有后续 Tab 拆分所依赖的基础模块，完成后提交一次代码。

---

#### 任务 A1：抽取 Admin 类型定义

**文件**：`web/src/types/admin.ts`（新建）

**操作**：将 `AdminDashboard.vue` 中第 8-76 行的所有 interface 抽取出来，统一导出。

```ts
// web/src/types/admin.ts

export interface Page<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface AdminUser {
  id: number
  username: string
  email: string
  role: number
  status: number
  credits: number
  created_at: string
}

export interface CreditLog {
  id: number
  user_id: number
  type: number
  amount: number
  balance: number
  remark: string
  created_at: string
}

export interface Generation {
  id: number
  prompt: string
  quality: string
  size: string
  status: number
  image_url: string
  created_at: string
}

export interface PromptTemplate {
  id: number
  category: string
  label: string
  prompt: string
  sort_order: number
  status: number
}

export interface Channel {
  id: number
  name: string
  base_url: string
  api_key?: string
  headers?: string
  status: number
  weight: number
  remark?: string
}

export interface MonitorSummary {
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

export interface CreditForm {
  amount: number
  remark: string
}

export type TemplateCategory = 'style' | 'sample' | 'default' | 'repair'
```

**自测**：
- [ ] 文件无 TypeScript 编译错误
- [ ] 所有类型均有导出
- [ ] `npm run build` 通过

---

#### 任务 A2：抽取 Admin API 模块

**文件**：`web/src/api/admin.ts`（新建）

**操作**：将 `AdminDashboard.vue` 中散落的 `api.get/post/put/delete` 调用全部收归到一个模块。

```ts
// web/src/api/admin.ts
import api from '@/api'
import type {
  AdminUser, Channel, CreditForm, CreditLog,
  Generation, MonitorSummary, Page, PromptTemplate,
} from '@/types/admin'

// ===== 用户 =====
export function fetchUsers(params: { keyword?: string; page?: number; pageSize?: number }) {
  return api.get<Page<AdminUser>>('/admin/users', { params })
}

export function fetchUserGenerations(userId: number, params: { page?: number; pageSize?: number }) {
  return api.get<Page<Generation>>(`/admin/users/${userId}/generations`, { params })
}

export function updateUserStatus(userId: number, status: number) {
  return api.put(`/admin/users/${userId}/status`, { status })
}

export function updateUserRole(userId: number, role: number) {
  return api.put(`/admin/users/${userId}/role`, { role })
}

export function topupCredits(userId: number, form: CreditForm) {
  return api.post(`/admin/users/${userId}/credits`, form)
}

// ===== 积分 =====
export function fetchCreditLogs(params: { page?: number; pageSize?: number }) {
  return api.get<Page<CreditLog>>('/admin/credits/logs', { params })
}

// ===== 模板 =====
export function fetchTemplates() {
  return api.get<{ items: PromptTemplate[] }>('/admin/prompt-templates')
}

export function createTemplate(payload: Omit<PromptTemplate, 'id'>) {
  return api.post('/admin/prompt-templates', payload)
}

export function updateTemplate(id: number, payload: Partial<PromptTemplate>) {
  return api.put(`/admin/prompt-templates/${id}`, payload)
}

export function deleteTemplate(id: number) {
  return api.delete(`/admin/prompt-templates/${id}`)
}

// ===== 渠道 =====
export function fetchChannels() {
  return api.get<{ items: Channel[] }>('/admin/channels')
}

export function createChannel(payload: Omit<Channel, 'id'>) {
  return api.post('/admin/channels', payload)
}

export function updateChannel(id: number, payload: Partial<Channel>) {
  return api.put(`/admin/channels/${id}`, payload)
}

export function deleteChannel(id: number) {
  return api.delete(`/admin/channels/${id}`)
}

export function testChannel(id: number) {
  return api.post<{ ok: boolean; status?: number; error?: string }>(`/admin/channels/${id}/test`)
}

// ===== 设置 =====
export function fetchSettings() {
  return api.get<{ items: Record<string, string> }>('/admin/settings')
}

export function saveSettings(items: Record<string, string>) {
  return api.put('/admin/settings', { items })
}

// ===== 监控 =====
export function fetchMonitorSummary() {
  return api.get<MonitorSummary>('/admin/monitor/summary')
}

export function triggerMonitorCheck() {
  return api.post<{ sent: boolean }>('/admin/monitor/check')
}
```

**自测**：
- [ ] 文件无编译错误
- [ ] 所有函数签名与后端接口对应
- [ ] `npm run build` 通过

---

#### 任务 A3：Toast 通知系统

**文件**：
- `web/src/composables/useToast.ts`（新建）
- `web/src/components/ui/AppToast.vue`（新建）
- `web/src/App.vue`（修改：挂载 AppToast）

**功能**：
- 支持 `success` / `error` / `info` 三种类型
- 自动消失（success 3秒，error 5秒）
- 右上角堆叠显示，最多 5 条
- provide/inject 模式，全局可用

```ts
// web/src/composables/useToast.ts
import { inject, provide, reactive } from 'vue'

export interface ToastItem {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

export interface ToastContext {
  items: ToastItem[]
  success: (message: string) => void
  error: (message: string) => void
  info: (message: string) => void
  remove: (id: number) => void
}

const TOAST_KEY = Symbol('toast')
let nextId = 0

export function createToastContext(): ToastContext {
  const items = reactive<ToastItem[]>([])

  function add(type: ToastItem['type'], message: string) {
    const id = ++nextId
    items.push({ id, type, message })
    if (items.length > 5) items.shift()
    const delay = type === 'error' ? 5000 : 3000
    setTimeout(() => remove(id), delay)
  }

  function remove(id: number) {
    const index = items.findIndex(item => item.id === id)
    if (index !== -1) items.splice(index, 1)
  }

  return {
    items,
    success: (msg) => add('success', msg),
    error: (msg) => add('error', msg),
    info: (msg) => add('info', msg),
    remove,
  }
}

export function provideToast() {
  const ctx = createToastContext()
  provide(TOAST_KEY, ctx)
  return ctx
}

export function useToast(): ToastContext {
  const ctx = inject<ToastContext>(TOAST_KEY)
  if (!ctx) throw new Error('useToast() requires provideToast() in ancestor')
  return ctx
}
```

```vue
<!-- web/src/components/ui/AppToast.vue -->
<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const toast = useToast()

const colorMap = {
  success: 'border-emerald-200 bg-emerald-50 text-emerald-800',
  error: 'border-red-200 bg-red-50 text-red-800',
  info: 'border-blue-200 bg-blue-50 text-blue-800',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed right-4 top-4 z-[100] flex flex-col gap-2">
      <TransitionGroup
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-x-8 opacity-0"
        enter-to-class="translate-x-0 opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="translate-x-8 opacity-0"
      >
        <div
          v-for="item in toast.items"
          :key="item.id"
          class="flex min-w-64 max-w-sm items-center gap-3 rounded-lg border px-4 py-3 text-sm shadow-lg"
          :class="colorMap[item.type]"
        >
          <span class="flex-1">{{ item.message }}</span>
          <button class="opacity-50 hover:opacity-100" type="button" @click="toast.remove(item.id)">&times;</button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
```

**修改 `App.vue`**：
- 在 `<script setup>` 中调用 `provideToast()`
- 在 `<template>` 最后加 `<AppToast />`

**自测**：
- [ ] 在任意子组件中 `useToast().success('测试')` 能看到右上角弹出绿色 Toast
- [ ] 3秒后自动消失
- [ ] error 类型 5秒消失，显示红色
- [ ] 点击 × 可立即关闭
- [ ] 同时触发多条 Toast 能正确堆叠

---

#### 任务 A4：确认弹窗组件

**文件**：`web/src/components/ui/ConfirmDialog.vue`（新建）

**功能**：
- Props: `open`, `title`, `message`, `confirmText`, `confirmColor`（默认 red）
- Emits: `confirm`, `cancel`
- 背景遮罩 + 居中弹窗 + 动画

```vue
<!-- web/src/components/ui/ConfirmDialog.vue -->
<script setup lang="ts">
defineProps<{
  open: boolean
  title?: string
  message: string
  confirmText?: string
  confirmColor?: 'red' | 'blue' | 'default'
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const colorMap = {
  red: 'bg-red-600 text-white hover:bg-red-700',
  blue: 'bg-blue-600 text-white hover:bg-blue-700',
  default: 'bg-slate-900 text-white hover:bg-slate-800',
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-150"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="emit('cancel')">
        <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
          <h3 class="text-lg font-semibold text-slate-900">{{ title || '确认操作' }}</h3>
          <p class="mt-2 text-sm text-slate-600">{{ message }}</p>
          <div class="mt-6 flex justify-end gap-3">
            <button
              class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50"
              type="button"
              @click="emit('cancel')"
            >
              取消
            </button>
            <button
              class="rounded-lg px-4 py-2 text-sm font-medium"
              :class="colorMap[confirmColor || 'red']"
              type="button"
              @click="emit('confirm')"
            >
              {{ confirmText || '确认' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
```

**自测**：
- [ ] `open=true` 时弹出遮罩 + 弹窗
- [ ] 点遮罩/取消按钮触发 `cancel`
- [ ] 点确认按钮触发 `confirm`
- [ ] 三种 `confirmColor` 显示不同颜色
- [ ] 进出有淡入淡出动画

---

#### 任务 A5：通用分页器组件

**文件**：`web/src/components/ui/Pagination.vue`（新建）

**功能**：
- Props: `page`, `pageSize`, `total`
- Emits: `update:page`
- 显示"第 X-Y 条 / 共 Z 条" + 上一页/下一页按钮
- 首页/末页禁用状态

```vue
<!-- web/src/components/ui/Pagination.vue -->
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const rangeStart = computed(() => Math.min((props.page - 1) * props.pageSize + 1, props.total))
const rangeEnd = computed(() => Math.min(props.page * props.pageSize, props.total))
const hasPrev = computed(() => props.page > 1)
const hasNext = computed(() => props.page < totalPages.value)
</script>

<template>
  <div class="flex items-center justify-between gap-4 text-sm text-slate-600">
    <span>第 {{ rangeStart }}-{{ rangeEnd }} 条 / 共 {{ total }} 条</span>
    <div class="flex gap-1">
      <button
        class="rounded-lg border border-slate-200 px-3 py-1.5 disabled:opacity-40"
        :disabled="!hasPrev"
        type="button"
        @click="emit('update:page', page - 1)"
      >
        上一页
      </button>
      <span class="flex items-center px-3 text-slate-500">{{ page }} / {{ totalPages }}</span>
      <button
        class="rounded-lg border border-slate-200 px-3 py-1.5 disabled:opacity-40"
        :disabled="!hasNext"
        type="button"
        @click="emit('update:page', page + 1)"
      >
        下一页
      </button>
    </div>
  </div>
</template>
```

**自测**：
- [ ] `total=0` 时显示"第 0-0 条 / 共 0 条"，两个按钮都禁用
- [ ] 第 1 页时"上一页"禁用
- [ ] 最后一页时"下一页"禁用
- [ ] 点击翻页 emit 正确的 page 值

---

#### 任务 A6：空状态组件 + 骨架屏组件

**文件**：
- `web/src/components/ui/EmptyState.vue`（新建）
- `web/src/components/ui/SkeletonCard.vue`（新建）

```vue
<!-- web/src/components/ui/EmptyState.vue -->
<script setup lang="ts">
defineProps<{
  icon?: string   // SVG path
  title: string
  description?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center rounded-xl border border-dashed border-slate-300 bg-white p-12 text-center">
    <div class="mb-4 flex size-16 items-center justify-center rounded-full bg-slate-100">
      <svg class="size-8 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="icon || 'M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4'" />
      </svg>
    </div>
    <p class="text-base font-medium text-slate-700">{{ title }}</p>
    <p v-if="description" class="mt-1 max-w-sm text-sm text-slate-500">{{ description }}</p>
    <div class="mt-4"><slot /></div>
  </div>
</template>
```

```vue
<!-- web/src/components/ui/SkeletonCard.vue -->
<script setup lang="ts">
defineProps<{
  lines?: number
}>()
</script>

<template>
  <div class="animate-pulse rounded-xl border border-slate-200 bg-white p-5">
    <div class="h-4 w-1/3 rounded bg-slate-200"></div>
    <div class="mt-4 space-y-3">
      <div v-for="i in (lines || 3)" :key="i" class="h-3 rounded bg-slate-100" :class="i % 2 === 0 ? 'w-full' : 'w-2/3'"></div>
    </div>
  </div>
</template>
```

**自测**：
- [ ] EmptyState 传入 title 和 description 显示正确
- [ ] EmptyState 的 slot 可以放按钮
- [ ] SkeletonCard 显示脉冲动画
- [ ] `lines` prop 控制骨架条数

---

### 阶段 A 完成标准

- [ ] 以上 A1-A6 全部自测通过
- [ ] `npm run build` 零错误零警告
- [ ] `npm run lint` 通过（如有 ESLint）
- [ ] 提交代码，commit message: `feat(admin): extract shared types, API module, and UI components`
- [ ] 更新本文档进度表

---

### 阶段 B：布局拆分与路由（AdminLayout + Sidebar + 路由子页面）

将 AdminDashboard.vue 从单文件巨组件改为布局组件 + 7 个子 Tab 组件。

---

#### 任务 B1：AdminLayout 布局组件

**文件**：`web/src/components/admin/AdminLayout.vue`（新建）

**职责**：
- 左侧导航栏（桌面端固定，移动端底部 Sheet 或顶部横向滚动）
- 右侧内容区显示当前 Tab
- 管理 `activeTab` 状态
- 权限校验（非管理员跳转首页）
- 初始化数据加载

**关键设计**：
- 移动端侧边栏改为顶部横向滚动标签栏，每个 Tab 只显示图标 + 短名称
- 桌面端保持左侧固定 260px 侧边栏
- 内容区有 `<component :is="currentTabComponent" />` 动态渲染

```vue
<!-- web/src/components/admin/AdminLayout.vue -->
<script setup lang="ts">
import { computed, markRaw, onMounted, provide, ref, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useToast } from '@/composables/useToast'
import AdminSidebar from './AdminSidebar.vue'
import OverviewTab from './OverviewTab.vue'
import UsersTab from './UsersTab.vue'
import ChannelsTab from './ChannelsTab.vue'
import TemplatesTab from './TemplatesTab.vue'
import SettingsTab from './SettingsTab.vue'
import CreditsTab from './CreditsTab.vue'
import MonitorTab from './MonitorTab.vue'

const router = useRouter()
const userStore = useUserStore()
const toast = useToast()
const activeTab = ref('overview')
const ready = ref(false)

const tabComponents: Record<string, ReturnType<typeof markRaw>> = {
  overview: markRaw(OverviewTab),
  users: markRaw(UsersTab),
  channels: markRaw(ChannelsTab),
  templates: markRaw(TemplatesTab),
  settings: markRaw(SettingsTab),
  credits: markRaw(CreditsTab),
  monitor: markRaw(MonitorTab),
}

const currentComponent = computed(() => tabComponents[activeTab.value] || tabComponents.overview)
const isAdmin = computed(() => (userStore.user?.role || 0) >= 10)

// 提供 toast 给子组件（已由 App.vue 提供，这里只是类型安全导入）
provide('adminToast', toast)

onMounted(async () => {
  await userStore.fetchUser()
  if (!isAdmin.value) {
    await router.push('/')
    return
  }
  ready.value = true
})
</script>

<template>
  <section v-if="ready" class="min-h-[calc(100vh-65px)] bg-slate-50 text-slate-950">
    <div class="grid gap-0 lg:grid-cols-[260px_1fr]">
      <AdminSidebar v-model:active-tab="activeTab" />
      <main class="min-w-0 p-4 sm:p-6 lg:p-8">
        <component :is="currentComponent" />
      </main>
    </div>
  </section>
</template>
```

**自测**：
- [ ] 页面加载后显示侧边栏 + 内容区
- [ ] 非管理员自动跳转首页
- [ ] 切换 Tab 内容区正确切换

---

#### 任务 B2：AdminSidebar 侧边栏组件

**文件**：`web/src/components/admin/AdminSidebar.vue`（新建）

**功能**：
- 桌面端：左侧 260px 固定导航
- 移动端：顶部横向滚动标签栏
- 每个 Tab 有图标 + 名称 + 描述（桌面端显示描述，移动端隐藏）
- 当前选中 Tab 高亮样式
- 底部显示当前管理员邮箱 + 返回前台按钮

**Tab 图标映射**（使用 Heroicons 风格内联 SVG）：

| Tab | 图标 | 名称 | 描述 |
|-----|------|------|------|
| overview | ChartBarIcon | 概览 | 核心指标与运行状态 |
| users | UsersIcon | 用户 | 账号、角色与充值 |
| channels | ServerIcon | 渠道 | API 渠道配置与测试 |
| templates | DocumentTextIcon | 模板 | 提示词模板管理 |
| settings | CogIcon | 设置 | 系统开关和配置 |
| credits | CreditCardIcon | 积分 | 积分流水审计 |
| monitor | BellIcon | 监控 | 每日指标和告警 |

**自测**：
- [ ] 桌面端 (>1024px) 显示左侧固定栏，Tab 有名称和描述
- [ ] 移动端 (<1024px) 显示顶部横向滚动栏，只有图标和短名称
- [ ] 点击 Tab 触发 `update:activeTab` 事件
- [ ] 当前 Tab 有明显高亮
- [ ] 底部可见管理员邮箱

---

#### 任务 B3：重写 AdminDashboard.vue 为入口壳

**文件**：`web/src/views/admin/AdminDashboard.vue`（改写）

**操作**：将 659 行的巨组件改为仅引入 AdminLayout 的薄壳。

```vue
<!-- web/src/views/admin/AdminDashboard.vue -->
<script setup lang="ts">
import AdminLayout from '@/components/admin/AdminLayout.vue'
</script>

<template>
  <AdminLayout />
</template>
```

**自测**：
- [ ] 路由 `/console/admin` 正常加载新布局
- [ ] 原有所有功能暂时不可用是预期的（各 Tab 尚未实现）
- [ ] 页面不白屏，侧边栏可见

---

### 阶段 B 完成标准

- [ ] B1-B3 全部自测通过
- [ ] 布局在桌面端/移动端均正确渲染
- [ ] `npm run build` 通过
- [ ] 提交代码，commit message: `feat(admin): split layout into AdminLayout + AdminSidebar`
- [ ] 更新本文档进度表

---

### 阶段 C：各 Tab 页面实现（逐个拆分并优化 UI）

每完成一个 Tab 即可自测提交，无需等所有 Tab 完成。

---

#### 任务 C1：OverviewTab 概览页

**文件**：`web/src/components/admin/OverviewTab.vue`（新建）

**现有功能迁移**：
- 四张数据卡片（今日生成、新增用户、积分消耗、启用渠道）
- 渠道状态列表（展示前 4 个）
- 今日监控摘要

**UI 优化点**：
1. **数据卡片增加图标**：每张卡片左上角加一个圆形图标（不同颜色）
2. **数据卡片增加趋势标识**：hint 区域如果有"成功/失败"等关键词，用绿色/红色小标签
3. **渠道状态列表**：每行加绿色/灰色圆点表示启用/禁用
4. **骨架屏加载**：数据加载中显示 SkeletonCard
5. **快捷操作区**：概览页底部加"快捷操作"按钮组（跳转到用户管理、渠道管理等）

**数据加载**：
- `onMounted` 时调用 `fetchUsers`, `fetchChannels`, `fetchMonitorSummary`
- 使用 loading ref 控制骨架屏/内容切换

**自测**：
- [ ] 数据加载中显示骨架屏
- [ ] 数据加载后显示 4 张卡片、渠道列表、监控摘要
- [ ] 渠道启用/禁用状态颜色正确
- [ ] "检查告警"按钮正常工作
- [ ] 快捷操作按钮能切换到对应 Tab（通过 emit 或 inject）

---

#### 任务 C2：UsersTab 用户管理页

**文件**：`web/src/components/admin/UsersTab.vue`（新建）

**现有功能迁移**：
- 用户搜索
- 用户列表表格
- 用户操作（查看记录、充值、封禁/解封、设角色）
- 充值表单
- 用户生成记录

**UI 优化点**：
1. **搜索栏升级**：搜索框加搜索图标，回车即搜索，搜索时显示加载态
2. **表格响应式**：
   - 桌面端 (≥768px)：正常表格
   - 移动端 (<768px)：每个用户显示为卡片，操作按钮在卡片底部
3. **操作按钮改为下拉菜单**：桌面端 4 个按钮挤在一列太拥挤，改为"操作"下拉菜单包含：查看记录、充值、封禁/解封、设置角色
4. **充值表单改为抽屉**：点"充值"弹出右侧抽屉面板（或 Modal），不在页面中间展示
5. **分页器**：使用 Pagination 组件，支持翻页
6. **删除/封禁确认**：使用 ConfirmDialog 组件，不用 `window.confirm`
7. **生成记录弹窗优化**：点"查看记录"弹出 Modal，显示用户生成记录列表

**自测**：
- [ ] 搜索用户功能正常，回车可搜索
- [ ] 桌面端表格正常显示
- [ ] 移动端 (<768px) 切换为卡片布局（使用浏览器 DevTools 模拟）
- [ ] 操作下拉菜单正常弹出/收起
- [ ] 充值弹窗正常打开/关闭/提交
- [ ] 封禁/解封弹出确认弹窗，确认后执行
- [ ] 分页器翻页正常
- [ ] 所有操作成功/失败显示 Toast 通知

---

#### 任务 C3：ChannelsTab 渠道管理页

**文件**：`web/src/components/admin/ChannelsTab.vue`（新建）

**现有功能迁移**：
- 渠道列表
- 渠道测试
- 新增/编辑渠道表单
- 删除渠道

**UI 优化点**：
1. **渠道卡片升级**：
   - 状态指示灯：启用绿色脉冲点，禁用灰色静态点
   - 测试结果：成功绿色 badge、失败红色 badge、测试中旋转 spinner
   - API Key 状态：显示"已配置 sk-...xxx"（脱敏最后 4 位）或"未配置"红色警告
2. **表单改为右侧抽屉**：点击"新增渠道"或"编辑"弹出右侧抽屉面板
3. **删除确认**：使用 ConfirmDialog
4. **批量操作**：顶部添加"全部测试"按钮，一键测试所有渠道
5. **空状态**：无渠道时使用 EmptyState 组件

**自测**：
- [ ] 渠道列表正常显示，状态颜色正确
- [ ] 测试按钮点击后显示加载态，结果正确
- [ ] 新增渠道：打开抽屉 → 填写表单 → 保存 → 列表刷新
- [ ] 编辑渠道：打开抽屉并回填数据 → 修改 → 保存
- [ ] 删除渠道：弹出确认 → 确认后删除 → 列表刷新
- [ ] 无渠道时显示空状态插画
- [ ] "全部测试"按钮正常工作

---

#### 任务 C4：TemplatesTab 模板管理页

**文件**：`web/src/components/admin/TemplatesTab.vue`（新建）

**现有功能迁移**：
- 模板列表（分类、名称、状态、prompt 预览）
- 新增/编辑/删除模板
- 表单

**UI 优化点**：
1. **分类 Tab 筛选**：顶部加横向标签切换 全部/首页风格预设/首页推荐样例/默认标签/修复标签
2. **模板卡片**：改为卡片网格（每行 2-3 列），每张卡片显示名称、分类标签、prompt 预览（3 行截断）、启用/禁用 badge
3. **表单改为 Modal**：点"新增"或"编辑"弹出居中 Modal
4. **Prompt 预览**：编辑 Modal 中 prompt 文本框下方实时显示字符数
5. **拖拽排序**（可选，不强制）：如果时间允许，模板可拖拽调整 sort_order
6. **删除确认**：使用 ConfirmDialog

**自测**：
- [ ] 分类 Tab 切换正确过滤
- [ ] 卡片网格布局正常
- [ ] 新增模板：Modal → 填写 → 保存 → 列表刷新
- [ ] 编辑模板：Modal 回填 → 修改 → 保存
- [ ] 删除确认后删除
- [ ] 启用/禁用标签颜色正确

---

#### 任务 C5：SettingsTab 系统设置页

**文件**：`web/src/components/admin/SettingsTab.vue`（新建）

**现有功能迁移**：
- 设置项双列表单
- 保存按钮

**UI 优化点**：
1. **分组显示**：将设置项按逻辑分组
   - **存储配置**：r2_endpoint, r2_access_key, r2_secret_key, r2_bucket, r2_public_url
   - **生成配置**：image_model, enabled_image_sizes
   - **安全配置**：captcha_enabled, turnstile_site_key, turnstile_secret, register_enabled, ip_blacklist
2. **每组一个卡片**：组标题 + 组内设置项
3. **敏感字段交互**：password 类型字段加"显示/隐藏"切换按钮（眼睛图标）
4. **保存确认**：点保存弹出 ConfirmDialog"确认保存所有设置？修改后立即生效。"
5. **字段说明**：help 文案显示在输入框下方，浅灰色小字
6. **骨架屏加载**：加载中显示骨架

**设置分组映射**：

```ts
const settingGroups = [
  {
    title: '存储配置',
    description: 'Cloudflare R2 对象存储',
    keys: ['r2_endpoint', 'r2_access_key', 'r2_secret_key', 'r2_bucket', 'r2_public_url'],
  },
  {
    title: '生成配置',
    description: '图片生成模型与参数',
    keys: ['image_model', 'enabled_image_sizes'],
  },
  {
    title: '安全配置',
    description: '验证码、注册与访问控制',
    keys: ['captcha_enabled', 'turnstile_site_key', 'turnstile_secret', 'register_enabled', 'ip_blacklist'],
  },
]
```

**自测**：
- [ ] 设置项按组分卡片显示
- [ ] 敏感字段默认 `***` 显示，点眼睛图标切换明文
- [ ] 保存按钮弹出确认弹窗
- [ ] 确认后保存成功，Toast 提示
- [ ] 加载中显示骨架屏

---

#### 任务 C6：CreditsTab 积分流水页

**文件**：`web/src/components/admin/CreditsTab.vue`（新建）

**现有功能迁移**：
- 积分流水表格

**UI 优化点**：
1. **type 字段可读化**：将数字映射为可读文案 + 颜色标签

| type 值 | 文案 | 颜色 |
|---------|------|------|
| 1 | 充值 | 绿色 |
| 2 | 消费 | 蓝色 |
| 3 | 退还 | 黄色 |
| 4 | 系统赠送 | 紫色 |
| 其他 | 未知 (原始值) | 灰色 |

2. **金额颜色**：正数绿色 +，负数红色
3. **分页器**：使用 Pagination 组件
4. **搜索筛选**：顶部加用户 ID 搜索 + 类型下拉筛选（可选）
5. **表格响应式**：移动端隐藏部分列，只保留核心列（用户、类型、金额、时间）

**自测**：
- [ ] type 字段显示中文标签，颜色正确
- [ ] 金额正负颜色区分
- [ ] 分页翻页正常
- [ ] 移动端不出现横向滚动条
- [ ] 空数据显示 EmptyState

---

#### 任务 C7：MonitorTab 监控告警页

**文件**：`web/src/components/admin/MonitorTab.vue`（新建）

**现有功能迁移**：
- 4 张数据卡片（与概览相同）
- 告警阈值和状态
- 检查告警按钮

**UI 优化点**：
1. **告警状态醒目化**：
   - 未触发：绿色"正常"大 badge + 绿色边框卡片
   - 已触发：红色"告警"大 badge + 红色边框卡片 + 脉冲动画
2. **阈值可视化**：用进度条显示 "当前积分消耗 / 阈值"
3. **数据卡片同步概览**：复用概览的卡片数据（通过 props 或共享 store）
4. **检查告警按钮**：操作后显示 Toast 结果

**自测**：
- [ ] 告警未触发：绿色正常显示
- [ ] 告警已触发：红色脉冲告警
- [ ] 阈值进度条百分比正确
- [ ] "检查告警"按钮正常工作
- [ ] Toast 显示告警检查结果

---

### 阶段 C 完成标准

- [ ] C1-C7 全部自测通过
- [ ] 所有 Tab 功能等同或超过旧版
- [ ] 移动端体验正常（无表格溢出、无横向滚动、操作可触达）
- [ ] 所有删除操作使用 ConfirmDialog
- [ ] 所有操作反馈使用 Toast
- [ ] `npm run build` 通过
- [ ] 提交代码，commit message: `feat(admin): implement all 7 tab pages with improved UX`
- [ ] 更新本文档进度表

---

### 阶段 D：集成验证与收尾

---

#### 任务 D1：删除旧代码 + 清理

**操作**：
- 确认旧 `AdminDashboard.vue` 中的所有代码已迁移到新组件
- 清理旧 AdminDashboard.vue 中的旧逻辑（已在 B3 中替换为壳组件）
- 检查是否有无用的 import
- 确认无 `console.log` 残留

**自测**：
- [ ] 所有页面功能正常
- [ ] 无 console 警告或错误
- [ ] `npm run build` 零错误

---

#### 任务 D2：全流程回归测试

**测试矩阵**：

| Tab | 测试项 | 桌面端 | 移动端 |
|-----|--------|--------|--------|
| 概览 | 数据卡片加载 | [ ] | [ ] |
| 概览 | 渠道状态列表 | [ ] | [ ] |
| 概览 | 检查告警 | [ ] | [ ] |
| 用户 | 搜索用户 | [ ] | [ ] |
| 用户 | 充值操作 | [ ] | [ ] |
| 用户 | 封禁/解封 | [ ] | [ ] |
| 用户 | 设置角色 | [ ] | [ ] |
| 用户 | 查看记录 | [ ] | [ ] |
| 用户 | 分页翻页 | [ ] | [ ] |
| 渠道 | 新增渠道 | [ ] | [ ] |
| 渠道 | 编辑渠道 | [ ] | [ ] |
| 渠道 | 删除渠道 | [ ] | [ ] |
| 渠道 | 测试渠道 | [ ] | [ ] |
| 模板 | 分类筛选 | [ ] | [ ] |
| 模板 | 新增模板 | [ ] | [ ] |
| 模板 | 编辑模板 | [ ] | [ ] |
| 模板 | 删除模板 | [ ] | [ ] |
| 设置 | 分组显示 | [ ] | [ ] |
| 设置 | 敏感字段隐藏 | [ ] | [ ] |
| 设置 | 保存确认 | [ ] | [ ] |
| 积分 | 流水列表 | [ ] | [ ] |
| 积分 | type 可读化 | [ ] | [ ] |
| 积分 | 分页翻页 | [ ] | [ ] |
| 监控 | 告警状态 | [ ] | [ ] |
| 监控 | 检查告警 | [ ] | [ ] |
| 通用 | Toast 通知 | [ ] | [ ] |
| 通用 | 确认弹窗 | [ ] | [ ] |
| 通用 | 骨架屏加载 | [ ] | [ ] |
| 通用 | 空状态显示 | [ ] | [ ] |

---

#### 任务 D3：最终提交

- [ ] 所有回归测试通过
- [ ] `npm run build` 通过
- [ ] 提交代码，commit message: `feat(admin): complete admin dashboard redesign with component architecture`
- [ ] 更新本文档状态为"已完成"

---

## 五、进度跟踪

| 阶段 | 任务 | 状态 | 完成日期 | 备注 |
|------|------|------|----------|------|
| A | A1 类型定义 | 待开发 | | |
| A | A2 API 模块 | 待开发 | | |
| A | A3 Toast 通知 | 待开发 | | |
| A | A4 确认弹窗 | 待开发 | | |
| A | A5 分页器 | 待开发 | | |
| A | A6 空状态+骨架屏 | 待开发 | | |
| A | **阶段提交** | 待提交 | | |
| B | B1 AdminLayout | 待开发 | | |
| B | B2 AdminSidebar | 待开发 | | |
| B | B3 重写入口壳 | 待开发 | | |
| B | **阶段提交** | 待提交 | | |
| C | C1 概览页 | 待开发 | | |
| C | C2 用户管理页 | 待开发 | | |
| C | C3 渠道管理页 | 待开发 | | |
| C | C4 模板管理页 | 待开发 | | |
| C | C5 系统设置页 | 待开发 | | |
| C | C6 积分流水页 | 待开发 | | |
| C | C7 监控告警页 | 待开发 | | |
| C | **阶段提交** | 待提交 | | |
| D | D1 清理旧代码 | 待开发 | | |
| D | D2 全流程回归 | 待开发 | | |
| D | D3 最终提交 | 待提交 | | |

---

## 六、技术约束

1. **不新增后端接口**：所有 API 复用现有 `/admin/*` 接口
2. **不引入新依赖**：使用 Vue 3 内置能力（Teleport, TransitionGroup, provide/inject）
3. **样式统一**：全部使用 Tailwind CSS，不写自定义 CSS（除非动画需要 keyframes）
4. **TypeScript 严格**：所有新文件 TypeScript strict 模式
5. **渐进式迁移**：拆分过程中可以临时保留旧代码，阶段 D 统一清理

---

## 七、验收标准

### 功能验收
- [ ] 所有 7 个 Tab 功能完整，等同或超过旧版
- [ ] 所有 CRUD 操作正常（用户/渠道/模板/设置/积分）
- [ ] 所有删除操作有确认弹窗
- [ ] 所有操作有 Toast 反馈
- [ ] 分页功能正常

### UI 验收
- [ ] 桌面端 (≥1280px)：双栏布局，左侧固定导航
- [ ] 平板端 (768-1279px)：布局自适应
- [ ] 移动端 (<768px)：表格改卡片，导航栏改横向滚动
- [ ] 无横向滚动条
- [ ] 加载状态有骨架屏
- [ ] 空状态有占位组件

### 代码验收
- [ ] `AdminDashboard.vue` 不超过 20 行（壳组件）
- [ ] 新组件文件数量 ≥ 15 个
- [ ] 所有类型从 `types/admin.ts` 导入
- [ ] 所有 API 从 `api/admin.ts` 调用
- [ ] `npm run build` 零错误零警告
- [ ] 无 `console.log` 残留
- [ ] 无 `window.confirm` 残留
