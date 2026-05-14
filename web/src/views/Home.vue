<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import api from '@/api'
import { fetchSiteConfig, type CreditCosts } from '@/api/site'
import CreditExhaustedGuide from '@/components/CreditExhaustedGuide.vue'
import GenerationProgress from '@/components/GenerationProgress.vue'
import SceneCard from '@/components/SceneCard.vue'
import { useToast } from '@/composables/useToast'
import { useUserStore } from '@/stores/user'
import { downloadImage } from '@/utils/download'

type Quality = 'low' | 'medium' | 'high'
type GenerationMode = 'generate' | 'edit'
type CreditErrorType = 'free_trial_exhausted' | 'insufficient_credits' | 'credits_expired'

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
  credit_cost?: number
}

const aspectRatioFallbacks: Record<string, SizeOption> = {
  square: { value: 'square', label: '方形', ratio: '1:1', credit_cost: 1 },
  portrait_3_4: { value: 'portrait_3_4', label: '竖版', ratio: '3:4', credit_cost: 2 },
  story: { value: 'story', label: '故事版', ratio: '9:16', credit_cost: 2 },
  landscape_4_3: { value: 'landscape_4_3', label: '横版', ratio: '4:3', credit_cost: 2 },
  widescreen: { value: 'widescreen', label: '宽屏', ratio: '16:9', credit_cost: 2 },
}

