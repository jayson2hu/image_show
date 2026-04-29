<script setup lang="ts">
defineProps<{
  url: string
}>()
</script>

<template>
  <div class="rounded border border-slate-200 bg-white p-4">
    <img class="aspect-square w-full rounded object-contain" :src="url" alt="生成结果" />
    <div class="mt-4 flex gap-2">
      <a class="rounded bg-teal px-3 py-2 text-sm text-white" :href="url" download>下载</a>
    </div>
  </div>
</template>
