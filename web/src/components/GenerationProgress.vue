<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  generationId: number
}>()

const emit = defineEmits<{
  completed: [url: string]
  failed: [message: string]
  cancelled: []
  cancel: []
}>()

const currentStatus = ref(0)
const backendMessage = ref('')
const pointerActive = ref(false)
const pulseActive = ref(false)
const canvasStyle = ref<Record<string, string>>({
  '--pointer-x': '50%',
  '--pointer-y': '50%',
})
let source: EventSource | null = null

const statusCopy: Record<number, { title: string; detail: string }> = {
  0: { title: '任务已创建', detail: '正在进入生成队列' },
  1: { title: '正在生成图片', detail: '模型正在处理提示词和画面内容' },
  2: { title: '正在保存结果', detail: '正在处理图片文件并准备展示' },
  3: { title: '生成完成', detail: '图片已经准备好' },
  4: { title: '生成失败', detail: '任务没有完成，请查看错误信息' },
  5: { title: '任务已取消', detail: '本次生成已经停止' },
}

const phases = [
  { status: 0, label: '创建' },
  { status: 1, label: '生成' },
  { status: 2, label: '保存' },
  { status: 3, label: '完成' },
]

const currentCopy = computed(() => statusCopy[currentStatus.value] || { title: '处理中', detail: backendMessage.value || '请稍候' })
const progressPercent = computed(() => {
  if (currentStatus.value >= 3) {
    return 100
  }
  if (currentStatus.value <= 0) {
    return 12
  }
  if (currentStatus.value === 1) {
    return 56
  }
  return 84
})

onMounted(() => {
  source = new EventSource(`/api/generations/${props.generationId}/stream`)
  source.addEventListener('status', (event) => {
    const payload = JSON.parse((event as MessageEvent).data)
    currentStatus.value = payload.status
    if (payload.message) {
      backendMessage.value = payload.message
    }
    if (payload.status === 3) {
      emit('completed', payload.image_url)
      close()
    }
    if (payload.status === 4) {
      emit('failed', payload.error || '生成失败，请重试')
      close()
    }
    if (payload.status === 5) {
      emit('cancelled')
      close()
    }
  })
  source.onerror = () => {
    emit('failed', '连接中断，请稍后重试')
    close()
  }
})

onUnmounted(close)

function close() {
  source?.close()
  source = null
}

function updatePointer(event: PointerEvent) {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  canvasStyle.value = {
    '--pointer-x': `${((event.clientX - rect.left) / rect.width) * 100}%`,
    '--pointer-y': `${((event.clientY - rect.top) / rect.height) * 100}%`,
  }
  pointerActive.value = true
}

function clearPointer() {
  pointerActive.value = false
}

function pulseCanvas() {
  pulseActive.value = false
  window.requestAnimationFrame(() => {
    pulseActive.value = true
    window.setTimeout(() => {
      pulseActive.value = false
    }, 520)
  })
}
</script>

<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-700 dark:bg-slate-900">
    <div
      class="generation-canvas relative flex h-80 cursor-crosshair items-center justify-center overflow-hidden rounded-xl border border-slate-200 bg-slate-100 dark:border-slate-700 dark:bg-slate-950"
      :class="{ 'is-pointer-active': pointerActive, 'is-pulsing': pulseActive }"
      :style="canvasStyle"
      role="img"
      aria-label="生成中的图片预览占位"
      @pointermove="updatePointer"
      @pointerleave="clearPointer"
      @pointerdown="pulseCanvas"
    >
      <div class="relative z-10 flex size-24 items-center justify-center rounded-full border border-white/50 bg-white/45 shadow-xl shadow-slate-900/10 backdrop-blur-md dark:border-white/10 dark:bg-white/10">
        <div class="size-12 animate-spin rounded-full border-4 border-violet-200 border-t-violet-600 dark:border-violet-400/20 dark:border-t-violet-300"></div>
      </div>
      <div class="pointer-glow" aria-hidden="true"></div>
    </div>
    <div class="mt-4 space-y-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-slate-800 dark:text-slate-100">{{ currentCopy.title }}</p>
          <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ currentCopy.detail }}</p>
        </div>
        <button class="min-h-10 shrink-0 rounded-xl border border-slate-300 px-3 py-2 text-sm transition hover:border-violet-300 hover:text-violet-700 dark:border-slate-600 dark:hover:border-violet-400 dark:hover:text-violet-200" type="button" @click="emit('cancel')">
          取消
        </button>
      </div>
      <div class="h-2 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
        <div class="h-full rounded-full bg-violet-600 transition-all duration-500 ease-out dark:bg-violet-400" :style="{ width: `${progressPercent}%` }"></div>
      </div>
      <div class="grid grid-cols-4 gap-2 text-xs">
        <div v-for="phase in phases" :key="phase.status" class="flex items-center gap-2 text-slate-500 dark:text-slate-400" :class="{ 'text-violet-700 dark:text-violet-200': currentStatus >= phase.status }">
          <span class="size-2 rounded-full bg-slate-300 dark:bg-slate-700" :class="{ 'bg-violet-600 dark:bg-violet-300': currentStatus >= phase.status }"></span>
          <span>{{ phase.label }}</span>
        </div>
      </div>
      <p v-if="backendMessage" class="text-xs text-slate-400 dark:text-slate-500">后端状态：{{ backendMessage }}</p>
    </div>
  </div>