const aspectRatioRealSizes: Record<string, string> = {
  square: '1024x1024',
  portrait_3_4: '1152x1536',
  story: '1008x1792',
  landscape_4_3: '1536x1152',
  widescreen: '1792x1008',
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

interface ScenarioPrompt {
  id: string | number
  name: string
  icon: string
  description: string
  prompt_template: string
  recommended_ratio: string
  credit_cost: number
}

interface PromptTemplate {
  id?: number
  category: string
  label: string
  prompt: string
}

interface GenerationDraft {
  prompt: string
  selectedStyle: string
  size: string
  generationMode: GenerationMode
}

const generationDraftKey = 'image_show_generation_draft'
const userStore = useUserStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()
const prompt = ref('')
const promptInput = ref<HTMLTextAreaElement | null>(null)
const generationMode = ref<GenerationMode>('generate')
const isImageEditEnabled = false
const sourceImageFile = ref<File | null>(null)
const sourceImagePreview = ref('')
const selectedStyle = ref('')
const quality: Quality = 'medium'
const size = ref('square')
const defaultSizeValues = ['square', 'portrait_3_4', 'story', 'landscape_4_3', 'widescreen']
const creditCosts = ref<CreditCosts>({
  square: 1,
  portrait: 2,
  story: 2,
  landscape: 2,
  widescreen: 2,
})
const sizeOptions = ref<SizeOption[]>([
  ...defaultSizeValues.map(makeSizeOption),
])
const isSamplesExpanded = ref(false)
const generationId = ref<number | null>(null)
const imageURL = ref('')
const error = ref('')
const creditError = ref<CreditErrorType | null>(null)
const loading = ref(false)
const lastRequest = ref<GenerationPayload | null>(null)
const captchaEnabled = ref(false)
const captchaSiteKey = ref('')
const captchaToken = ref('')
const captchaEl = ref<HTMLElement | null>(null)
const captchaWidgetId = ref<string | null>(null)
const isPromptPanelCollapsed = ref(false)
const toggleNudge = ref(false)
const isFullscreenOpen = ref(false)
const fullscreenEl = ref<HTMLElement | null>(null)
const isBillingGuideOpen = ref(false)
const resultNotice = ref('')
const selectedSceneId = ref<string | number | null>(null)
const typewriterTimer = ref<number | null>(null)

const defaultStylePresets: StylePreset[] = [
  { id: 'style-realistic', name: '写实', prompt: '写实摄影风格，细节丰富，自然光影，真实材质，高质量商业摄影' },
  { id: 'style-anime', name: '动漫', prompt: '动漫插画风格，清晰线稿，高饱和色彩，精致角色设计，干净背景' },
  { id: 'style-fantasy', name: '幻想', prompt: '幻想艺术风格，史诗氛围，电影级构图，丰富层次，强烈空间感' },
  { id: 'style-cyberpunk', name: '赛博朋克', prompt: '赛博朋克风格，霓虹灯光，未来城市质感，高对比光影，雨夜氛围' },
  { id: 'style-watercolor', name: '水彩', prompt: '水彩画风格，柔和笔触，温暖色调，纸张纹理，轻盈通透' },
  { id: 'style-abstract', name: '抽象', prompt: '抽象艺术风格，流动光影，紫蓝渐变，几何节奏，现代视觉表达' },
  { id: 'style-illustration', name: '插画', prompt: '现代商业插画风格，清晰轮廓，柔和配色，细腻纹理，画面干净有层次，适合封面、海报和内容配图' },
]
const defaultSamplePrompts: SamplePrompt[] = [
  { id: 'sample-fantasy', title: '幻想风景', prompt: '沙漠中的神秘传送门，远处有漂浮的古代遗迹，超现实主义场景，金色夕阳，电影级构图，4K 高清细节' },
  { id: 'sample-cyberpunk', title: '赛博朋克城市', prompt: '未来城市夜景，湿润街道反射霓虹灯，密集高楼与飞行交通，赛博朋克风格，强烈蓝紫色光影' },
  { id: 'sample-watercolor', title: '水彩小屋', prompt: '森林中的小木屋，清晨薄雾，温暖阳光穿过树叶，柔和水彩画风格，安静治愈氛围' },
  { id: 'sample-abstract', title: '抽象艺术', prompt: '流动的光影和透明几何结构，紫蓝渐变，细腻颗粒质感，现代抽象艺术海报' },
]
const defaultScenarioPrompts: ScenarioPrompt[] = [
  { id: 'scene-redbook-cover', name: '小红书封面', icon: '📸', description: '精致生活、穿搭、美食风格封面', prompt_template: '小红书封面图，精致生活方式视觉，一眼能看懂主题，清晰大标题留白，明亮干净的构图，适合手机竖屏浏览', recommended_ratio: 'portrait_3_4', credit_cost: 2 },
  { id: 'scene-product', name: '商品展示图', icon: '🛒', description: '白底或场景化商品展示', prompt_template: '电商商品展示图，主体突出，干净背景，真实材质，高级商业摄影光影，适合商品主图', recommended_ratio: 'square', credit_cost: 1 },
  { id: 'scene-avatar', name: '社交头像', icon: '👤', description: '精致人物或动漫风格头像', prompt_template: '精致社交头像，主体居中，五官清晰，背景简洁，有辨识度，适合作为社交平台头像', recommended_ratio: 'square', credit_cost: 1 },
  { id: 'scene-poster', name: '海报设计', icon: '🎨', description: '活动、促销、艺术创意海报', prompt_template: '活动宣传海报视觉，主题突出，层次清晰，保留文字排版空间，适合促销活动和创意传播', recommended_ratio: 'portrait_3_4', credit_cost: 2 },
  { id: 'scene-wallpaper', name: '手机壁纸', icon: '📷', description: '风景、抽象、治愈系壁纸', prompt_template: '高清手机壁纸画面，风景治愈氛围，视觉舒适，构图开阔，细节丰富，适合手机屏幕背景', recommended_ratio: 'story', credit_cost: 2 },
  { id: 'scene-free', name: '自由创作', icon: '✨', description: '不填充提示词，自由输入', prompt_template: '', recommended_ratio: 'square', credit_cost: 1 },
]
const stylePresets = ref<StylePreset[]>([...defaultStylePresets])
const samplePrompts = ref<SamplePrompt[]>([...defaultSamplePrompts])
const scenarioPrompts = ref<ScenarioPrompt[]>([...defaultScenarioPrompts])

const selectedStylePrompt = computed(() => stylePresets.value.find((item) => item.id === selectedStyle.value)?.prompt || '')
const canRetry = computed(() => !!lastRequest.value && !loading.value)
const canGenerate = computed(() => prompt.value.trim().length > 0 && !loading.value && (generationMode.value === 'generate' || (!!userStore.user && !!sourceImageFile.value)))
const selectedSizeOption = computed(() => sizeOptions.value.find((item) => item.value === size.value))
const estimatedCreditCost = computed(() => selectedSizeOption.value?.credit_cost ?? creditCostForSize(size.value))
const selectedSizeDisplay = computed(() => formatSizeOption(selectedSizeOption.value, size.value))
const generationModeText = computed(() => (generationMode.value === 'edit' ? '图片编辑' : '文本生成'))
const generateHint = computed(() => {
  if (loading.value) {
    return '正在处理当前任务'
  }
  if (!prompt.value.trim()) {
    return generationMode.value === 'edit' ? '填写编辑描述后即可开始' : '输入提示词后即可开始'
  }
  if (generationMode.value === 'edit' && !userStore.user) {
    return ''
  }
  if (generationMode.value === 'edit' && !sourceImageFile.value) {
    return '上传要编辑的图片后即可开始'
  }
  return `${generationModeText.value} · ${selectedSizeDisplay.value} · 预计消耗 ${estimatedCreditCost.value} 积分`
})
const displayName = computed(() => userStore.user?.email.split('@')[0] || '访客')
const creditBalanceText = computed(() => (userStore.user ? `${userStore.user.credits} 积分` : '1 次免费试用'))
const creditStatusText = computed(() => (userStore.user ? '当前可用额度' : '登录后可获得更多积分'))
const heroTitle = computed(() => (userStore.user ? '继续创作你的下一张图' : '输入一句话，生成封面、商品图和头像'))
const heroSubtitle = computed(() =>
  userStore.user
    ? '复用你的想法，选择合适比例，生成完成后可以下载并在历史记录中找回。'
    : '游客可免费生成 1 次；注册后获得积分，作品会保存在历史记录里，方便继续修改和下载。',
)
const creditBenefitText = computed(() =>
  userStore.user ? '生成会按尺寸消耗积分，失败任务不会让你白白扣费。' : '游客免费试用 1 次，注册后获得积分并保存历史作品。',
)
const canAffordCurrentGeneration = computed(() => !userStore.user || userStore.user.credits >= estimatedCreditCost.value)
const creditCostTone = computed(() => (canAffordCurrentGeneration.value ? 'text-emerald-700 bg-emerald-50 border-emerald-100' : 'text-amber-700 bg-amber-50 border-amber-100'))

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string
      reset: (widgetId?: string | null) => void
    }
  }
}

