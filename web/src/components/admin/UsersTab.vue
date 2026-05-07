<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchUserGenerations, fetchUsers, topupCredits, updateUserRole, updateUserStatus } from '@/api/admin'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { AdminUser, Generation, Page } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const keyword = ref('')
const users = ref<Page<AdminUser>>({ items: [], total: 0, page: 1, pageSize: 20 })
const selectedUser = ref<AdminUser | null>(null)
const userGenerations = ref<Page<Generation>>({ items: [], total: 0, page: 1, pageSize: 10 })
const recordsOpen = ref(false)
const creditOpen = ref(false)
const creditAmount = ref(1)
const creditRemark = ref('')
const confirmState = ref<{ open: boolean; user: AdminUser | null; action: 'status' | 'role' | null }>({ open: false, user: null, action: null })

const totalPages = computed(() => Math.max(1, Math.ceil(users.value.total / users.value.pageSize)))

onMounted(() => loadUsers(1))

async function loadUsers(page = users.value.page) {
  loading.value = true
  try {
    const response = await fetchUsers({ keyword: keyword.value.trim() || undefined, page, pageSize: users.value.pageSize })
    users.value = response.data
  } catch (error: any) {
    toast.error(error.response?.data?.error || '用户列表加载失败')
  } finally {
    loading.value = false
  }
}

async function openRecords(user: AdminUser) {
  selectedUser.value = user
  recordsOpen.value = true
  try {
    const response = await fetchUserGenerations(user.id, { page: 1, pageSize: 10 })
    userGenerations.value = response.data
  } catch (error: any) {
    toast.error(error.response?.data?.error || '用户记录加载失败')
  }
}

function openCredit(user: AdminUser) {
  selectedUser.value = user
  creditAmount.value = 1
  creditRemark.value = ''
  creditOpen.value = true
}

async function submitCredit() {
  if (!selectedUser.value) {
    return
  }
  try {
    await topupCredits(selectedUser.value.id, { amount: creditAmount.value, remark: creditRemark.value })
    toast.success('充值成功')
    creditOpen.value = false
    await loadUsers(users.value.page)
  } catch (error: any) {
    toast.error(error.response?.data?.error || '充值失败')
  }
}

function askStatus(user: AdminUser) {
  confirmState.value = { open: true, user, action: 'status' }
}

function askRole(user: AdminUser) {
  confirmState.value = { open: true, user, action: 'role' }
}

async function confirmAction() {
  const { user, action } = confirmState.value
  if (!user || !action) {
    return
  }
  try {
    if (action === 'status') {
      await updateUserStatus(user.id, user.status === 1 ? 2 : 1)
      toast.success(user.status === 1 ? '用户已禁用' : '用户已启用')
    } else {
      await updateUserRole(user.id, user.role >= 10 ? 1 : 10)
      toast.success(user.role >= 10 ? '已设为普通用户' : '已设为管理员')
    }
    confirmState.value = { open: false, user: null, action: null }
    await loadUsers(users.value.page)
  } catch (error: any) {
    toast.error(error.response?.data?.error || '操作失败')
  }
}

function statusText(status: number) {
  return status === 1 ? '正常' : '禁用'
}

function roleText(role: number) {
  return role >= 10 ? '管理员' : '普通用户'
}

