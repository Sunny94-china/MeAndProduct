import request from './request'

export type Founder = {
  id: number
  slug: string
  display_name: string
  title?: string
  bio?: string
  avatar_url?: string
  location?: string
  website_url?: string
  github_url?: string
  twitter_url?: string
  tags?: { id: number; name: string }[]
  favorite_count: number
  view_count: number
}

export function getFounderList(params: Record<string, any>) {
  return request.get<any, Founder[]>('/api/v1/founders', { params })
}

export function getFounderDetail(slug: string) {
  return request.get<any, Founder>(`/api/v1/founders/${slug}`)
}

export function searchFounders(q: string, params?: Record<string, any>) {
  return request.post<any, Founder[]>(`/api/v1/founders/search`, { q, ...params })
}
