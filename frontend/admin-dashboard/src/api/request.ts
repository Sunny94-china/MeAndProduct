import axios, { AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/',
  timeout: 30000,
})

service.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.accessToken) {
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

service.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) {
        const data = body.data
        if (data && typeof data === 'object') {
          if (Array.isArray(data.list) && data.items === undefined) data.items = data.list
          if (Array.isArray(data.items) && data.list === undefined) data.list = data.items
          if (data.page && typeof data.page === 'object' && data.page.total !== undefined && data.total === undefined) {
            data.total = data.page.total
          }
        }
        return data
      }
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(body)
    }
    return body
  },
  (err) => {
    const status = err.response?.status
    if (status === 401) {
      const auth = useAuthStore()
      return auth
        .refreshSilent()
        .then(() => service.request(err.config as AxiosRequestConfig))
        .catch(() => {
          auth.clear()
          ElMessage.error('请先登录')
          return Promise.reject(err)
        })
    }
    ElMessage.error(err.response?.data?.message || err.message || '请求失败')
    return Promise.reject(err)
  },
)

export default service
