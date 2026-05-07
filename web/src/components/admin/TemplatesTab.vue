<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { createTemplate, deleteTemplate, fetchTemplates, updateTemplate } from '@/api/admin'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { useToast } from '@/composables/useToast'
import type { PromptTemplate } from '@/types/admin'

const toast = useToast()
const loading = ref(false)
const saving = ref(false)
const templates = ref<PromptTemplate[]>([])
const category = ref('')
const modalOpen = ref(false)
const editingTemplate = ref<PromptTemplate | null>(null)
const deleteTarget = ref<PromptTemplate | null>(null)
const form = ref<PromptTemplate>({ id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 })

const categories = [
  { value: '', label: '全部模板' },
  { value: 'style', label: '风格' },
  { value: 'sample', label: '样例' },
  { value: 'default', label: '默认' },
  { value: 'repair', label: '修复' },
]

const filteredTemplates = computed(() => (category.value ? templates.value.filter((item) => item.category === category.value) : templates.value))

onMounted(loadTemplates)

async function loadTemplates() {
  loading.value = true
  try {
    const response = await fetchTemplates()
    templates.value = response.data.items || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || '模板加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingTemplate.value = null
  form.value = { id: 0, category: 'style', label: '', prompt: '', sort_order: 0, status: 1 }
  modalOpen.value = true
}

function openEdit(item: PromptTemplate) {
  editingTemplate.value = item
  form.value = { ...item }
  modalOpen.value = true
}

async function saveTemplate() {
  saving.value = true
  try {
    if (editingTemplate.value) {
      await updateTemplate(editingTemplate.value.id, form.value)
      toast.success('模板已更新')
    } else {
      await createTemplate(form.value)
      toast.success('模板已创建')
    }
    modalOpen.value = false
    await loadTemplates()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '模板保存失败')
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return
  }
  try {
    await deleteTemplate(deleteTarget.value.id)
    toast.success('模板已删除')
    deleteTarget.value = null
    await loadTemplates()
  } catch (error: any) {
    toast.error(error.response?.data?.error || '删除模板失败')
  }
}

function categoryLabel(value: string) {
  return categories.find((item) => item.value === value)?.label || value
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-teal">Templates</p>
        <h2 class="mt-1 text-2xl font-semibold text-slate-950">提示词模板</h2>
        <p class="mt-2 text-sm text-slate-500">维护前台风格、样例和默认提示词。</p>
      </div>
      <div class="flex gap-2">
        <select v-model="category" class="rounded-2xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-teal focus:ring-2 focus:ring-teal/20">
          <option v-for="item in categories" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
        <button class="rounded-2xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800" type="button" @click="openCreate">新增模板</button>
      </div>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <SkeletonCard v-for="item in 4" :key="item" />
    </div>

    <EmptyState v-else-if="!filteredTemplates.length" title="暂无模板" description="新增模板后可在前台展示给用户选择。" />

    <div v-else class="grid gap-4 xl:grid-cols-2">
      <article v-for="item in filteredTemplates" :key="item.id" class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="truncate text-lg font-semibold text-slate-950">{{ item.label }}</h3>
              <span class="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-600">{{ categoryLabel(item.category) }}</span>
              <span class="rounded-full px-2.5 py-1 text-xs" :class="item.status === 1 ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'">{{ item.status === 1 ? '启用' : '禁用' }}</span>
            </div>
            <p class="mt-2 line-clamp-3 text-sm leading-6 text-slate-500">{{ item.prompt }}</p>
          </div>
          <span class="shrink-0 rounded-full bg-slate-50 px-2.5 py-1 text-xs text-slate-500">排序 {{ item.sort_order }}</span>
        </div>
        <div class="mt-5 flex gap-2">
          <button class="template-btn" type="button" @click="openEdit(item)">编辑</button>
          <button class="template-btn text-rose-600" type="button" @click="deleteTarget = item">删除</button>
        </div>
      </article>
    </div>

    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="modalOpen = false">
      <form class="max-h-[90vh] w-full max-w-2xl overflow-auto rounded-3xl bg-white p-5 shadow-2xl" @submit.prevent="saveTemplate">
        <h3 class="text-lg font-semibold text-slate-950">{{ editingTemplate ? '编辑模板' : '新增模板' }}</h3>
        <div class="mt-5 grid gap-4">
          <div class="grid gap-3 sm:grid-cols-3">
            <label>
              <span class="text-sm font-medium text-slate-700">分类</span>
              <select v-model="form.category" class="template-input">
                <option value="style">风格</option>
                <option value="sample">样例</option>
                <option value="default">默认</option>
                <option value="repair">修复</option>
              </select>
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">排序</span>
              <input v-model.number="form.sort_order" class="template-input" type="number" />
            </label>
            <label>
              <span class="text-sm font-medium text-slate-700">状态</span>
              <select v-model.number="form.status" class="template-input">
                <option :value="1">启用</option>
                <option :value="2">禁用</option>
              </select>
            </label>
          </div>
          <label>
            <span class="text-sm font-medium text-slate-700">名称</span>
            <input v-model="form.label" class="template-input" required />
          </label>
          <label>
            <span class="text-sm font-medium text-slate-700">提示词</span>
            <textarea v-model="form.prompt" class="template-input min-h-36 py-3" required></textarea>
          </label>
        </div>
        <div class="mt-6 flex justify-end gap-2">
          <button class="rounded-xl border border-slate-200 px-4 py-2 text-sm" type="button" @click="modalOpen = false">取消</button>
          <button class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" type="submit" :disabled="saving">{{ saving ? '保存中' : '保存' }}</button>
        </div>
      </form>
    </div>

    <ConfirmDialog
      :open="!!deleteTarget"
      title="确认删除模板"
      :message="`确认删除模板 ${deleteTarget?.label || ''}？`"
      confirm-text="删除"
      confirm-color="red"
      @cancel="deleteTarget = null"
      @confirm="confirmDelete"
    />
  </section>
</template>

<style scoped>
.template-btn {
  border-radius: 0.75rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(71 85 105);
  transition: background-color 0.2s ease;
}

.template-btn:hover {
  background: rgb(248 250 252);
}

.template-input {
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

.template-input:focus {
  border-color: rgb(20 184 166);
  box-shadow: 0 0 0 2px rgb(20 184 166 / 0.2);
}
</style>
