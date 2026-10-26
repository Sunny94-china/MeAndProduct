<template>
  <div class="studio-page">
    <el-card style="max-width:640px;">
      <el-form label-width="100px">
        <el-form-item label="当前账号">
          <el-input :model-value="account" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-input :model-value="roleLabel" disabled />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">保存密码</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const account = computed(() => auth.user?.email || auth.user?.username || '')
const roleLabel = computed(() => {
  const role = auth.user?.role
  if (role === 'admin') return '管理员'
  if (role === 'founder') return '创业者'
  return role || ''
})

function save() {
  ElMessage.warning('修改密码还没接上，当前密码保持不变')
}
</script>
<style scoped>.studio-page{ padding: 8px 4px 24px; }</style>
