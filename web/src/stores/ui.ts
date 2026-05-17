import { defineStore } from 'pinia'

const SIDEBAR_COLLAPSED_KEY = 'sidebar_collapsed'

function readSidebarCollapsed() {
  if (typeof window === 'undefined') return false
  return window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true'
}

export const useUiStore = defineStore('ui', {
  state: () => ({
    sidebarCollapsed: readSidebarCollapsed(),
  }),
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(this.sidebarCollapsed))
    },
  },
})
