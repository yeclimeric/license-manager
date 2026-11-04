<template>
  <div class="forbidden-page">
    <el-result icon="warning" title="403" sub-title="当前账号没有访问该页面的权限">
      <template #extra>
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="goBack">返回上一页</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElResult, ElButton } from 'element-plus'

const router = useRouter()

const goHome = () => {
  void router.replace('/dashboard')
}

const goBack = () => {
  if (window.history.length > 1) router.back()
  else void router.replace('/dashboard')
}
</script>

<style scoped>
.forbidden-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-bg-color-page, #f5f7fa);
}
</style>
