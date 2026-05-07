<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { createChannel, deleteChannel, fetchChannels, testChannel, updateChannel } from '@/api/admin'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { Channel } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const saving = ref(false)
const channels = ref<Channel[]>([])
const modalOpen = ref(false)
const editingChannel = ref<Channel | null>(null)
const deleteTarget = ref<Channel | null>(null)
const testing = ref<Record<number, boolean>>({})
const form = ref<Channel>({ id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' })

onMounted(loadChannels)

async function loadChannels() {
  loading.value = true
  try {
    const response = await fetchChannels()
    channels.value = response.data.items || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || '渠道列表加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingChannel.value = null
  form.value = { id: 0, name: '', base_url: '', api_key: '', headers: '', status: 1, weight: 1, remark: '' }
  modalOpen.value = true
}

function openEdit(channel: Channel) {
  editingChannel.value = channel
  form.value = { ...channel }
  modalOpen.value = true
}

async function saveChannel() {
  saving.value = true
  try {
    if (editingChannel.value) {
      await updateChannel(editingChannel.value.id, form.value)
      toast.success('渠道已更新')
    } else {
      await createChannel(form.value)
      toast.success('渠道已创建')
    }
    modalOpen.value = false
    await loadChannels()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '渠道保存失败')
  } finally {
    saving.value = false
  }
}

async function runTest(channel: Channel) {
  testing.value[channel.id] = true
  try {
    const response = await testChannel(channel.id)
    if (response.data.ok) {
      toast.success(`${channel.name} 测试通过`)
    } else {
      toast.error(response.data.error || `${channel.name} 测试失败`)
    }
    await loadChannels()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '渠道测试失败')
  } finally {
    testing.value[channel.id] = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return
  }
  try {
    await deleteChannel(deleteTarget.value.id)
    toast.success('渠道已删除')
    deleteTarget.value = null
    await loadChannels()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '删除渠道失败')
  }
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Channels</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">渠道管理</h2>
        <p class="mt-2 text-sm text-slate-500">维护 Sub2API 渠道、权重、状态和测试结果。</p>
      </div>
      <button class="rounded-2xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800" type="button" @click="openCreate">新增渠道</button>
    </div>

    <div v-if="loading" class="grid gap-4 lg:grid-cols-2">
      <SkeletonCard v-for="item in 4" :key="item" />
    </div>

    <EmptyState v-else-if="!channels.length" title="暂无渠道" description="新增渠道后，系统可按权重选择可用渠道。">
      <button class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-medium text-white" type="button" @click="openCreate">新增渠道</button>
    </EmptyState>

    <div v-else class="grid gap-4 xl:grid-cols-2">
      <article v-for="channel in channels" :key="channel.id" class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <span class="size-2.5 rounded-full" :class="channel.status === 1 ? 'bg-emerald-500' : 'bg-slate-300'"></span>
              <h3 class="truncate text-lg font-semibold text-slate-950">{{ channel.name }}</h3>
            </div>
            <p class="mt-1 truncate text-sm text-slate-500">{{ channel.base_url }}</p>
          </div>
          <span class="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-600">权重 {{ channel.weight }}</span>
        </div>

        <div class="mt-5 grid gap-3 sm:grid-cols-3">
          <div class="rounded-2xl bg-slate-50 p-3 text-sm">
            <p class="text-xs text-slate-500">状态</p>
            <p class="mt-1 font-medium text-slate-900">{{ channel.status === 1 ? '启用' : '禁用' }}</p>
          </div>
          <div class="rounded-2xl bg-slate-50 p-3 text-sm">
            <p class="text-xs text-slate-500">最近测试</p>
            <p class="mt-1 font-medium" :class="channel.last_test_success ? 'text-emerald-700' : 'text-slate-900'">{{ channel.last_test_success ? '正常' : '待确认' }}</p>
          </div>
          <div class="rounded-2xl bg-slate-50 p-3 text-sm">
            <p class="text-xs text-slate-500">失败率</p>
            <p class="mt-1 font-medium text-slate-900">{{ channel.recent_failure_rate ?? 0 }}%</p>
          </div>
        </div>

        <p v-if="channel.last_test_error" class="mt-4 rounded-2xl bg-rose-50 px-3 py-2 text-sm text-rose-700">{{ channel.last_test_error }}</p>
        <p v-else-if="channel.remark" class="mt-4 rounded-2xl bg-slate-50 px-3 py-2 text-sm text-slate-500">{{ channel.remark }}</p>

        <div class="mt-5 flex flex-wrap gap-2">
          <button class="channel-btn" type="button" :disabled="testing[channel.id]" @click="runTest(channel)">{{ testing[channel.id] ? '测试中' : '测试' }}</button>
          <button class="channel-btn" type="button" @click="openEdit(channel)">编辑</button>
          <button class="channel-btn text-rose-600" type="button" @click="deleteTarget = channel">删除</button>
        </div>
      </article>
    </div>

    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="modalOpen = false">
      <form class="max-h-[90vh] w-full max-w-2xl overflow-auto rounded-3xl bg-white p-5 shadow-2xl" @submit.prevent="saveChannel">
        <h3 class="text-lg font-semibold text-slate-950">{{ editingChannel ? '编辑渠道' : '新增渠道' }}</h3>
        <div class="mt-5 grid gap-4">
          <label class="block">
            <span class="text-sm font-medium text-slate-700">名称</span>
            <input v-model="form.name" class="channel-input" required />
          </label>
          <label class="block">
            <span class="text-sm font-medium text-slate-700">Base URL</span>
            <input v-model="form.base_url" class="channel-input" required placeholder="https://example.com" />
          </label>
          <label class="block">
            <span class="text-sm font-medium text-slate-700">API Key</span>
            <input v-model="form.api_key" class="channel-input" placeholder="sk-..." />
          </label>
          <label class="block">
            <span class="text-sm font-medium text-slate-700">Headers</span>
            <textarea v-model="form.headers" class="channel-input min-h-24 py-3" placeholder='{"X-Custom":"value"}'></textarea>
          </label>
          <div class="grid gap-3 sm:grid-cols-3">
            <label>
              <span class="text-sm font-medium text-slate-700">状态</span>
              <select v-model.number="form.status" class="channel-input">
                <option :value="1">启用</option>
                <option :value="2">禁用</option>
              </select>
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">权重</span>
              <input v-model.number="form.weight" class="channel-input" min="1" type="number" />
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">备注</span>
              <input v-model="form.remark" class="channel-input" />
            </label>
          </div>
        </div>
        <div class="mt-6 flex justify-end gap-2">
          <button class="rounded-xl border border-slate-200 px-4 py-2 text-sm" type="button" @click="modalOpen = false">取消</button>
          <button class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" type="submit" :disabled="saving">{{ saving ? '保存中' : '保存' }}</button>
        </div>
      </form>
    </div>

    <ConfirmDialog
      :open="!!deleteTarget"
      title="确认删除渠道"
      :message="`确认删除渠道 ${deleteTarget?.name || ''}？删除后不可恢复。`"
      confirm-text="删除"
      confirm-color="red"
      @cancel="deleteTarget = null"
      @confirm="confirmDelete"
    />
  </section>
</template>

<style scoped>
.channel-btn {
  border-radius: 0.75rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(71 85 105);
  transition: background-color 0.2s ease;
}

.channel-btn:hover:not(:disabled) {
  background: rgb(248 250 252);
}

.channel-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.channel-input {
  margin-top: 0.5rem;
  min-height: 2.75rem;
  width: 100%;
  border-radius: 1rem;
  border: 1px solid rgb(226 232 240);
  padding-left: 1rem;
  padding-right: 1rem;
  font-size: 0.875rem;
  outline: none;
}

.channel-input:focus {
  border-color: rgb(20 184 166);
  box-shadow: 0 0 0 2px rgb(20 184 166 / 0.2);
}
</style>
