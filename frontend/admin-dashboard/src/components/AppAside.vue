<template>
  <el-aside width="220px" class="aside hide-sm">
    <div class="brand">
      <el-icon :size="22" color="#DC2626"><Lock /></el-icon>
      <span>管理后台</span>
    </div>
    <el-menu
      :default-active="$route.path"
      router
      background-color="#1E1B4B"
      text-color="#C7D2FE"
      active-text-color="#ffffff"
      class="menu"
    >
      <el-menu-item v-for="item in items" :key="item.path" :index="item.path">
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </el-menu-item>
    </el-menu>
  </el-aside>

  <el-drawer
    v-model="drawerOpen"
    direction="ltr"
    size="78%"
    :with-header="false"
  >
    <div class="drawer-panel">
      <div class="brand">
        <el-icon :size="22" color="#DC2626"><Lock /></el-icon>
        <span>管理后台</span>
      </div>
      <el-menu
        :default-active="$route.path"
        router
        background-color="#1E1B4B"
        text-color="#C7D2FE"
        active-text-color="#ffffff"
        class="menu"
        @select="close"
      >
        <el-menu-item v-for="item in items" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import {
  Lock, DataAnalysis, User, Goods, Warning, CollectionTag, Present, Avatar,
} from '@element-plus/icons-vue'
import { useAdminNav } from '@/composables/useAdminNav'

const { drawerOpen, close } = useAdminNav()

const items = [
  { path: '/', label: '数据概览', icon: DataAnalysis },
  { path: '/founders', label: '创业者管理', icon: User },
  { path: '/products', label: '产品管理', icon: Goods },
  { path: '/reports', label: '举报处理', icon: Warning },
  { path: '/tags', label: '标签管理', icon: CollectionTag },
  { path: '/invites', label: '邀请码', icon: Present },
  { path: '/users', label: '用户角色', icon: Avatar },
]
</script>

<style scoped>
.aside {
  background: #1e1b4b;
  color: #fff;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.brand {
  padding: 20px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  font-size: 16px;
  border-bottom: 1px solid #312e81;
}

.menu {
  border-right: none;
  flex: 1;
}

.drawer-panel {
  margin: -20px;
  min-height: calc(100% + 40px);
  background: #1e1b4b;
}

.hide-sm {
  display: block;
}

@media (max-width: 768px) {
  .hide-sm {
    display: none !important;
  }
}
</style>
