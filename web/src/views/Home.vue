<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import api from '@/api'
import GenerationProgress from '@/components/GenerationProgress.vue'
import ImagePreview from '@/components/ImagePreview.vue'
import PromptTags from '@/components/PromptTags.vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const health = ref('检查中')
const prompt = ref('')
const quality = ref<'low' | 'medium' | 'high'>('medium')
const size = ref('1024x1024')
const generationId = ref<number | null>(null)
const imageURL = ref('')
const error = ref('')
const loading = ref(false)
const lastRequest = ref<{ prompt: string; quality: 'low' | 'medium' | 'high'; size: string } | null>(null)
const costs = { low: 0.2, medium: 1, high: 4 }
const cost = computed(() => costs[quality.value])
const canRetry = computed(() => !!lastRequest.value && !loading.value)

onMounted(async () => {
  try {
    const response = await api.get('/health')
    health.value = response.data.status === 'ok' ? '后端已连接' : '后端响应异常'
  } catch {
    health.value = '后端未连接'
  }
})

function appendPrompt(value: string) {
  prompt.value = prompt.value ? `${prompt.value}，${value}` : value
}

async function generate() {
  await createGeneration({
    prompt: prompt.value,
    quality: quality.value,
    size: size.value,
  })
}

async function retry() {
  if (!lastRequest.value) {
    return
  }
  prompt.value = lastRequest.value.prompt
  quality.value = lastRequest.value.quality
  size.value = lastRequest.value.size
  await createGeneration(lastRequest.value)
}

async function createGeneration(payload: { prompt: string; quality: 'low' | 'medium' | 'high'; size: string }) {
  error.value = ''
  imageURL.value = ''
  generationId.value = null
  loading.value = true
  lastRequest.value = { ...payload }
  try {
    const response = await api.post('/generations', payload)
    generationId.value = response.data.id
  } catch (err: any) {
    error.value = err.response?.data?.error || '创建生成任务失败'
    loading.value = false
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
  }
}

function completed(url: string) {
  imageURL.value = url
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
}

function failed(message: string) {
  error.value = message
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
}

function cancelled() {
  error.value = '任务已取消'
  generationId.value = null
  loading.value = false
  userStore.fetchUser()
}
</script>

<template>
  <section class="grid gap-6 lg:grid-cols-[1fr_320px]">
    <div class="space-y-4">
      <div>
        <h1 class="text-2xl font-semibold">图片生成</h1>
        <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">输入提示词，选择质量和尺寸后生成图片。</p>
      </div>
      <textarea
        v-model="prompt"
        class="min-h-52 w-full resize-y rounded border border-slate-300 bg-white p-4 text-base outline-none focus:border-teal dark:border-slate-700 dark:bg-slate-900"
        placeholder="描述你想生成的图片"
      />
      <PromptTags @select="appendPrompt" />
      <div class="grid gap-3 sm:grid-cols-2">
        <label class="text-sm font-medium">
          质量
          <select v-model="quality" class="mt-1 min-h-11 w-full rounded border border-slate-300 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
            <option value="low">low - 0.2 积分</option>
            <option value="medium">medium - 1 积分</option>
            <option value="high">high - 4 积分</option>
          </select>
        </label>
        <label class="text-sm font-medium">
          尺寸
          <select v-model="size" class="mt-1 min-h-11 w-full rounded border border-slate-300 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
            <option value="1024x1024">1024x1024</option>
            <option value="1024x1536">1024x1536</option>
            <option value="1536x1024">1536x1024</option>
          </select>
        </label>
      </div>
      <p v-if="!userStore.user" class="text-sm text-slate-600 dark:text-slate-300">未登录可免费试用 1 次，生成质量会自动使用 low。</p>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex flex-col gap-2 sm:flex-row">
        <button class="min-h-11 rounded bg-coral px-4 py-2 text-white disabled:opacity-60" type="button" :disabled="loading || !prompt" @click="generate">
          {{ loading ? '生成中...' : `生成图片（${cost} 积分）` }}
        </button>
        <button v-if="canRetry" class="min-h-11 rounded border border-slate-300 px-4 py-2 dark:border-slate-600" type="button" @click="retry">
          重试
        </button>
      </div>
      <GenerationProgress
        v-if="generationId"
        :generation-id="generationId"
        @completed="completed"
        @failed="failed"
        @cancelled="cancelled"
        @cancel="cancelGeneration"
      />
      <ImagePreview v-if="imageURL" :url="imageURL" />
    </div>

    <aside class="h-fit rounded border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900">
      <h2 class="text-base font-medium">状态</h2>
      <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">{{ health }}</p>
    </aside>
  </section>
</template>