onMounted(async () => {
  restoreGenerationDraft()
  applyRoutePrefill()
  await Promise.all([loadSiteCreditCosts(), loadPromptTemplates(), loadScenes(), loadGenerationOptions(), loadCaptcha()])
})

onUnmounted(() => {
  clearSourceImage()
  clearTypewriter()
})

watch(
  () => userStore.user?.id,
  () => {
    loadGenerationOptions()
  },
)

watch([prompt, selectedStyle, size, generationMode], persistGenerationDraft)

watch(isPromptPanelCollapsed, (collapsed) => {
  if (!collapsed) {
    toggleNudge.value = false
    return
  }
  toggleNudge.value = true
  window.setTimeout(() => {
    toggleNudge.value = false
  }, 3000)
})

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

async function loadScenes() {
  try {
    const response = await api.get('/generation/scenes')
    const items = Array.isArray(response.data.items) ? response.data.items : []
    if (items.length > 0) {
      scenarioPrompts.value = items.map((item: any) => ({
        id: item.id || item.name,
        name: item.name,
        icon: item.icon || '✨',
        description: item.description || '',
        prompt_template: item.prompt_template || '',
        recommended_ratio: item.recommended_ratio || 'square',
        credit_cost: Number(item.credit_cost) || creditCostForSize(item.recommended_ratio || 'square'),
      }))
    }
  } catch {
    scenarioPrompts.value = defaultScenarioPrompts.map((item) => ({ ...item, credit_cost: creditCostForSize(item.recommended_ratio) }))
  }
}

async function loadGenerationOptions() {
  try {
    const response = await api.get('/generation/options')
    if (Array.isArray(response.data.size_options) && response.data.size_options.length > 0) {
      sizeOptions.value = response.data.size_options.map((item: SizeOption) => ({
        ...item,
        credit_cost: item.credit_cost ?? creditCostForSize(item.value),
      }))
      if (!sizeOptions.value.some((item) => item.value === size.value)) {
        size.value = sizeOptions.value[0].value
      }
    } else if (Array.isArray(response.data.sizes) && response.data.sizes.length > 0) {
      sizeOptions.value = response.data.sizes.map(makeSizeOption)
      if (!sizeOptions.value.some((item) => item.value === size.value)) {
        size.value = sizeOptions.value[0].value
      }
    }
  } catch {
    sizeOptions.value = defaultSizeValues.map(makeSizeOption)
    if (!sizeOptions.value.some((item) => item.value === size.value)) {
      size.value = sizeOptions.value[0].value
    }
  }
}

async function loadSiteCreditCosts() {
  try {
    const response = await fetchSiteConfig()
    const costs = response.data.credit_costs
    if (costs) {
      creditCosts.value = {
        square: normalizeCreditCost(costs.square, 1),
        portrait: normalizeCreditCost(costs.portrait, 2),
        story: normalizeCreditCost(costs.story, 2),
        landscape: normalizeCreditCost(costs.landscape, 2),
        widescreen: normalizeCreditCost(costs.widescreen, 2),
      }
      sizeOptions.value = sizeOptions.value.map((item) => ({
        ...item,
        credit_cost: creditCostForSize(item.value),
      }))
    }
  } catch {
    // Generation options still carries credit costs; keep local defaults if site config is unavailable.
  }
}

function normalizeCreditCost(value: unknown, fallback: number) {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue) || numberValue < 1) {
    return fallback
  }
  return Math.ceil(numberValue)
}

