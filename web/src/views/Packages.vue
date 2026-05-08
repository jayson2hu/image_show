<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import api from '@/api'

interface CreditPackage {
  id: number
  name: string
  credits: number
  price: number
  valid_days: number
}

interface SupportContact {
  manual_recharge_enabled: string
  manual_recharge_wechat_id: string
  manual_recharge_wechat_qrcode_url: string
  manual_recharge_qq: string
  manual_recharge_note: string
}

const packages = ref<CreditPackage[]>([])
const contact = ref<SupportContact>({
  manual_recharge_enabled: 'true',
  manual_recharge_wechat_id: '',
  manual_recharge_wechat_qrcode_url: '',
  manual_recharge_qq: '',
  manual_recharge_note: '',
})
const selectedPackage = ref<CreditPackage | null>(null)
const loading = ref(false)
const error = ref('')
const copied = ref('')
const router = useRouter()

const rechargeEnabled = computed(() => contact.value.manual_recharge_enabled !== 'false')
const hasContact = computed(() => Boolean(wechatId.value || wechatQRCodeURL.value || qqContact.value))
const wechatId = computed(() => contact.value.manual_recharge_wechat_id.trim())
const wechatQRCodeURL = computed(() => contact.value.manual_recharge_wechat_qrcode_url.trim())
const qqContact = computed(() => contact.value.manual_recharge_qq.trim())
const rechargeNote = computed(() => contact.value.manual_recharge_note.trim() || '请联系管理员人工充值，并备注账号邮箱和需要开通的套餐。')

onMounted(loadPage)

async function loadPage() {
  loading.value = true
  error.value = ''
  try {
    const [packageResponse, contactResponse] = await Promise.all([
      api.get('/packages'),
      api.get('/support/contact'),
    ])
    packages.value = packageResponse.data.items || []
    contact.value = {
      manual_recharge_enabled: contactResponse.data.manual_recharge_enabled || 'true',
      manual_recharge_wechat_id: contactResponse.data.manual_recharge_wechat_id || '',
      manual_recharge_wechat_qrcode_url: contactResponse.data.manual_recharge_wechat_qrcode_url || '',
      manual_recharge_qq: contactResponse.data.manual_recharge_qq || '',
      manual_recharge_note: contactResponse.data.manual_recharge_note || '',
    }
    selectedPackage.value = packages.value[0] || null
  } catch (err: any) {
    error.value = err.response?.data?.error || '套餐和充值方式加载失败'
  } finally {
    loading.value = false
  }
}

function selectPackage(item: CreditPackage) {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }
  selectedPackage.value = item
}

function standardImageCount(item: CreditPackage) {
  return Math.max(0, Math.floor(Number(item.credits) || 0))
}

function openCreditLogs() {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }
  router.push('/credits')
}

async function copyText(label: string, value: string) {
  if (!value) return
  try {
    await navigator.clipboard.writeText(value)
    copied.value = `${label}已复制`
  } catch {
    copied.value = '复制失败，请手动复制'
  }
  window.setTimeout(() => {
    copied.value = ''
  }, 1800)
}
</script>

