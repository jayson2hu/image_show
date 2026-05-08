<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'
import { downloadImage } from '@/utils/download'

interface Generation {
  id: number
  prompt: string
  quality: string
  size: string
  status: number
  image_url: string
  error_msg: string
  created_at: string
}

const router = useRouter()
const userStore = useUserStore()
const items = ref<Generation[]>([])
const selected = ref<Generation | null>(null)
const page = ref(1)
const pageSize = 12
const total = ref(0)
const loading = ref(false)
const error = ref('')
const notice = ref('')
const keyword = ref('')
const statusFilter = ref('')
const sizeFilter = ref('')

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: '3', label: '已完成' },
  { value: '4', label: '失败' },
  { value: '1', label: '生成中' },
  { value: '0', label: '排队中' },
  { value: '5', label: '已取消' },
]

const sizeOptions = [
  { value: '', label: '全部尺寸' },
  { value: '1024x1024', label: '方形 1:1' },
  { value: '1536x1024', label: '横版 3:2' },
  { value: '1024x1536', label: '竖版 2:3' },
  { value: '1792x1008', label: '宽屏 16:9' },
  { value: '1008x1792', label: '故事版 9:16' },
  { value: '1536x1152', label: '横版 4:3' },
  { value: '1152x1536', label: '竖版 3:4' },
]

const generationDraftKey = 'image_show_generation_draft'

async function load(reset = false) {
  if (loading.value) {
    return
  }
  loading.value = true
  error.value = ''
  try {
    if (reset) {
      page.value = 1
      items.value = []
    }
    const response = await api.get('/generations', {
      params: {
        page: page.value,
        pageSize,
        keyword: keyword.value.trim() || undefined,
        status: statusFilter.value || undefined,
        size: sizeFilter.value || undefined,
      },
    })
    total.value = response.data.total
    items.value.push(...response.data.items)
    page.value += 1
  } catch {
    error.value = '历史记录加载失败'
  } finally {
    loading.value = false
  }
}

async function openDetail(item: Generation) {
  try {
    const response = await api.get(`/generations/${item.id}`)
    selected.value = response.data.item
  } catch {
    error.value = '图片详情加载失败'
  }
}

async function deleteItem(item: Generation) {
  if (!window.confirm('确认删除这张图片？')) {
    return
  }
  try {
    await api.delete(`/generations/${item.id}`)
    items.value = items.value.filter((existing) => existing.id !== item.id)
    total.value -= 1
    if (selected.value?.id === item.id) {
      selected.value = null
    }
  } catch {
    error.value = '删除失败'
  }
}

function download(url: string, id?: number) {
  downloadImage(url, `image-show-${id || Date.now()}.png`)
}

async function copyPrompt(item: Generation) {
  try {
    await navigator.clipboard.writeText(item.prompt || '')
    notice.value = '提示词已复制'
  } catch {
    error.value = '复制失败，请手动选择提示词复制'
  }
}

async function reuseGeneration(item: Generation) {
  localStorage.setItem(
    generationDraftKey,
    JSON.stringify({
      prompt: item.prompt || '',
      selectedStyle: '',
      size: item.size || 'square',
      generationMode: 'generate',
    }),
  )
  await router.push('/')
}

function fmtTime(value: string) {
  return value ? new Date(value).toLocaleString() : ''
}

function statusText(status: number) {
  const map: Record<number, string> = { 0: '排队中', 1: '生成中', 2: '保存中', 3: '已完成', 4: '失败', 5: '已取消' }
  return map[status] || `状态 ${status}`
}

function friendlyError(message: string) {
  if (!message) {
    return '生成没有完成，可以带回首页重试。'
  }
  const text = message.toLowerCase()
  if (text.includes('timeout') || text.includes('524')) {
    return '上游生成超时，可以稍后重试。'
  }
  if (text.includes('502') || text.includes('503') || text.includes('service unavailable')) {
    return '上游服务暂时不可用，可以稍后重试。'
  }
  if (text.includes('r2') || text.includes('save') || text.includes('upload')) {
    return '图片保存失败，积分会按规则退回，可以重新生成。'
  }
  return '生成失败，可以调整提示词或稍后重试。'
}

function clearFilters() {
  keyword.value = ''
  statusFilter.value = ''
  sizeFilter.value = ''
  void load(true)
}

onMounted(async () => {
  await userStore.fetchUser()
  if (!userStore.user) {
    await router.push('/login')
    return
  }
  await load(true)
})
</script>

