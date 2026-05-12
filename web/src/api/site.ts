import api from '@/api'

export interface SiteConfig {
  site_title: string
  site_about: string
  seo_title: string
  seo_keywords: string
  seo_description: string
  credit_costs?: CreditCosts
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
