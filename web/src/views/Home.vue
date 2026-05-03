<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

import api from '@/api'
import GenerationProgress from '@/components/GenerationProgress.vue'
import { useUserStore } from '@/stores/user'
import { downloadImage } from '@/utils/download'

type Quality = 'low' | 'medium' | 'high'
type GenerationMode = 'generate' | 'edit'

interface GenerationPayload {
  prompt: string
  quality: Quality
  size: string
  mode: GenerationMode
  sourceImage?: File | null
}

interface SizeOption {
  value: string
  label: string
  ratio: string
}

interface StylePreset {
  id: string
  name: string
  prompt: string
}

interface SamplePrompt {
  id: string
  title: string
  prompt: string
}

interface PromptTemplate {
  id?: number
  category: string
  label: string
  prompt: string
}

const userStore = useUserStore()
const health = ref('检查中')
const prompt = ref('')
const generationMode = ref<GenerationMode>('generate')
const sourceImageFile = ref<File | null>(null)
const sourceImagePreview = ref('')
const selectedStyle = ref('')
const quality = ref<Quality>('medium')
const size = ref('1024x1024')
const sizeOptions = ref<SizeOption[]>([
  { value: '1280x720', label: '16:9', ratio: '16:9' },
  { value: '720x1280', label: '9:16', ratio: '9:16' },
  { value: '1024x1024', label: '1:1', ratio: '1:1' },
])
const imageCount = ref(4)
const creativity = ref(0.7)
const steps = ref(30)
const cfgScale = ref(7)
const isAdvancedExpanded = ref(false)
const isSamplesExpanded = ref(false)
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
const isPromptPanelCollapsed = ref(false)
const isFullscreenOpen = ref(false)
const fullscreenEl = ref<HTMLElement | null>(null)

const costs: Record<Quality, number> = { low: 0.2, medium: 1, high: 4 }
const qualityLabels: Record<Quality, string> = { low: '快速', medium: '标准', high: '高清' }
const defaultStylePresets: StylePreset[] = [
  { id: 'style-realistic', name: '写实', prompt: '写实摄影风格，细节丰富，自然光影，真实材质，高质量商业摄影' },
  { id: 'style-anime', name: '动漫', prompt: '动漫插画风格，清晰线稿，高饱和色彩，精致角色设计，干净背景' },
  { id: 'style-fantasy', name: '幻想', prompt: '幻想艺术风格，史诗氛围，电影级构图，丰富层次，强烈空间感' },
  { id: 'style-cyberpunk', name: '赛博朋克', prompt: '赛博朋克风格，霓虹灯光，未来城市质感，高对比光影，雨夜氛围' },
  { id: 'style-watercolor', name: '水彩', prompt: '水彩画风格，柔和笔触，温暖色调，纸张纹理，轻盈通透' },
  { id: 'style-abstract', name: '抽象', prompt: '抽象艺术风格，流动光影，紫蓝渐变，几何节奏，现代视觉表达' },
]
const defaultSamplePrompts: SamplePrompt[] = [
  { id: 'sample-fantasy', title: '幻想风景', prompt: '沙漠中的神秘传送门，远处有漂浮的古代遗迹，超现实主义场景，金色夕阳，电影级构图，4K 高清细节' },
  { id: 'sample-cyberpunk', title: '赛博朋克城市', prompt: '未来城市夜景，湿润街道反射霓虹灯，密集高楼与飞行交通，赛博朋克风格，强烈蓝紫色光影' },
  { id: 'sample-watercolor', title: '水彩小屋', prompt: '森林中的小木屋，清晨薄雾，温暖阳光穿过树叶，柔和水彩画风格，安静治愈氛围' },
  { id: 'sample-abstract', title: '抽象艺术', prompt: '流动的光影和透明几何结构，紫蓝渐变，细腻颗粒质感，现代抽象艺术海报' },
]
const stylePresets = ref<StylePreset[]>([...defaultStylePresets])
const samplePrompts = ref<SamplePrompt[]>([...defaultSamplePrompts])

