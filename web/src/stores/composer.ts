import { defineStore } from 'pinia'

type Draft = {
  prompt: string
  size: string
  quality: string
  style_id: string
  scene_id: string
  layered: boolean
  layer_count: number
  attachment: File | null
}

const defaultDraft = (): Draft => ({
  prompt: '',
  size: 'square',
  quality: 'medium',
  style_id: '',
  scene_id: '',
  layered: false,
  layer_count: 5,
  attachment: null,
})

export const useComposerStore = defineStore('composer', {
  state: () => ({
    draft: defaultDraft(),
    focusTick: 0,
  }),
  actions: {
    setDraft(patch: Partial<Draft>) {
      Object.assign(this.draft, patch)
    },
    focusInput() {
      this.focusTick += 1
    },
    reset() {
      this.draft = defaultDraft()
      this.focusInput()
    },
  },
})
