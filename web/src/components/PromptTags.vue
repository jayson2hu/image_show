<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import api from '@/api'

interface PromptTemplate {
  id?: number
  category: 'default' | 'repair' | 'style' | 'sample'
  label: string
  prompt: string
}

const emit = defineEmits<{
  select: [prompt: string]
}>()

const templates = ref<PromptTemplate[]>([])
const groups = computed(() => [
  { key: 'default', label: '默认', items: templates.value.filter((item) => item.category === 'default') },
  { key: 'repair', label: '修复', items: templates.value.filter((item) => item.category === 'repair') },
  { key: 'style', label: '风格', items: templates.value.filter((item) => item.category === 'style') },
])

onMounted(async () => {
  const response = await api.get('/prompt-templates')
  templates.value = response.data.items
})
</script>

<template>
  <div v-if="templates.length" class="space-y-3">
    <div v-for="group in groups" :key="group.key" class="space-y-2">
      <h3 v-if="group.items.length" class="text-sm font-medium text-slate-600">{{ group.label }}</h3>
      <div v-if="group.items.length" class="flex flex-wrap gap-2">
        <button
          v-for="item in group.items"
          :key="`${item.category}-${item.label}`"
          class="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-sm hover:border-violet-300"
          type="button"
          @click="emit('select', item.prompt)"
        >
          {{ item.label }}
        </button>
      </div>
    </div>
  </div>
</template>
