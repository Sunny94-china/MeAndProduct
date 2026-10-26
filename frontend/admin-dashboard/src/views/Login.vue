<template>
  <div class="bg">
    <el-card class="card" shadow="hover">
      <div class="logo"><el-icon :size="28" color="#DC2626"><Lock/></el-icon><span>MeAndProduct · 管理后台</span></div>
      <el-alert type="warning" :closable="false" show-icon style="margin-bottom: 16px;"
        title="仅管理员可登录"
        description="非管理员账号将自动跳回登录页。" />
      <el-form label-width="0" size="large" @submit.prevent="submit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="管理员用户名 / 邮箱" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" show-password placeholder="密码" />
        </el-form-item>
        <el-form-item>
          <el-button type="danger" class="w100" :loading="loading" @click="submit">登录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function submit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名/邮箱和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(form.username.trim(), form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.bg{ min-height:100vh; background: linear-gradient(135deg,#FEE2E2,#FEF3C7); display:flex; align-items:center; justify-content:center; padding:20px;}
.card{ width: 100%; max-width: 420px; }
.logo{ display:flex; align-items:center; gap:8px; font-weight:700; margin-bottom:18px; font-size:18px;}
.w100{ width:100%; }
</style>
