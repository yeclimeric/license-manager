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
import AuditIcon from './svg/AuditIcon.vue'
import CustomerIcon from './svg/CustomerIcon.vue'
import DashboardIcon from './svg/DashboardIcon.vue'
import KeyIcon from './svg/KeyIcon.vue'
import LicensesIcon from './svg/LicensesIcon.vue'
import MenuIcon from './svg/MenuIcon.vue'
import OfferingIcon from './svg/OfferingIcon.vue'
import RoleIcon from './svg/RoleIcon.vue'
import SettingIcon from './svg/SettingIcon.vue'
import TemplateIcon from './svg/TemplateIcon.vue'

interface Props {
  name: string
  active?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  active: false
})

/**
 * 图标映射。
 *
 * 键必须与 Layout.vue 中 iconFor / iconFallback 实际产出的名字一致
 * （dashboard / licenses / customers / products / templates / users /
 * menu / audit / settings / api-keys），区分大小写。之前这里用的是
 * DataLine/Lock/Menu 等 Element Plus 风格名，与上游完全对不上，
 * 导致所有菜单都回退到 DashboardIcon，图标全一样。
 * 新增菜单时按这里的键加一项。
 */
const iconMap: Record<string, unknown> = {
  dashboard: DashboardIcon,
  licenses: LicensesIcon,
  customers: CustomerIcon,
  products: OfferingIcon,
  templates: TemplateIcon,
  users: RoleIcon,
  menu: MenuIcon,
  audit: AuditIcon,
  settings: SettingIcon,
  'api-keys': KeyIcon
}
const iconComponent = computed(() => {
  return iconMap[props.name] || DashboardIcon
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
