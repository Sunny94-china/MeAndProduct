import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const ACCESS_KEY = 'mp_admin_access'
const REFRESH_KEY = 'mp_admin_refresh'

const API_BASE = (import.meta.env.VITE_API_BASE || '/api/v1').replace(/\/$/, '')

interface LoginResp {
  access_token: string
  refresh_token: string
  user?: any
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>(localStorage.getItem(ACCESS_KEY) || '')
  const refreshToken = ref<string>(localStorage.getItem(REFRESH_KEY) || '')
  const user = ref<any>(JSON.parse(localStorage.getItem('mp_admin_user') || 'null'))
  const isRefreshing = ref(false)
  const pendingQueue: Array<() => void> = []

  function setTokens(access: string, refresh: string, u?: any) {
    accessToken.value = access
    refreshToken.value = refresh
    localStorage.setItem(ACCESS_KEY, access)
    localStorage.setItem(REFRESH_KEY, refresh)
    if (u !== undefined) {
      user.value = u
      localStorage.setItem('mp_admin_user', JSON.stringify(u))
    }
  }
  function clear() {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
    localStorage.removeItem('mp_admin_user')
  }
  function isLoggedIn() {
    return accessToken.value.length > 0
  }
  function isAdmin() {
    return (user.value && user.value.role === 'admin') || false
  }

  async function login(usernameOrEmail: string, password: string): Promise<LoginResp> {
    const resp = await axios.post<{ code: number; message: string; data: LoginResp }>(
      `${API_BASE}/auth/login`,
      { username_or_email: usernameOrEmail, password },
    )
    if (resp.data.code !== 0) throw new Error(resp.data.message || '登录失败')
    setTokens(resp.data.data.access_token, resp.data.data.refresh_token, resp.data.data.user || undefined)
    if (!isAdmin()) {
      clear()
      throw new Error('您不是管理员，无法登录本后台')
    }
    return resp.data.data
  }

  async function logout(): Promise<void> {
    if (refreshToken.value) {
      try {
        await axios.post(
          `${API_BASE}/auth/logout`,
          { refresh_token: refreshToken.value },
          { headers: { Authorization: `Bearer ${accessToken.value}` } },
        )
      } catch (_e) {}
    }
    clear()
  }

  async function refreshSilent(): Promise<void> {
    if (!refreshToken.value) {
      clear()
      throw new Error('no refresh token')
    }
    if (isRefreshing.value) {
      return new Promise<void>((resolve, reject) => {
        pendingQueue.push(() => (accessToken.value ? resolve() : reject(new Error('refresh failed'))))
      })
    }
    isRefreshing.value = true
    try {
      const resp = await axios.post<{ code: number; message: string; data: LoginResp }>(
        `${API_BASE}/auth/refresh`,
        { refresh_token: refreshToken.value },
      )
      if (resp.data.code !== 0) throw new Error(resp.data.message || 'refresh failed')
      setTokens(resp.data.data.access_token, resp.data.data.refresh_token, resp.data.data.user || undefined)
      if (!isAdmin()) {
        clear()
        throw new Error('您不是管理员')
      }
      pendingQueue.splice(0, pendingQueue.length).forEach((fn) => fn())
    } catch (e) {
      clear()
      pendingQueue.splice(0, pendingQueue.length)
      try {
        const router = useRouter()
        router.push('/login')
      } catch (_) {}
      ElMessage.error('登录已过期，请重新登录')
      throw e
    } finally {
      isRefreshing.value = false
    }
  }

  async function fetchMe(): Promise<any> {
    if (!accessToken.value) throw new Error('not logged in')
    const resp = await axios.get<{ code: number; message: string; data: any }>(`${API_BASE}/auth/me`, {
      headers: { Authorization: `Bearer ${accessToken.value}` },
    })
    if (resp.data.code !== 0) throw new Error(resp.data.message || '获取用户信息失败')
    user.value = resp.data.data
    localStorage.setItem('mp_admin_user', JSON.stringify(user.value))
    return resp.data.data
  }

  return {
    accessToken,
    refreshToken,
    user,
    setTokens,
    clear,
    login,
    logout,
    refreshSilent,
    fetchMe,
    isLoggedIn,
    isAdmin,
  }
})
