<template>
  <div class="report-page">
    <el-card shadow="never" class="card">
      <template #header>
        <div class="card-head">
          <div>
            <h1 class="title"><el-icon :size="22" color="#F59E0B"><Warning /></el-icon> 提交举报</h1>
            <p class="sub">我们会认真对待每一份举报，核实后 48 小时内处理并给你反馈（如提供邮箱）</p>
          </div>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
        label-position="top"
        class="form"
        @submit.prevent="submit"
      >
        <el-row :gutter="20">
          <el-col :sm="12" :xs="24">
            <el-form-item label="举报对象类型" prop="target_type">
              <el-select v-model="form.target_type" placeholder="请选择" style="width: 100%">
                <el-option label="创业者名片" value="founder" />
                <el-option label="产品" value="product" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :sm="12" :xs="24">
            <el-form-item label="举报对象 ID" prop="target_id">
              <el-input-number
                v-model="form.target_id"
                :min="1"
                controls-position="right"
                style="width: 100%"
                placeholder="输入被举报的名片或产品 ID"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="举报原因分类" prop="reason_type">
          <el-radio-group v-model="form.reason_type">
            <el-radio-button value="spam">垃圾营销 / 刷屏</el-radio-button>
            <el-radio-button value="fake">虚假信息 / 欺诈</el-radio-button>
            <el-radio-button value="inappropriate">色情 / 暴力 / 违法</el-radio-button>
            <el-radio-button value="copyright">侵权 / 冒用身份</el-radio-button>
            <el-radio-button value="other">其他</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="详细说明（20-500字）" prop="reason_detail">
          <el-input
            v-model="form.reason_detail"
            type="textarea"
            :rows="6"
            maxlength="500"
            show-word-limit
            placeholder="请详细描述举报内容的问题，提供必要的证据链接或说明..."
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :sm="12" :xs="24">
            <el-form-item label="你的称呼（可选）">
              <el-input v-model="form.reporter_name" placeholder="方便我们联系你" />
            </el-form-item>
          </el-col>
          <el-col :sm="12" :xs="24">
            <el-form-item label="联系邮箱（可选）" prop="contact_email">
              <el-input v-model="form.contact_email" placeholder="example@mail.com" />
            </el-form-item>
          </el-col>
        </el-row>

        <div class="actions">
          <el-button @click="back">返回上一页</el-button>
          <el-button type="primary" :loading="submitting" @click="submit">提交举报</el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Warning } from '@element-plus/icons-vue'
import { submitReport } from '@/api/content'

const route = useRoute()
const router = useRouter()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive({
  target_type: 'founder' as 'founder' | 'product',
  target_id: 1,
  reason_type: 'spam' as string,
  reason_detail: '',
  contact_email: '',
  reporter_name: '',
})

const rules: FormRules = {
  target_type: [{ required: true, message: '请选择举报对象类型', trigger: 'change' }],
  target_id: [{ required: true, type: 'number', min: 1, message: '举报对象 ID 无效', trigger: 'blur' }],
  reason_type: [{ required: true, message: '请选择举报原因分类', trigger: 'change' }],
  reason_detail: [
    { required: true, message: '请填写详细说明', trigger: 'blur' },
    { min: 20, max: 500, message: '长度需在 20 到 500 个字符之间', trigger: 'blur' },
  ],
  contact_email: [
    {
      type: 'email',
      message: '请输入正确的邮箱格式',
      trigger: 'blur',
    },
  ],
}

onMounted(() => {
  if (route.query.target_type === 'founder' || route.query.target_type === 'product') {
    form.target_type = route.query.target_type as any
  }
  const tid = Number(route.query.target_id)
  if (tid > 0) form.target_id = tid
})

async function submit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch (_e) {
    return
  }
  submitting.value = true
  try {
    const r = await submitReport({
      target_type: form.target_type,
      target_id: Number(form.target_id),
      reason_type: form.reason_type,
      reason_detail: form.reason_detail,
      contact_email: form.contact_email || undefined,
      reporter_name: form.reporter_name || undefined,
    })
    ElMessage.success(`举报已提交，编号 #${r?.id || '—'}，我们会尽快处理`)
    back()
  } catch (e: any) {
    ElMessage.error(e?.message || '提交失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

function back() {
  const to = route.query.back as string | undefined
  if (to) router.push(to)
  else router.push('/')
}
</script>

<style scoped>
.report-page { max-width: 780px; margin: clamp(16px, 4vw, 40px) auto; padding: 0 2px; }
.card { border-radius: 14px; border: 1px solid #EEF2F7; }
.card-head { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; }
.title { margin: 0; font-size: 22px; display: inline-flex; align-items: center; gap: 10px; }
.sub { color:#64748b; margin: 6px 0 0; }
.form { padding-top: 12px; }
.actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 12px; }
</style>
