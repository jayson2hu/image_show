<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/api'

interface CreditPackage {
  id: number
  name: string
  credits: number
  price: number
  valid_days: number
}

interface CreditLog {
  id: number
  type: number
  amount: number
  balance: number
  remark: string
  created_at: string
}

const packages = ref<CreditPackage[]>([])
const loading = ref(false)
const error = ref('')
const payingId = ref<number | null>(null)
const creditLogs = ref<CreditLog[]>([])
const creditLogsTotal = ref(0)
const creditLogsLoading = ref(false)
const creditLogsOpen = ref(false)
const router = useRouter()

onMounted(async () => {
  loading.value = true
  try {
    const response = await api.get('/packages')
    packages.value = response.data.items
  } catch {
    error.value = '套餐加载失败'
  } finally {
    loading.value = false
  }
})

async function buyPackage(item: CreditPackage) {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }
  error.value = ''
  payingId.value = item.id
  try {
    const response = await api.post('/orders', {
      package_id: item.id,
      pay_method: 'alipay',
    })
    const payURL = response.data.pay_url
    const params = response.data.params || {}
    const form = document.createElement('form')
    form.method = 'POST'
    form.action = payURL
    form.style.display = 'none'
    Object.entries(params).forEach(([key, value]) => {
      const input = document.createElement('input')
      input.type = 'hidden'
      input.name = key
      input.value = String(value)
      form.appendChild(input)
    })
    document.body.appendChild(form)
    form.submit()
  } catch (err: any) {
    error.value = err.response?.data?.error || '创建订单失败'
  } finally {
    payingId.value = null
  }
}

function standardImageCount(item: CreditPackage) {
  return Math.max(0, Math.floor(Number(item.credits) || 0))
}

function creditTypeText(type: number) {
  const map: Record<number, string> = { 1: '注册赠送', 2: '生成消耗', 3: '人工充值', 4: '失败退回', 5: '注册赠送' }
  return map[type] || `类型 ${type}`
}

function formatTime(value: string) {
  if (!value) {
    return '-'
  }
  return new Date(value).toLocaleString()
}

async function openCreditLogs() {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }
  creditLogsOpen.value = true
  creditLogsLoading.value = true
  error.value = ''
  try {
    const response = await api.get('/credits/logs', { params: { page: 1, pageSize: 10 } })
    creditLogs.value = response.data.items || []
    creditLogsTotal.value = response.data.total || 0
  } catch (err: any) {
    error.value = err.response?.data?.error || '积分流水加载失败'
  } finally {
    creditLogsLoading.value = false
  }
}
</script>

<template>
  <section class="space-y-6">
    <div class="border-b border-slate-200 pb-4 dark:border-slate-800">
      <h1 class="text-2xl font-semibold">积分套餐</h1>
      <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">选择适合的积分包，创建订单后跳转到支付平台完成付款。</p>
      <button class="mt-3 rounded-full border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-900" type="button" @click="openCreditLogs">
        查看我的积分流水
      </button>
    </div>

    <p v-if="error" class="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
    <p v-if="loading" class="text-sm text-slate-600 dark:text-slate-300">加载中</p>

    <div class="grid gap-4 md:grid-cols-3">
      <article v-for="item in packages" :key="item.id" class="rounded border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-900">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold">{{ item.name }}</h2>
            <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">有效期 {{ item.valid_days }} 天</p>
          </div>
          <span class="rounded bg-teal px-2 py-1 text-sm text-white">{{ item.credits }} 积分</span>
        </div>
        <div class="mt-5 rounded-xl border border-teal/15 bg-teal/5 px-4 py-3 text-sm text-slate-700 dark:border-teal/25 dark:bg-teal/10 dark:text-slate-200">
          <p class="font-medium text-teal">约可生成 {{ standardImageCount(item) }} 张标准图</p>
          <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">按 1024 x 1024、1 积分/张估算</p>
        </div>
        <div class="mt-6 flex items-end gap-1">
          <span class="text-3xl font-semibold">¥{{ item.price }}</span>
        </div>
        <button
          class="mt-5 min-h-11 w-full rounded bg-coral px-4 py-2 text-white disabled:cursor-not-allowed disabled:opacity-70"
          type="button"
          :disabled="payingId === item.id"
          @click="buyPackage(item)"
        >
          {{ payingId === item.id ? '创建订单中' : '立即购买' }}
        </button>
      </article>
    </div>

    <section class="rounded-2xl border border-slate-200 bg-white p-5 dark:border-slate-700 dark:bg-slate-900">
      <h2 class="text-lg font-semibold">积分使用规则</h2>
      <div class="mt-4 grid gap-3 md:grid-cols-3">
        <div class="rounded-xl bg-slate-50 p-4 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">
          <p class="font-medium text-slate-900 dark:text-slate-100">按尺寸计费</p>
          <p class="mt-2 leading-6">1024 x 1024 为 1 积分，更大尺寸按像素量向上取整。</p>
        </div>
        <div class="rounded-xl bg-slate-50 p-4 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">
          <p class="font-medium text-slate-900 dark:text-slate-100">有效期</p>
          <p class="mt-2 leading-6">套餐有效期以购买成功后写入账号的到期时间为准，请在有效期内使用。</p>
        </div>
        <div class="rounded-xl bg-slate-50 p-4 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">
          <p class="font-medium text-slate-900 dark:text-slate-100">失败退回</p>
          <p class="mt-2 leading-6">生成失败或开始前取消时，系统会按后端规则退回本次扣除的积分。</p>
        </div>
      </div>
    </section>

    <div v-if="creditLogsOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/55 px-4" role="dialog" aria-modal="true" aria-labelledby="credit-logs-title" @click.self="creditLogsOpen = false">
      <div class="w-full max-w-2xl rounded-2xl bg-white p-5 text-slate-900 shadow-xl dark:bg-slate-900 dark:text-slate-100">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h2 id="credit-logs-title" class="text-lg font-semibold">我的积分流水</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">最近 10 条记录，共 {{ creditLogsTotal }} 条。</p>
          </div>
          <button class="rounded-full border border-slate-200 px-3 py-1 text-sm text-slate-500 transition hover:bg-slate-50 dark:border-slate-700 dark:hover:bg-slate-800" type="button" @click="creditLogsOpen = false">关闭</button>
        </div>
        <div class="mt-5 overflow-hidden rounded-xl border border-slate-200 dark:border-slate-700">
          <div v-if="creditLogsLoading" class="p-5 text-sm text-slate-500">加载中...</div>
          <div v-else-if="creditLogs.length === 0" class="p-5 text-sm text-slate-500">暂无积分流水。</div>
          <table v-else class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs text-slate-500 dark:bg-slate-950 dark:text-slate-400">
              <tr>
                <th class="px-4 py-3 font-medium">时间</th>
                <th class="px-4 py-3 font-medium">类型</th>
                <th class="px-4 py-3 font-medium">变动</th>
                <th class="px-4 py-3 font-medium">余额</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-slate-700">
              <tr v-for="log in creditLogs" :key="log.id">
                <td class="px-4 py-3 text-slate-500 dark:text-slate-400">{{ formatTime(log.created_at) }}</td>
                <td class="px-4 py-3">{{ creditTypeText(log.type) }}</td>
                <td class="px-4 py-3 font-medium" :class="log.amount >= 0 ? 'text-emerald-600' : 'text-rose-600'">{{ log.amount >= 0 ? '+' : '' }}{{ log.amount }}</td>
                <td class="px-4 py-3">{{ log.balance }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </section>
</template>
