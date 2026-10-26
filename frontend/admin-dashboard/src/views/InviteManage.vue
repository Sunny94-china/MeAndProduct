<template>
  <el-container class="d">
    <app-aside />
    <el-container>
      <app-header crumb="邀请码发放" />
      <el-main>
        <div class="bar">
          <el-input-number v-model="n" :min="1" :max="100" label="数量" />
          <el-select v-model="max" label="次数" style="width:140px"><el-option label="1 次" :value="1" /><el-option label="多次" :value="10" /></el-select>
          <el-button type="primary">批量生成邀请码</el-button>
        </div>
        <el-empty description="P5 阶段实现" />
      </el-main>
    </el-container>
  </el-container>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import AppAside from '@/components/AppAside.vue'
import AppHeader from '@/components/AppHeader.vue'
const n = ref(10); const max = ref(1)
</script>
<style scoped>.d{ min-height:100vh;} .bar{ margin-bottom:12px; display:flex; gap:10px; align-items:center;}</style>
