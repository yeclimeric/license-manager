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
import TemplateIcon from './svg/TemplateIcon.vue'
import RoleIcon from './svg/RoleIcon.vue'
import SettingIcon from './svg/SettingIcon.vue'

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
 * 键是后端 menus.icon 的取值，不是页面路径：后端下发的是一串
 * Element Plus 风格的图标名，用路径名当键会全部落空，只剩一个空占位。
 * 新增菜单时按这里的键加一项，不要改成路径。
 */
const iconMap: Record<string, unknown> = {
  DataLine: DashboardIcon,
  Lock: LicensesIcon,
  User: CustomerIcon,
  Box: OfferingIcon,
  Files: TemplateIcon,
  Goods: OfferingIcon,
  Setting: SettingIcon,
  UserFilled: RoleIcon,
  Menu: MenuIcon,
  Document: AuditIcon,
  Key: KeyIcon
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
