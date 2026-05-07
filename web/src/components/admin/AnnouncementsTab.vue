<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { createAnnouncement, deleteAnnouncement, fetchAnnouncements, updateAnnouncement } from '@/api/admin'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { Announcement } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const saving = ref(false)
const announcements = ref<Announcement[]>([])
const modalOpen = ref(false)
const editing = ref<Announcement | null>(null)
const deleteTarget = ref<Announcement | null>(null)
const form = ref<Announcement>({ id: 0, title: '', content: '', notify_mode: 'silent', target: 'all', sort_order: 0, status: 1, starts_at: '', ends_at: '', created_at: '', updated_at: '' })

onMounted(loadAnnouncements)

async function loadAnnouncements() {
  loading.value = true
  try {
    const response = await fetchAnnouncements()
    announcements.value = response.data.items || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || '公告加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.value = { id: 0, title: '', content: '', notify_mode: 'silent', target: 'all', sort_order: 0, status: 1, starts_at: '', ends_at: '', created_at: '', updated_at: '' }
  modalOpen.value = true
}

function openEdit(item: Announcement) {
  editing.value = item
  form.value = { ...item, starts_at: toInputDate(item.starts_at), ends_at: toInputDate(item.ends_at) }
  modalOpen.value = true
}

async function saveAnnouncement() {
  saving.value = true
  try {
    const payload = {
      ...form.value,
      starts_at: toRFC3339(form.value.starts_at),
      ends_at: toRFC3339(form.value.ends_at),
    }
    if (editing.value) {
      await updateAnnouncement(editing.value.id, payload)
      toast.success('公告已更新')
    } else {
      await createAnnouncement(payload)
      toast.success('公告已发布')
    }
    modalOpen.value = false
    await loadAnnouncements()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '公告保存失败')
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return
  }
  try {
    await deleteAnnouncement(deleteTarget.value.id)
    toast.success('公告已删除')
    deleteTarget.value = null
    await loadAnnouncements()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '公告删除失败')
  }
}

function toInputDate(value?: string | null) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
}

function toRFC3339(value?: string | null) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toISOString()
}

function targetText(target: string) {
  const map: Record<string, string> = { all: '全部', guest: '游客', user: '用户', admin: '管理员' }
  return map[target] || target
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Announcements</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">公告通知</h2>
        <p class="mt-2 text-sm text-slate-500">发布前台通知，支持静默公告和弹窗公告。</p>
      </div>
      <button class="rounded-2xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800" type="button" @click="openCreate">发布公告</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <SkeletonCard v-for="item in 4" :key="item" />
    </div>

    <EmptyState v-else-if="!announcements.length" title="暂无公告" description="发布后会在前台生成页展示给用户。" />

    <div v-else class="grid gap-4 xl:grid-cols-2">
      <article v-for="item in announcements" :key="item.id" class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="truncate text-lg font-semibold text-slate-950">{{ item.title }}</h3>
              <span class="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-600">{{ targetText(item.target) }}</span>
              <span class="rounded-full px-2.5 py-1 text-xs" :class="item.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'">{{ item.status === 1 ? '启用' : '禁用' }}</span>
            </div>
            <p class="mt-2 line-clamp-3 text-sm leading-6 text-slate-500">{{ item.content }}</p>
          </div>
          <span class="shrink-0 rounded-full bg-slate-50 px-2.5 py-1 text-xs text-slate-500">已读 {{ item.read_count || 0 }}</span>
        </div>
        <div class="mt-5 flex gap-2">
          <button class="announcement-btn" type="button" @click="openEdit(item)">编辑</button>
          <button class="announcement-btn text-rose-600" type="button" @click="deleteTarget = item">删除</button>
        </div>
      </article>
    </div>

    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="modalOpen = false">
      <form class="max-h-[90vh] w-full max-w-2xl overflow-auto rounded-3xl bg-white p-5 shadow-2xl" @submit.prevent="saveAnnouncement">
        <h3 class="text-lg font-semibold text-slate-950">{{ editing ? '编辑公告' : '发布公告' }}</h3>
        <div class="mt-5 grid gap-4">
          <label>
            <span class="text-sm font-medium text-slate-700">标题</span>
            <input v-model="form.title" class="announcement-input" required />
          </label>
          <label>
            <span class="text-sm font-medium text-slate-700">内容</span>
            <textarea v-model="form.content" class="announcement-input min-h-36 py-3" required></textarea>
          </label>
          <div class="grid gap-3 sm:grid-cols-4">
            <label>
              <span class="text-sm font-medium text-slate-700">通知方式</span>
              <select v-model="form.notify_mode" class="announcement-input">
                <option value="silent">静默</option>
                <option value="popup">弹窗</option>
              </select>
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">目标</span>
              <select v-model="form.target" class="announcement-input">
                <option value="all">全部</option>
                <option value="guest">游客</option>
                <option value="user">用户</option>
                <option value="admin">管理员</option>
              </select>
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">状态</span>
              <select v-model.number="form.status" class="announcement-input">
                <option :value="1">启用</option>
                <option :value="2">禁用</option>
              </select>
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">排序</span>
              <input v-model.number="form.sort_order" class="announcement-input" type="number" />
            </label>
          </div>
          <div class="grid gap-3 sm:grid-cols-2">
            <label>
              <span class="text-sm font-medium text-slate-700">开始时间</span>
              <input v-model="form.starts_at" class="announcement-input" type="datetime-local" />
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">结束时间</span>
              <input v-model="form.ends_at" class="announcement-input" type="datetime-local" />
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
      title="确认删除公告"
      :message="`确认删除公告 ${deleteTarget?.title || ''}？`"
      confirm-text="删除"
      confirm-color="red"
      @cancel="deleteTarget = null"
      @confirm="confirmDelete"
    />
  </section>
</template>

<style scoped>
.announcement-btn {
  border-radius: 0.75rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(71 85 105);
  transition: background-color 0.2s ease;
}

.announcement-btn:hover {
  background: rgb(248 250 252);
}

.announcement-input {
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

.announcement-input:focus {
  border-color: rgb(20 184 166);
  box-shadow: 0 0 0 2px rgb(20 184 166 / 0.2);
}
</style>