<template>
  <section class="space-y-6">
    <RouterLink class="inline-flex rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-50" to="/account">
      返回个人中心
    </RouterLink>
    <div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-sm font-medium text-teal">Credits</p>
          <h1 class="mt-1 text-2xl font-semibold text-slate-950">积分套餐</h1>
          <p class="mt-2 max-w-2xl text-sm leading-6 text-slate-500">
            选择需要的积分套餐后，通过后台配置的联系方式联系管理员人工开通。当前版本不会自动跳转第三方支付。
          </p>
        </div>
        <button class="rounded-2xl border border-slate-200 px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50" type="button" @click="openCreditLogs">
          查看积分流水
        </button>
      </div>
    </div>

    <p v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</p>
    <p v-if="copied" class="rounded-2xl border border-teal/20 bg-teal/10 px-4 py-3 text-sm text-teal">{{ copied }}</p>
    <p v-if="loading" class="rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-500">加载中...</p>

    <div v-else class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
      <div class="grid gap-4 md:grid-cols-3">
        <article
          v-for="item in packages"
          :key="item.id"
          class="rounded-3xl border bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
          :class="selectedPackage?.id === item.id ? 'border-teal ring-2 ring-teal/15' : 'border-slate-200'"
        >
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-lg font-semibold text-slate-950">{{ item.name }}</h2>
              <p class="mt-1 text-sm text-slate-500">有效期 {{ item.valid_days }} 天</p>
            </div>
            <span class="rounded-full bg-teal/10 px-3 py-1 text-sm font-semibold text-teal">{{ item.credits }} 积分</span>
          </div>
          <div class="mt-5 rounded-2xl border border-slate-100 bg-slate-50 px-4 py-3 text-sm text-slate-600">
            <p class="font-medium text-slate-900">约可生成 {{ standardImageCount(item) }} 张标准图</p>
            <p class="mt-1 text-xs text-slate-500">按 1024 x 1024、1 积分/张估算</p>
          </div>
          <div class="mt-6 flex items-end gap-1">
            <span class="text-3xl font-semibold text-slate-950">¥{{ item.price }}</span>
          </div>
          <button
            class="mt-5 min-h-11 w-full rounded-2xl bg-coral px-4 py-2 text-sm font-semibold text-white transition hover:bg-coral/90"
            type="button"
            @click="selectPackage(item)"
          >
            {{ selectedPackage?.id === item.id ? '已选择' : '选择套餐' }}
          </button>
        </article>
      </div>

      <aside class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
        <div>
          <p class="text-sm font-medium text-teal">Manual recharge</p>
          <h2 class="mt-1 text-xl font-semibold text-slate-950">联系管理员充值</h2>
          <p class="mt-2 text-sm leading-6 text-slate-500">{{ rechargeNote }}</p>
        </div>

        <div v-if="selectedPackage" class="mt-5 rounded-2xl border border-slate-100 bg-slate-50 p-4">
          <p class="text-xs font-medium uppercase tracking-wide text-slate-400">已选套餐</p>
          <p class="mt-2 text-base font-semibold text-slate-950">{{ selectedPackage.name }}</p>
          <p class="mt-1 text-sm text-slate-500">¥{{ selectedPackage.price }} / {{ selectedPackage.credits }} 积分</p>
        </div>

        <div v-if="!rechargeEnabled" class="mt-5 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
          当前暂未开放人工充值，请稍后再试。
        </div>

        <div v-else-if="!hasContact" class="mt-5 rounded-2xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
          暂未配置联系方式，请联系站点管理员配置充值方式。
        </div>

        <div v-else class="mt-5 space-y-4">
          <img v-if="wechatQRCodeURL" class="aspect-square w-full rounded-2xl border border-slate-200 object-cover" :src="wechatQRCodeURL" alt="充值微信二维码" />

          <div v-if="wechatId" class="rounded-2xl border border-slate-200 p-4">
            <p class="text-xs text-slate-400">微信号</p>
            <div class="mt-2 flex items-center justify-between gap-3">
              <span class="break-all text-sm font-semibold text-slate-900">{{ wechatId }}</span>
              <button class="shrink-0 rounded-xl border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50" type="button" @click="copyText('微信号', wechatId)">复制</button>
            </div>
          </div>

          <div v-if="qqContact" class="rounded-2xl border border-slate-200 p-4">
            <p class="text-xs text-slate-400">QQ</p>
            <div class="mt-2 flex items-center justify-between gap-3">
              <span class="break-all text-sm font-semibold text-slate-900">{{ qqContact }}</span>
              <button class="shrink-0 rounded-xl border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50" type="button" @click="copyText('QQ', qqContact)">复制</button>
            </div>
          </div>
        </div>
      </aside>
    </div>

    <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
      <h2 class="text-lg font-semibold text-slate-950">积分使用规则</h2>
      <div class="mt-4 grid gap-3 md:grid-cols-3">
        <div class="rounded-2xl bg-slate-50 p-4 text-sm text-slate-600">
          <p class="font-medium text-slate-900">按尺寸计费</p>
          <p class="mt-2 leading-6">1024 x 1024 为 1 积分，更大尺寸按像素量向上取整。</p>
        </div>
        <div class="rounded-2xl bg-slate-50 p-4 text-sm text-slate-600">
          <p class="font-medium text-slate-900">有效期</p>
          <p class="mt-2 leading-6">套餐有效期以人工开通后写入账号的到期时间为准。</p>
        </div>
        <div class="rounded-2xl bg-slate-50 p-4 text-sm text-slate-600">
          <p class="font-medium text-slate-900">失败退回</p>
          <p class="mt-2 leading-6">生成失败或开始前取消时，系统会按后端规则退回本次扣除的积分。</p>
        </div>
      </div>
    </section>
  </section>
</template>
