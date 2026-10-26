<template>
  <div class="studio-page">
    <el-row :gutter="20">
      <el-col :xs="12" :md="6">
        <el-card shadow="hover" class="stat"><div class="label">名片浏览</div><div class="num">{{ profile?.view_count ?? 0 }}</div></el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card shadow="hover" class="stat"><div class="label">被收藏</div><div class="num">{{ profile?.favorite_count ?? 0 }}</div></el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card shadow="hover" class="stat"><div class="label">产品数</div><div class="num">{{ productTotal }}</div></el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card shadow="hover" class="stat"><div class="label">公开状态</div><div class="num small">{{ profile ? '已创建' : '未完善' }}</div></el-card>
      </el-col>
    </el-row>
    <el-card class="mt20" shadow="never">
      <template #header>欢迎来到工作台</template>
      <p>去「名片编辑」完善个人资料，再在「产品管理」添加产品。保存后，同一站点上的公开页面会展示这些内容。</p>
      <div class="actions">
        <el-button type="primary" @click="$router.push('/studio/profile')">编辑名片</el-button>
        <el-button @click="$router.push('/studio/products')">管理产品</el-button>
        <el-button v-if="profile?.slug" link type="primary" @click="$router.push(`/@${profile.slug}`)">查看我的公开名片</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getMyProfile, listMyProducts, type FounderProfileItemVO } from '@/api/content'

const profile = ref<FounderProfileItemVO | null>(null)
const productTotal = ref(0)

onMounted(async () => {
  try {
    profile.value = await getMyProfile()
  } catch {
    profile.value = null
  }
  try {
    const page = await listMyProducts({ page: 1, page_size: 1 })
    productTotal.value = page?.total || 0
  } catch {
    productTotal.value = 0
  }
})
</script>

<style scoped>
.studio-page { padding: 8px 4px 24px; }
.stat { margin-bottom: 12px; border-radius: 14px; }
.stat .label { color: #64748B; font-size: 13px; }
.stat .num { font-size: clamp(1.35rem, 4vw, 1.75rem); font-weight: 800; margin-top: 6px; }
.stat .num.small { font-size: 18px; }
.mt20 { margin-top: 20px; }
.actions { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
@media (max-width: 768px) {
  .actions .el-button { flex: 1 1 auto; }
}
</style>
