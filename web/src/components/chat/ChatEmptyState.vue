<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'

import { fallbackSceneItems, listScenes } from '@/api/scene'
import type { Scene } from '@/api/types'
import { useComposerStore } from '@/stores/composer'
import { useUserStore } from '@/stores/user'

import Composer from './Composer.vue'

const userStore = useUserStore()
const composerStore = useComposerStore()
const scenes = ref<Scene[]>([])

const greeting = computed(() => {
  const hour = new Date().getHours()
  const name = userStore.user?.username || userStore.user?.email?.split('@')[0] || ''
  const timePart = hour < 6 ? '夜深了' : hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'
  return name ? `${timePart}，${name}，准备好创作了吗？` : '你好，准备好创作了吗？'
})

const displayScenes = computed(() => [...scenes.value].sort((a, b) => a.sort_order - b.sort_order).slice(0, 5))

onMounted(async () => {
  try {
    const response = await listScenes()
    scenes.value = response.items
  } catch (error) {
    console.error('Load scenes failed', error)
    scenes.value = fallbackSceneItems()
  } finally {
    nextTick(() => composerStore.focusInput())
  }
})

function onSceneClick(scene: Scene) {
  composerStore.setDraft({
    prompt: scene.prompt_template,
    size: scene.recommended_ratio,
    scene_id: String(scene.id),
    layered: Boolean(scene.default_layered),
    layer_count: scene.default_layer_count || 5,
  })
  nextTick(() => composerStore.focusInput())
}
</script>

<template>
  <div class="flex h-full items-center justify-center px-4 py-8 sm:px-6">
    <div class="w-full max-w-[720px]">
      <div class="mb-6 text-center">
        <h2 class="text-[22px] font-semibold leading-tight text-ink">{{ greeting }}</h2>
        <p class="mt-1.5 text-[13px] text-slate-500">输入一句话，或选下方的场景快速开始</p>
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
          <span v-if="scene.default_layered" class="rounded-full bg-coral/10 px-1.5 py-0.5 text-[10px] font-medium text-coral">自动分层</span>
        </button>
      </div>

      <Composer compact />

      <div class="mt-3 text-center text-[11px] text-slate-500">
        Enter 发送 · Shift+Enter 换行 · 余
        <span class="font-semibold text-ink">{{ userStore.user?.credits ?? 0 }}</span>
        积分 ·
        <RouterLink class="text-teal hover:underline" to="/packages">计费规则</RouterLink>
      </div>
    </div>
  </div>
</template>