</template>

<style scoped>
.generation-canvas {
  --pointer-x: 50%;
  --pointer-y: 50%;
  background:
    radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(124, 58, 237, 0.18), transparent 26%),
    radial-gradient(circle at 20% 20%, rgba(14, 165, 233, 0.16), transparent 24%),
    linear-gradient(135deg, #f8fafc 0%, #eef2ff 46%, #f1f5f9 100%);
}

.generation-canvas::before,
.generation-canvas::after {
  content: '';
  position: absolute;
  inset: -22%;
  pointer-events: none;
}

.generation-canvas::before {
  background-image:
    radial-gradient(circle, rgba(15, 23, 42, 0.16) 1px, transparent 1px),
    radial-gradient(circle, rgba(124, 58, 237, 0.12) 1px, transparent 1px);
  background-position:
    0 0,
    13px 17px;
  background-size:
    28px 28px,
    34px 34px;
  opacity: 0.46;
  animation: dot-drift 12s linear infinite;
}

.generation-canvas::after {
  background:
    repeating-radial-gradient(ellipse at 50% 50%, transparent 0 18px, rgba(99, 102, 241, 0.08) 19px 21px, transparent 22px 42px);
  opacity: 0.62;
  animation: wave-breathe 6s ease-in-out infinite;
}

.generation-canvas.is-pointer-active .pointer-glow {
  opacity: 1;
  transform: translate(-50%, -50%) scale(1);
}

.generation-canvas.is-pulsing .pointer-glow {
  animation: click-pulse 520ms ease-out;
}

.pointer-glow {
  position: absolute;
  left: var(--pointer-x);
  top: var(--pointer-y);
  z-index: 1;
  width: 180px;
  height: 180px;
  border-radius: 9999px;
  background: radial-gradient(circle, rgba(124, 58, 237, 0.24), rgba(14, 165, 233, 0.1) 42%, transparent 70%);
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.82);
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

:global(.dark) .generation-canvas {
  background:
    radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(167, 139, 250, 0.2), transparent 28%),
    radial-gradient(circle at 20% 20%, rgba(14, 165, 233, 0.12), transparent 24%),
    linear-gradient(135deg, #020617 0%, #111827 52%, #0f172a 100%);
}

:global(.dark) .generation-canvas::before {
  background-image:
    radial-gradient(circle, rgba(248, 250, 252, 0.24) 1px, transparent 1px),
    radial-gradient(circle, rgba(167, 139, 250, 0.18) 1px, transparent 1px);
  opacity: 0.4;
}

@keyframes dot-drift {
  from {
    transform: translate3d(0, 0, 0);
  }
  to {
    transform: translate3d(28px, 34px, 0);
  }
}

@keyframes wave-breathe {
  0%,
  100% {
    transform: scale(0.98) rotate(0deg);
    opacity: 0.46;
  }
  50% {
    transform: scale(1.05) rotate(2deg);
    opacity: 0.7;
  }
}

@keyframes click-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.32);
  }
  100% {
    box-shadow: 0 0 0 32px rgba(124, 58, 237, 0);
  }
}
</style>
