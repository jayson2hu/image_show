<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

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
const guestCredits = ref(5)

const fallbackGreeting = computed(() => {
  const hour = new Date().getHours()
  const name = userStore.user?.username || userStore.user?.email?.split('@')[0] || ''
  const timePart = hour < 6 ? '凌晨好' : hour < 12 ? '上午好' : hour < 18 ? '下午好' : '晚上好'
  return name ? `${timePart}，${name}，准备好创作了吗？` : '准备好开始创作了吗？'
})
const greeting = computed(() => greetingText.value.trim() || fallbackGreeting.value)
const displayScenes = computed(() => [...scenes.value].sort((a, b) => a.sort_order - b.sort_order).slice(0, 5))
const guestCreditsExhausted = computed(() => !userStore.user && guestCredits.value <= 0)

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
    guestCredits.value = response.data.guest_free_credits ?? 5
  } catch {
    guestCredits.value = 5
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
    <div v-if="guestCreditsExhausted" class="w-full max-w-md text-center">
      <div class="mx-auto flex size-14 items-center justify-center rounded-full bg-coral/10 text-coral">
        <svg class="size-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M4.93 19h14.14a2 2 0 0 0 1.73-3L13.73 4a2 2 0 0 0-3.46 0L3.2 16a2 2 0 0 0 1.73 3Z" />
        </svg>
      </div>
      <h2 class="mt-4 text-xl font-semibold text-ink">体验积分已用完</h2>
      <p class="mt-2 text-sm leading-6 text-slate-500">登录或注册后可以继续使用完整创作能力，并查看历史作品。</p>
      <div class="mt-6 flex justify-center gap-3">
        <RouterLink class="rounded-lg bg-teal px-4 py-2 text-sm font-semibold text-white hover:bg-ink" to="/login">注册</RouterLink>
        <RouterLink class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-ink hover:bg-slate-50" to="/login">登录</RouterLink>
      </div>
    </div>

    <div v-else class="w-full max-w-[720px]">
      <div class="mb-6 text-center">
        <div v-if="!userStore.user" class="mb-3 inline-flex rounded-full bg-teal/10 px-3 py-1 text-xs font-semibold text-teal">
          体验积分：{{ guestCredits }}
        </div>
        <h2 class="text-[22px] font-semibold leading-tight text-ink">{{ greeting }}</h2>
        <p class="mt-1.5 text-[13px] text-slate-500">选择一个场景，或直接描述你想生成的画面。</p>
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
        <span class="font-semibold text-ink">{{ userStore.user?.credits ?? guestCredits }}</span>
        积分 ·
        <RouterLink class="text-teal hover:underline" to="/packages">查看套餐</RouterLink>
      </div>
    </div>
  </div>
</template>
