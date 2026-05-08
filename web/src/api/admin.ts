import api from '@/api'
import type {
  AdminUser,
  AdminUserForm,
  Announcement,
  AnnouncementRead,
  Channel,
  CreditPackage,
  CreditForm,
  CreditLog,
  Generation,
  MonitorSummary,
  Page,
  PromptTemplate,
} from '@/types/admin'

export function fetchUsers(params: { keyword?: string; page?: number; pageSize?: number }) {
  return api.get<Page<AdminUser>>('/admin/users', { params })
}

export function createUser(payload: AdminUserForm) {
  return api.post<AdminUser>('/admin/users', payload)
}

export function fetchUserGenerations(userId: number, params: { page?: number; pageSize?: number }) {
  return api.get<Page<Generation>>(`/admin/users/${userId}/generations`, { params })
}

export function updateUserStatus(userId: number, status: number) {
  return api.put(`/admin/users/${userId}/status`, { status })
}

export function updateUserRole(userId: number, role: number) {
  return api.put(`/admin/users/${userId}/role`, { role })
}

export function topupCredits(userId: number, form: CreditForm) {
  return api.post(`/admin/users/${userId}/credits`, form)
}

export function fetchCreditLogs(params: { user_id?: number; page?: number; pageSize?: number }) {
  return api.get<Page<CreditLog>>('/admin/credits/logs', { params })
}

export function fetchTemplates() {
  return api.get<{ items: PromptTemplate[] }>('/admin/prompt-templates')
}

export function createTemplate(payload: Omit<PromptTemplate, 'id'>) {
  return api.post('/admin/prompt-templates', payload)
}

export function updateTemplate(id: number, payload: Partial<PromptTemplate>) {
  return api.put(`/admin/prompt-templates/${id}`, payload)
}

export function deleteTemplate(id: number) {
  return api.delete(`/admin/prompt-templates/${id}`)
}

export function fetchChannels() {
  return api.get<{ items: Channel[] }>('/admin/channels')
}

export function createChannel(payload: Omit<Channel, 'id'>) {
  return api.post('/admin/channels', payload)
}

export function updateChannel(id: number, payload: Partial<Channel>) {
  return api.put(`/admin/channels/${id}`, payload)
}

export function deleteChannel(id: number) {
  return api.delete(`/admin/channels/${id}`)
}

export function testChannel(id: number) {
  return api.post<{ ok: boolean; status?: number; error?: string }>(`/admin/channels/${id}/test`)
}

export function fetchSettings() {
  return api.get<{ items: Record<string, string> }>('/admin/settings')
}

export function saveSettings(items: Record<string, string>) {
  return api.put('/admin/settings', { items })
}

export function fetchMonitorSummary() {
  return api.get<MonitorSummary>('/admin/monitor/summary')
}

export function triggerMonitorCheck() {
  return api.post<{ sent: boolean }>('/admin/monitor/check')
}

export function fetchAnnouncements() {
  return api.get<{ items: Announcement[] }>('/admin/announcements')
}

export function createAnnouncement(payload: Omit<Announcement, 'id' | 'created_at' | 'updated_at'>) {
  return api.post('/admin/announcements', payload)
}

export function updateAnnouncement(id: number, payload: Partial<Announcement>) {
  return api.put(`/admin/announcements/${id}`, payload)
}

export function deleteAnnouncement(id: number) {
  return api.delete(`/admin/announcements/${id}`)
}

export function fetchAnnouncementReads(id: number) {
  return api.get<{ items: AnnouncementRead[] }>(`/admin/announcements/${id}/reads`)
}

export function fetchAdminGenerations(params: { user_id?: number; status?: number; page?: number; pageSize?: number }) {
  return api.get<Page<Generation>>('/admin/generations', { params })
}

export function deleteAdminGenerations(ids: number[], deleteR2 = false) {
  return api.delete('/admin/generations/batch', { data: { ids, delete_r2: deleteR2 } })
}

export function fetchAdminPackages() {
  return api.get<{ items: CreditPackage[] }>('/admin/packages')
}

export function createPackage(payload: Omit<CreditPackage, 'id' | 'created_at' | 'updated_at'>) {
  return api.post<CreditPackage>('/admin/packages', payload)
}

export function updatePackage(id: number, payload: Partial<CreditPackage>) {
  return api.put(`/admin/packages/${id}`, payload)
}

export function deletePackage(id: number) {
  return api.delete(`/admin/packages/${id}`)
}
