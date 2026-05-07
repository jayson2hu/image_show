<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchSettings, saveSettings } from '@/api/admin'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const loading = ref(false)
const saving = ref(false)
const activeGroup = ref('account')
const settings = ref<Record<string, string>>({})
const revealed = ref<Record<string, boolean>>({})

const groups = [
  {
    id: 'account',
    title: '账号与额度',
    description: '注册开关、新用户赠送积分和额度用完后的联系提示。',
    keys: ['register_enabled', 'register_gift_credits', 'credit_exhausted_message', 'credit_exhausted_wechat_qrcode_url', 'credit_exhausted_qq'],
  },
  {
    id: 'wechat',
    title: '微信登录',
    description: '公众号二维码、验证码服务地址和访问凭证。',
    keys: ['wechat_auth_enabled', 'wechat_qrcode_url', 'wechat_server_address', 'wechat_server_token'],
  },
  {
    id: 'generation',
    title: '图像生成',
    description: '模型名称和前台可选尺寸比例。',
    keys: ['image_model', 'enabled_image_sizes'],
  },
  {
    id: 'storage',
    title: '图片存储',
    description: 'Cloudflare R2 上传和公开访问地址。',
    keys: ['r2_endpoint', 'r2_access_key', 'r2_secret_key', 'r2_bucket', 'r2_public_url'],
  },
  {
    id: 'captcha',
    title: '人机验证',
    description: 'Cloudflare Turnstile 验证开关和密钥。',
    keys: ['captcha_enabled', 'turnstile_site_key', 'turnstile_secret'],
  },
  {
    id: 'security',
    title: '安全与监控',
    description: 'IP 黑名单和每日消耗告警。',
    keys: ['ip_blacklist', 'monitor_daily_credit_threshold', 'monitor_alert_last_date'],
  },
]

const activeSettingGroup = computed(() => groups.find((item) => item.id === activeGroup.value) || groups[0])

onMounted(loadSettings)

async function loadSettings() {
  loading.value = true
  try {
    const response = await fetchSettings()
    settings.value = response.data.items || {}
  } catch (error: any) {
    toast.error(error.response?.data?.error || '设置加载失败')
  } finally {
    loading.value = false
  }
}

async function submitSettings() {
  saving.value = true
  try {
    await saveSettings(settings.value)
    toast.success('设置已保存')
    await loadSettings()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '设置保存失败')
  } finally {
    saving.value = false
  }
}

function settingLabel(key: string) {
  const map: Record<string, string> = {
    register_enabled: '注册开关',
    register_gift_credits: '新用户赠送积分',
    credit_exhausted_message: '额度用完提示',
    credit_exhausted_wechat_qrcode_url: '联系二维码',
    credit_exhausted_qq: '联系 QQ',
    wechat_auth_enabled: '微信登录开关',
    wechat_qrcode_url: '公众号二维码',
    wechat_server_address: 'WeChat Server 地址',
    wechat_server_token: 'WeChat Server Token',
    image_model: '图像模型',
    enabled_image_sizes: '启用尺寸',
    r2_endpoint: 'R2 Endpoint',
    r2_access_key: 'R2 Access Key',
    r2_secret_key: 'R2 Secret Key',
    r2_bucket: 'R2 Bucket',
    r2_public_url: 'R2 Public URL',
    captcha_enabled: '验证码开关',
    turnstile_site_key: 'Turnstile Site Key',
    turnstile_secret: 'Turnstile Secret',
    ip_blacklist: 'IP 黑名单',
    monitor_daily_credit_threshold: '每日积分告警阈值',
    monitor_alert_last_date: '最近告警日期',
  }
  return map[key] || key
}

function settingHelp(key: string) {
  const map: Record<string, string> = {
    register_gift_credits: '示例：10 表示注册即送 10 积分；0 表示不赠送。',
    enabled_image_sizes: '示例：square,portrait_3_4,story,landscape_4_3,widescreen',
    ip_blacklist: '一行一个 IP，或用英文逗号分隔。例如：1.2.3.4, 5.6.7.8',
    monitor_daily_credit_threshold: '示例：500 表示当天积分消耗超过 500 时触发告警。',
    wechat_server_address: '示例：https://your-domain.com/wechat',
    r2_public_url: '示例：https://img.example.com',
  }
  return map[key] || ''
}

function isSensitive(key: string) {
  return key.includes('secret') || key.includes('token') || key.includes('access_key')
}

function isTextarea(key: string) {
  return key.includes('message') || key.includes('headers') || key === 'ip_blacklist' || key === 'enabled_image_sizes'
}

function inputType(key: string) {
  if (isSensitive(key) && !revealed.value[key]) {
    return 'password'
  }
  if (key.includes('credits') || key.includes('threshold')) {
    return 'number'
  }
  return 'text'
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Settings</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">系统设置</h2>
        <p class="mt-2 text-sm text-slate-500">按场景维护注册、微信、生成、存储和安全配置。</p>
      </div>
      <button class="rounded-2xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:opacity-60" type="button" :disabled="saving" @click="submitSettings">
        {{ saving ? '保存中' : '保存设置' }}
      </button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <SkeletonCard v-for="item in 4" :key="item" />
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-[260px_minmax(0,1fr)]">
      <aside class="rounded-3xl border border-slate-200 bg-white p-3 shadow-sm">
        <button
          v-for="group in groups"
          :key="group.id"
          class="block w-full rounded-2xl px-4 py-3 text-left transition"
          :class="activeGroup === group.id ? 'bg-slate-950 text-white' : 'text-slate-600 hover:bg-slate-50'"
          type="button"
          @click="activeGroup = group.id"
        >
          <span class="block text-sm font-semibold">{{ group.title }}</span>
          <span class="mt-1 block text-xs opacity-70">{{ group.description }}</span>
        </button>
      </aside>

      <form class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm" @submit.prevent="submitSettings">
        <div class="border-b border-slate-100 pb-4">
          <h3 class="text-lg font-semibold text-slate-950">{{ activeSettingGroup.title }}</h3>
          <p class="mt-1 text-sm text-slate-500">{{ activeSettingGroup.description }}</p>
        </div>

        <div class="mt-5 grid gap-5">
          <label v-for="key in activeSettingGroup.keys" :key="key" class="block">
            <span class="text-sm font-medium text-slate-700">{{ settingLabel(key) }}</span>
            <span v-if="settingHelp(key)" class="mt-1 block text-xs leading-5 text-slate-500">{{ settingHelp(key) }}</span>
            <div class="mt-2 flex gap-2">
              <textarea v-if="isTextarea(key)" v-model="settings[key]" class="setting-input min-h-24 py-3"></textarea>
              <input v-else v-model="settings[key]" class="setting-input" :type="inputType(key)" />
              <button v-if="isSensitive(key)" class="shrink-0 rounded-xl border border-slate-200 px-3 text-sm text-slate-600 transition hover:bg-slate-50" type="button" @click="revealed[key] = !revealed[key]">
                {{ revealed[key] ? '隐藏' : '显示' }}
              </button>
            </div>
          </label>
        </div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.setting-input {
  min-height: 2.75rem;
  width: 100%;
  border-radius: 1rem;
  border: 1px solid rgb(226 232 240);
  padding-left: 1rem;
  padding-right: 1rem;
  font-size: 0.875rem;
  outline: none;
}

.setting-input:focus {
  border-color: rgb(20 184 166);
  box-shadow: 0 0 0 2px rgb(20 184 166 / 0.2);
}
</style>
