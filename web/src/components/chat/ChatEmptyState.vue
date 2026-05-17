<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { fallbackSceneItems, listScenes } from '@/api/scene'
import { fetchSiteConfig } from '@/api/site'
import type { Scene } from '@/api/types'
import { useComposerStore } from '@/stores/composer'
import { useUserStore } from '@/stores/user'

import Composer from './Composer.vue'

const userStore = useUserStore()
const composerStore = useComposerStore()
const scenes = ref<Scene[]>([])
const greetingText = ref('')
const guestLimit = ref(5)

const fallbackGreeting = computed(() => {
  const hour = new Date().getHours()
  const name = userStore.user?.username || userStore.user?.email?.split('@')[0] || ''
  const timePart = hour < 6 ? '夜深了' : hour < 12 ? '上午好' : hour < 18 ? '下午好' : '晚上好'
  return name ? `${timePart}，${name}` : '想生成什么图片？'
})
const greeting = computed(() => greetingText.value.trim() || fallbackGreeting.value)
const displayScenes = computed(() => [...scenes.value].sort((a, b) => a.sort_order - b.sort_order).slice(0, 5))

onMounted(async () => {
  await Promise.all([loadScenes(), loadSite()])
  nextTick(() => composerStore.focusInput())
})

async function loadScenes() {
  try {
    const response = await listScenes()
    scenes.value = response.items
  } catch (error) {
    console.error('Load scenes failed', error)
    scenes.value = fallbackSceneItems()
  }
}

async function loadSite() {
  try {
    const response = await fetchSiteConfig()
    greetingText.value = response.data.greeting_text || ''
    guestLimit.value = response.data.guest_generation_limit ?? response.data.guest_free_credits ?? 5
  } catch {
    guestLimit.value = 5
  }
}

function onSceneClick(scene: Scene) {
  composerStore.setDraft({
    prompt: scene.prompt_template,
    size: scene.recommended_ratio,
    scene_id: String(scene.id),
    layer_count: scene.default_layer_count || 5,
  })
  nextTick(() => composerStore.focusInput())
}
</script>

<template>
  <div class="flex h-full items-center justify-center px-4 py-8 sm:px-6">
    <div class="w-full max-w-[720px]">
      <div class="mb-6 text-center">
        <div v-if="!userStore.user" class="mb-3 inline-flex rounded-full bg-teal/10 px-3 py-1 text-xs font-semibold text-teal">
          游客可生成 {{ guestLimit }} 次
        </div>
        <h2 class="text-[22px] font-semibold leading-tight text-ink">{{ greeting }}</h2>
        <p class="mt-1.5 text-[13px] text-slate-500">输入提示词，或选择一个场景开始生成。</p>
      </div>

      <div v-if="displayScenes.length" class="mb-3 flex flex-wrap items-center justify-center gap-2">
        <button
          v-for="scene in displayScenes"
          :key="scene.id"
          class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-2 text-[13px] text-ink ring-1 ring-slate-200 transition hover:-translate-y-px hover:ring-teal/45"
          type="button"
          @click="onSceneClick(scene)"
        >
          <span>{{ scene.icon }}</span>
          <span>{{ scene.name }}</span>
        </button>
      </div>

      <Composer compact />

      <div class="mt-3 text-center text-[11px] text-slate-500">
        Enter 发送 · Shift+Enter 换行 · 当前
        <span class="font-semibold text-ink">{{ userStore.user?.credits ?? guestLimit }}</span>
        {{ userStore.user ? '点数' : '次' }}
      </div>
    </div>
  </div>
</template>
