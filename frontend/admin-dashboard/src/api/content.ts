import request from './request'

export type PageVO<T> = {
  items: T[]
  total: number
  page: number
  page_size: number
}

export type FounderProfileItemVO = {
  id: number
  user_id: number
  slug: string
  display_name: string
  title?: string
  bio?: string
  bio_long?: string
  avatar_url?: string
  cover_url?: string
  location?: string
  public_email?: string
  tags?: { id: number; name: string; type: string }[]
  custom_links?: { id?: number; link_type: string; name: string; url: string }[]
  favorite_count: number
  view_count: number
  is_visible: boolean
  created_at?: string
}

export type ProductVO = {
  id: number
  founder_id: number
  founder_display_name?: string
  founder_slug?: string
  slug: string
  name: string
  tagline: string
  description?: string
  logo_url?: string
  cover_url?: string
  official_url?: string
  status: string
  tags?: { id: number; name: string; type: string }[]
  custom_links?: { id?: number; link_type: string; name: string; url: string }[]
  favorite_count: number
  view_count: number
  is_visible: boolean
  deleted_at?: string
  created_at?: string
}

export type TagVO = {
  id: number
  name: string
  slug: string
  type: 'industry' | 'stage' | 'role' | 'other'
  description?: string
  founder_count: number
  product_count: number
  is_active: boolean
  created_at?: string
}

export type ReportVO = {
  id: number
  reporter_id?: number
  reporter_name?: string
  target_type: 'founder' | 'product'
  target_id: number
  target_display_name?: string
  target_meta?: {
    slug?: string
    founder_slug?: string
    display_name?: string
    name?: string
    is_visible?: number | boolean
    [k: string]: any
  }
  reason_type: string
  reason_detail: string
  contact_email?: string
  status: 'pending' | 'processing' | 'resolved' | 'ignored'
  handler_id?: number
  handler_note?: string
  created_at?: string
  handled_at?: string
}

export type UserVO = {
  id: number
  username: string
  email?: string
  phone?: string
  display_name?: string
  role: 'guest' | 'founder' | 'admin'
  status: 'pending' | 'active' | 'banned' | 'suspended'
  created_at?: string
}

export function listFounders(params: { page?: number; page_size?: number; sort?: string; keyword?: string }) {
  return request.get<any, PageVO<FounderProfileItemVO>>('/api/v1/admin/founders', { params })
}

export function adminUpdateFounder(
  id: number,
  payload: {
    is_visible?: boolean
    display_name?: string
    slug?: string
    title?: string
    bio?: string
    public_email?: string
    location?: string
  },
) {
  return request.put<any, FounderProfileItemVO>(`/api/v1/admin/founders/${id}`, payload)
}

export function listProducts(params: {
  page?: number
  page_size?: number
  sort?: string
  keyword?: string
  status?: string
  is_visible?: 0 | 1 | boolean
  include_deleted?: 0 | 1
}) {
  return request.get<any, PageVO<ProductVO>>('/api/v1/admin/products', { params })
}

export function adminSetProductStatus(
  id: number,
  payload: {
    is_visible?: boolean
    status?: string
    action?: 'soft_delete' | 'restore'
  },
) {
  return request.put<any, ProductVO>(`/api/v1/admin/products/${id}/status`, payload)
}

export function listTags(params: {
  page?: number
  page_size?: number
  type?: string
  is_active?: boolean
  keyword?: string
}) {
  return request.get<any, PageVO<TagVO>>('/api/v1/admin/tags', { params })
}

export function createTag(payload: {
  name: string
  slug: string
  type: TagVO['type']
  description?: string
  is_active?: boolean
}) {
  return request.post<any, TagVO>('/api/v1/admin/tags', payload)
}

export function updateTag(
  id: number,
  payload: {
    name?: string
    slug?: string
    type?: TagVO['type']
    description?: string
    is_active?: boolean
  },
) {
  return request.put<any, TagVO>(`/api/v1/admin/tags/${id}`, payload)
}

export function deactivateTag(id: number) {
  return request.post<any, TagVO>(`/api/v1/admin/tags/${id}/deactivate`)
}

export function listReports(params: {
  page?: number
  page_size?: number
  status?: ReportVO['status']
  target_type?: string
  reason_type?: string
}) {
  return request.get<any, PageVO<ReportVO>>('/api/v1/admin/reports', { params })
}

export function handleReport(
  id: number,
  payload: {
    action: 'hide_target' | 'resolve' | 'ignore' | 'reject' | 'processing'
    admin_note?: string
  },
) {
  return request.post<any, ReportVO>(`/api/v1/admin/reports/${id}/handle`, payload)
}

export function listUsers(params: {
  page?: number
  page_size?: number
  role?: string
  status?: string
  keyword?: string
}) {
  return request.get<any, PageVO<UserVO>>('/api/v1/admin/users', { params })
}

export function adminUpdateUser(
  id: number,
  payload: {
    role?: UserVO['role']
    status?: UserVO['status']
  },
) {
  return request.put<any, UserVO>(`/api/v1/admin/users/${id}`, payload)
}
