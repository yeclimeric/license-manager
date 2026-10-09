<template>
  <div class="app-layout" :class="{ 'app-layout--collapsed': appStore.sidebarCollapsed }">
    <div
      v-if="appStore.isMobile && !appStore.sidebarCollapsed"
      class="layout-overlay"
      aria-hidden="true"
      @click="closeMobileSidebar"
    />
    <Sidebar :nav-items="navItems" @nav-click="handleNavClick" />
    <div class="layout-main">
      <NavContent />
      <main class="layout-content"><slot /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/store/modules/app'
import Sidebar from './Sidebar.vue'
import NavContent from './NavContent.vue'
import { getUserMenus } from '@/api/rbac'

defineProps<{ appName?: string; pageTitle?: string }>()

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

interface NavItem {
  id: string
  label: string
  href: string
  icon?: string
  active?: boolean
  children?: NavItem[]
}

interface FallbackEntry {
  id: string
  href: string
  icon: string
  labelKey: string
}

// 菜单树由后端按角色返回，标签存 i18n 键，切换语言时随之变化。
// 取不到时退回静态清单：登录页等未认证场景也要能渲染侧边栏。
const staticFallback: FallbackEntry[] = [
  { id: 'dashboard', href: '/dashboard', icon: 'dashboard', labelKey: 'navigation.menu.dashboard' },
  { id: 'licenses', href: '/licenses', icon: 'licenses', labelKey: 'navigation.menu.licenses' },
  { id: 'customers', href: '/customers', icon: 'customers', labelKey: 'navigation.menu.customers' },
  { id: 'products', href: '/products', icon: 'products', labelKey: 'navigation.menu.productOfferings' },
  { id: 'feature-templates', href: '/feature-templates', icon: 'licenses', labelKey: 'navigation.menu.featureTemplates' },
  { id: 'audit-logs', href: '/audit-logs', icon: 'audit', labelKey: 'navigation.menu.auditLogs' },
  { id: 'roles', href: '/roles', icon: 'users', labelKey: 'rbac.page.roleTitle' },
  { id: 'menus', href: '/menus', icon: 'menu', labelKey: 'rbac.page.menuTitle' }
]

interface SourceMenu {
  id: string
  name_key: string
  path: string
  icon: string
  type: 'menu' | 'directory'
  children: SourceMenu[] | null
}

const remoteMenus = ref<SourceMenu[] | null>(null)

const iconFor = (icon: string, id: string): string => {
  if (icon && iconNames.has(icon)) return icon
  // 种子数据使用 Material Symbols 名称，回退到既有图标集
  return id.includes('role') || id.includes('user') ? 'users' : id.includes('menu') ? 'menu' : iconFallback(id)
}

const iconFallback = (id: string): string => {
  if (id.includes('dashboard')) return 'dashboard'
  if (id.includes('license')) return 'licenses'
  if (id.includes('customer')) return 'customers'
  if (id.includes('product')) return 'products'
  if (id.includes('audit') || id.includes('feature')) return 'licenses'
  return 'dashboard'
}

const iconNames = new Set([
  'dashboard', 'licenses', 'customers', 'products', 'users', 'menu', 'audit'
])

// 分组目录本身没有路径，不能作为可点击项，但仍需保留其子菜单。
// 因此过滤条件是"有路径 或 有子菜单"，而不是只看路径。
const toNavItems = (list: SourceMenu[]): NavItem[] =>
  list
    .map((item) => {
      const children = toNavItems(item.children || [])
      if (!item.path && children.length === 0) return null

      const base: NavItem = {
        id: item.id,
        label: t(item.name_key),
        href: item.path,
        icon: iconFor(item.icon, item.id),
        active: !!item.path &&
          (route.path === item.path || route.path.startsWith(`${item.path}/`))
      }
      if (children.length > 0) base.children = children
      return base
    })
    .filter((item): item is NavItem => item !== null)

const sourceMenus = computed<SourceMenu[]>(() => {
  if (remoteMenus.value && remoteMenus.value.length > 0) return remoteMenus.value
  return staticFallback.map((item) => ({
    id: item.id,
    name_key: item.labelKey,
    path: item.href,
    icon: item.icon || '',
    type: 'menu' as const,
    children: null
  }))
})

const navItems = computed(() => toNavItems(sourceMenus.value))

// 拉取用户菜单。失败时保留静态清单，不影响其他页面使用。
const loadMenus = async () => {
  try {
    const res = await getUserMenus()
    if (res.data?.list?.length) {
      remoteMenus.value = res.data.list as SourceMenu[]
    }
  } catch {
    remoteMenus.value = null
  }
}

onMounted(() => {
  loadMenus()
})

const closeMobileSidebar = () => appStore.setSidebarCollapsed(true)
const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && appStore.isMobile && !appStore.sidebarCollapsed) {
    closeMobileSidebar()
  }
}
const handleNavClick = async (item: NavItem) => {
  if (route.path !== item.href) await router.push(item.href)
  if (appStore.isMobile) closeMobileSidebar()
}

watch(
  () => appStore.isMobile && !appStore.sidebarCollapsed,
  isOpen => document.body.classList.toggle('sidebar-drawer-open', isOpen),
  { immediate: true }
)
onMounted(() => window.addEventListener('keydown', handleKeydown))
onUnmounted(() => {
  document.body.classList.remove('sidebar-drawer-open')
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style lang="scss" scoped>
.app-layout {
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: var(--app-bg-color);
}
.layout-main {
  height: 100vh;
  margin-left: var(--layout-sidebar-width);
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: margin-left 0.2s ease;
}
.app-layout--collapsed .layout-main { margin-left: var(--layout-sidebar-collapsed-width); }
.layout-content {
  flex: 1;
  min-height: 0;
  padding-top: var(--layout-header-height);
  overflow: auto;
  background: var(--app-bg-color);
}
.layout-overlay {
  position: fixed;
  inset: 0;
  z-index: 1998;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
}
@media (min-width: 768px) and (max-width: 1023px) {
  .layout-main, .app-layout--collapsed .layout-main { margin-left: var(--layout-sidebar-collapsed-width); }
}
@media (max-width: 767px) {
  .layout-main, .app-layout--collapsed .layout-main { margin-left: 0; }
}
@media print {
  .layout-main { margin-left: 0 !important; }
  .layout-content { padding-top: 0; overflow: visible; }
}
</style>

<style>
body.sidebar-drawer-open { overflow: hidden; }
</style>
