<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

interface CreditLog {
  id: number
  type: number
  amount: number
  balance: number
  remark: string
  created_at: string
}

const router = useRouter()
const userStore = useUserStore()
const logs = ref<CreditLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const error = ref('')
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

onMounted(async () => {
  if (!userStore.token) {
    await router.push('/login')
    return
  }
  await loadLogs()
})

async function loadLogs(targetPage = page.value) {
  loading.value = true
  error.value = ''
  try {
    const response = await api.get('/credits/logs', { params: { page: targetPage, pageSize } })
    logs.value = response.data.items || []
    total.value = response.data.total || 0
    page.value = response.data.page || targetPage
  } catch (err: any) {
    error.value = err.response?.data?.error || '积分流水加载失败'
  } finally {
    loading.value = false
  }
}

function creditTypeText(type: number) {
  const map: Record<number, string> = { 1: '充值', 2: '生成消耗', 3: '人工充值', 4: '失败退回', 5: '注册赠送' }
  return map[type] || `类型 ${type}`
}

function formatTime(value: string) {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

<template>
  <section class="space-y-5">
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p class="text-sm font-medium text-teal">Credits</p>
          <h1 class="mt-1 text-2xl font-semibold text-slate-950">我的积分流水</h1>
          <p class="mt-2 text-sm text-slate-500">查看积分获取、消耗和退回记录。</p>
        </div>
        <div class="rounded-2xl bg-slate-50 px-4 py-3 text-right">
          <div class="text-xs text-slate-500">当前余额</div>
          <div class="mt-1 text-2xl font-semibold text-slate-950">{{ userStore.user?.credits ?? 0 }}</div>
        </div>
      </div>
    </div>

    <p v-if="error" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</p>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div v-if="loading" class="p-8 text-center text-sm text-slate-500">加载中...</div>
      <div v-else-if="logs.length === 0" class="p-8 text-center text-sm text-slate-500">暂无积分流水</div>
      <table v-else class="w-full text-left text-sm">
        <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
          <tr>
            <th class="px-4 py-3 font-medium">时间</th>
            <th class="px-4 py-3 font-medium">类型</th>
            <th class="px-4 py-3 font-medium">变动</th>
            <th class="px-4 py-3 font-medium">余额</th>
            <th class="px-4 py-3 font-medium">备注</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="log in logs" :key="log.id">
            <td class="px-4 py-3 text-slate-500">{{ formatTime(log.created_at) }}</td>
            <td class="px-4 py-3">{{ creditTypeText(log.type) }}</td>
            <td class="px-4 py-3 font-semibold" :class="log.amount >= 0 ? 'text-emerald-600' : 'text-rose-600'">{{ log.amount >= 0 ? '+' : '' }}{{ log.amount }}</td>
            <td class="px-4 py-3">{{ log.balance }}</td>
            <td class="max-w-md truncate px-4 py-3 text-slate-500">{{ log.remark || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex items-center justify-between">
      <button class="rounded-full border border-slate-300 px-4 py-2 text-sm disabled:opacity-50" type="button" :disabled="page <= 1 || loading" @click="loadLogs(page - 1)">上一页</button>
      <span class="text-sm text-slate-500">第 {{ page }} / {{ totalPages }} 页，共 {{ total }} 条</span>
      <button class="rounded-full border border-slate-300 px-4 py-2 text-sm disabled:opacity-50" type="button" :disabled="page >= totalPages || loading" @click="loadLogs(page + 1)">下一页</button>
    </div>
  </section>
</template>
