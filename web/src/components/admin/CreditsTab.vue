<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { fetchCreditLogs } from '@/api/admin'
import EmptyState from '@/components/ui/EmptyState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { CreditLog, Page } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const logs = ref<Page<CreditLog>>({ items: [], total: 0, page: 1, pageSize: 20 })

onMounted(() => loadLogs(1))

async function loadLogs(page = logs.value.page) {
  loading.value = true
  try {
    const response = await fetchCreditLogs({ page, pageSize: logs.value.pageSize })
    logs.value = response.data
  } catch (error: any) {
    toast.error(error.response?.data?.error || '积分流水加载失败')
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
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Credits</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">积分流水</h2>
        <p class="mt-2 text-sm text-slate-500">审计用户充值、生成消耗、失败退回和注册赠送记录。</p>
      </div>
      <button class="rounded-2xl border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-50" type="button" @click="loadLogs(1)">刷新</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <SkeletonCard v-for="item in 4" :key="item" />
    </div>

    <EmptyState v-else-if="!logs.items.length" title="暂无积分流水" description="用户充值、生成或退回后会出现在这里。" />

    <div v-else class="space-y-4">
      <div class="hidden overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm md:block">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3 font-medium">时间</th>
              <th class="px-4 py-3 font-medium">用户 ID</th>
              <th class="px-4 py-3 font-medium">类型</th>
              <th class="px-4 py-3 font-medium">变动</th>
              <th class="px-4 py-3 font-medium">余额</th>
              <th class="px-4 py-3 font-medium">备注</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in logs.items" :key="item.id">
              <td class="px-4 py-3 text-slate-500">{{ formatTime(item.created_at) }}</td>
              <td class="px-4 py-3">{{ item.user_id }}</td>
              <td class="px-4 py-3">{{ creditTypeText(item.type) }}</td>
              <td class="px-4 py-3 font-semibold" :class="item.amount >= 0 ? 'text-emerald-600' : 'text-rose-600'">{{ item.amount >= 0 ? '+' : '' }}{{ item.amount }}</td>
              <td class="px-4 py-3">{{ item.balance }}</td>
              <td class="max-w-md truncate px-4 py-3 text-slate-500">{{ item.remark || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="grid gap-3 md:hidden">
        <article v-for="item in logs.items" :key="item.id" class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="font-semibold text-slate-950">{{ creditTypeText(item.type) }}</p>
              <p class="mt-1 text-xs text-slate-500">{{ formatTime(item.created_at) }}</p>
            </div>
            <p class="font-semibold" :class="item.amount >= 0 ? 'text-emerald-600' : 'text-rose-600'">{{ item.amount >= 0 ? '+' : '' }}{{ item.amount }}</p>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-2 text-sm">
            <div class="rounded-2xl bg-slate-50 p-3">用户：{{ item.user_id }}</div>
            <div class="rounded-2xl bg-slate-50 p-3">余额：{{ item.balance }}</div>
          </div>
          <p class="mt-3 text-sm text-slate-500">{{ item.remark || '-' }}</p>
        </article>
      </div>

      <Pagination :page="logs.page" :page-size="logs.pageSize" :total="logs.total" @update:page="loadLogs" />
    </div>
  </section>
</template>
