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

const packages = ref<CreditPackage[]>([])
const loading = ref(false)
const error = ref('')
const payingId = ref<number | null>(null)
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
</script>

<template>
  <section class="space-y-6">
    <div class="border-b border-slate-200 pb-4 dark:border-slate-800">
      <h1 class="text-2xl font-semibold">积分套餐</h1>
      <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">选择适合的积分包，创建订单后跳转到支付平台完成付款。</p>
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
  </section>
</template>