const selectedStylePrompt = computed(() => stylePresets.value.find((item) => item.id === selectedStyle.value)?.prompt || '')
const canRetry = computed(() => !!lastRequest.value && !loading.value)
const canGenerate = computed(() => prompt.value.trim().length > 0 && !loading.value && (generationMode.value === 'generate' || !!sourceImageFile.value))
const displayName = computed(() => userStore.user?.email.split('@')[0] || '访客')
const creditText = computed(() => (userStore.user ? `${userStore.user.credits} 积分` : '免费试用'))
const selectedSizeOption = computed(() => sizeOptions.value.find((item) => item.value === size.value))

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
  await Promise.all([loadPromptTemplates(), loadGenerationOptions(), loadCaptcha()])
})

onUnmounted(() => {
  clearSourceImage()
})

watch(
  () => userStore.user?.id,
  () => {
    loadGenerationOptions()
  },
)

async function loadPromptTemplates() {
  try {
    const response = await api.get('/prompt-templates')
    const items: PromptTemplate[] = Array.isArray(response.data.items) ? response.data.items : []
    const styles = items.filter((item) => item.category === 'style')
    const samples = items.filter((item) => item.category === 'sample')
    if (styles.length > 0) {
      stylePresets.value = styles.map((item) => ({
        id: `style-${item.id || item.label}`,
        name: item.label,
        prompt: item.prompt,
      }))
    }
    if (samples.length > 0) {
      samplePrompts.value = samples.map((item) => ({
        id: `sample-${item.id || item.label}`,
        title: item.label,
        prompt: item.prompt,
      }))
    }
  } catch {
    stylePresets.value = [...defaultStylePresets]
    samplePrompts.value = [...defaultSamplePrompts]
  }
}

async function loadGenerationOptions() {
  try {
    const response = await api.get('/generation/options')
    if (Array.isArray(response.data.size_options) && response.data.size_options.length > 0) {
      sizeOptions.value = response.data.size_options
      if (!sizeOptions.value.some((item) => item.value === size.value)) {
        size.value = sizeOptions.value[0].value
      }
    } else if (Array.isArray(response.data.sizes) && response.data.sizes.length > 0) {
      sizeOptions.value = response.data.sizes.map((item: string) => ({ value: item, label: sizeRatioLabel(item), ratio: sizeRatioLabel(item) }))
      if (!sizeOptions.value.some((item) => item.value === size.value)) {
        size.value = sizeOptions.value[0].value
      }
    }
  } catch {
    const fallback = userStore.user
      ? ['1280x720', '720x1280', '1024x1024', '1152x768', '768x1152', '1536x1024', '1024x1536', '2048x2048']
      : ['1280x720', '720x1280', '1024x1024', '1152x768', '768x1152']
    sizeOptions.value = fallback.map((item) => ({ value: item, label: sizeRatioLabel(item), ratio: sizeRatioLabel(item) }))
    if (!sizeOptions.value.some((item) => item.value === size.value)) {
      size.value = sizeOptions.value[0].value
    }
  }
}

function sizeRatioLabel(value: string) {
  const [width, height] = value.split('x').map((item) => Number(item))
  if (!width || !height) {
    return value.replace('x', ' x ')
  }
  const divisor = gcd(width, height)
  return `${width / divisor}:${height / divisor}`
}

function gcd(a: number, b: number): number {
  return b === 0 ? a : gcd(b, a % b)
}

function useSample(sample: SamplePrompt) {
  prompt.value = sample.prompt
}

function buildPrompt() {
  const parts = [prompt.value.trim()]
  if (selectedStylePrompt.value) {
    parts.push(selectedStylePrompt.value)
  }
  return parts.filter(Boolean).join('\n')
}

function rangeWidth(value: number, min: number, max: number) {
  return `${((value - min) / (max - min)) * 100}%`
}

