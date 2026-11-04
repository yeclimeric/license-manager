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
      <main class="layout-content">
        <!-- 路由级父布局：子页面经 router children 渲染于此，
             布局与侧边栏状态跨页面跳转保持，不再随页面重建 -->
        <router-view />
      </main>
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
import { getPublicConfig } from '@/api/systemConfig'

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

// 菜单树由后端按角色返回（/v1/menus），标签存 i18n 键，切换语言时随之变化。
//
// 接口失败时的兜底只保留公共页。绝不静态兜底 products / feature-templates /
// audit-logs / roles / menus 等管理端菜单：接口失败时无法确认当前角色，
// 展示管理端入口等于把越权可见的菜单主动送给普通用户。后端菜单树才是权威源，
// 接口恢复后会自动按角色重新渲染完整菜单。
const staticFallback: FallbackEntry[] = [
  { id: 'dashboard', href: '/dashboard', icon: 'dashboard', labelKey: 'navigation.menu.dashboard' },
  { id: 'licenses', href: '/licenses', icon: 'licenses', labelKey: 'navigation.menu.licenses' },
  { id: 'customers', href: '/customers', icon: 'customers', labelKey: 'navigation.menu.customers' }
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

const iconNames = new Set([
  'dashboard', 'licenses', 'customers', 'products', 'templates', 'users', 'menu', 'audit', 'settings', 'api-keys'
])

// 后端种子菜单（rbac_seed.go）的 Icon 存的是 Element Plus 图标名。
// 这里做一层显式映射，落到前端 SVG 图标集的键上；
// User 同名用于客户与用户两个菜单，需结合 id 区分。
const epIconMap: Record<string, string> = {
  DataLine: 'dashboard',
  Lock: 'licenses',
  Box: 'products',
  Goods: 'products',
  Files: 'templates',
  Setting: 'settings',
  UserFilled: 'users',
  Menu: 'menu',
  Document: 'audit',
  Key: 'api-keys'
}

const iconFor = (icon: string, id: string): string => {
  if (icon && iconNames.has(icon)) return icon
  if (icon && epIconMap[icon]) {
    // User 图标客户/用户菜单共用，按 id 区分
    if (icon === 'User') return id.includes('customer') ? 'customers' : 'users'
    return epIconMap[icon]
  }
  // 按 id 精确回退。注意所有种子菜单 id 都以 menu- 开头，
  // 必须先匹配更具体的业务词，menu 自身放最后。
  return iconFallback(id)
}

const iconFallback = (id: string): string => {
  if (id.includes('dashboard')) return 'dashboard'
  if (id.includes('license')) return 'licenses'
  if (id.includes('customer')) return 'customers'
  if (id.includes('template')) return 'templates'
  if (id.includes('api-key')) return 'api-keys'
  if (id.includes('setting') || id.includes('config')) return 'settings'
  if (id.includes('audit') || id.includes('log')) return 'audit'
  if (id.includes('role') || id.includes('user')) return 'users'
  if (id.includes('product') || id.includes('offering')) return 'products'
  if (id === 'menu' || id.endsWith('-menu') || id.startsWith('menu-')) return 'menu'
  return 'dashboard'
}

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

// 站点名称来自系统配置 system.site_name（公开接口，未认证也可读）。
// 供侧边栏品牌位展示；失败时保持 store 内置品牌。
const loadSiteName = async () => {
  try {
    const res = await getPublicConfig()
    appStore.setSiteName(res.data?.['system.site_name'] || '')
  } catch {
    // 保持内置品牌
  }
}

onMounted(() => {
  loadMenus()
  loadSiteName()
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
