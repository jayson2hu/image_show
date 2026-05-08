import api from '@/api'

export interface SiteConfig {
  site_title: string
  site_about: string
  seo_title: string
  seo_keywords: string
  seo_description: string
}

export function fetchSiteConfig() {
  return api.get<SiteConfig>('/site/config')
}