<template>
  <section class="space-y-5">
    <RouterLink class="inline-flex rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-50" to="/account">
      返回个人中心
    </RouterLink>
    <div class="flex items-end justify-between border-b border-slate-200 pb-4">
      <div>
        <h1 class="text-2xl font-semibold">图片历史</h1>
        <p class="mt-1 text-sm text-slate-600">查看、下载和删除已完成的生成记录。</p>
      </div>
      <button class="rounded border border-slate-300 px-3 py-2 text-sm" type="button" @click="load(true)">刷新</button>
    </div>

    <form class="grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 sm:grid-cols-[1fr_160px_180px_auto] dark:border-slate-700 dark:bg-slate-900" @submit.prevent="load(true)">
      <input
        v-model="keyword"
        class="min-h-11 rounded-xl border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-700 dark:bg-slate-950"
        placeholder="搜索提示词"
      />
      <select v-model="statusFilter" class="min-h-11 rounded-xl border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-700 dark:bg-slate-950">
        <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
      <select v-model="sizeFilter" class="min-h-11 rounded-xl border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20 dark:border-slate-700 dark:bg-slate-950">
        <option v-for="option in sizeOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
      <div class="flex gap-2">
        <button class="min-h-11 rounded-xl bg-teal px-4 text-sm font-medium text-white transition hover:bg-teal/90" type="submit">筛选</button>
        <button class="min-h-11 rounded-xl border border-slate-300 px-4 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800" type="button" @click="clearFilters">清空</button>
      </div>
    </form>

    <p v-if="error" class="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
    <p v-if="notice" class="rounded border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{{ notice }}</p>

    <div v-if="items.length === 0 && !loading" class="rounded border border-slate-200 bg-white p-8 text-center text-sm text-slate-600">
      暂无符合条件的历史图片
    </div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <article v-for="item in items" :key="item.id" class="overflow-hidden rounded border border-slate-200 bg-white">
        <button class="block aspect-square w-full bg-slate-100" type="button" @click="openDetail(item)">
          <img v-if="item.image_url" :src="item.image_url" class="h-full w-full object-cover" alt="生成图片" />
        </button>
        <div class="space-y-2 p-3 text-sm">
          <div class="flex items-center justify-between text-slate-600">
            <span>#{{ item.id }} · {{ statusText(item.status) }}</span>
            <span>{{ fmtTime(item.created_at) }}</span>
          </div>
          <p class="line-clamp-2 min-h-10 text-slate-800">{{ item.prompt }}</p>
          <p v-if="item.status === 4" class="rounded-lg border border-amber-200 bg-amber-50 px-2 py-1 text-xs text-amber-700">{{ friendlyError(item.error_msg) }}</p>
          <div class="flex gap-2">
            <button class="rounded border border-slate-300 px-2 py-1" type="button" @click="openDetail(item)">查看</button>
            <button class="rounded border border-slate-300 px-2 py-1" type="button" @click="copyPrompt(item)">复制提示词</button>
            <button class="rounded border border-teal/30 px-2 py-1 text-teal" type="button" @click="reuseGeneration(item)">再次生成</button>
            <button class="rounded border border-slate-300 px-2 py-1" type="button" @click="download(item.image_url, item.id)">下载</button>
            <button class="rounded border border-red-200 px-2 py-1 text-red-600" type="button" @click="deleteItem(item)">删除</button>
          </div>
        </div>
      </article>
    </div>

    <div v-if="items.length < total" class="text-center">
      <button class="rounded bg-teal px-4 py-2 text-white disabled:opacity-60" type="button" :disabled="loading" @click="load(false)">
        {{ loading ? '加载中' : '加载更多' }}
      </button>
    </div>

    <div v-if="selected" class="fixed inset-0 z-20 flex items-center justify-center bg-black/70 p-4" @click.self="selected = null">
      <div class="max-h-full w-full max-w-4xl overflow-auto rounded bg-white p-4">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div class="min-w-0">
            <h2 class="font-semibold">#{{ selected.id }}</h2>
            <p class="truncate text-sm text-slate-600">{{ selected.prompt }}</p>
            <p v-if="selected.status === 4" class="mt-1 text-sm text-amber-700">{{ friendlyError(selected.error_msg) }}</p>
          </div>
          <button class="rounded border border-slate-300 px-3 py-1.5" type="button" @click="selected = null">关闭</button>
        </div>
        <img :src="selected.image_url" class="max-h-[70vh] w-full object-contain" alt="生成图片详情" />
        <div class="mt-3 flex gap-2">
          <button class="rounded border border-slate-300 px-4 py-2" type="button" @click="copyPrompt(selected)">复制提示词</button>
          <button class="rounded border border-teal/30 px-4 py-2 text-teal" type="button" @click="reuseGeneration(selected)">再次生成</button>
          <button class="rounded bg-teal px-4 py-2 text-white" type="button" @click="download(selected.image_url, selected.id)">下载</button>
          <button class="rounded border border-red-200 px-4 py-2 text-red-600" type="button" @click="deleteItem(selected)">删除</button>
        </div>
      </div>
    </div>
  </section>
</template>
