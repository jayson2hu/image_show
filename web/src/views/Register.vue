<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

onMounted(() => {
  router.replace('/login')
})
</script>

<template>
  <div class="py-12 text-center text-sm text-slate-500">正在跳转到登录页...</div>
</template>
