<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import api from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const email = ref('')
const code = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const sending = ref(false)
const cooldown = ref(0)

const canSendCode = computed(() => email.value && cooldown.value === 0 && !sending.value)

async function sendCode() {
  if (!canSendCode.value) {
    return
  }
  error.value = ''
  sending.value = true
  try {
    await api.post('/auth/send-code', { email: email.value })
    cooldown.value = 60
    const timer = window.setInterval(() => {
      cooldown.value -= 1
      if (cooldown.value <= 0) {
        window.clearInterval(timer)
      }
    }, 1000)
  } catch {
    error.value = '验证码发送失败或过于频繁'
  } finally {
    sending.value = false
  }
}

async function submit() {
  error.value = ''
  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  loading.value = true
  try {
    await userStore.register(email.value, password.value, code.value)
    await router.push('/')
  } catch {
    error.value = '注册失败，请检查验证码或稍后再试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-md rounded border border-slate-200 bg-white p-6">
    <h1 class="text-xl font-semibold">注册</h1>
    <form class="mt-6 space-y-4" @submit.prevent="submit">
      <label class="block text-sm font-medium">
        邮箱
        <input v-model="email" class="mt-1 w-full rounded border border-slate-300 px-3 py-2" type="email" autocomplete="email" required />
      </label>
      <label class="block text-sm font-medium">
        验证码
        <div class="mt-1 flex gap-2">
          <input v-model="code" class="min-w-0 flex-1 rounded border border-slate-300 px-3 py-2" type="text" inputmode="numeric" required />
          <button class="rounded border border-slate-300 px-3 py-2 text-sm disabled:opacity-60" type="button" :disabled="!canSendCode" @click="sendCode">
            {{ cooldown > 0 ? `${cooldown}s` : '发送' }}
          </button>
        </div>
      </label>
      <label class="block text-sm font-medium">
        密码
        <input
          v-model="password"
          class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
          type="password"
          autocomplete="new-password"
          minlength="8"
          required
        />
      </label>
      <label class="block text-sm font-medium">
        确认密码
        <input
          v-model="confirmPassword"
          class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
          type="password"
          autocomplete="new-password"
          minlength="8"
          required
        />
      </label>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="w-full rounded bg-coral px-4 py-2 text-white disabled:opacity-60" type="submit" :disabled="loading">
        {{ loading ? '注册中...' : '注册' }}
      </button>
    </form>
  </section>
</template>