function formatTime(value?: string | null) {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Users</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">用户管理</h2>
        <p class="mt-2 text-sm text-slate-500">搜索用户、查看记录、调整角色和人工充值。</p>
      </div>
      <form class="flex gap-2" @submit.prevent="loadUsers(1)">
        <input v-model="keyword" class="min-h-11 w-64 rounded-2xl border border-slate-200 bg-white px-4 text-sm outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20" placeholder="搜索邮箱或用户名" />
        <button class="rounded-2xl bg-slate-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800" type="submit">搜索</button>
      </form>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <SkeletonCard v-for="item in 4" :key="item" />
    </div>

    <div v-else-if="!users.items.length">
      <EmptyState title="暂无用户" description="调整关键词后重新搜索，或等待新用户注册。" />
    </div>

    <div v-else class="space-y-4">
      <div class="hidden overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm md:block">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-4 py-3 font-medium">用户</th>
              <th class="px-4 py-3 font-medium">角色</th>
              <th class="px-4 py-3 font-medium">状态</th>
              <th class="px-4 py-3 font-medium">积分</th>
              <th class="px-4 py-3 font-medium">最近登录</th>
              <th class="px-4 py-3 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="user in users.items" :key="user.id">
              <td class="px-4 py-3">
                <p class="font-medium text-slate-900">{{ user.username || user.email }}</p>
                <p class="text-xs text-slate-500">{{ user.email }}</p>
              </td>
              <td class="px-4 py-3">{{ roleText(user.role) }}</td>
              <td class="px-4 py-3">
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="user.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'">{{ statusText(user.status) }}</span>
              </td>
              <td class="px-4 py-3 font-semibold">{{ user.credits }}</td>
              <td class="px-4 py-3 text-slate-500">{{ formatTime(user.last_login_at) }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-2">
                  <button class="admin-mini-btn" type="button" @click="openRecords(user)">记录</button>
                  <button class="admin-mini-btn" type="button" @click="openCredit(user)">充值</button>
                  <button class="admin-mini-btn" type="button" @click="askRole(user)">{{ user.role >= 10 ? '降级' : '设管理员' }}</button>
                  <button class="admin-mini-btn text-rose-600" type="button" @click="askStatus(user)">{{ user.status === 1 ? '禁用' : '启用' }}</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="grid gap-3 md:hidden">
        <article v-for="user in users.items" :key="user.id" class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="truncate font-semibold text-slate-950">{{ user.username || user.email }}</p>
              <p class="mt-1 truncate text-sm text-slate-500">{{ user.email }}</p>
            </div>
            <span class="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-600">{{ roleText(user.role) }}</span>
          </div>
          <div class="mt-4 grid grid-cols-2 gap-2 text-sm">
            <div class="rounded-2xl bg-slate-50 p-3">积分：{{ user.credits }}</div>
            <div class="rounded-2xl bg-slate-50 p-3">状态：{{ statusText(user.status) }}</div>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <button class="admin-mini-btn" type="button" @click="openRecords(user)">记录</button>
            <button class="admin-mini-btn" type="button" @click="openCredit(user)">充值</button>
            <button class="admin-mini-btn" type="button" @click="askRole(user)">{{ user.role >= 10 ? '降级' : '设管理员' }}</button>
            <button class="admin-mini-btn text-rose-600" type="button" @click="askStatus(user)">{{ user.status === 1 ? '禁用' : '启用' }}</button>
          </div>
        </article>
      </div>

      <Pagination :page="users.page" :page-size="users.pageSize" :total="users.total" @update:page="loadUsers" />
      <p class="text-center text-xs text-slate-400">共 {{ users.total }} 个用户，{{ totalPages }} 页</p>
    </div>

    <div v-if="recordsOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="recordsOpen = false">
      <div class="max-h-[85vh] w-full max-w-3xl overflow-auto rounded-3xl bg-white p-5 shadow-2xl">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-semibold text-slate-950">用户记录</h3>
            <p class="text-sm text-slate-500">{{ selectedUser?.email }}</p>
          </div>
          <button class="rounded-xl border border-slate-200 px-3 py-2 text-sm" type="button" @click="recordsOpen = false">关闭</button>
        </div>
        <div v-if="!userGenerations.items.length" class="rounded-2xl bg-slate-50 p-6 text-center text-sm text-slate-500">暂无生成记录</div>
        <div v-else class="space-y-2">
          <div v-for="item in userGenerations.items" :key="item.id" class="rounded-2xl bg-slate-50 p-3 text-sm">
            <p class="line-clamp-2 text-slate-800">{{ item.prompt }}</p>
            <p class="mt-1 text-xs text-slate-500">{{ item.size }} · {{ formatTime(item.created_at) }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="creditOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="creditOpen = false">
      <form class="w-full max-w-md rounded-3xl bg-white p-5 shadow-2xl" @submit.prevent="submitCredit">
        <h3 class="text-lg font-semibold text-slate-950">人工充值</h3>
        <p class="mt-1 text-sm text-slate-500">{{ selectedUser?.email }}</p>
        <label class="mt-5 block">
          <span class="text-sm font-medium text-slate-700">积分数量</span>
          <input v-model.number="creditAmount" class="mt-2 min-h-11 w-full rounded-2xl border border-slate-200 px-4 text-sm outline-none focus:border-teal focus:ring-2 focus:ring-teal/20" min="0.01" step="0.01" type="number" />
        </label>
        <label class="mt-4 block">
          <span class="text-sm font-medium text-slate-700">备注</span>
          <input v-model="creditRemark" class="mt-2 min-h-11 w-full rounded-2xl border border-slate-200 px-4 text-sm outline-none focus:border-teal focus:ring-2 focus:ring-teal/20" placeholder="人工充值原因" />
        </label>
        <div class="mt-6 flex justify-end gap-2">
          <button class="rounded-xl border border-slate-200 px-4 py-2 text-sm" type="button" @click="creditOpen = false">取消</button>
          <button class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-semibold text-white" type="submit">确认充值</button>
        </div>
      </form>
    </div>

    <ConfirmDialog
      :open="confirmState.open"
      :title="confirmState.action === 'role' ? '确认调整角色' : '确认调整状态'"
      :message="confirmState.action === 'role' ? '确认修改该用户角色？' : '确认修改该用户状态？'"
      confirm-text="确认"
      :confirm-color="confirmState.action === 'status' ? 'red' : 'blue'"
      @cancel="confirmState = { open: false, user: null, action: null }"
      @confirm="confirmAction"
    />
  </section>
</template>

<style scoped>
.admin-mini-btn {
  border-radius: 0.75rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.375rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: rgb(71 85 105);
  transition: background-color 0.2s ease;
}

.admin-mini-btn:hover {
  background: rgb(248 250 252);
}
</style>
