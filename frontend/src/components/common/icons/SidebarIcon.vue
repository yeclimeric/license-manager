<!--
 * @Author: 13895237362 2205451508@qq.com
 * @Date: 2025-08-12 00:00:00
 * @LastEditors: 13895237362 2205451508@qq.com
 * @LastEditTime: 2025-08-12 00:00:00
 * @FilePath: /frontend/src/components/common/icons/SidebarIcon.vue
 * @Description: 侧边栏图标组件
-->
<template>
  <div class="sidebar-icon" :class="iconClass">
    <component :is="iconComponent" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DashboardIcon from './svg/DashboardIcon.vue'
import CustomersIcon from './svg/CustomersIcon.vue'
import LicensesIcon from './svg/LicensesIcon.vue'
import ProductsIcon from './svg/ProductsIcon.vue'
import AuditIcon from './svg/AuditIcon.vue'
import UserIcon from './svg/UserIcon.vue'

interface Props {
  name: string
  active?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  active: false
})

const iconMap = {
  'dashboard': DashboardIcon,
  'customers': CustomersIcon,
  'licenses': LicensesIcon,
  'products': ProductsIcon,
  'audit': AuditIcon,
  'users': UserIcon,
  'menu': UserIcon
}

const iconComponent = computed(() => {
  return iconMap[props.name as keyof typeof iconMap] || DashboardIcon
})

const iconClass = computed(() => ({
  'sidebar-icon--active': props.active
}))
</script>

<style scoped>
.sidebar-icon {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sidebar-icon :deep(svg) {
  width: 100%;
  height: 100%;
}

.sidebar-icon :deep(svg) { color:inherit; }
.sidebar-icon :deep(path),
.sidebar-icon :deep(circle),
.sidebar-icon :deep(rect),
.sidebar-icon :deep(line),
.sidebar-icon :deep(polyline) { transition:stroke .18s ease,fill .18s ease; }
</style>
