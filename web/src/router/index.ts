import { createRouter, createWebHistory } from 'vue-router'

import { useUserStore } from '@/stores/user'
import Home from '@/views/Home.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: Home },
    { path: '/login', name: 'login', component: () => import('@/views/Login.vue') },
    { path: '/register', name: 'register', component: () => import('@/views/Register.vue') },
    { path: '/console/admin/login', name: 'admin-login', component: () => import('@/views/admin/AdminLogin.vue') },
    { path: '/console/admin', name: 'admin', component: () => import('@/views/admin/AdminDashboard.vue') },
    { path: '/history', name: 'history', component: () => import('@/views/History.vue') },
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
  if (to.name === 'login' && isAdmin) {
    return { name: 'admin' }
  }
  if (to.name === 'admin-login') {
    if (isAdmin) {
      return { name: 'admin' }
    }
    if (userStore.user && !isAdmin) {
      return { name: 'home' }
    }
  }
})

export default router
