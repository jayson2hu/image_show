<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await userStore.login(email.value, password.value)
    await router.push('/')
  } catch {
    error.value = '邮箱或密码不正确'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-md rounded border border-slate-200 bg-white p-6">
    <h1 class="text-xl font-semibold">登录</h1>
    <form class="mt-6 space-y-4" @submit.prevent="submit">
      <label class="block text-sm font-medium">
        邮箱
        <input v-model="email" class="mt-1 w-full rounded border border-slate-300 px-3 py-2" type="email" autocomplete="email" required />
      </label>
      <label class="block text-sm font-medium">
        密码
        <input
          v-model="password"
          class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
          type="password"
          autocomplete="current-password"
          required
        />
      </label>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="w-full rounded bg-teal px-4 py-2 text-white disabled:opacity-60" type="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>
  </section>
</template>
