import request from './request'

export type PageVO<T> = {
  items: T[]
  total: number
  page: number
  page_size: number
}

export type FounderVO = {
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
  contact_email?: string
  website_url?: string
  tags?: { id: number; name: string; type: string }[]
  custom_links?: { id: number; link_type: string; name: string; url: string; click_count: number }[]
  favorite_count: number
  view_count: number
  created_at?: string
  updated_at?: string
}

export type ProductVO = {
  id: number
  founder_id: number
  founder_slug?: string
  founder_display_name?: string
  slug: string
  name: string
  tagline: string
  description?: string
  logo_url?: string
  cover_url?: string
  official_url?: string
  status: string
  tags?: { id: number; name: string; type: string }[]
  custom_links?: { id: number; link_type: string; name: string; url: string; click_count: number }[]
  favorite_count: number
  view_count: number
  sort_order: number
  is_visible: boolean
  created_at?: string
  updated_at?: string
}

export type TagVO = {
  id: number
  name: string
  slug: string
  type: 'industry' | 'stage' | 'role' | 'other'
  founder_count: number
  product_count: number
  is_active: boolean
  created_at?: string
}

export type TagGroupedVO = {
  industry?: TagVO[]
  stage?: TagVO[]
  role?: TagVO[]
  other?: TagVO[]
}

export type FavoriteToggleResp = {
  target_type: string
  target_id: number
  favorited: boolean
  favorite_count: number
}

export type FavoriteItemVO = {
  id: number
  target_type: 'founder' | 'product'
  target_id: number
  target_snapshot?: {
    display_name?: string
    name?: string
    slug?: string
    founder_slug?: string
    tagline?: string
    avatar_url?: string
    logo_url?: string
  }
  created_at?: string
}

export function listFounders(params: { page?: number; page_size?: number; sort?: string; tags?: string; keyword?: string }) {
  return request.get<any, PageVO<FounderVO>>('/api/v1/public/founders', { params })
}

export function getFounderDetail(slug: string) {
  return request.get<any, FounderVO>(`/api/v1/public/founders/${slug}`)
}

export function listFounderProducts(slug: string, params: { page?: number; page_size?: number; sort?: string }) {
  return request.get<any, PageVO<ProductVO>>(`/api/v1/public/founders/${slug}/products`, { params })
}

export function getFounderProductDetail(slug: string, pSlug: string) {
  return request.get<any, ProductVO>(`/api/v1/public/founders/${slug}/products/${pSlug}`)
}

export function listProducts(params: { page?: number; page_size?: number; sort?: string; stage?: string; tags?: string; keyword?: string }) {
  return request.get<any, PageVO<ProductVO>>('/api/v1/public/products', { params })
}

export function listActiveTags() {
  return request.get<any, TagVO[]>('/api/v1/public/tags')
}

export function getTagsGrouped() {
  return request.get<any, TagGroupedVO>('/api/v1/public/tags/grouped')
}

export function submitReport(payload: {
  target_type: 'founder' | 'product'
  target_id: number
  reason_type: string
  reason_detail: string
  contact_email?: string
  reporter_name?: string
}) {
  return request.post<any, { id: number; status: string }>('/api/v1/public/reports', payload)
}

export function trackLinkClick(id: number | string) {
  return request.post<any, { click_count: number }>(`/api/v1/public/links/${id}/click`)
}

export function toggleFavorite(payload: { target_type: 'founder' | 'product'; target_id: number }) {
  return request.post<any, FavoriteToggleResp>('/api/v1/user/favorites/toggle', payload)
}

export function listMyFavorites(params: { page?: number; page_size?: number; target_type?: 'founder' | 'product' | 'all' }) {
  return request.get<any, PageVO<FavoriteItemVO>>('/api/v1/user/favorites', { params })
}

export function checkFavorite(targetType: string, targetId: number | string) {
  return request.get<any, { favorited: boolean; favorite_count: number }>(
    `/api/v1/user/favorites/exists/${targetType}/${targetId}`,
  )
}

export type FounderProfileItemVO = FounderVO

export function applyProfile(payload: {
  slug?: string
  display_name: string
  title?: string
  bio?: string
  bio_long?: string
  avatar_url?: string
  cover_url?: string
  location?: string
  public_email?: string
  custom_links?: { link_type: string; name: string; url: string }[]
  tag_ids?: number[]
}) {
  return request.post<any, FounderProfileItemVO>('/api/v1/founder/profiles/apply', payload)
}

export function getMyProfile() {
  return request.get<any, FounderProfileItemVO>('/api/v1/founder/profiles/me', { skipErrorToast: true } as any)
}

export function updateMyProfile(payload: {
  display_name?: string
  title?: string
  bio?: string
  bio_long?: string
  slug?: string
  avatar_url?: string
  cover_url?: string
  location?: string
  public_email?: string
  custom_links?: { id?: number; link_type: string; name: string; url: string }[]
  tag_ids?: number[]
}) {
  return request.put<any, FounderProfileItemVO>('/api/v1/founder/profiles/me', payload)
}

export function createProduct(payload: {
  name: string
  tagline: string
  slug?: string
  description?: string
  logo_url?: string
  cover_url?: string
  official_url?: string
  status?: string
  tag_ids?: number[]
  custom_links?: { link_type: string; name: string; url: string }[]
}) {
  return request.post<any, ProductVO>('/api/v1/founder/products', payload)
}

export function listMyProducts(params: { page?: number; page_size?: number; sort?: string }) {
  return request.get<any, PageVO<ProductVO>>('/api/v1/founder/products', { params })
}

export function getMyProduct(id: number | string) {
  return request.get<any, ProductVO>(`/api/v1/founder/products/${id}`)
}

export function updateMyProduct(
  id: number | string,
  payload: {
    name?: string
    tagline?: string
    slug?: string
    description?: string
    logo_url?: string
    cover_url?: string
    official_url?: string
    status?: string
    is_visible?: boolean
    tag_ids?: number[]
    custom_links?: { id?: number; link_type: string; name: string; url: string }[]
  },
) {
  return request.put<any, ProductVO>(`/api/v1/founder/products/${id}`, payload)
}

export function deleteMyProduct(id: number | string) {
  return request.delete<any, void>(`/api/v1/founder/products/${id}`)
}

export function reorderProducts(ordered_ids: (number | string)[]) {
  return request.post<any, void>('/api/v1/founder/products/reorder', { ordered_ids })
}

export function listMyInvites() {
  return request.get<any, { items: any[]; total: number }>('/api/v1/founder/invites')
}

export function createInvite(payload?: { max_uses?: number; expires_days?: number }) {
  return request.post<any, { invite: any }>('/api/v1/founder/invites', payload || { max_uses: 5, expires_days: 30 })
}

export function revokeInvite(invite_id: number) {
  return request.post<any, void>('/api/v1/founder/invites/revoke', { invite_id })
}
