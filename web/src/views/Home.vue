<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import api from '@/api'
import GenerationProgress from '@/components/GenerationProgress.vue'
import PromptTags from '@/components/PromptTags.vue'
import { useUserStore } from '@/stores/user'

type Quality = 'low' | 'medium' | 'high'

interface GenerationPayload {
  prompt: string
  quality: Quality
  size: string
}

interface StylePreset {
  id: string
  name: string
  prompt: string
}

interface SamplePrompt {
  title: string
  prompt: string
  style: string
}

const userStore = useUserStore()
const health = ref('检查中')
const prompt = ref('')
const negativePrompt = ref('')
const selectedStyle = ref('realistic')
const quality = ref<Quality>('medium')
const size = ref('1024x1024')
const imageCount = ref(1)
const creativity = ref(0.7)
const steps = ref(30)
const cfgScale = ref(7)
const generationId = ref<number | null>(null)
const imageURL = ref('')
const error = ref('')
const loading = ref(false)
const lastRequest = ref<GenerationPayload | null>(null)
const captchaEnabled = ref(false)
const captchaSiteKey = ref('')
const captchaToken = ref('')
const captchaEl = ref<HTMLElement | null>(null)
const captchaWidgetId = ref<string | null>(null)

const costs: Record<Quality, number> = { low: 0.2, medium: 1, high: 4 }
const qualityLabels: Record<Quality, string> = { low: '快速', medium: '标准', high: '高清' }
const stylePresets: StylePreset[] = [
  { id: 'realistic', name: '写实', prompt: '写实摄影风格，细节丰富，自然光影' },
  { id: 'anime', name: '动漫', prompt: '动漫插画风格，清晰线稿，高饱和色彩' },
  { id: 'fantasy', name: '幻想', prompt: '幻想艺术风格，史诗氛围，电影级构图' },
  { id: 'cyberpunk', name: '赛博朋克', prompt: '赛博朋克风格，霓虹灯光，未来城市质感' },
  { id: 'watercolor', name: '水彩', prompt: '水彩画风格，柔和笔触，温暖色调' },
  { id: 'abstract', name: '抽象', prompt: '抽象艺术风格，流动光影，紫蓝渐变' },
]
const samplePrompts: SamplePrompt[] = [
  { title: '幻想风景', prompt: '沙漠中的神秘传送门，超现实主义，4K高清', style: 'fantasy' },
  { title: '赛博朋克城市', prompt: '未来城市夜景，霓虹灯，赛博朋克风格', style: 'cyberpunk' },
  { title: '水彩画', prompt: '森林中的小木屋，温暖色调，水彩画风格', style: 'watercolor' },
  { title: '抽象艺术', prompt: '流动的光影，紫蓝渐变，抽象艺术', style: 'abstract' },
]

const selectedStylePrompt = computed(() => stylePresets.find((item) => item.id === selectedStyle.value)?.prompt || '')
const cost = computed(() => costs[quality.value])
const canRetry = computed(() => !!lastRequest.value && !loading.value)
const displayName = computed(() => userStore.user?.email.split('@')[0] || '访客')
const creditText = computed(() => (userStore.user ? `${userStore.user.credits} 积分` : '免费试用 1 次'))
const canGenerate = computed(() => prompt.value.trim().length > 0 && !loading.value)

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string
      reset: (widgetId?: string | null) => void
    }
  }
}

onMounted(async () => {
  try {
    const response = await api.get('/health')
    health.value = response.data.status === 'ok' ? '后端已连接' : '后端响应异常'
  } catch {
    health.value = '后端未连接'
  }
  await loadCaptcha()
})

function appendPrompt(value: string) {
  prompt.value = prompt.value ? `${prompt.value}，${value}` : value
}

function useSample(sample: SamplePrompt) {
  prompt.value = sample.prompt
  selectedStyle.value = sample.style
}

