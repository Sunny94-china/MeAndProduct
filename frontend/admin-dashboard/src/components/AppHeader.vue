<template>
  <el-header class="hdr">
    <div class="left">
      <button class="menu-btn show-sm" type="button" aria-label="打开菜单" @click="open">
        <el-icon :size="20"><Menu /></el-icon>
      </button>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/' }">管理后台</el-breadcrumb-item>
        <el-breadcrumb-item>{{ crumb }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="right">
      <el-dropdown @command="onCmd">
        <div class="user">
          <el-avatar :size="32" style="background:#DC2626;">A</el-avatar>
          <span class="n hide-sm">管理员</span>
          <el-icon><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="logout" divided>退出</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>

<script setup lang="ts">
import { ArrowDown, Menu } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { useAdminNav } from '@/composables/useAdminNav'

defineProps<{ crumb: string }>()
const router = useRouter()
const { open } = useAdminNav()

function onCmd() {
  localStorage.removeItem('mp_admin_access')
  router.push('/login')
}
</script>

<style scoped>
.hdr {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  background: #fff;
  border-bottom: 1px solid #e2e8f0;
  height: auto !important;
  min-height: 56px;
  padding: 10px 16px;
}

.left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.menu-btn {
  display: none;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #fff;
  cursor: pointer;
  flex: none;
}

.right {
  display: flex;
  align-items: center;
  gap: 14px;
  flex: none;
}

.user {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.n {
  font-size: 14px;
}

@media (max-width: 768px) {
  .menu-btn {
    display: inline-flex;
  }

  .hide-sm {
    display: none !important;
  }

  :deep(.el-breadcrumb) {
    font-size: 13px;
  }
}
</style>
