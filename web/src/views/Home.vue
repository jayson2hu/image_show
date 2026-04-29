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
const costs = { low: 0.2, medium: 1, high: 4 }
const cost = computed(() => costs[quality.value])

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
  error.value = ''
  imageURL.value = ''
  generationId.value = null
  loading.value = true
  try {
    const response = await api.post('/generations', {
      prompt: prompt.value,
      quality: quality.value,
      size: size.value,
    })
    generationId.value = response.data.id
  } catch (err: any) {
    error.value = err.response?.data?.error || '创建生成任务失败'
    loading.value = false
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
</script>

<template>
  <section class="grid gap-6 lg:grid-cols-[1fr_320px]">
    <div class="space-y-4">
      <div>
        <h1 class="text-2xl font-semibold">图片生成</h1>
        <p class="mt-2 text-sm text-slate-600">输入提示词，选择质量和尺寸后生成图片。</p>
      </div>
      <textarea
        v-model="prompt"
        class="min-h-44 w-full resize-y rounded border border-slate-300 bg-white p-4 text-base outline-none focus:border-teal"
        placeholder="描述你想生成的图片"
      />
      <PromptTags @select="appendPrompt" />
      <div class="grid gap-3 sm:grid-cols-2">
        <label class="text-sm font-medium">
          质量
          <select v-model="quality" class="mt-1 w-full rounded border border-slate-300 bg-white px-3 py-2">
            <option value="low">low - 0.2 积分</option>
            <option value="medium">medium - 1 积分</option>
            <option value="high">high - 4 积分</option>
          </select>
        </label>
        <label class="text-sm font-medium">
          尺寸
          <select v-model="size" class="mt-1 w-full rounded border border-slate-300 bg-white px-3 py-2">
            <option value="1024x1024">1024x1024</option>
            <option value="1024x1536">1024x1536</option>
            <option value="1536x1024">1536x1024</option>
          </select>
        </label>
      </div>
      <p v-if="!userStore.user" class="text-sm text-slate-600">未登录可免费试用 1 次，生成质量会自动使用 low。</p>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="rounded bg-coral px-4 py-2 text-white disabled:opacity-60" type="button" :disabled="loading || !prompt" @click="generate">
        {{ loading ? '生成中...' : `生成图片（${cost} 积分）` }}
      </button>
      <GenerationProgress v-if="generationId" :generation-id="generationId" @completed="completed" @failed="failed" />
      <ImagePreview v-if="imageURL" :url="imageURL" />
    </div>

    <aside class="h-fit rounded border border-slate-200 bg-white p-4">
      <h2 class="text-base font-medium">状态</h2>
      <p class="mt-2 text-sm text-slate-600">{{ health }}</p>
    </aside>
  </section>
</template>