function buildPrompt() {
  const parts = [prompt.value.trim()]
  if (selectedStylePrompt.value) {
    parts.push(selectedStylePrompt.value)
  }
  if (negativePrompt.value.trim()) {
    parts.push(`负面提示词：${negativePrompt.value.trim()}`)
  }
  return parts.filter(Boolean).join('\n')
}

async function generate() {
  await createGeneration({
    prompt: buildPrompt(),
    quality: quality.value,
    size: size.value,
  })
}

async function retry() {
  if (!lastRequest.value) {
    return
  }
  await createGeneration(lastRequest.value)
}

async function createGeneration(payload: GenerationPayload) {
  error.value = ''
  if (captchaEnabled.value && !captchaToken.value) {
    error.value = '请先完成人机验证'
    return
  }
  imageURL.value = ''
  generationId.value = null
  loading.value = true
  lastRequest.value = { ...payload }
  try {
    const response = await api.post('/generations', { ...payload, captcha_token: captchaToken.value })
    generationId.value = response.data.id
  } catch (err: any) {
    error.value = err.response?.data?.error || '创建生成任务失败'
    loading.value = false
    resetCaptcha()
  }
}

async function cancelGeneration() {
  if (!generationId.value) {
    return
  }
  try {
    await api.post(`/generations/${generationId.value}/cancel`)
    error.value = '任务已取消'
  } catch (err: any) {
    error.value = err.response?.data?.error || '取消失败'
  } finally {
    generationId.value = null
    loading.value = false
    userStore.fetchUser()
    resetCaptcha()
  }
}

function completed(url: string) {
  imageURL.value = url
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
  resetCaptcha()
}

function failed(message: string) {
  error.value = message
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
  resetCaptcha()
}

function cancelled() {
  error.value = '任务已取消'
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
  resetCaptcha()
}

async function loadCaptcha() {
  const response = await api.get('/captcha/config')
  captchaEnabled.value = response.data.enabled
  captchaSiteKey.value = response.data.site_key
  if (!captchaEnabled.value || !captchaSiteKey.value) {
    return
  }
  await loadTurnstileScript()
  await nextTick()
  renderCaptcha()
}

function loadTurnstileScript() {
  if (window.turnstile) {
    return Promise.resolve()
  }
  return new Promise<void>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>('script[data-turnstile]')
    if (existing) {
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', () => reject(new Error('turnstile load failed')), { once: true })
      return
    }
    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.dataset.turnstile = 'true'
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('turnstile load failed'))
    document.head.appendChild(script)
  })
}

function renderCaptcha() {
  if (!captchaEl.value || !window.turnstile || captchaWidgetId.value) {
    return
  }
  captchaWidgetId.value = window.turnstile.render(captchaEl.value, {
    sitekey: captchaSiteKey.value,
    callback: (token: string) => {
      captchaToken.value = token
    },
    'expired-callback': () => {
      captchaToken.value = ''
    },
    'error-callback': () => {
      captchaToken.value = ''
    },
  })
}

function resetCaptcha() {
  captchaToken.value = ''
  if (window.turnstile && captchaWidgetId.value) {
    window.turnstile.reset(captchaWidgetId.value)
  }
}
</script>

