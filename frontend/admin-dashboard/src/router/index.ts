import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login',     name: 'login',      component: () => import('@/views/Login.vue') },
  { path: '/',          name: 'dash',       component: () => import('@/views/Dashboard.vue'),     meta: { auth: true, role: 'admin' } },
  { path: '/founders',  name: 'founders',   component: () => import('@/views/FounderManage.vue'), meta: { auth: true, role: 'admin' } },
  { path: '/products',  name: 'products',   component: () => import('@/views/ProductManage.vue'), meta: { auth: true, role: 'admin' } },
  { path: '/reports',   name: 'reports',    component: () => import('@/views/ReportManage.vue'),  meta: { auth: true, role: 'admin' } },
  { path: '/tags',      name: 'tags',       component: () => import('@/views/TagManage.vue'),     meta: { auth: true, role: 'admin' } },
  { path: '/invites',   name: 'invites',    component: () => import('@/views/InviteManage.vue'),  meta: { auth: true, role: 'admin' } },
  { path: '/users',     name: 'users',      component: () => import('@/views/UserManage.vue'),    meta: { auth: true, role: 'admin' } },
]

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes,
})

router.beforeEach((to, _f, next) => {
  const auth = useAuthStore()
  const needAuth = !!to.meta.auth
  if (needAuth && !auth.isLoggedIn()) {
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (needAuth && !auth.isAdmin()) {
    auth.clear()
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (to.path === '/login' && auth.isLoggedIn() && auth.isAdmin()) {
    return next('/')
  }
  next()
})

export default router
