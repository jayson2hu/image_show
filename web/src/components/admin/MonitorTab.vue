<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchMonitorSummary, triggerMonitorCheck } from '@/api/admin'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { MonitorSummary } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const checking = ref(false)
const monitor = ref<MonitorSummary | null>(null)

const metrics = computed(() => [
  { label: '生成总数', value: monitor.value?.generation_count ?? 0 },
  { label: '完成数', value: monitor.value?.completed_count ?? 0 },
  { label: '失败数', value: monitor.value?.failed_count ?? 0 },
  { label: '失败率', value: `${monitor.value?.failure_rate ?? 0}%` },
  { label: '积分消耗', value: monitor.value?.credits_consumed ?? 0 },
  { label: '新增用户', value: monitor.value?.new_users ?? 0 },
  { label: '支付订单', value: monitor.value?.paid_order_count ?? 0 },
  { label: '支付金额', value: `¥${monitor.value?.paid_order_amount ?? 0}` },
])

onMounted(loadMonitor)

async function loadMonitor() {
  loading.value = true
  try {
    const response = await fetchMonitorSummary()
    monitor.value = response.data
  } catch (error: any) {
    toast.error(error.response?.data?.error || '监控数据加载失败')
  } finally {
    loading.value = false
  }
}

async function checkAlert() {
  checking.value = true
  try {
    const response = await triggerMonitorCheck()
    toast.success(response.data.sent ? '告警已发送' : '当前未触发告警')
    await loadMonitor()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '告警检查失败')
  } finally {
    checking.value = false
  }
}

function formatTime(value: string) {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Monitor</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">监控告警</h2>
        <p class="mt-2 text-sm text-slate-500">查看每日生成、失败原因、积分消耗和告警状态。</p>
      </div>
      <div class="flex gap-2">
        <button class="rounded-2xl border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-50" type="button" @click="loadMonitor">刷新</button>
        <button class="rounded-2xl bg-slate-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:opacity-60" type="button" :disabled="checking" @click="checkAlert">
          {{ checking ? '检查中' : '检查告警' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <SkeletonCard v-for="item in 8" :key="item" :lines="2" />
    </div>

    <template v-else-if="monitor">
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article v-for="item in metrics" :key="item.label" class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <p class="text-sm text-slate-500">{{ item.label }}</p>
          <p class="mt-3 text-3xl font-semibold text-slate-950">{{ item.value }}</p>
        </article>
      </div>

      <div class="rounded-3xl border p-5 shadow-sm" :class="monitor.alert_triggered ? 'border-rose-200 bg-rose-50 text-rose-800' : 'border-emerald-200 bg-emerald-50 text-emerald-800'">
        <h3 class="text-lg font-semibold">{{ monitor.alert_triggered ? '已触发告警' : '当前运行正常' }}</h3>
        <p class="mt-2 text-sm">告警阈值：{{ monitor.alert_threshold }} 积分，当前消耗：{{ monitor.credits_consumed }} 积分。</p>
      </div>

      <div class="grid gap-6 xl:grid-cols-2">
        <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 class="text-lg font-semibold text-slate-950">失败原因</h3>
          <div v-if="monitor.failure_reasons?.length" class="mt-4 space-y-3">
            <div v-for="item in monitor.failure_reasons" :key="item.category" class="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3 text-sm">
              <span class="font-medium text-slate-800">{{ item.label }}</span>
              <span class="rounded-full bg-white px-2.5 py-1 text-xs text-slate-500">{{ item.count }} 次</span>
            </div>
          </div>
          <EmptyState v-else class="mt-4" title="暂无失败原因" description="当天没有失败记录或尚未生成统计。" />
        </section>

        <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 class="text-lg font-semibold text-slate-950">最近失败任务</h3>
          <div v-if="monitor.recent_failures?.length" class="mt-4 space-y-3">
            <div v-for="item in monitor.recent_failures" :key="item.id" class="rounded-2xl bg-slate-50 p-4 text-sm">
              <div class="flex items-start justify-between gap-3">
                <p class="line-clamp-2 font-medium text-slate-800">{{ item.prompt || '未填写提示词' }}</p>
                <span class="shrink-0 rounded-full bg-white px-2.5 py-1 text-xs text-slate-500">{{ item.label }}</span>
              </div>
              <p class="mt-2 text-xs text-slate-500">{{ item.size }} · {{ formatTime(item.created_at) }}</p>
              <p class="mt-2 line-clamp-2 text-xs text-rose-600">{{ item.error }}</p>
            </div>
          </div>
          <EmptyState v-else class="mt-4" title="暂无失败任务" description="最近没有失败任务。" />
        </section>
      </div>
    </template>

    <EmptyState v-else title="暂无监控数据" description="刷新后仍无数据时，请检查后端监控接口。" />
  </section>
</template>
