<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api'

type CreditGuideType = 'free_trial_exhausted' | 'insufficient_credits' | 'credits_expired'

const props = defineProps<{
  type: CreditGuideType
}>()

const emit = defineEmits<{
  dismiss: []
}>()

const router = useRouter()
const supportContact = ref({
  credit_exhausted_message: '',
  credit_exhausted_wechat_qrcode_url: '',
  credit_exhausted_qq: '',
})

const guide = computed(() => {
  const map = {
    free_trial_exhausted: {
      title: '这次免费体验已经用完',
      description: '注册账号后会赠送新用户积分，你刚才填写的提示词和尺寸会保留，登录后可以继续生成，也能在历史记录里找回作品。',
      primaryText: '注册领取积分',
      primaryRoute: '/register',
      secondaryText: '已有账号？去登录',
      secondaryRoute: '/login',
      iconPath: 'M12 3.75l1.93 3.91 4.32.63-3.13 3.05.74 4.3L12 13.61l-3.86 2.03.74-4.3-3.13-3.05 4.32-.63L12 3.75z',
    },
    insufficient_credits: {
      title: '积分不足',
      description: '当前余额不够生成这张图。你可以先换成消耗更低的尺寸，或购买积分后继续生成。',
      primaryText: '查看积分套餐',
      primaryRoute: '/packages',
      secondaryText: '',
      secondaryRoute: '',
      iconPath: 'M4.5 7.5h15A1.5 1.5 0 0121 9v7.5a1.5 1.5 0 01-1.5 1.5h-15A1.5 1.5 0 013 16.5V6a1.5 1.5 0 011.5-1.5h12M17.25 13.5h.01',
    },
    credits_expired: {
      title: '积分已过期',
      description: '你的积分已经过期，当前不能继续生成。购买新的积分包后，可以继续创作和下载作品。',
      primaryText: '查看积分套餐',
      primaryRoute: '/packages',
      secondaryText: '',
      secondaryRoute: '',
      iconPath: 'M12 6v6l3 2m6-2a9 9 0 11-18 0 9 9 0 0118 0z',
    },
  }
  return map[props.type]
})

const contactMessage = computed(() => supportContact.value.credit_exhausted_message.trim())
const wechatQRCodeURL = computed(() => supportContact.value.credit_exhausted_wechat_qrcode_url.trim())
const qqContact = computed(() => supportContact.value.credit_exhausted_qq.trim())

function goPrimary() {
  router.push(guide.value.primaryRoute)
}

function goSecondary() {
  if (guide.value.secondaryRoute) {
    router.push(guide.value.secondaryRoute)
  }
}

async function loadSupportContact() {
  try {
    const response = await api.get('/support/contact')
    supportContact.value = {
      credit_exhausted_message: response.data.credit_exhausted_message || '',
      credit_exhausted_wechat_qrcode_url: response.data.credit_exhausted_wechat_qrcode_url || '',
      credit_exhausted_qq: response.data.credit_exhausted_qq || '',
    }
  } catch {
    supportContact.value = {
      credit_exhausted_message: '',
      credit_exhausted_wechat_qrcode_url: '',
      credit_exhausted_qq: '',
    }
  }
}

onMounted(loadSupportContact)
</script>

<template>
  <div class="relative w-full max-w-md rounded-2xl border border-violet-200 bg-gradient-to-br from-violet-50 to-blue-50 p-6 shadow-xl shadow-violet-900/10 dark:border-violet-400/30 dark:from-slate-900 dark:to-slate-950 dark:shadow-black/30">
    <button
      class="absolute right-3 top-3 flex size-8 items-center justify-center rounded-full text-slate-400 transition hover:bg-white/80 hover:text-slate-700 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-200"
      type="button"
      aria-label="关闭引导"
      @click="emit('dismiss')"
    >
      ×
    </button>

    <div class="mb-4 flex size-12 items-center justify-center rounded-2xl bg-white text-violet-700 shadow-sm dark:bg-violet-500/15 dark:text-violet-200">
      <svg class="size-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" :d="guide.iconPath" />
      </svg>
    </div>

    <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ guide.title }}</h2>
    <p class="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">{{ guide.description }}</p>

    <div v-if="contactMessage || wechatQRCodeURL || qqContact" class="mt-4 rounded-xl border border-white/80 bg-white/75 p-4 text-left shadow-sm dark:border-slate-700 dark:bg-slate-900/70">
      <p v-if="contactMessage" class="text-sm leading-6 text-slate-700 dark:text-slate-200">{{ contactMessage }}</p>
      <div v-if="wechatQRCodeURL || qqContact" class="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
        <img v-if="wechatQRCodeURL" class="size-28 rounded-lg border border-slate-200 bg-white object-contain p-1 dark:border-slate-700" :src="wechatQRCodeURL" alt="微信联系二维码" />
        <div v-if="qqContact" class="rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-700 dark:bg-slate-800 dark:text-slate-200">
          <span class="text-slate-500 dark:text-slate-400">QQ 联系：</span>
          <span class="font-semibold">{{ qqContact }}</span>
        </div>
      </div>
    </div>

    <button class="mt-5 w-full rounded-xl bg-violet-600 py-3 font-semibold text-white transition hover:bg-violet-700" type="button" @click="goPrimary">
      {{ guide.primaryText }}
    </button>
    <button v-if="guide.secondaryText" class="mt-3 w-full text-center text-sm font-medium text-violet-600 transition hover:text-violet-800 dark:text-violet-300 dark:hover:text-violet-100" type="button" @click="goSecondary">
      {{ guide.secondaryText }}
    </button>
  </div>
</template>
