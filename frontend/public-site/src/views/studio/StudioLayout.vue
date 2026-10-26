<template>
  <div class="studio">
    <aside class="aside hide-sm">
      <div class="aside-title">我的工作台</div>
      <el-menu
        :default-active="active"
        router
        background-color="#0F172A"
        text-color="#CBD5E1"
        active-text-color="#ffffff"
        class="menu"
      >
        <el-menu-item v-for="item in items" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <div class="studio-body">
      <div class="studio-top show-sm">
        <button class="menu-btn" type="button" aria-label="工作台菜单" @click="drawerOpen = true">
          <el-icon :size="20"><Menu /></el-icon>
          <span>{{ currentLabel }}</span>
        </button>
      </div>

      <main class="studio-main">
        <router-view />
      </main>
    </div>

    <el-drawer v-model="drawerOpen" direction="ltr" size="78%" :with-header="false">
      <div class="drawer-panel">
        <div class="aside-title dark">我的工作台</div>
        <el-menu
          :default-active="active"
          router
          background-color="#0F172A"
          text-color="#CBD5E1"
          active-text-color="#ffffff"
          class="menu"
          @select="drawerOpen = false"
        >
          <el-menu-item v-for="item in items" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </el-menu>
      </div>
    </el-drawer>

    <nav class="studio-tabbar show-sm">
      <router-link
        v-for="item in tabItems"
        :key="item.path"
        :to="item.path"
        class="tab-item"
        :class="{ active: active === item.path || (item.path === '/studio/products' && active.startsWith('/studio/products')) }"
      >
        <el-icon :size="20"><component :is="item.icon" /></el-icon>
        <span>{{ item.short }}</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { DataAnalysis, Edit, Goods, Present, Star, Setting, Menu } from '@element-plus/icons-vue'

const route = useRoute()
const drawerOpen = ref(false)

const items = [
  { path: '/studio', label: '概览', icon: DataAnalysis, short: '概览' },
  { path: '/studio/profile', label: '名片编辑', icon: Edit, short: '名片' },
  { path: '/studio/products', label: '产品管理', icon: Goods, short: '产品' },
  { path: '/studio/invites', label: '邀请码', icon: Present, short: '邀请' },
  { path: '/favorites', label: '我的收藏', icon: Star, short: '收藏' },
  { path: '/studio/settings', label: '账号设置', icon: Setting, short: '设置' },
]

const tabItems = items.filter((i) =>
  ['/studio', '/studio/profile', '/studio/products', '/studio/invites', '/studio/settings'].includes(i.path),
)

const active = computed(() => {
  if (route.path.startsWith('/studio/products')) return '/studio/products'
  return route.path
})

const currentLabel = computed(() => {
  const hit = items.find((i) => i.path === active.value)
  return hit?.label || '工作台'
})

watch(
  () => route.fullPath,
  () => {
    drawerOpen.value = false
  },
)
</script>

<style scoped>
.studio {
  min-height: calc(100vh - var(--mp-nav-h));
  background: var(--mp-bg);
  display: flex;
}

.aside {
  width: 220px;
  flex: none;
  background: #0f172a;
  color: #fff;
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - var(--mp-nav-h));
}

.aside-title {
  padding: 18px 16px 10px;
  font-weight: 700;
  font-size: 14px;
  color: #e2e8f0;
}

.aside-title.dark {
  background: #0f172a;
  margin: -20px -20px 8px;
  padding: 24px 20px 14px;
}

.menu {
  border-right: none;
  flex: 1;
}

.studio-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.studio-top {
  position: sticky;
  top: var(--mp-nav-h);
  z-index: 20;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--mp-border);
  padding: 8px var(--mp-pad);
}

.menu-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--mp-border);
  background: #fff;
  border-radius: 999px;
  padding: 8px 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--mp-text);
  cursor: pointer;
}

.studio-main {
  flex: 1;
  padding: 20px 24px 32px;
}

.drawer-panel {
  margin: -20px;
  min-height: calc(100% + 40px);
  background: #0f172a;
}

.studio-tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(12px);
  border-top: 1px solid var(--mp-border);
  padding: 6px 4px calc(6px + env(safe-area-inset-bottom));
}

.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  text-decoration: none;
  color: #64748b;
  font-size: 11px;
  font-weight: 560;
  padding: 6px 2px;
  border-radius: 10px;
}

.tab-item.active {
  color: var(--mp-primary);
}

@media (max-width: 768px) {
  .studio-main {
    padding: 14px var(--mp-pad) calc(78px + env(safe-area-inset-bottom));
  }
}
</style>
