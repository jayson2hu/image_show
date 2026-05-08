export interface Page<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface AdminUser {
  id: number
  username: string
  email: string
  role: number
  status: number
  credits: number
  created_at: string
  last_login_at?: string | null
}

export interface CreditLog {
  id: number
  user_id: number
  type: number
  amount: number
  balance: number
  remark: string
  created_at: string
}

export interface Generation {
  id: number
  prompt: string
  quality: string
  size: string
  status: number
  image_url: string
  created_at: string
}

export interface FailureReasonSummary {
  category: string
  label: string
  count: number
}

export interface RecentFailure {
  id: number
  user_id?: number | null
  prompt: string
  size: string
  error: string
  category: string
  label: string
  created_at: string
}

export interface PromptTemplate {
  id: number
  category: TemplateCategory | string
  label: string
  prompt: string
  sort_order: number
  status: number
}

export interface CreditPackage {
  id: number
  name: string
  credits: number
  price: number
  valid_days: number
  sort_order: number
  status: number
  created_at?: string
  updated_at?: string
}

export interface Channel {
  id: number
  name: string
  base_url: string
  api_key?: string
  headers?: string
  status: number
  weight: number
  remark?: string
  last_test_at?: string | null
  last_test_success?: boolean
  last_test_status?: number
  last_test_error?: string
  recent_success_count?: number
  recent_failed_count?: number
  recent_failure_rate?: number
}

export interface Announcement {
  id: number
  title: string
  content: string
  status: number
  notify_mode: 'silent' | 'popup'
  target: 'all' | 'guest' | 'user' | 'admin'
  sort_order: number
  starts_at?: string | null
  ends_at?: string | null
  read_count?: number
  created_at: string
  updated_at: string
}

export interface AnnouncementRead {
  user_id: number
  email: string
  username: string
  role: number
  read_at: string
}

export interface MonitorSummary {
  date: string
  generation_count: number
  completed_count: number
  failed_count: number
  failure_rate: number
  credits_consumed: number
  new_users: number
  paid_order_count: number
  paid_order_amount: number
  alert_threshold: number
  alert_triggered: boolean
  failure_reasons?: FailureReasonSummary[]
  recent_failures?: RecentFailure[]
}

export interface CreditForm {
  amount: number
  remark: string
}

export interface AdminUserForm {
  email: string
  username: string
  password: string
  role: number
  status: number
  credits: number
}

export type TemplateCategory = 'style' | 'sample' | 'default' | 'repair'
