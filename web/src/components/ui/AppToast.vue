<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const toast = useToast()

const colorMap = {
  success: 'border-emerald-200 bg-emerald-50 text-emerald-800',
  error: 'border-red-200 bg-red-50 text-red-800',
  info: 'border-blue-200 bg-blue-50 text-blue-800',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed right-4 top-4 z-[100] flex w-[calc(100vw-2rem)] max-w-sm flex-col gap-2">
      <TransitionGroup
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-x-8 opacity-0"
        enter-to-class="translate-x-0 opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="translate-x-8 opacity-0"
      >
        <div
          v-for="item in toast.items"
          :key="item.id"
          class="flex items-start gap-3 rounded-xl border px-4 py-3 text-sm shadow-lg"
          :class="colorMap[item.type]"
          role="status"
        >
          <span class="min-w-0 flex-1 leading-6">{{ item.message }}</span>
          <button class="rounded px-1 text-lg leading-5 opacity-60 transition hover:opacity-100" type="button" aria-label="关闭通知" @click="toast.remove(item.id)">
            &times;
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
