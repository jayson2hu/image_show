<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { createPackage, deletePackage, fetchAdminPackages, updatePackage } from '@/api/admin'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { CreditPackage } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const saving = ref(false)
const packages = ref<CreditPackage[]>([])
const modalOpen = ref(false)
const editingPackage = ref<CreditPackage | null>(null)
const deleteTarget = ref<CreditPackage | null>(null)
const form = ref<CreditPackage>({
  id: 0,
  name: '',
  credits: 10,
  price: 9.9,
  valid_days: 30,
  sort_order: 1,
  status: 1,
})

onMounted(loadPackages)

async function loadPackages() {
  loading.value = true
  try {
    const response = await fetchAdminPackages()
    packages.value = response.data.items || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || '套餐加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingPackage.value = null
  form.value = {
    id: 0,
    name: '',
    credits: 10,
    price: 9.9,
    valid_days: 30,
    sort_order: packages.value.length + 1,
    status: 1,
  }
  modalOpen.value = true
}

function openEdit(item: CreditPackage) {
  editingPackage.value = item
  form.value = { ...item }
  modalOpen.value = true
}

async function savePackage() {
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      credits: Number(form.value.credits),
      price: Number(form.value.price),
      valid_days: Number(form.value.valid_days),
      sort_order: Number(form.value.sort_order),
      status: Number(form.value.status),
    }
    if (editingPackage.value) {
      await updatePackage(editingPackage.value.id, payload)
      toast.success('套餐已更新')
    } else {
      await createPackage(payload)
      toast.success('套餐已创建')
    }
    modalOpen.value = false
    await loadPackages()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '套餐保存失败')
  } finally {
    saving.value = false
  }
}

async function togglePackage(item: CreditPackage) {
  try {
    await updatePackage(item.id, { ...item, status: item.status === 1 ? 2 : 1 })
    toast.success(item.status === 1 ? '套餐已停用' : '套餐已启用')
    await loadPackages()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '套餐状态更新失败')
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return
  }
  try {
    await deletePackage(deleteTarget.value.id)
    toast.success('套餐已删除')
    deleteTarget.value = null
    await loadPackages()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '删除套餐失败')
  }
}

function standardImageCount(item: CreditPackage) {
  return Math.floor(item.credits)
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Packages</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">套餐管理</h2>
        <p class="mt-2 text-sm text-slate-500">维护前台可购买的积分套餐、价格、有效期和展示顺序。</p>
      </div>
      <button class="rounded-2xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800" type="button" @click="openCreate">新增套餐</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <SkeletonCard v-for="item in 3" :key="item" />
    </div>

    <EmptyState v-else-if="!packages.length" title="暂无套餐" description="新增套餐后，用户可以在购买中心选择对应积分包。">
      <button class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-medium text-white" type="button" @click="openCreate">新增套餐</button>
    </EmptyState>

    <div v-else class="grid gap-4 xl:grid-cols-3">
      <article v-for="item in packages" :key="item.id" class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-slate-950">{{ item.name }}</h3>
            <p class="mt-1 text-sm text-slate-500">排序 {{ item.sort_order }} · 有效期 {{ item.valid_days }} 天</p>
          </div>
          <span class="rounded-full px-2.5 py-1 text-xs" :class="item.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'">{{ item.status === 1 ? '启用' : '停用' }}</span>
        </div>

        <div class="mt-5 grid grid-cols-2 gap-3">
          <div class="rounded-2xl bg-slate-50 p-3">
            <p class="text-xs text-slate-500">售价</p>
            <p class="mt-1 text-xl font-semibold text-slate-950">¥{{ item.price }}</p>
          </div>
          <div class="rounded-2xl bg-slate-50 p-3">
            <p class="text-xs text-slate-500">积分</p>
            <p class="mt-1 text-xl font-semibold text-slate-950">{{ item.credits }}</p>
          </div>
        </div>

        <p class="mt-4 rounded-2xl bg-teal/10 px-3 py-2 text-sm text-teal">约可生成 {{ standardImageCount(item) }} 张标准图</p>

        <div class="mt-5 flex flex-wrap gap-2">
          <button class="package-btn" type="button" @click="openEdit(item)">编辑</button>
          <button class="package-btn" type="button" @click="togglePackage(item)">{{ item.status === 1 ? '停用' : '启用' }}</button>
          <button class="package-btn text-rose-600" type="button" @click="deleteTarget = item">删除</button>
        </div>
      </article>
    </div>

    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="modalOpen = false">
      <form class="w-full max-w-2xl rounded-3xl bg-white p-5 shadow-2xl" @submit.prevent="savePackage">
        <h3 class="text-lg font-semibold text-slate-950">{{ editingPackage ? '编辑套餐' : '新增套餐' }}</h3>
        <div class="mt-5 grid gap-4">
          <label>
            <span class="text-sm font-medium text-slate-700">套餐名称</span>
            <input v-model="form.name" class="package-input" placeholder="例如：标准包" required />
          </label>
          <div class="grid gap-3 sm:grid-cols-3">
            <label>
              <span class="text-sm font-medium text-slate-700">积分数量</span>
              <input v-model.number="form.credits" class="package-input" min="0.01" step="0.01" type="number" required />
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">售价</span>
              <input v-model.number="form.price" class="package-input" min="0.01" step="0.01" type="number" required />
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">有效期天数</span>
              <input v-model.number="form.valid_days" class="package-input" min="1" type="number" required />
            </label>
          </div>
          <div class="grid gap-3 sm:grid-cols-2">
            <label>
              <span class="text-sm font-medium text-slate-700">排序</span>
              <input v-model.number="form.sort_order" class="package-input" type="number" />
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">状态</span>
              <select v-model.number="form.status" class="package-input">
                <option :value="1">启用</option>
                <option :value="2">停用</option>
              </select>
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
      title="确认删除套餐"
      :message="`确认删除套餐 ${deleteTarget?.name || ''}？删除后不可恢复。`"
      confirm-text="删除"
      confirm-color="red"
      @cancel="deleteTarget = null"
      @confirm="confirmDelete"
    />
  </section>
</template>

<style scoped>
.package-btn {
  border-radius: 0.75rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(71 85 105);
  transition: background-color 0.2s ease;
}

.package-btn:hover {
  background: rgb(248 250 252);
}

.package-input {
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

.package-input:focus {
  border-color: rgb(20 184 166);
  box-shadow: 0 0 0 2px rgb(20 184 166 / 0.2);
}
</style>