function restoreGenerationDraft() {
  const raw = localStorage.getItem(generationDraftKey)
  if (!raw) {
    return
  }
  try {
    const draft = JSON.parse(raw) as Partial<GenerationDraft>
    prompt.value = typeof draft.prompt === 'string' ? draft.prompt : ''
    selectedStyle.value = typeof draft.selectedStyle === 'string' ? draft.selectedStyle : ''
    if (typeof draft.size === 'string' && draft.size) {
      size.value = draft.size
    }
    if (draft.generationMode === 'generate' || (isImageEditEnabled && draft.generationMode === 'edit')) {
      generationMode.value = draft.generationMode
    } else {
      generationMode.value = 'generate'
    }
  } catch {
    localStorage.removeItem(generationDraftKey)
  }
}

function persistGenerationDraft() {
  const draft: GenerationDraft = {
    prompt: prompt.value,
    selectedStyle: selectedStyle.value,
    size: size.value,
    generationMode: generationMode.value,
  }
  localStorage.setItem(generationDraftKey, JSON.stringify(draft))
}

function makeSizeOption(value: string): SizeOption {
  const fallback = aspectRatioFallbacks[value]
  if (fallback) {
    return { ...fallback }
  }
  const ratio = sizeRatioLabel(value)
  return {
    value,
    label: ratio,
    ratio,
    credit_cost: creditCostForSize(value),
  }
}

function creditCostForSize(value: string) {
  const ratioCost = creditCostForRatioValue(value)
  if (ratioCost !== null) {
    return ratioCost
  }
  const realSize = aspectRatioRealSizes[value] || value
  const [width, height] = realSize.toLowerCase().split('x').map((part) => Number.parseInt(part.trim(), 10))
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) {
    return 1
  }
  return Math.max(1, Math.ceil((width * height) / (1024 * 1024)))
}

function creditCostForRatioValue(value: string) {
  switch (value) {
    case 'square':
    case '1024x1024':
      return creditCosts.value.square
    case 'portrait':
    case 'portrait_3_4':
    case '1152x1536':
      return creditCosts.value.portrait
    case 'story':
    case '1008x1792':
      return creditCosts.value.story
    case 'landscape':
    case 'landscape_4_3':
    case '1536x1152':
      return creditCosts.value.landscape
    case 'widescreen':
    case '1792x1008':
      return creditCosts.value.widescreen
    default:
      return null
  }
}

