<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

interface Generation {
  id: number
  prompt: string
  quality: string
  size: string
  status: number
  image_url: string
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
    const response = await api.get('/generations', { params: { page: page.value, pageSize } })
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

function download(url: string) {
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = 'image-show.png'
  anchor.click()
}

function fmtTime(value: string) {
  return value ? new Date(value).toLocaleString() : ''
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
    <div class="flex items-end justify-between border-b border-slate-200 pb-4">
      <div>
        <h1 class="text-2xl font-semibold">图片历史</h1>
        <p class="mt-1 text-sm text-slate-600">查看、下载和删除已完成的生成记录。</p>
      </div>
      <button class="rounded border border-slate-300 px-3 py-2 text-sm" type="button" @click="load(true)">刷新</button>
    </div>

    <p v-if="error" class="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="items.length === 0 && !loading" class="rounded border border-slate-200 bg-white p-8 text-center text-sm text-slate-600">
      暂无历史图片
    </div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <article v-for="item in items" :key="item.id" class="overflow-hidden rounded border border-slate-200 bg-white">
        <button class="block aspect-square w-full bg-slate-100" type="button" @click="openDetail(item)">
          <img v-if="item.image_url" :src="item.image_url" class="h-full w-full object-cover" alt="生成图片" />
        </button>
        <div class="space-y-2 p-3 text-sm">
          <div class="flex items-center justify-between text-slate-600">
            <span>#{{ item.id }} · {{ item.quality }}</span>
            <span>{{ fmtTime(item.created_at) }}</span>
          </div>
          <p class="line-clamp-2 min-h-10 text-slate-800">{{ item.prompt }}</p>
          <div class="flex gap-2">
            <button class="rounded border border-slate-300 px-2 py-1" type="button" @click="openDetail(item)">查看</button>
            <button class="rounded border border-slate-300 px-2 py-1" type="button" @click="download(item.image_url)">下载</button>
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
          </div>
          <button class="rounded border border-slate-300 px-3 py-1.5" type="button" @click="selected = null">关闭</button>
        </div>
        <img :src="selected.image_url" class="max-h-[70vh] w-full object-contain" alt="生成图片详情" />
        <div class="mt-3 flex gap-2">
          <button class="rounded bg-teal px-4 py-2 text-white" type="button" @click="download(selected.image_url)">下载</button>
          <button class="rounded border border-red-200 px-4 py-2 text-red-600" type="button" @click="deleteItem(selected)">删除</button>
        </div>
      </div>
    </div>
  </section>
</template>
