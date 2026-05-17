import api from '@/api'

export interface SiteConfig {
  site_title: string
  site_about: string
  seo_title: string
  seo_keywords: string
  seo_description: string
  register_enabled?: boolean
  credit_costs?: CreditCosts
  greeting_text?: string
  guest_free_credits?: number
  guest_generation_limit?: number
  guest_layered_generation_limit?: number
  user_generation_limit?: number
  user_layered_generation_limit?: number
}

export interface CreditCosts {
  square: number
  portrait: number
  story: number
  landscape: number
  widescreen: number
}

export function fetchSiteConfig() {
  return api.get<SiteConfig>('/site/config')
}
