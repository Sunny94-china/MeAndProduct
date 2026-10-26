<template>
  <el-dialog
    v-model="visible"
    :title="tab === 'login' ? '登录' : '邀请码注册'"
    :width="dialogWidth"
    :fullscreen="isMobile"
    :close-on-click-modal="false"
    destroy-on-close
    align-center
    @close="onClose"
  >
    <el-tabs v-model="tab">
      <el-tab-pane label="登录" name="login">
        <el-form :model="loginForm" label-width="0" size="default">
          <el-form-item>
            <el-input v-model="loginForm.username" placeholder="用户名 / 邮箱" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="loginForm.password" type="password" show-password placeholder="密码" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" class="w100" :loading="loading" @click="submitLogin">登录</el-button>
          </el-form-item>
          <div class="hint">
            还没有账号？
            <el-link type="primary" @click="tab = 'invite'">使用邀请码注册</el-link>
          </div>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="邀请码注册" name="invite">
        <el-alert
          type="info"
          :closable="false"
          show-icon
          title="注册后即为创业者，可以编辑自己的名片和产品"
          style="margin-bottom: 12px;"
        />
        <el-form :model="inviteForm" label-width="0" size="default">
          <el-form-item>
            <el-input v-model="inviteForm.code" placeholder="邀请码" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="inviteForm.username" placeholder="用户名（3-50位字母数字_-）" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="inviteForm.email" placeholder="邮箱" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="inviteForm.password" type="password" show-password placeholder="密码（8-64位非空白）" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="inviteForm.displayName" placeholder="昵称（可选）" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" class="w100" :loading="loading" @click="submitInvite">注册并登录</el-button>
          </el-form-item>
          <div class="hint">
            已有账号？
            <el-link type="primary" @click="tab = 'login'">返回登录</el-link>
          </div>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean]; success: [] }>()

const auth = useAuthStore()
const winW = ref(typeof window !== 'undefined' ? window.innerWidth : 1024)
const isMobile = computed(() => winW.value < 640)
const dialogWidth = computed(() => (isMobile.value ? '100%' : '460px'))

function onResize() {
  winW.value = window.innerWidth
}
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))

const visible = ref(props.modelValue)
watch(
  () => props.modelValue,
  (v) => (visible.value = v),
)
watch(visible, (v) => emit('update:modelValue', v))

const tab = ref<'login' | 'invite'>('login')
const loading = ref(false)
const loginForm = reactive({ username: '', password: '' })
const inviteForm = reactive({ code: '', username: '', email: '', password: '', displayName: '' })

const API_BASE = (import.meta.env.VITE_API_BASE || '/api/v1').replace(/\/$/, '')

function onClose() {
  tab.value = 'login'
  loginForm.username = ''
  loginForm.password = ''
}

async function submitLogin() {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning('请输入用户名/邮箱和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(loginForm.username.trim(), loginForm.password)
    ElMessage.success('登录成功')
    emit('success')
    visible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}

async function submitInvite() {
  const f = inviteForm
  if (!f.code || !f.username || !f.email || !f.password) {
    ElMessage.warning('请完整填写邀请码、用户名、邮箱和密码')
    return
  }
  loading.value = true
  try {
    const resp = await axios.post<{ code: number; message: string; data: any }>(`${API_BASE}/auth/register/invite`, {
      invite_code: f.code.trim(),
      username: f.username.trim(),
      email: f.email.trim(),
      password: f.password,
      display_name: f.displayName.trim() || f.username.trim(),
    })
    if (resp.data.code !== 0) throw new Error(resp.data.message || '注册失败')
    const d = resp.data.data
    auth.setTokens(d.access_token, d.refresh_token, d.user)
    ElMessage.success('注册成功，你现在是创业者')
    emit('success')
    visible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.w100 { width: 100%; }
.hint { color: #64748B; font-size: 13px; text-align: center; }
</style>