function sizeRatioLabel(value: string) {
  const fallback = aspectRatioFallbacks[value]
  if (fallback) {
    return fallback.ratio
  }
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

function formatSizeOption(option: SizeOption | undefined, fallbackValue: string) {
  const item = option || makeSizeOption(fallbackValue)
  if (item.label && item.ratio && item.label !== item.ratio) {
    return `${item.label} ${item.ratio}`
  }
  return item.label || item.ratio || fallbackValue.replace('x', ' x ')
}

function useSample(sample: SamplePrompt) {
  prompt.value = sample.prompt
}

async function useScenario(scenario: ScenarioPrompt) {
  clearTypewriter()
  if (selectedSceneId.value === scenario.id) {
    selectedSceneId.value = null
    prompt.value = ''
    selectedStyle.value = ''
    size.value = 'square'
    await focusPrompt()
    return
  }
  selectedSceneId.value = scenario.id
  selectedStyle.value = ''
  if (sizeOptions.value.some((item) => item.value === scenario.recommended_ratio)) {
    size.value = scenario.recommended_ratio
  }
  await focusPrompt()
  if (!scenario.prompt_template) {
    prompt.value = ''
    return
  }
  prompt.value = ''
  await new Promise((resolve) => window.setTimeout(resolve, 150))
  typePrompt(scenario.prompt_template)
}

function typePrompt(value: string) {
  let index = 0
  clearTypewriter()
  typewriterTimer.value = window.setInterval(() => {
    prompt.value = value.slice(0, index + 1)
    index += 1
    if (index >= value.length) {
      clearTypewriter()
    }
  }, 40)
}

function clearTypewriter() {
  if (typewriterTimer.value !== null) {
    window.clearInterval(typewriterTimer.value)
    typewriterTimer.value = null
  }
}

async function focusPrompt() {
  await nextTick()
  promptInput.value?.focus()
}

function applyRoutePrefill() {
  const queryPrompt = typeof route.query.prompt === 'string' ? route.query.prompt : ''
  const queryRatio = typeof route.query.ratio === 'string' ? route.query.ratio : ''
  if (!queryPrompt && !queryRatio) {
    return
  }
  if (queryPrompt) {
    prompt.value = queryPrompt
  }
  if (queryRatio) {
    size.value = queryRatio
  }
  selectedSceneId.value = null
  router.replace({ name: 'home' })
  toast.info('已回填历史参数，确认后点击生成')
}

function buildPrompt() {
  const parts = [prompt.value.trim()]
  if (selectedStylePrompt.value) {
    parts.push(selectedStylePrompt.value)
  }
  return parts.filter(Boolean).join('\n')
}

async function generate() {
  await createGeneration({
    prompt: buildPrompt(),
    quality,
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
  creditError.value = null
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
        : await api.post('/generations', {
            prompt: payload.prompt,
            quality: payload.quality,
            size: payload.size,
            captcha_token: captchaToken.value,
          })
    generationId.value = response.data.id
  } catch (err: any) {
    const mappedCreditError = resolveCreditError(err)
    if (mappedCreditError) {
      creditError.value = mappedCreditError
      error.value = ''
    } else {
      error.value = err.response?.data?.message || err.response?.data?.error || err.message || '创建生成任务失败'
    }
    loading.value = false
    resetCaptcha()
  }
}

function isCreditError(value: unknown): value is CreditErrorType {
  return value === 'free_trial_exhausted' || value === 'insufficient_credits' || value === 'credits_expired'
}

function resolveCreditError(err: any): CreditErrorType | null {
  const status = err.response?.status
  const errCode = err.response?.data?.error
  if (status === 402 && isCreditError(errCode)) {
    return errCode
  }
  if (typeof errCode === 'string') {
    const normalized = errCode.toLowerCase()
    if (normalized.includes('free trial used') || normalized.includes('please register')) {
      return 'free_trial_exhausted'
    }
    if (status === 402 && normalized.includes('insufficient credits')) {
      return 'insufficient_credits'
    }
  }
  return null
}

function createImageEditRequest(payload: GenerationPayload) {
  if (!userStore.user) {
    throw new Error('请先登录后再使用上传图像编辑')
  }
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
  if (!userStore.user) {
    error.value = ''
    input.value = ''
    clearSourceImage()
    return
  }
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
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
  resetCaptcha()
}

function downloadCurrentImage() {
  void downloadImage(imageURL.value, `image-show-${Date.now()}`)
}

async function copyCurrentPrompt() {
  try {
    await navigator.clipboard.writeText(lastRequest.value?.prompt || buildPrompt())
    resultNotice.value = '提示词已复制'
  } catch {
    error.value = '复制失败，请手动选择提示词复制'
  }
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
  <section class="home-shell h-auto min-h-[calc(100vh-65px)] overflow-hidden bg-gray-50 text-gray-950 transition-colors dark:bg-slate-950 dark:text-slate-100 lg:h-[calc(100vh-65px)]">
    <div class="flex h-full min-h-[calc(100vh-65px)] flex-col lg:flex-row">
      <main class="home-main min-h-[560px] flex-1 overflow-y-auto bg-gray-50 p-5 transition-colors dark:bg-slate-950 sm:p-8 lg:h-[calc(100vh-65px)]">
        <div v-if="generationId" class="-m-5 flex h-[calc(100vh-65px)] min-h-[560px] flex-col sm:-m-8">
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
            <div class="pointer-events-none absolute bottom-5 left-5 z-20 max-w-[min(520px,calc(100%-120px))] sm:bottom-8 sm:left-8">
              <div class="pointer-events-auto rounded-2xl border border-white/10 bg-black/28 px-4 py-3 text-white shadow-2xl shadow-black/20 backdrop-blur-md">
                <h2 class="text-base font-medium">生成结果</h2>
                <p class="mt-1 text-sm text-white/70">{{ selectedSizeDisplay }} · {{ estimatedCreditCost }} 积分</p>
                <p class="mt-1 truncate text-xs text-white/55">{{ lastRequest?.prompt || buildPrompt() }}</p>
                <p v-if="resultNotice" class="mt-1 text-xs text-emerald-200">{{ resultNotice }}</p>
              </div>
            </div>
            <div class="absolute bottom-5 right-5 z-30 flex items-center gap-1.5 rounded-full border border-white/12 bg-black/25 p-1.5 text-white opacity-80 shadow-2xl shadow-black/20 backdrop-blur-md transition hover:bg-black/45 hover:opacity-100 focus-within:bg-black/45 focus-within:opacity-100 sm:bottom-8 sm:right-8">
              <button
                class="inline-flex size-10 items-center justify-center rounded-full text-white/85 transition hover:bg-white/15 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/40"
                type="button"
                title="复制提示词"
                aria-label="复制提示词"
                @click="copyCurrentPrompt"
              >
                <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h10a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2Z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7V5a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2" />
                </svg>
              </button>
              <button
                v-if="canRetry"
                class="inline-flex size-10 items-center justify-center rounded-full bg-white/12 text-white/90 transition hover:bg-violet-500 hover:text-white focus:outline-none focus:ring-2 focus:ring-violet-200/60"
                type="button"
                title="再生成一次"
                aria-label="再生成一次"
                @click="retry"
              >
                <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 11a8.1 8.1 0 0 0-15.5-2M4 5v4h4" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 13a8.1 8.1 0 0 0 15.5 2M20 19v-4h-4" />
                </svg>
              </button>
              <button
                class="inline-flex size-10 items-center justify-center rounded-full text-white/90 transition hover:bg-white/15 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/40"
                type="button"
                title="下载"
                aria-label="下载"
                @click="downloadCurrentImage"
              >
                <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v12m0 0 4-4m-4 4-4-4" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 21h14" />
                </svg>
              </button>
              <button
                class="inline-flex h-10 items-center gap-2 rounded-full bg-white px-3.5 text-sm font-semibold text-slate-950 shadow-lg shadow-black/15 transition hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-white/60"
                type="button"
                title="下载全部"
                aria-label="下载全部"
                @click="downloadCurrentImage"
              >
                <svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 3h8m-4 0v10m0 0 3.5-3.5M12 13 8.5 9.5" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 17v2a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-2" />
                </svg>
                <span class="hidden sm:inline">全部</span>
              </button>
            </div>
          </div>
        </div>

        <div v-else-if="creditError" class="flex h-[calc(100vh-200px)] min-h-[460px] items-center justify-center p-6">
          <CreditExhaustedGuide :type="creditError" @dismiss="creditError = null" />
        </div>

        <div v-else class="flex h-[calc(100vh-200px)] min-h-[460px] items-center justify-center">
          <div class="text-center">
            <div class="mx-auto mb-4 flex size-20 items-center justify-center rounded-full bg-gray-200 transition-colors dark:bg-slate-800">
              <svg class="size-10 text-gray-400 dark:text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2 1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <p class="mb-2 text-xl text-gray-700 dark:text-slate-200">{{ heroTitle }}</p>
            <p class="mx-auto max-w-md text-gray-500 dark:text-slate-400">{{ heroSubtitle }}</p>
          </div>
        </div>
      </main>

      <aside
        class="home-panel relative w-full shrink-0 border-t border-gray-200 bg-white transition-[width] duration-300 dark:border-slate-800 dark:bg-slate-950 lg:h-[calc(100vh-65px)] lg:border-l lg:border-t-0"
        :class="isPromptPanelCollapsed ? 'lg:w-0 lg:border-l-0' : 'lg:w-[420px]'"
      >
        <button
          class="absolute top-1/2 z-30 hidden min-h-14 -translate-y-1/2 flex-col items-center justify-center gap-2 border border-slate-200 bg-white/90 px-2 text-slate-500 shadow-xl shadow-slate-900/15 backdrop-blur transition-all duration-200 hover:bg-white hover:text-violet-700 hover:shadow-violet-900/20 dark:border-slate-700 dark:bg-slate-900/82 dark:text-slate-300 dark:hover:bg-slate-900 dark:hover:text-violet-200 lg:flex"
          :class="[
            isPromptPanelCollapsed ? '-left-10 w-10 rounded-l-xl rounded-r-none opacity-70 hover:-left-12 hover:w-12 hover:opacity-100 hover:shadow-2xl' : '-left-5 w-10 rounded-l-xl rounded-r-none opacity-80 hover:opacity-100',
            toggleNudge ? 'panel-toggle-nudge' : '',
          ]"
          type="button"
          :aria-label="isPromptPanelCollapsed ? '展开参数面板' : '收起参数面板'"
          :title="isPromptPanelCollapsed ? '展开参数面板' : '收起参数面板'"
          @click="isPromptPanelCollapsed = !isPromptPanelCollapsed"
        >
          <svg
            class="size-5 transition-transform duration-200"
            :class="isPromptPanelCollapsed ? 'rotate-180' : ''"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            aria-hidden="true"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </button>

        <div v-show="!isPromptPanelCollapsed" class="h-full overflow-y-auto">
          <div class="space-y-6 p-6 pb-36">
          <div class="home-card rounded-2xl border border-violet-100 bg-violet-50 px-4 py-4 shadow-sm shadow-violet-900/5">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <p class="text-sm font-semibold text-gray-900">{{ displayName }}</p>
                <p class="mt-1 text-xs text-gray-500">{{ creditStatusText }}</p>
              </div>
              <div class="shrink-0 rounded-full border border-white/70 bg-white px-3 py-1 text-sm font-semibold text-violet-700 shadow-sm">
                {{ creditBalanceText }}
              </div>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3">
              <p class="text-xs text-violet-700">{{ creditBenefitText }}</p>
              <button class="rounded-full border border-violet-200 bg-white px-3 py-1.5 text-xs font-medium text-violet-700 transition hover:bg-violet-100" type="button" @click="isPromptPanelCollapsed = true">收起</button>
            </div>
          </div>

          <div v-if="isImageEditEnabled">
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
                @click="generationMode = 'edit'; error = ''"
              >
                上传图像
              </button>
            </div>
          </div>

          <div v-if="isImageEditEnabled && generationMode === 'edit'" class="rounded-xl border border-dashed border-gray-300 bg-gray-50 p-4">
            <p v-if="!userStore.user" class="mb-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
              上传图像编辑仅对登录用户开放，请先登录或注册后使用。
            </p>
            <label
              class="flex flex-col items-center justify-center rounded-lg bg-white px-4 py-6 text-center transition"
              :class="userStore.user ? 'cursor-pointer hover:bg-violet-50' : 'cursor-not-allowed opacity-70'"
            >
              <input class="sr-only" type="file" accept="image/png,image/jpeg,image/webp" :disabled="!userStore.user" @change="handleSourceImageChange" />
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
              ref="promptInput"
              v-model="prompt"
              class="home-input h-32 w-full resize-none rounded-xl border border-gray-300 px-4 py-3 outline-none transition focus:border-transparent focus:ring-2 focus:ring-violet-500"
              :placeholder="generationMode === 'edit' ? '描述你想要怎样修改这张图片，例如：把背景换成星空，保留人物姿势...' : '描述你想要生成的图片，例如：一只在星空下的猫咪，水彩画风格...'"
            />
          </div>

          <section v-if="generationMode === 'generate'" class="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <div class="mb-3 flex items-center justify-between gap-3">
              <div>
                <h3 class="text-sm font-semibold text-gray-900">场景入口</h3>
                <p class="mt-1 text-xs text-gray-500">点击后自动填充提示词、切换推荐比例并同步积分预估。</p>
              </div>
            </div>
            <div class="scene-scroll grid gap-3 md:grid-cols-2 lg:grid-cols-3">
              <SceneCard
                v-for="scenario in scenarioPrompts"
                :key="scenario.id"
                :scene="scenario"
                :selected="selectedSceneId === scenario.id"
                :ratio-label="formatSizeOption(sizeOptions.find((item) => item.value === scenario.recommended_ratio), scenario.recommended_ratio)"
                @click="useScenario(scenario)"
              />
            </div>
          </section>

          <div>
            <label class="mb-2 block text-gray-900">风格预设</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                class="home-button rounded-lg border px-4 py-2 transition"
                :class="selectedStyle === '' ? 'border-violet-600 bg-violet-600 text-white' : 'border-gray-300 bg-white text-gray-700 hover:border-violet-400'"
                @click="selectedStyle = ''"
              >
                无
              </button>
              <button
                v-for="style in stylePresets"
                :key="style.id"
                type="button"
                class="home-button rounded-lg border px-4 py-2 transition"
                :class="selectedStyle === style.id ? 'border-violet-600 bg-violet-600 text-white' : 'border-gray-300 bg-white text-gray-700 hover:border-violet-400'"
                @click="selectedStyle = style.id"
              >
                {{ style.name }}
              </button>
            </div>
          </div>

          <div class="grid gap-3">
            <div class="text-sm font-medium text-gray-700">
              <div class="mb-2 flex items-center justify-between">
                <span>尺寸比例</span>
                <span class="text-xs font-normal text-gray-500">{{ selectedSizeDisplay }} · {{ estimatedCreditCost }} 积分</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <button
                  v-for="item in sizeOptions"
                  :key="item.value"
                  class="home-button min-h-11 rounded-xl border px-3 py-2 text-sm transition"
                  :class="size === item.value ? 'border-violet-600 bg-violet-600 text-white shadow-sm shadow-violet-500/20' : 'border-gray-300 bg-white text-gray-700 hover:border-violet-400'"
                  type="button"
                  @click="size = item.value"
                >
                  <span class="block font-medium">{{ item.label }}</span>
                  <span class="mt-0.5 block text-[11px] opacity-75">{{ item.ratio }}</span>
                  <span class="mt-1 block text-[11px] font-semibold opacity-90">{{ item.credit_cost ?? creditCostForSize(item.value) }} 积分</span>
                </button>
              </div>
            </div>
          </div>

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

          <button v-if="canRetry && !imageURL" class="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm font-medium text-gray-700 transition hover:border-violet-300" type="button" @click="retry">
            重新生成上一次
          </button>

          <p v-if="!userStore.user" class="rounded-xl border border-violet-100 bg-violet-50 px-3 py-2 text-center text-sm text-violet-700">
            游客可免费生成 1 次，注册后获得积分并保存历史记录。
          </p>
          </div>
          <div class="home-panel sticky bottom-0 border-t border-gray-200 bg-white/95 p-4 shadow-2xl shadow-slate-900/10 backdrop-blur dark:border-slate-800 dark:bg-slate-950/95 dark:shadow-black/30">
            <div class="mb-3 rounded-2xl border border-violet-100 bg-violet-50 px-4 py-3 text-sm shadow-sm shadow-violet-900/5 dark:border-violet-400/20 dark:bg-violet-500/10 dark:shadow-none">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="font-medium text-slate-700 dark:text-slate-200">本次预计消耗</p>
                  <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ generationModeText }} · {{ selectedSizeDisplay }}</p>
                </div>
                <span class="rounded-full border px-3 py-1 text-sm font-semibold shadow-sm dark:border-violet-400/20 dark:bg-slate-900 dark:text-violet-200" :class="creditCostTone">{{ estimatedCreditCost }} 积分</span>
              </div>
              <div class="mt-3 flex items-center justify-between rounded-xl bg-white/70 px-3 py-2 text-xs text-slate-600 dark:bg-slate-950/40 dark:text-slate-300">
                <span>{{ userStore.user ? '可用余额' : '当前额度' }}</span>
                <span class="font-semibold" :class="canAffordCurrentGeneration ? 'text-emerald-700 dark:text-emerald-300' : 'text-amber-700 dark:text-amber-300'">{{ creditBalanceText }}</span>
              </div>
              <div class="mt-2 flex items-center justify-between gap-3">
                <p v-if="generateHint" class="text-xs text-slate-500 dark:text-slate-400">{{ generateHint }}</p>
                <button class="shrink-0 text-xs font-medium text-violet-700 transition hover:text-violet-900 dark:text-violet-300 dark:hover:text-violet-100" type="button" @click="isBillingGuideOpen = true">
                  查看计费规则
                </button>
              </div>
            </div>
            <button
              class="w-full rounded-xl bg-gradient-to-r from-violet-600 to-blue-600 py-4 text-white shadow-lg shadow-violet-500/30 transition hover:from-violet-700 hover:to-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
              type="button"
              :disabled="!canGenerate"
              @click="generate"
            >
              {{ loading ? '处理中...' : '开始生成' }}
            </button>
          </div>
        </div>
      </aside>
    </div>

    <div v-show="isFullscreenOpen && imageURL" ref="fullscreenEl" class="fixed inset-0 z-50 flex flex-col bg-black" @click.self="closeFullscreen">
      <div class="flex min-h-16 items-center justify-between gap-3 border-b border-white/10 px-4 py-3 text-white sm:px-6">
        <div class="min-w-0">
          <p class="truncate text-sm font-medium">生成结果</p>
          <p class="text-xs text-white/60">{{ selectedSizeDisplay }} · {{ estimatedCreditCost }} 积分</p>
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

    <div v-if="isBillingGuideOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/55 px-4" role="dialog" aria-modal="true" aria-labelledby="billing-guide-title" @click.self="isBillingGuideOpen = false">
      <div class="w-full max-w-md rounded-2xl bg-white p-5 text-slate-900 shadow-xl dark:bg-slate-900 dark:text-slate-100">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 id="billing-guide-title" class="text-lg font-semibold">积分怎么计算</h3>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">按最终图片像素计费，生成前会先展示预计消耗。</p>
          </div>
          <button class="rounded-full border border-slate-200 px-3 py-1 text-sm text-slate-500 transition hover:bg-slate-50 dark:border-slate-700 dark:hover:bg-slate-800" type="button" @click="isBillingGuideOpen = false">关闭</button>
        </div>
        <div class="mt-5 space-y-3 text-sm leading-6 text-slate-600 dark:text-slate-300">
          <div class="rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950">
            <p class="font-medium text-slate-900 dark:text-slate-100">标准图</p>
            <p class="mt-1">1024 x 1024 记为 1 积分。</p>
          </div>
          <div class="rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950">
            <p class="font-medium text-slate-900 dark:text-slate-100">更大尺寸</p>
            <p class="mt-1">按像素量向上取整。例如 1536 x 1024 约为 1.5 张标准图，按 2 积分计算。</p>
          </div>
          <div class="rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-950">
            <p class="font-medium text-slate-900 dark:text-slate-100">失败任务</p>
            <p class="mt-1">如果生成失败或你在开始前取消任务，系统会按当前后端规则退回已扣积分。</p>
          </div>
        </div>
        <button class="mt-5 w-full rounded-xl bg-violet-600 py-3 text-sm font-semibold text-white transition hover:bg-violet-700" type="button" @click="isBillingGuideOpen = false">
          我知道了
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
@keyframes nudge-expand {
  0%,
  100% {
    transform: translateY(-50%) translateX(0);
  }
  50% {
    transform: translateY(-50%) translateX(-4px);
  }
}

.panel-toggle-nudge {
  animation: nudge-expand 1.5s ease-in-out 2;
}

@media (max-width: 767px) {
  .scene-scroll {
    display: flex;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    padding-bottom: 0.25rem;
  }
}
</style>
