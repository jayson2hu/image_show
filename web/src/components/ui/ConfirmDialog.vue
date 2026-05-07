<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    message: string
    confirmText?: string
    confirmColor?: 'red' | 'blue' | 'default'
  }>(),
  {
    title: '确认操作',
    confirmText: '确认',
    confirmColor: 'red',
  },
)

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const colorMap = {
  red: 'bg-red-600 text-white hover:bg-red-700',
  blue: 'bg-blue-600 text-white hover:bg-blue-700',
  default: 'bg-slate-900 text-white hover:bg-slate-800',
}
</script>

<template>
  <Teleport to="body">
    <Transition enter-active-class="transition duration-200" enter-from-class="opacity-0" leave-active-class="transition duration-150" leave-to-class="opacity-0">
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="emit('cancel')">
        <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
          <h3 class="text-lg font-semibold text-slate-900">{{ props.title }}</h3>
          <p class="mt-2 text-sm leading-6 text-slate-600">{{ message }}</p>
          <div class="mt-6 flex justify-end gap-3">
            <button class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-50" type="button" @click="emit('cancel')">
              取消
            </button>
            <button class="rounded-lg px-4 py-2 text-sm font-medium transition" :class="colorMap[props.confirmColor]" type="button" @click="emit('confirm')">
              {{ props.confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
