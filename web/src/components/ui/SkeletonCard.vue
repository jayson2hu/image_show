<script setup lang="ts">
withDefaults(
  defineProps<{
    lines?: number
  }>(),
  {
    lines: 3,
  },
)
</script>

<template>
  <div class="animate-pulse rounded-xl border border-slate-200 bg-white p-5">
    <div class="h-4 w-1/3 rounded bg-slate-200"></div>
    <div class="mt-4 space-y-3">
      <div v-for="line in lines" :key="line" class="h-3 rounded bg-slate-100" :class="line % 2 === 0 ? 'w-full' : 'w-2/3'"></div>
    </div>
  </div>
</template>
