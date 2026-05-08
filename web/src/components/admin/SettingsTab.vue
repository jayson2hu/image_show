<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchSettings, saveSettings } from '@/api/admin'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const loading = ref(false)
const saving = ref(false)
const activeGroup = ref('site')
const settings = ref<Record<string, string>>({})
const revealed = ref<Record<string, boolean>>({})

const groups = [
  {
    id: 'site',
    title: '站点与 SEO',
    description: '配置网站标题、关于网站、搜索展示标题、关键词和描述。',
    keys: ['site_title', 'site_about', 'seo_title', 'seo_keywords', 'seo_description'],
  },
  {
    id: 'registration',
    title: '注册策略',
    description: '控制前台注册入口和允许注册的邮箱后缀，后台创建用户不受这里限制。',
    keys: ['register_enabled', 'register_email_domain_allowlist'],
  },
  {
    id: 'account',
    title: '账号与额度',
    description: '注册开关、新用户赠送积分和额度用完后的联系提示。',
    keys: ['register_gift_credits', 'credit_exhausted_message', 'credit_exhausted_wechat_qrcode_url', 'credit_exhausted_qq'],
  },
  {
    id: 'manual-recharge',
    title: '人工充值',
    description: '配置购买中心展示的人工充值联系方式，第一版不接真实支付渠道。',
    keys: ['manual_recharge_enabled', 'manual_recharge_wechat_id', 'manual_recharge_wechat_qrcode_url', 'manual_recharge_qq', 'manual_recharge_note'],
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
    id: 'avatar-storage',
    title: '头像存储',
    description: '配置用户头像上传限制。当前版本仅使用本地存储，R2 迁移后续单独执行。',
    keys: ['avatar_storage_driver', 'avatar_max_size_mb', 'avatar_allowed_types'],
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
    const payload = activeSettingGroup.value.keys.reduce<Record<string, string>>((items, key) => {
      items[key] = settings.value[key] || ''
      return items
    }, {})
    await saveSettings(payload)
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
    manual_recharge_enabled: '启用人工充值',
    manual_recharge_wechat_id: '充值微信号',
    manual_recharge_wechat_qrcode_url: '充值微信二维码 URL',
    manual_recharge_qq: '充值 QQ',
    manual_recharge_note: '充值说明',
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
    avatar_storage_driver: '头像存储方式',
    avatar_max_size_mb: '头像最大大小 MB',
    avatar_allowed_types: '允许头像格式',
    captcha_enabled: '验证码开关',
    turnstile_site_key: 'Turnstile Site Key',
    turnstile_secret: 'Turnstile Secret',
    ip_blacklist: 'IP 黑名单',
    monitor_daily_credit_threshold: '每日积分告警阈值',
    monitor_alert_last_date: '最近告警日期',
  }
  Object.assign(map, {
    site_title: '网站标题',
    site_about: '关于网站',
    seo_title: 'SEO 标题',
    seo_keywords: 'SEO 关键词',
    seo_description: 'SEO 描述',
    register_enabled: '注册开关',
    register_email_domain_allowlist: '允许注册邮箱后缀',
    register_gift_credits: '新用户赠送积分',
    credit_exhausted_message: '额度用完提示',
    credit_exhausted_wechat_qrcode_url: '联系二维码',
    credit_exhausted_qq: '联系 QQ',
    manual_recharge_enabled: '启用人工充值',
    manual_recharge_wechat_id: '充值微信号',
    manual_recharge_wechat_qrcode_url: '充值微信二维码 URL',
    manual_recharge_qq: '充值 QQ',
    manual_recharge_note: '充值说明',
    image_model: '图像模型',
    enabled_image_sizes: '启用尺寸',
    captcha_enabled: '验证码开关',
    ip_blacklist: 'IP 黑名单',
    monitor_daily_credit_threshold: '每日积分告警阈值',
    monitor_alert_last_date: '最近告警日期',
  })
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
  Object.assign(map, {
    site_title: '展示在浏览器标题、顶部品牌和默认 SEO 中，例如：来看看巴。',
    site_about: '用于关于网站介绍，建议一句话说明网站能帮用户做什么。',
    seo_title: '搜索结果标题，建议包含站点名称和核心能力。',
    seo_keywords: '多个关键词用英文逗号分隔，例如：AI图片生成,AI绘画,图片编辑。',
    seo_description: '搜索结果描述，建议 50-120 字，说明用户能在这里完成什么。',
    register_enabled: 'true 表示允许前台注册，false 表示关闭前台注册；管理员后台创建用户不受影响。',
    register_email_domain_allowlist: '留空表示不限制；多个后缀用英文逗号或换行分隔，例如：qq.com, gmail.com。',
    register_gift_credits: '例如：10 表示注册即送 10 积分，0 表示不赠送。',
    manual_recharge_enabled: 'true 表示在购买中心展示人工充值入口；false 表示暂时隐藏充值联系方式。',
    manual_recharge_wechat_id: '填写客服微信号，留空则前台不展示微信号。',
    manual_recharge_wechat_qrcode_url: '填写可公开访问的二维码图片地址，留空则前台不展示二维码。',
    manual_recharge_qq: '填写 QQ 号码，留空则前台不展示 QQ。',
    manual_recharge_note: '用于说明充值备注、到账时间、客服时间等，避免用户误以为已经自动支付。',
    enabled_image_sizes: '例如：square,portrait_3_4,story,landscape_4_3,widescreen。',
    ip_blacklist: '一行一个 IP 或 CIDR，例如：1.2.3.4 或 10.0.0.0/24。',
    monitor_daily_credit_threshold: '例如：500 表示当天积分消耗超过 500 时触发告警。',
    wechat_server_address: '例如：https://your-domain.com/wechat。',
    r2_public_url: '例如：https://img.example.com。',
    avatar_storage_driver: '当前只支持 local。本地头像会保存到后端 uploads/avatars 目录；后续启用 R2 前需要先做迁移。',
    avatar_max_size_mb: '例如：2 表示最大 2MB。后端会限制最高 10MB，避免上传过大文件。',
    avatar_allowed_types: '例如：jpg,jpeg,png,webp。多个格式用英文逗号或换行分隔。',
  })
  return map[key] || ''
}