async function generate() {
  await createGeneration({
    prompt: buildPrompt(),
    quality: quality.value,
    size: size.value,
    mode: generationMode.value,
    sourceImage: sourceImageFile.value,
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
  isFullscreenOpen.value = false
  generationId.value = null
  loading.value = true
  lastRequest.value = { ...payload }
  try {
    const response =
      payload.mode === 'edit'
        ? await createImageEditRequest(payload)
        : await api.post('/generations', { prompt: payload.prompt, quality: payload.quality, size: payload.size, captcha_token: captchaToken.value })
    generationId.value = response.data.id
  } catch (err: any) {
    error.value = err.response?.data?.error || '创建生成任务失败'
    loading.value = false
    resetCaptcha()
  }
}

function createImageEditRequest(payload: GenerationPayload) {
  if (!payload.sourceImage) {
    throw new Error('image file required')
  }
  const form = new FormData()
  form.append('prompt', payload.prompt)
  form.append('quality', payload.quality)
  form.append('size', payload.size)
  form.append('captcha_token', captchaToken.value)
  form.append('image', payload.sourceImage)
  return api.post('/generations/edit', form)
}

function handleSourceImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    clearSourceImage()
    return
  }
  if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
    error.value = '仅支持 PNG、JPG、WebP 图片'
    input.value = ''
    clearSourceImage()
    return
  }
  if (file.size > 50 * 1024 * 1024) {
    error.value = '图片不能超过 50MB'
    input.value = ''
    clearSourceImage()
    return
  }
  error.value = ''
  sourceImageFile.value = file
  if (sourceImagePreview.value) {
    URL.revokeObjectURL(sourceImagePreview.value)
  }
  sourceImagePreview.value = URL.createObjectURL(file)
}

function clearSourceImage() {
  sourceImageFile.value = null
  if (sourceImagePreview.value) {
    URL.revokeObjectURL(sourceImagePreview.value)
  }
  sourceImagePreview.value = ''
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
  isPromptPanelCollapsed.value = true
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
  resetCaptcha()
}

function downloadCurrentImage() {
  void downloadImage(imageURL.value, `image-show-${Date.now()}`)
}

async function openFullscreen() {
  if (imageURL.value) {
    isFullscreenOpen.value = true
    await nextTick()
    try {
      await fullscreenEl.value?.requestFullscreen?.()
    } catch {
      // Keep the in-app fullscreen preview visible if the browser rejects native fullscreen.
    }
  }
}

