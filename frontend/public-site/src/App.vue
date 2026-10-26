<template>
  <el-config-provider :locale="locale">
    <div class="app-root">
      <header class="nav" v-if="showNav">
        <div class="nav-inner">
          <button
            v-if="!isStudio"
            class="icon-btn show-sm"
            type="button"
            aria-label="打开菜单"
            @click="drawerOpen = true"
          >
            <el-icon :size="22"><Menu /></el-icon>
          </button>

          <router-link to="/" class="brand" @click="drawerOpen = false">
            <el-icon :size="22" color="#4F46E5"><Link /></el-icon>
            <span>MeAndProduct</span>
          </router-link>

          <nav class="nav-links hide-sm">
            <router-link to="/" class="nav-link">首页</router-link>
            <router-link to="/tags" class="nav-link">
              <el-icon :size="16"><PriceTag /></el-icon>标签
            </router-link>
            <router-link to="/search" class="nav-link">
              <el-icon :size="16"><Search /></el-icon>搜索
            </router-link>
            <router-link v-if="auth.isLoggedIn()" to="/favorites" class="nav-link">
              <el-icon :size="16"><Star /></el-icon>我的收藏
            </router-link>
            <router-link v-if="auth.isFounder()" to="/studio" class="nav-link">工作台</router-link>
          </nav>

          <div class="nav-actions">
            <template v-if="!auth.isLoggedIn()">
              <el-button type="primary" class="login-btn" @click="loginOpen = true">登录</el-button>
            </template>
            <template v-else>
              <el-dropdown trigger="click" @command="onUserCmd">
                <span class="user-chip">
                  <el-avatar :size="32" :src="avatar">{{ avatarText }}</el-avatar>
                  <span class="user-name hide-sm">{{ auth.user?.display_name || auth.user?.username || '我' }}</span>
                  <el-icon class="hide-sm"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="auth.isFounder()" command="founder">
                      <el-icon><Setting /></el-icon>我的工作台
                    </el-dropdown-item>
                    <el-dropdown-item v-if="auth.user?.role === 'admin'" command="admin">
                      <el-icon><UserFilled /></el-icon>管理后台
                    </el-dropdown-item>
                    <el-dropdown-item command="profile" v-if="profileSlug">
                      <el-icon><View /></el-icon>访问我的名片
                    </el-dropdown-item>
                    <el-dropdown-item command="favorites" v-if="auth.isLoggedIn()">
                      <el-icon><Star /></el-icon>我的收藏
                    </el-dropdown-item>
                    <el-dropdown-item divided command="logout">
                      <el-icon><SwitchButton /></el-icon>退出登录
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </div>
        </div>
      </header>

      <el-drawer
        v-model="drawerOpen"
        direction="ltr"
        size="78%"
        :with-header="false"
        class="nav-drawer"
      >
        <div class="drawer-body">
          <div class="drawer-brand">
            <el-icon :size="22" color="#4F46E5"><Link /></el-icon>
            <span>MeAndProduct</span>
          </div>
          <nav class="drawer-links">
            <router-link to="/" class="drawer-link" @click="drawerOpen = false">首页</router-link>
            <router-link to="/tags" class="drawer-link" @click="drawerOpen = false">标签探索</router-link>
            <router-link to="/search" class="drawer-link" @click="drawerOpen = false">搜索</router-link>
            <router-link
              v-if="auth.isLoggedIn()"
              to="/favorites"
              class="drawer-link"
              @click="drawerOpen = false"
            >我的收藏</router-link>
            <router-link
              v-if="auth.isFounder()"
              to="/studio"
              class="drawer-link"
              @click="drawerOpen = false"
            >我的工作台</router-link>
          </nav>
          <el-button
            v-if="!auth.isLoggedIn()"
            type="primary"
            class="drawer-login"
            @click="drawerOpen = false; loginOpen = true"
          >登录 / 注册</el-button>
        </div>
      </el-drawer>

      <main :class="['main', { 'main-studio': isStudio }]">
        <router-view v-slot="{ Component }">
          <component :is="Component" />
        </router-view>
      </main>

      <LoginModal v-model="loginOpen" @success="onLoginSuccess" />
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import {
  Link, ArrowDown, Setting, UserFilled, View, SwitchButton,
  PriceTag, Search, Star, Menu,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import LoginModal from '@/views/LoginModal.vue'

const locale = zhCn
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const loginOpen = ref(false)
const drawerOpen = ref(false)
const isStudio = computed(() => route.matched.some((r) => r.meta.studio))

watch(
  () => route.query.login,
  (v) => {
    if (v === '1' && !auth.isLoggedIn()) loginOpen.value = true
  },
  { immediate: true },
)

watch(
  () => route.fullPath,
  () => {
    drawerOpen.value = false
  },
)

const showNav = computed(() => true)

const profileSlug = computed(() => {
  return auth.user?.founder?.slug || auth.user?.slug || ''
})
const avatar = computed(() => auth.user?.avatar_url || auth.user?.founder?.avatar_url || '')
const avatarText = computed(() => {
  const s = auth.user?.display_name || auth.user?.username || 'U'
  return s.slice(0, 1).toUpperCase()
})

function onUserCmd(cmd: string) {
  switch (cmd) {
    case 'founder':
      router.push('/studio')
      break
    case 'admin':
      window.open('/admin/', '_blank')
      break
    case 'profile':
      if (profileSlug.value) router.push(`/@${profileSlug.value}`)
      break
    case 'favorites':
      router.push('/favorites')
      break
    case 'logout':
      auth.logout().then(() => {
        ElMessage.success('已退出登录')
      })
      break
  }
}

async function onLoginSuccess() {
  try {
    await auth.fetchMe()
  } catch (_e) {}
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  if (redirect.startsWith('/') && !redirect.startsWith('//')) {
    router.replace(redirect)
  }
}
</script>

<style>
.app-root {
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

.nav {
  position: sticky;
  top: 0;
  z-index: 50;
  height: var(--mp-nav-h);
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--mp-border);
}

.nav-inner {
  max-width: var(--mp-max);
  height: 100%;
  margin: 0 auto;
  padding: 0 var(--mp-pad);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--mp-text);
  cursor: pointer;
  flex: none;
}