function isSensitive(key: string) {
  return key.includes('secret') || key.includes('token') || key.includes('access_key')
}

function isTextarea(key: string) {
  return key.includes('message') || key.includes('headers') || key === 'manual_recharge_note' || key === 'avatar_allowed_types' || key === 'ip_blacklist' || key === 'enabled_image_sizes' || key === 'site_about' || key === 'seo_description' || key === 'register_email_domain_allowlist'
}

function inputType(key: string) {
  if (isSensitive(key) && !revealed.value[key]) {
    return 'password'
  }
  if (key.includes('credits') || key.includes('threshold') || key === 'avatar_max_size_mb') {
    return 'number'
  }
  return 'text'
}

function settingPlaceholder(key: string) {
  const map: Record<string, string> = {
    site_title: '来看看巴',
    site_about: '把想法变成一张好图。',
    seo_title: '来看看巴 - AI 图片生成',
    seo_keywords: 'AI图片生成,AI绘画,图片编辑',
    seo_description: '输入提示词，选择比例，持续查看生成进度，直到作品完成。',
    register_email_domain_allowlist: 'qq.com\ngmail.com\ncompany.com',
    manual_recharge_enabled: 'true',
    manual_recharge_wechat_id: 'image-show-admin',
    manual_recharge_wechat_qrcode_url: 'https://img.example.com/recharge-wechat.png',
    manual_recharge_qq: '123456',
    manual_recharge_note: '添加客服后请备注账号邮箱和套餐名称，工作日 10:00-19:00 处理。',
    avatar_storage_driver: 'local',
    avatar_max_size_mb: '2',
    avatar_allowed_types: 'jpg,jpeg,png,webp',
  }
  return map[key] || ''
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
              <textarea v-if="isTextarea(key)" v-model="settings[key]" class="setting-input min-h-24 py-3" :placeholder="settingPlaceholder(key)"></textarea>
              <input v-else v-model="settings[key]" class="setting-input" :type="inputType(key)" :placeholder="settingPlaceholder(key)" />
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
