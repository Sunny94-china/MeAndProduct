import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/Home.vue'),
  },
  {
    path: '/search',
    name: 'search',
    component: () => import('@/views/SearchResult.vue'),
  },
  {
    path: '/tags',
    name: 'tags',
    component: () => import('@/views/TagExplore.vue'),
  },
  {
    path: '/@:slug',
    name: 'founder-detail',
    component: () => import('@/views/FounderDetail.vue'),
    props: true,
  },
  {
    path: '/@:slug/:productSlug',
    name: 'product-detail',
    component: () => import('@/views/ProductDetail.vue'),
    props: true,
  },
  {
    path: '/report',
    name: 'report',
    component: () => import('@/views/Report.vue'),
  },
  {
    path: '/favorites',
    name: 'favorites',
    component: () => import('@/views/MyFavorites.vue'),
    meta: { auth: false },
  },
  {
    path: '/studio',
    component: () => import('@/views/studio/StudioLayout.vue'),
    meta: { studio: true, auth: true, role: 'founder' },
    children: [
      { path: '', name: 'studio-home', component: () => import('@/views/studio/StudioHome.vue') },
      { path: 'profile', name: 'studio-profile', component: () => import('@/views/studio/ProfileEdit.vue') },
      { path: 'products', name: 'studio-products', component: () => import('@/views/studio/ProductList.vue') },
      { path: 'products/new', name: 'studio-product-new', component: () => import('@/views/studio/ProductEdit.vue') },
      { path: 'products/:id/edit', name: 'studio-product-edit', component: () => import('@/views/studio/ProductEdit.vue'), props: true },
      { path: 'invites', name: 'studio-invites', component: () => import('@/views/studio/InviteManage.vue') },
      { path: 'settings', name: 'studio-settings', component: () => import('@/views/studio/Settings.vue') },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'notfound-redirect',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory('/'),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  const needAuth = to.matched.some((r) => r.meta.auth)
  const needFounder = to.matched.some((r) => r.meta.role === 'founder')
  if (needAuth && !auth.isLoggedIn()) {
    return { path: '/', query: { login: '1', redirect: to.fullPath } }
  }
  if (needFounder && !auth.isFounder()) {
    ElMessage.warning('工作台只对创业者开放')
    return { path: '/' }
  }
})

export default router
