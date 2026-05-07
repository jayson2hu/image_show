<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const rangeStart = computed(() => (props.total === 0 ? 0 : Math.min((props.page - 1) * props.pageSize + 1, props.total)))
const rangeEnd = computed(() => (props.total === 0 ? 0 : Math.min(props.page * props.pageSize, props.total)))
const hasPrev = computed(() => props.page > 1)
const hasNext = computed(() => props.page < totalPages.value)
</script>

<template>
  <div class="flex flex-col gap-3 text-sm text-slate-600 sm:flex-row sm:items-center sm:justify-between">
    <span>第 {{ rangeStart }}-{{ rangeEnd }} 条 / 共 {{ total }} 条</span>
    <div class="flex items-center gap-1">
      <button class="rounded-lg border border-slate-200 px-3 py-1.5 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40" :disabled="!hasPrev" type="button" @click="emit('update:page', page - 1)">
        上一页
      </button>
      <span class="flex min-w-20 justify-center px-3 text-slate-500">{{ page }} / {{ totalPages }}</span>
      <button class="rounded-lg border border-slate-200 px-3 py-1.5 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40" :disabled="!hasNext" type="button" @click="emit('update:page', page + 1)">
        下一页
      </button>
    </div>
  </div>
</template>
