<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchChannels, fetchMonitorSummary, fetchUsers, triggerMonitorCheck } from '@/api/admin'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { Channel, MonitorSummary, Page, AdminUser } from '@/types/admin'

const emit = defineEmits<{
  'change-tab': [tab: string]
}>()

const toast = useToast()
const loading = ref(true)
const checking = ref(false)
const users = ref<Page<AdminUser> | null>(null)
const channels = ref<Channel[]>([])
const monitor = ref<MonitorSummary | null>(null)

const enabledChannels = computed(() => channels.value.filter((item) => item.status === 1).length)
const metricCards = computed(() => [
  {
    label: '今日生成',
    value: monitor.value?.generation_count ?? 0,
    hint: `完成 ${monitor.value?.completed_count ?? 0} / 失败 ${monitor.value?.failed_count ?? 0}`,
    tone: 'blue',
    icon: 'M4 16l4-4 4 4 8-8',
  },
  {
    label: '新增用户',
    value: monitor.value?.new_users ?? 0,
    hint: `用户总数 ${users.value?.total ?? 0}`,
    tone: 'emerald',
    icon: 'M16 11a4 4 0 1 0-8 0m8 0c2.5.8 4 2.2 4 4H4c0-1.8 1.5-3.2 4-4',
  },
  {
    label: '积分消耗',
    value: monitor.value?.credits_consumed ?? 0,
    hint: `告警阈值 ${monitor.value?.alert_threshold ?? 0}`,
    tone: 'violet',
    icon: 'M4 7h16v10H4V7Zm0 3h16M8 15h3',
  },
  {
    label: '启用渠道',
    value: enabledChannels.value,
    hint: `全部渠道 ${channels.value.length}`,
    tone: 'amber',
    icon: 'M4 6h16M4 12h16M4 18h16',
  },
])

onMounted(loadOverview)

async function loadOverview() {
  loading.value = true
  try {
    const [userResp, channelResp, monitorResp] = await Promise.all([
      fetchUsers({ pageSize: 1 }),
      fetchChannels(),
      fetchMonitorSummary(),
    ])
    users.value = userResp.data
    channels.value = channelResp.data.items || []
    monitor.value = monitorResp.data
  } catch (error: any) {
    toast.error(error.response?.data?.error || '概览数据加载失败')
  } finally {
    loading.value = false
  }
}

async function checkAlert() {
  checking.value = true
  try {
    const response = await triggerMonitorCheck()
    toast.success(response.data.sent ? '告警已发送' : '当前未触发告警')
    await loadOverview()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '告警检查失败')
  } finally {
    checking.value = false
  }
}

function toneClass(tone: string) {
  const map: Record<string, string> = {
    blue: 'bg-blue-50 text-blue-700',
    emerald: 'bg-emerald-50 text-emerald-700',
    violet: 'bg-violet-50 text-violet-700',
    amber: 'bg-amber-50 text-amber-700',
  }
  return map[tone] || map.blue
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Overview</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">运营概览</h2>
        <p class="mt-2 text-sm text-slate-500">查看今日生成、用户增长、渠道状态和告警摘要。</p>
      </div>
      <button class="rounded-2xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:opacity-60" type="button" :disabled="checking" @click="checkAlert">
        {{ checking ? '检查中...' : '检查告警' }}
      </button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <SkeletonCard v-for="item in 4" :key="item" :lines="2" />
    </div>

    <template v-else>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article v-for="card in metricCards" :key="card.label" class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-sm text-slate-500">{{ card.label }}</p>
              <p class="mt-3 text-3xl font-semibold text-slate-950">{{ card.value }}</p>
            </div>
            <span class="flex size-11 items-center justify-center rounded-2xl" :class="toneClass(card.tone)">
              <svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" :d="card.icon" />
              </svg>
            </span>
          </div>
          <p class="mt-4 inline-flex rounded-full bg-slate-50 px-3 py-1 text-xs text-slate-500">{{ card.hint }}</p>
        </article>
      </div>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h3 class="text-lg font-semibold text-slate-950">渠道状态</h3>
              <p class="mt-1 text-sm text-slate-500">展示最近 4 个渠道的启用和测试状态。</p>
            </div>
            <button class="rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:bg-slate-50" type="button" @click="emit('change-tab', 'channels')">管理渠道</button>
          </div>
          <div v-if="channels.length" class="mt-5 divide-y divide-slate-100">
            <div v-for="channel in channels.slice(0, 4)" :key="channel.id" class="flex items-center justify-between gap-4 py-3">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="size-2.5 rounded-full" :class="channel.status === 1 ? 'bg-emerald-500' : 'bg-slate-300'"></span>
                  <p class="truncate font-medium text-slate-900">{{ channel.name }}</p>
                </div>
                <p class="mt-1 truncate text-xs text-slate-500">{{ channel.base_url }}</p>
              </div>
              <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="channel.last_test_success ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'">
                {{ channel.last_test_success ? '测试正常' : '待测试' }}
              </span>
            </div>
          </div>
          <EmptyState v-else class="mt-5" title="暂无渠道" description="添加渠道后可以在这里查看启用状态。">
            <button class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-medium text-white" type="button" @click="emit('change-tab', 'channels')">去添加渠道</button>
          </EmptyState>
        </section>

        <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 class="text-lg font-semibold text-slate-950">今日监控</h3>
          <div class="mt-5 space-y-3 text-sm">
            <div class="flex justify-between rounded-2xl bg-slate-50 px-4 py-3">
              <span class="text-slate-500">失败率</span>
              <span class="font-semibold text-slate-900">{{ monitor?.failure_rate ?? 0 }}%</span>
            </div>
            <div class="flex justify-between rounded-2xl bg-slate-50 px-4 py-3">
              <span class="text-slate-500">支付订单</span>
              <span class="font-semibold text-slate-900">{{ monitor?.paid_order_count ?? 0 }} 单</span>
            </div>
            <div class="flex justify-between rounded-2xl bg-slate-50 px-4 py-3">
              <span class="text-slate-500">订单金额</span>
              <span class="font-semibold text-slate-900">¥{{ monitor?.paid_order_amount ?? 0 }}</span>
            </div>
            <div class="rounded-2xl px-4 py-3" :class="monitor?.alert_triggered ? 'bg-rose-50 text-rose-700' : 'bg-emerald-50 text-emerald-700'">
              {{ monitor?.alert_triggered ? '当前已触发告警，请检查积分消耗。' : '当前未触发告警。' }}
            </div>
          </div>
        </section>
      </div>

      <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
        <h3 class="text-lg font-semibold text-slate-950">快捷操作</h3>
        <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <button class="rounded-2xl border border-slate-200 px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50" type="button" @click="emit('change-tab', 'users')">用户管理</button>
          <button class="rounded-2xl border border-slate-200 px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50" type="button" @click="emit('change-tab', 'channels')">渠道配置</button>
          <button class="rounded-2xl border border-slate-200 px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50" type="button" @click="emit('change-tab', 'settings')">系统设置</button>
          <button class="rounded-2xl border border-slate-200 px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50" type="button" @click="emit('change-tab', 'monitor')">监控告警</button>
        </div>
      </section>
    </template>
  </section>
</template>
