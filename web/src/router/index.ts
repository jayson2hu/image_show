import { createRouter, createWebHistory } from 'vue-router'

import { useUserStore } from '@/stores/user'
import Home from '@/views/Home.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'chat', component: () => import('@/views/Chat.vue') },
    { path: '/chat', redirect: '/' },
    { path: '/classic', name: 'classic', component: Home },
    { path: '/login', name: 'login', component: () => import('@/views/Login.vue') },
    { path: '/register', name: 'register', redirect: '/login' },
    { path: '/console/admin/login', name: 'admin-login', component: () => import('@/views/admin/AdminLogin.vue') },
    { path: '/console/admin', name: 'admin', component: () => import('@/views/admin/AdminDashboard.vue') },
    { path: '/account', name: 'account', component: () => import('@/views/Account.vue') },
    { path: '/history', name: 'history', component: () => import('@/views/History.vue') },
    { path: '/credits', name: 'credits', component: () => import('@/views/Credits.vue') },
    { path: '/packages', name: 'packages', component: () => import('@/views/Packages.vue') },
  ],
})

router.beforeEach(async (to) => {
  const userStore = useUserStore()
  if (userStore.token && !userStore.user) {
    await userStore.fetchUser()
  }

  const isAdmin = (userStore.user?.role || 0) >= 10
  if (to.name === 'admin' && !isAdmin) {
    return { name: 'admin-login' }
  }
  if (to.name === 'admin-login') {
    if (isAdmin) {
      return { name: 'admin' }
    }
  }
  if ((to.name === 'credits' || to.name === 'account') && !userStore.token) {
    return { name: 'login' }
  }
})

export default router
