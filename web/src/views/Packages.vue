<script setup lang="ts">
import { onMounted, ref } from 'vue'

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
</script>

<template>
  <section class="space-y-6">
    <div class="border-b border-slate-200 pb-4 dark:border-slate-800">
      <h1 class="text-2xl font-semibold">积分套餐</h1>
      <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">选择适合的积分包，支付接入后可直接购买。</p>
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
        <button class="mt-5 min-h-11 w-full rounded bg-coral px-4 py-2 text-white opacity-80" type="button" disabled>
          支付接入中
        </button>
      </article>
    </div>
  </section>
</template>