.icon-btn:active {
  background: var(--mp-primary-soft);
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 750;
  font-size: 17px;
  color: var(--mp-text);
  text-decoration: none;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  justify-content: center;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #334155;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  padding: 8px 12px;
  border-radius: 999px;
  transition: color 0.15s, background 0.15s;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: var(--mp-primary);
  background: var(--mp-primary-soft);
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.login-btn {
  border-radius: 999px !important;
  padding: 8px 16px !important;
}

.user-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px 4px 4px;
  border-radius: 999px;
  border: 1px solid var(--mp-border);
  background: #fff;
  cursor: pointer;
}

.user-name {
  font-weight: 500;
  font-size: 14px;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.main {
  flex: 1;
  max-width: var(--mp-max);
  width: 100%;
  margin: 0 auto;
  padding: 24px var(--mp-pad);
}

.main-studio {
  max-width: none;
  padding: 0;
}

.drawer-body {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 8px 4px 20px;
}

.drawer-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 750;
  font-size: 18px;
  padding: 12px 8px 20px;
  border-bottom: 1px solid var(--mp-border);
  margin-bottom: 12px;
}

.drawer-links {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.drawer-link {
  display: block;
  padding: 14px 12px;
  border-radius: 12px;
  text-decoration: none;
  color: #1e293b;
  font-weight: 560;
  font-size: 15px;
}

.drawer-link.router-link-active,
.drawer-link:hover {
  background: var(--mp-primary-soft);
  color: var(--mp-primary);
}

.drawer-login {
  margin-top: auto;
  width: 100%;
  border-radius: 12px !important;
  height: 44px;
}

@media (max-width: 768px) {
  .main {
    padding: 16px var(--mp-pad) 28px;
  }
}
</style>
