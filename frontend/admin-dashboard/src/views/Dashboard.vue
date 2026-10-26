<template>
  <el-container class="d">
    <app-aside />
    <el-container>
      <app-header crumb="数据概览" />
      <el-main>
        <el-row :gutter="16">
          <el-col v-for="s in stats" :key="s.name" :xs="12" :sm="12" :md="6">
            <el-card shadow="hover" class="stat" :body-style="{padding:'20px'}">
              <div class="label">{{ s.label }}</div>
              <div class="num">{{ s.value }}</div>
            </el-card>
          </el-col>
        </el-row>
        <el-row :gutter="16" style="margin-top:20px;">
          <el-col :xs="24" :md="12"><el-card><template #header>近期举报</template><el-empty description="暂无举报" /></el-card></el-col>
          <el-col :xs="24" :md="12"><el-card><template #header>近 30 天新增</template><el-empty description="P5 阶段实现图表" /></el-card></el-col>
        </el-row>
      </el-main>
    </el-container>
  </el-container>
</template>
<script setup lang="ts">
import AppAside from '@/components/AppAside.vue'
import AppHeader from '@/components/AppHeader.vue'
const stats = [
  { name:'u', label:'用户总数', value:0 },
  { name:'f', label:'创业者名片', value:0 },
  { name:'p', label:'产品数', value:0 },
  { name:'r', label:'待处理举报', value:0 },
]
</script>
<style scoped>
.d{ min-height:100vh; }
.stat{ margin-bottom: 12px; border-radius: 14px; }
.stat .label{ color:#64748B; font-size:13px;}
.stat .num{ font-size: clamp(1.4rem, 4vw, 1.75rem); font-weight: 800; margin-top: 4px;}
</style>
