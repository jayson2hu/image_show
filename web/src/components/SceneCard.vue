<script setup lang="ts">
defineProps<{
  scene: {
    id: string | number
    name: string
    icon: string
    description: string
    recommended_ratio: string
    credit_cost: number
  }
  selected: boolean
  ratioLabel: string
}>()

const emit = defineEmits<{
  select: []
}>()
</script>

<template>
  <button
    class="scene-card group relative min-h-[136px] scroll-snap-align-start rounded-xl border bg-white p-4 text-left shadow-sm transition-all duration-200 ease-out active:scale-[0.97]"
    :class="selected ? 'border-2 border-violet-500 bg-violet-50 shadow-lg shadow-violet-200/70' : 'border-slate-200 hover:border-violet-300 hover:shadow-md hover:scale-[1.02]'"
    type="button"
    @click="emit('select')"
  >
    <span v-if="selected" class="absolute left-2 top-2 flex size-5 items-center justify-center rounded-full bg-violet-600 text-xs font-bold text-white">✓</span>
    <span class="block text-2xl transition-transform duration-200 group-hover:-translate-y-0.5">{{ scene.icon || '✨' }}</span>
    <span class="mt-3 block text-sm font-semibold text-slate-950">{{ scene.name }}</span>
    <span class="mt-1 block min-h-9 text-xs leading-5 text-slate-500">{{ scene.description }}</span>
    <span class="mt-3 flex items-center justify-between gap-2">
      <span class="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-semibold text-slate-600">{{ ratioLabel }}</span>
      <span class="rounded-full bg-violet-100 px-2.5 py-1 text-[11px] font-semibold text-violet-700">{{ scene.credit_cost }} 积分</span>
    </span>
  </button>
</template>

<style scoped>
.scene-card {
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}

@media (max-width: 767px) {
  .scene-card {
    min-width: calc((100vw - 3rem) / 2.3);
    scroll-snap-align: start;
  }
}
</style>
