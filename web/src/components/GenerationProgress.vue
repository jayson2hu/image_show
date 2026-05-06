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
const displayStage = computed(() => {
  if (currentStatus.value >= 3) {
    return 3
  }
  if (currentStatus.value === 2) {
    return 2
  }
  if (currentStatus.value === 1) {
    return 1
  }
  return 0
})
const currentPhaseLabel = computed(() => phases.find((phase) => phase.status === displayStage.value)?.label || '创建')
const progressPercent = computed(() => {
  if (displayStage.value >= 3) {
    return 100
  }
  return displayStage.value * 33.333
})

onMounted(() => {
  source = new EventSource(`/api/generations/${props.generationId}/stream`)
  source.addEventListener('status', (event) => {
    const payload = JSON.parse((event as MessageEvent).data)
    scheduleStatus(payload)
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

function scheduleStatus(payload: any) {
  applyStatus(payload)
}

function applyStatus(payload: any) {
  currentStatus.value = payload.status
  backendMessage.value = payload.message || statusCopy[payload.status]?.title || '处理中'
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
  <div class="flex h-full min-h-[560px] flex-col bg-slate-50 text-slate-950">
    <div
      class="generation-canvas relative flex min-h-0 flex-1 cursor-crosshair items-center justify-center overflow-hidden p-6 sm:p-10"
      :class="{ 'is-pointer-active': pointerActive, 'is-pulsing': pulseActive }"
      :style="canvasStyle"
      role="img"
      aria-label="生成中的图片预览占位"
      @pointermove="updatePointer"
      @pointerleave="clearPointer"
      @pointerdown="pulseCanvas"
    >
      <div class="relative z-10 flex w-full max-w-xl flex-col items-center rounded-2xl border border-slate-200 bg-white/88 px-6 py-8 text-center shadow-xl shadow-slate-200/70 backdrop-blur sm:px-10 sm:py-12">
        <div class="mb-6 flex size-28 items-center justify-center rounded-full bg-slate-100 sm:size-32">
          <div class="size-16 animate-spin rounded-full border-[5px] border-slate-200 border-t-violet-600 sm:size-20"></div>
        </div>
        <p class="text-2xl font-medium tracking-tight sm:text-3xl">{{ currentCopy.title }}</p>
        <p class="mt-3 max-w-md text-sm leading-6 text-slate-500">{{ currentCopy.detail }}</p>
      </div>
      <div class="pointer-glow" aria-hidden="true"></div>
    </div>
    <div class="border-t border-slate-200 bg-white p-5 shadow-sm sm:p-7">
      <div class="mx-auto max-w-6xl space-y-5">
      <div class="flex items-center justify-between gap-4">
        <div class="min-w-0">
          <p class="text-sm font-medium text-slate-900">当前阶段：{{ currentPhaseLabel }}</p>
        </div>
        <button class="min-h-10 shrink-0 rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-violet-300 hover:text-violet-700" type="button" @click="emit('cancel')">
          取消
        </button>
      </div>
      <div class="h-2.5 overflow-hidden rounded-full bg-slate-100">
        <div class="h-full rounded-full bg-violet-600 transition-all duration-500 ease-out" :style="{ width: `${progressPercent}%` }"></div>
      </div>
      <div class="grid grid-cols-4 gap-3 text-xs sm:text-sm">
        <div
          v-for="phase in phases"
          :key="phase.status"
          class="flex items-center gap-2 text-slate-400"
          :class="{ 'text-violet-700': displayStage >= phase.status }"
        >
          <span
            class="size-2.5 rounded-full bg-slate-200 ring-0 ring-violet-100 transition"
            :class="{
              'bg-violet-600': displayStage >= phase.status,
              'ring-4': displayStage === phase.status,
            }"
          ></span>
          <span>{{ phase.label }}</span>
        </div>
      </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.generation-canvas {
  --pointer-x: 50%;
  --pointer-y: 50%;
  background:
    radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(124, 58, 237, 0.1), transparent 28%),
    linear-gradient(180deg, #f8fafc 0%, #eef2ff 100%);
}

.generation-canvas::before,
.generation-canvas::after {
  content: '';
  position: absolute;
  inset: -22%;
  pointer-events: none;
}

.generation-canvas::before {
  background: linear-gradient(135deg, transparent 0 48%, rgba(124, 58, 237, 0.07) 49% 51%, transparent 52% 100%);
  opacity: 0.4;
}

.generation-canvas::after {
  display: none;
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
  width: 240px;
  height: 240px;
  border-radius: 9999px;
  background: radial-gradient(circle, rgba(124, 58, 237, 0.16), rgba(14, 165, 233, 0.06) 42%, transparent 70%);
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.82);
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

:global(.dark) .generation-canvas {
  background:
    radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(167, 139, 250, 0.14), transparent 28%),
    linear-gradient(180deg, #f8fafc 0%, #eef2ff 100%);
}

:global(.dark) .generation-canvas::before {
  opacity: 0.36;
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