async function closeFullscreen() {
  isFullscreenOpen.value = false
  if (document.fullscreenElement) {
    try {
      await document.exitFullscreen()
    } catch {
      // The overlay is already closed; native fullscreen will be cleaned up by the browser.
    }
  }
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
  <section class="h-auto min-h-[calc(100vh-65px)] overflow-hidden bg-gray-50 text-gray-950 lg:h-[calc(100vh-65px)]">
    <div class="flex h-full min-h-[calc(100vh-65px)] flex-col lg:flex-row">
      <main class="min-h-[560px] flex-1 overflow-y-auto bg-gray-50 p-5 sm:p-8 lg:h-[calc(100vh-65px)]">
        <div v-if="generationId" class="mx-auto max-w-5xl">
          <GenerationProgress
            :generation-id="generationId"
            @completed="completed"
            @failed="failed"
            @cancelled="cancelled"
            @cancel="cancelGeneration"
          />
        </div>

        <div v-else-if="imageURL" class="-m-5 flex h-[calc(100vh-65px)] min-h-[560px] flex-col bg-slate-950 sm:-m-8">
          <div class="relative min-h-0 flex-1 overflow-hidden">
            <img class="absolute inset-0 size-full scale-110 object-cover opacity-35 blur-2xl" :src="imageURL" alt="" aria-hidden="true" />
            <div class="absolute inset-0 bg-slate-950/35"></div>
            <img class="relative z-10 size-full object-contain" :src="imageURL" alt="生成的图片" />
            <div class="absolute right-5 top-5 z-20 flex gap-2 sm:right-8 sm:top-8">
              <button
                class="inline-flex min-h-11 items-center justify-center rounded-full border border-white/20 bg-black/35 px-4 text-sm font-medium text-white backdrop-blur transition hover:bg-black/50"
                type="button"
                @click="openFullscreen"
              >
                全屏查看
              </button>
              <button
                class="inline-flex min-h-11 items-center justify-center rounded-full border border-white/20 bg-black/35 px-4 text-sm font-medium text-white backdrop-blur transition hover:bg-black/50 lg:hidden"
                type="button"
                @click="isPromptPanelCollapsed = !isPromptPanelCollapsed"
              >
                {{ isPromptPanelCollapsed ? '展开参数' : '收起参数' }}
              </button>
            </div>
            <div class="pointer-events-none absolute inset-x-0 bottom-0 z-20 bg-gradient-to-t from-black/75 via-black/30 to-transparent p-5 sm:p-8">
              <div class="pointer-events-auto flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div class="text-white">
                  <h2 class="text-lg font-medium">生成结果</h2>
                  <p class="mt-1 text-sm text-white/70">{{ selectedSizeOption?.label || size.replace('x', ' x ') }} · {{ qualityLabels[quality] }}</p>
                </div>
                <div class="flex gap-2">
                  <button
                    class="inline-flex min-h-11 items-center justify-center rounded-full border border-white/20 bg-white/15 px-4 text-sm font-medium text-white backdrop-blur transition hover:bg-white/25"
                    type="button"
                    @click="downloadCurrentImage"
                  >
                    下载
                  </button>
                  <button
                    class="inline-flex min-h-11 items-center justify-center rounded-full bg-white px-5 text-sm font-semibold text-slate-950 shadow-lg shadow-black/20 transition hover:bg-slate-100"
                    type="button"
                    @click="downloadCurrentImage"
                  >
                    下载全部
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="flex h-[calc(100vh-200px)] min-h-[460px] items-center justify-center">
          <div class="text-center">
            <div class="mx-auto mb-4 flex size-20 items-center justify-center rounded-full bg-gray-200">
              <svg class="size-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2 1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <p class="mb-2 text-xl text-gray-700">准备好了吗？</p>
            <p class="text-gray-500">在右侧输入提示词，开始创作你的 AI 图片</p>
          </div>
        </div>
      </main>

      <aside
        class="relative w-full shrink-0 border-t border-gray-200 bg-white transition-[width] duration-300 lg:h-[calc(100vh-65px)] lg:border-l lg:border-t-0"
        :class="isPromptPanelCollapsed ? 'lg:w-0' : 'lg:w-[420px]'"
      >
        <button
          class="absolute -left-4 top-1/2 z-30 hidden size-9 -translate-y-1/2 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-700 shadow-lg shadow-slate-900/15 transition hover:bg-violet-50 hover:text-violet-700 lg:flex"
          type="button"
          :aria-label="isPromptPanelCollapsed ? '展开参数面板' : '收起参数面板'"
          @click="isPromptPanelCollapsed = !isPromptPanelCollapsed"
        >
          <span class="text-lg leading-none">{{ isPromptPanelCollapsed ? '‹' : '›' }}</span>
        </button>

        <div v-show="!isPromptPanelCollapsed" class="h-full overflow-y-auto">
          <div class="space-y-6 p-6">
          <div class="flex items-center justify-between rounded-xl bg-violet-50 px-4 py-3">
            <div>
              <p class="text-sm font-medium text-gray-900">{{ displayName }}</p>
              <p class="text-xs text-gray-500">{{ health }}</p>
            </div>
            <div class="flex items-center gap-3">
              <p class="text-sm font-medium text-violet-900">{{ creditText }}</p>
              <button class="rounded-full border border-violet-200 bg-white px-3 py-1.5 text-xs font-medium text-violet-700 transition hover:bg-violet-100" type="button" @click="isPromptPanelCollapsed = true">
                收起
              </button>
            </div>
          </div>

          <div>
            <label class="mb-2 block text-gray-900">创作方式</label>
            <div class="grid grid-cols-2 gap-2 rounded-xl bg-gray-100 p-1">
              <button
                class="rounded-lg px-3 py-2 text-sm font-medium transition"
                :class="generationMode === 'generate' ? 'bg-white text-violet-700 shadow-sm' : 'text-gray-600 hover:text-gray-900'"
                type="button"
                @click="generationMode = 'generate'"
              >
                输入文本
              </button>
              <button
                class="rounded-lg px-3 py-2 text-sm font-medium transition"
                :class="generationMode === 'edit' ? 'bg-white text-violet-700 shadow-sm' : 'text-gray-600 hover:text-gray-900'"
                type="button"
                @click="generationMode = 'edit'"
              >
                上传图像
              </button>
            </div>
          </div>

          <div v-if="generationMode === 'edit'" class="rounded-xl border border-dashed border-gray-300 bg-gray-50 p-4">
            <label class="flex cursor-pointer flex-col items-center justify-center rounded-lg bg-white px-4 py-6 text-center transition hover:bg-violet-50">
              <input class="sr-only" type="file" accept="image/png,image/jpeg,image/webp" @change="handleSourceImageChange" />
              <template v-if="sourceImagePreview">
                <img class="mb-3 max-h-44 rounded-lg object-contain" :src="sourceImagePreview" alt="待编辑图片预览" />
                <span class="text-sm font-medium text-gray-900">{{ sourceImageFile?.name }}</span>
                <span class="mt-1 text-xs text-gray-500">点击可替换图片</span>
              </template>
              <template v-else>
                <span class="mb-2 flex size-12 items-center justify-center rounded-full bg-violet-100 text-xl text-violet-700">＋</span>
                <span class="text-sm font-medium text-gray-900">上传要编辑的图片</span>
                <span class="mt-1 text-xs text-gray-500">支持 PNG、JPG、WebP，最大 50MB</span>
              </template>
            </label>
            <button v-if="sourceImagePreview" class="mt-3 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-600 transition hover:border-red-200 hover:text-red-600" type="button" @click="clearSourceImage">
              移除图片
            </button>
          </div>

          <div>
            <label for="prompt" class="mb-2 block text-gray-900">{{ generationMode === 'edit' ? '编辑描述' : '提示词' }}</label>
            <textarea
              id="prompt"
              v-model="prompt"
              class="h-32 w-full resize-none rounded-xl border border-gray-300 px-4 py-3 outline-none transition focus:border-transparent focus:ring-2 focus:ring-violet-500"
              :placeholder="generationMode === 'edit' ? '描述你想要怎样修改这张图片，例如：把背景换成星空，保留人物姿势...' : '描述你想要生成的图片，例如：一只在星空下的猫咪，水彩画风格...'"
            />
          </div>

          <div>
            <label class="mb-2 block text-gray-900">风格预设</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                class="rounded-lg border px-4 py-2 transition"
                :class="selectedStyle === '' ? 'border-violet-600 bg-violet-600 text-white' : 'border-gray-300 bg-white text-gray-700 hover:border-violet-400'"
                @click="selectedStyle = ''"
              >
                无
              </button>
              <button
                v-for="style in stylePresets"
                :key="style.id"
                type="button"
                class="rounded-lg border px-4 py-2 transition"
                :class="selectedStyle === style.id ? 'border-violet-600 bg-violet-600 text-white' : 'border-gray-300 bg-white text-gray-700 hover:border-violet-400'"
                @click="selectedStyle = style.id"
              >
                {{ style.name }}
              </button>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
            <label class="text-sm font-medium text-gray-700">
              质量
              <select v-model="quality" class="mt-2 min-h-11 w-full rounded-xl border border-gray-300 bg-white px-3 py-2 outline-none focus:border-violet-400">
                <option value="low">快速 - 0.2 积分</option>
                <option value="medium">标准 - 1 积分</option>
                <option value="high">高清 - 4 积分</option>
              </select>
            </label>
            <div class="text-sm font-medium text-gray-700">
              <div class="mb-2 flex items-center justify-between">
                <span>尺寸比例</span>
                <span class="text-xs font-normal text-gray-500">{{ selectedSizeOption?.value.replace('x', ' x ') }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <button
                  v-for="item in sizeOptions"
                  :key="item.value"
                  class="min-h-11 rounded-xl border px-3 py-2 text-sm transition"
                  :class="size === item.value ? 'border-violet-600 bg-violet-600 text-white shadow-sm shadow-violet-500/20' : 'border-gray-300 bg-white text-gray-700 hover:border-violet-400'"
                  type="button"
                  @click="size = item.value"
                >
                  <span class="block font-medium">{{ item.label }}</span>
                  <span class="mt-0.5 block text-[11px] opacity-75">{{ item.value.replace('x', ' x ') }}</span>
                </button>
              </div>
            </div>
          </div>

          <section class="border-t border-gray-200 pt-4">
            <button class="mb-3 flex w-full items-center justify-between text-gray-900 transition hover:text-violet-700" type="button" @click="isAdvancedExpanded = !isAdvancedExpanded">
              <h3 class="text-lg font-medium">高级参数</h3>
              <span class="text-xl transition-transform" :class="isAdvancedExpanded ? 'rotate-180' : ''">⌄</span>
            </button>

            <div v-if="isAdvancedExpanded" class="space-y-4">
              <label class="block">
                <span class="mb-2 flex items-center justify-between text-gray-700"><span>生成数量</span><span class="text-violet-600">{{ imageCount }}</span></span>
                <span class="relative block">
                  <span class="block h-2 overflow-hidden rounded-full bg-gray-200">
                    <span class="block h-full bg-gradient-to-r from-violet-600 to-blue-600 transition-all" :style="{ width: rangeWidth(imageCount, 1, 4) }"></span>
                  </span>
                  <input v-model.number="imageCount" class="absolute inset-x-0 top-0 h-2 w-full cursor-pointer opacity-0" type="range" min="1" max="4" step="1" />
                </span>
                <span class="mt-1 flex justify-between text-xs text-gray-500"><span>1张</span><span>4张</span></span>
              </label>

              <label class="block">
                <span class="mb-2 flex items-center justify-between text-gray-700"><span>创意度</span><span class="text-violet-600">{{ creativity.toFixed(1) }}</span></span>
                <span class="relative block">
                  <span class="block h-2 overflow-hidden rounded-full bg-gray-200">
                    <span class="block h-full bg-gradient-to-r from-violet-600 to-blue-600 transition-all" :style="{ width: rangeWidth(creativity, 0, 1) }"></span>
                  </span>
                  <input v-model.number="creativity" class="absolute inset-x-0 top-0 h-2 w-full cursor-pointer opacity-0" type="range" min="0" max="1" step="0.1" />
                </span>
                <span class="mt-1 flex justify-between text-xs text-gray-500"><span>保守</span><span>创新</span></span>
              </label>

              <label class="block">
                <span class="mb-2 flex items-center justify-between text-gray-700"><span>生成步数</span><span class="text-violet-600">{{ steps }}</span></span>
                <span class="relative block">
                  <span class="block h-2 overflow-hidden rounded-full bg-gray-200">
                    <span class="block h-full bg-gradient-to-r from-violet-600 to-blue-600 transition-all" :style="{ width: rangeWidth(steps, 10, 50) }"></span>
                  </span>
                  <input v-model.number="steps" class="absolute inset-x-0 top-0 h-2 w-full cursor-pointer opacity-0" type="range" min="10" max="50" step="5" />
                </span>
                <span class="mt-1 flex justify-between text-xs text-gray-500"><span>快速</span><span>精细</span></span>
              </label>

              <label class="block">
                <span class="mb-2 flex items-center justify-between text-gray-700"><span>提示词相关度</span><span class="text-violet-600">{{ cfgScale }}</span></span>
                <span class="relative block">
                  <span class="block h-2 overflow-hidden rounded-full bg-gray-200">
                    <span class="block h-full bg-gradient-to-r from-violet-600 to-blue-600 transition-all" :style="{ width: rangeWidth(cfgScale, 1, 20) }"></span>
                  </span>
                  <input v-model.number="cfgScale" class="absolute inset-x-0 top-0 h-2 w-full cursor-pointer opacity-0" type="range" min="1" max="20" step="1" />
                </span>
                <span class="mt-1 flex justify-between text-xs text-gray-500"><span>宽松</span><span>严格</span></span>
              </label>
            </div>
          </section>

          <section class="border-t border-gray-200 pt-4">
            <button class="mb-3 flex w-full items-center justify-between text-gray-900 transition hover:text-violet-700" type="button" @click="isSamplesExpanded = !isSamplesExpanded">
              <h3 class="text-lg font-medium">推荐样例</h3>
              <span class="text-xl transition-transform" :class="isSamplesExpanded ? 'rotate-180' : ''">⌄</span>
            </button>
            <div v-if="isSamplesExpanded" class="space-y-2">
              <button
                v-for="sample in samplePrompts"
                :key="sample.id"
                type="button"
                class="group w-full rounded-lg border border-gray-200 px-4 py-3 text-left transition hover:border-violet-400 hover:bg-violet-50"
                @click="useSample(sample)"
              >
                <span class="flex items-center justify-between gap-3">
                  <span class="min-w-0">
                    <span class="block text-gray-900 group-hover:text-violet-700">{{ sample.title }}</span>
                    <span class="mt-1 block truncate text-sm text-gray-500">{{ sample.prompt }}</span>
                  </span>
                  <span class="text-gray-400 group-hover:text-violet-600">›</span>
                </span>
              </button>
            </div>
          </section>

          <div v-if="captchaEnabled" ref="captchaEl" class="min-h-[65px]"></div>
          <p v-if="error" class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-600">{{ error }}</p>

          <button
            class="w-full rounded-xl bg-gradient-to-r from-violet-600 to-blue-600 py-4 text-white shadow-lg shadow-violet-500/30 transition hover:from-violet-700 hover:to-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
            type="button"
            :disabled="!canGenerate"
            @click="generate"
          >
            {{ loading ? '处理中...' : generationMode === 'edit' ? '开始编辑' : '开始生成' }}
          </button>

          <button v-if="canRetry" class="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm font-medium text-gray-700 transition hover:border-violet-300" type="button" @click="retry">
            重新生成上一次
          </button>

          <p v-if="!userStore.user" class="text-center text-sm text-gray-500">未登录可免费试用 1 次；登录后可查看历史记录和积分。</p>
          <p class="text-center text-xs text-gray-400">{{ generationMode === 'edit' ? '图片编辑' : '文本生成' }} · {{ qualityLabels[quality] }}模式，本次预计消耗 {{ costs[quality] }} 积分。</p>
          </div>
        </div>
      </aside>
    </div>

    <div v-show="isFullscreenOpen && imageURL" ref="fullscreenEl" class="fixed inset-0 z-50 flex flex-col bg-black" @click.self="closeFullscreen">
      <div class="flex min-h-16 items-center justify-between gap-3 border-b border-white/10 px-4 py-3 text-white sm:px-6">
        <div class="min-w-0">
          <p class="truncate text-sm font-medium">生成结果</p>
          <p class="text-xs text-white/60">{{ selectedSizeOption?.label || size.replace('x', ' x ') }} · {{ qualityLabels[quality] }}</p>
        </div>
        <div class="flex shrink-0 gap-2">
          <button class="rounded-full border border-white/20 px-4 py-2 text-sm transition hover:bg-white/10" type="button" @click="downloadCurrentImage">
            下载
          </button>
          <button class="rounded-full bg-white px-4 py-2 text-sm font-medium text-slate-950 transition hover:bg-slate-100" type="button" @click="closeFullscreen">
            关闭
          </button>
        </div>
      </div>
      <div class="min-h-0 flex-1 p-3 sm:p-6">
        <img class="size-full object-contain" :src="imageURL" alt="生成的图片全屏预览" />
      </div>
    </div>
  </section>
</template>