<template>
  <section class="min-h-[calc(100vh-65px)] bg-gray-50 text-slate-950">
    <div class="flex min-h-[calc(100vh-65px)] flex-col lg:flex-row">
      <aside class="w-full shrink-0 border-b border-gray-200 bg-white lg:h-[calc(100vh-65px)] lg:w-[420px] lg:overflow-y-auto lg:border-b-0 lg:border-r">
        <div class="space-y-6 p-5 sm:p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <h1 class="text-2xl font-semibold">AI 图片生成器</h1>
              <p class="mt-1 text-sm text-gray-500">{{ health }}</p>
            </div>
            <div class="rounded-full bg-violet-50 px-3 py-1 text-sm font-medium text-violet-700">{{ creditText }}</div>
          </div>

          <div class="rounded-xl border border-gray-200 bg-gray-50 p-3 text-sm">
            <div class="flex items-center justify-between gap-3">
              <span class="font-medium text-gray-700">{{ displayName }}</span>
              <RouterLink v-if="!userStore.user" to="/login" class="rounded-lg bg-gradient-to-r from-violet-600 to-blue-600 px-3 py-2 text-white shadow-sm">
                登录 / 注册
              </RouterLink>
              <RouterLink v-else to="/history" class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-gray-700 hover:border-violet-300">
                历史记录
              </RouterLink>
            </div>
          </div>

          <label class="block space-y-2">
            <span class="text-sm font-medium text-gray-700">提示词</span>
            <textarea
              v-model="prompt"
              class="h-24 w-full resize-none rounded-xl border border-gray-200 bg-white p-3 text-sm outline-none transition focus:border-violet-400 focus:ring-4 focus:ring-violet-100"
              placeholder="描述你想生成的图片，例如：沙漠中的神秘传送门，超现实主义，4K高清"
            />
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-medium text-gray-700">负面提示词</span>
            <textarea
              v-model="negativePrompt"
              class="h-20 w-full resize-none rounded-xl border border-gray-200 bg-white p-3 text-sm outline-none transition focus:border-violet-400 focus:ring-4 focus:ring-violet-100"
              placeholder="不希望出现的元素，例如：模糊、低清晰度、文字水印"
            />
          </label>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h2 class="text-sm font-medium text-gray-700">风格预设</h2>
              <span class="text-xs text-gray-400">选择后会合并到提示词</span>
            </div>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="style in stylePresets"
                :key="style.id"
                type="button"
                class="min-h-11 rounded-xl border px-3 py-2 text-sm transition"
                :class="selectedStyle === style.id ? 'border-violet-500 bg-violet-50 text-violet-700 shadow-sm' : 'border-gray-200 bg-white text-gray-700 hover:border-violet-300'"
                @click="selectedStyle = style.id"
              >
                {{ style.name }}
              </button>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
            <label class="text-sm font-medium text-gray-700">
              质量
              <select v-model="quality" class="mt-2 min-h-11 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 outline-none focus:border-violet-400">
                <option value="low">快速 - 0.2 积分</option>
                <option value="medium">标准 - 1 积分</option>
                <option value="high">高清 - 4 积分</option>
              </select>
            </label>
            <label class="text-sm font-medium text-gray-700">
              尺寸
              <select v-model="size" class="mt-2 min-h-11 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 outline-none focus:border-violet-400">
                <option value="1024x1024">1024 x 1024</option>
                <option value="1024x1536">1024 x 1536</option>
                <option value="1536x1024">1536 x 1024</option>
              </select>
            </label>
          </div>

          <div class="space-y-4 rounded-xl border border-gray-200 bg-white p-4">
            <div class="flex items-center justify-between">
              <h2 class="text-sm font-medium text-gray-700">高级参数</h2>
              <span class="text-xs text-gray-400">当前模型参考</span>
            </div>
            <label class="block text-sm text-gray-600">
              <span class="flex justify-between"><span>图片数量</span><strong>{{ imageCount }}</strong></span>
              <input v-model.number="imageCount" class="figma-range mt-2 w-full" type="range" min="1" max="4" step="1" />
            </label>
            <label class="block text-sm text-gray-600">
              <span class="flex justify-between"><span>创造力</span><strong>{{ creativity.toFixed(1) }}</strong></span>
              <input v-model.number="creativity" class="figma-range mt-2 w-full" type="range" min="0" max="1" step="0.1" />
            </label>
            <label class="block text-sm text-gray-600">
              <span class="flex justify-between"><span>步数</span><strong>{{ steps }}</strong></span>
              <input v-model.number="steps" class="figma-range mt-2 w-full" type="range" min="10" max="50" step="1" />
            </label>
            <label class="block text-sm text-gray-600">
              <span class="flex justify-between"><span>CFG Scale</span><strong>{{ cfgScale }}</strong></span>
              <input v-model.number="cfgScale" class="figma-range mt-2 w-full" type="range" min="1" max="20" step="1" />
            </label>
          </div>

          <div class="space-y-3">
            <h2 class="text-sm font-medium text-gray-700">推荐示例</h2>
            <div class="space-y-2">
              <button
                v-for="sample in samplePrompts"
                :key="sample.title"
                type="button"
                class="flex w-full items-center justify-between gap-3 rounded-xl border border-gray-200 bg-white p-3 text-left transition hover:border-violet-300 hover:bg-violet-50"
                @click="useSample(sample)"
              >
                <span class="min-w-0">
                  <span class="block text-sm font-medium text-gray-800">{{ sample.title }}</span>
                  <span class="mt-1 block truncate text-xs text-gray-500">{{ sample.prompt }}</span>
                </span>
                <span class="text-lg text-gray-300">›</span>
              </button>
            </div>
          </div>

          <PromptTags @select="appendPrompt" />

          <div v-if="captchaEnabled" ref="captchaEl" class="min-h-[65px]"></div>
          <p v-if="!userStore.user" class="text-sm text-gray-500">未登录用户可免费试用 1 次，生成质量会自动使用快速模式。</p>
          <p v-if="error" class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-600">{{ error }}</p>

          <div class="space-y-2">
            <button
              class="w-full rounded-xl bg-gradient-to-r from-violet-600 to-blue-600 px-4 py-4 font-medium text-white shadow-lg shadow-violet-200 transition hover:from-violet-700 hover:to-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="!canGenerate"
              @click="generate"
            >
              {{ loading ? '生成中...' : `生成图片 · ${qualityLabels[quality]} · ${cost} 积分` }}
            </button>
            <button v-if="canRetry" class="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm font-medium text-gray-700 transition hover:border-violet-300" type="button" @click="retry">
              重新生成上一次
            </button>
          </div>
        </div>
      </aside>

      <main class="min-h-[560px] flex-1 overflow-y-auto p-5 sm:p-8 lg:h-[calc(100vh-65px)]">
        <div class="mx-auto max-w-5xl">
          <div class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-xl font-semibold text-gray-900">生成结果</h2>
              <p class="mt-1 text-sm text-gray-500">结果会在任务完成后展示在这里。</p>
            </div>
            <a v-if="imageURL" class="inline-flex min-h-11 items-center justify-center rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:border-violet-300" :href="imageURL" download>
              下载全部
            </a>
          </div>

          <GenerationProgress
            v-if="generationId"
            :generation-id="generationId"
            @completed="completed"
            @failed="failed"
            @cancelled="cancelled"
            @cancel="cancelGeneration"
          />

          <div v-else-if="imageURL" class="grid gap-4 md:grid-cols-2">
            <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm">
              <img class="aspect-square w-full bg-gray-100 object-contain" :src="imageURL" alt="生成结果" />
              <div class="flex items-center justify-between gap-3 p-4">
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium text-gray-800">{{ prompt || 'AI 生成图片' }}</p>
                  <p class="mt-1 text-xs text-gray-500">{{ size }} · {{ qualityLabels[quality] }}</p>
                </div>
                <a class="rounded-lg bg-gray-900 px-3 py-2 text-sm text-white" :href="imageURL" download>下载</a>
              </div>
            </div>
          </div>

          <div v-else class="flex min-h-[520px] items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-white">
            <div class="px-6 text-center">
              <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 text-3xl text-gray-400">✦</div>
              <h3 class="mt-5 text-lg font-semibold text-gray-800">准备好了吗？</h3>
              <p class="mt-2 max-w-sm text-sm leading-6 text-gray-500">填写提示词并选择风格后，生成结果会显示在右侧预览区。</p>
            </div>
          </div>
        </div>
      </main>
    </div>
  </section>
</template>
