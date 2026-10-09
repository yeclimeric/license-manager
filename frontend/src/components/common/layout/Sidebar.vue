<template>
  <aside class="sidebar" :class="{ 'sidebar--collapsed': isCollapsed, 'sidebar--mobile-open': isMobileOpen }">
    <div class="sidebar__brand" aria-label="Cedar-V">
      <svg class="brand-logo" viewBox="0 0 41 40" fill="none" aria-hidden="true">
        <path d="M26.3125 11.4814L22.25 19.5947V22.7148L27.1191 13.0576L29.7393 18.1777L18.7988 40H0L13.5938 22.8037H14L13.8125 23.1201L7.46875 33.541H17.8438V16.9111L7.1875 25.6475L18.0312 10.1406L11.625 14.4463V14.2588L20.4375 0L26.3125 11.4814Z" fill="#019C7C"/>
        <path d="M34.5498 39.9996H28.75L24.5938 32.8864L27.125 27.6246L34.5498 39.9996ZM41 39.9996H36.2705L27.9346 25.941L30.7188 20.1559L41 39.9996Z" fill="#146B59"/>
      </svg>
      <span v-if="!isCollapsed || appStore.isMobile" class="brand-name">Cedar-V</span>
    </div>
    <nav class="sidebar__nav" :aria-label="t('navigation.mainNavigation')">
      <el-tooltip
        v-for="item in navItems"
        :key="item.id"
        :content="item.label"
        placement="right"
        :disabled="!isCollapsed || appStore.isMobile"
      >
        <button
          type="button"
          class="nav-item"
          :class="{ 'nav-item--active': item.active }"
          :aria-current="item.active ? 'page' : undefined"
          @click="$emit('navClick', item, $event)"
        >
          <span class="nav-icon"><SidebarIcon v-if="item.icon" :name="item.icon" :active="item.active" /></span>
          <span v-if="!isCollapsed || appStore.isMobile" class="nav-label">{{ item.label }}</span>
        </button>
      </el-tooltip>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/store/modules/app'
import SidebarIcon from '@/components/common/icons/SidebarIcon.vue'

interface NavItem {
  id: string
  label: string
  href: string
  icon?: string
  active?: boolean
}

defineProps<{ navItems: NavItem[] }>()
defineEmits<{ navClick: [item: NavItem, event: Event] }>()
const { t } = useI18n()
const appStore = useAppStore()
const isCollapsed = computed(() => appStore.sidebarCollapsed)
const isMobileOpen = computed(() => appStore.isMobile && !appStore.sidebarCollapsed)
</script>

<style lang="scss" scoped>
.sidebar {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 2000;
  width: var(--layout-sidebar-width);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--app-sidebar-bg);
  border-right: 1px solid var(--app-border-light);
  transition: width 0.2s ease, transform 0.2s ease;
}
.sidebar--collapsed { width: var(--layout-sidebar-collapsed-width); }
.sidebar__brand {
  height: var(--layout-header-height);
  padding: 0 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--app-border-light);
}
.brand-logo { width: 34px; height: 34px; flex-shrink: 0; }
.brand-name { font-family: 'Swis721 BlkCn BT', sans-serif; font-size: 26px; color: var(--app-text-primary); white-space: nowrap; }
.sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: 18px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.nav-item {
  position: relative;
  width: 100%;
  min-height: 48px;
  padding: 0 15px;
  display: flex;
  align-items: center;
  gap: 13px;
  border: 0;
  border-radius: $border-radius-base;
  background: transparent;
  color: #586760;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  font-size: 15px;
  font-weight: 500;
  line-height: 1;
  transition: background-color 0.18s ease, color 0.18s ease;
}
.nav-item::before { content:''; position:absolute; inset:8px auto 8px 0; width:3px; background:#019c7c; opacity:0; transform:scaleY(.35); transition:opacity .18s ease,transform .18s ease; }
.nav-item:hover { background:#f2f8f6; color:#136b59; }
.nav-item:focus-visible { outline:2px solid rgba(1,156,124,.45); outline-offset:-2px; }
.nav-item--active { background:#eaf6f2; color:#017c63; font-weight:600; }
.nav-item--active::before { opacity:1; transform:scaleY(1); }
.nav-icon { width:22px; height:22px; display:inline-flex; align-items:center; justify-content:center; flex-shrink:0; color:#78877f; transition:color .18s ease,transform .18s ease; }
.nav-item:hover .nav-icon { color:#16816b; transform:translateX(1px); }
.nav-item--active .nav-icon { color:#019c7c; }
.nav-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sidebar--collapsed .nav-item { justify-content: center; padding: 0; }
[data-theme="dark"] .nav-item { color:#bec9c4; }
[data-theme="dark"] .nav-item:hover { color:#86dfc2; background:rgba(1,156,124,.12); }
[data-theme="dark"] .nav-item--active { color:#79e2bf; background:rgba(1,156,124,.18); }
@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar { width: var(--layout-sidebar-collapsed-width); }
  .sidebar__brand { padding: 0; }
  .brand-name, .nav-label { display: none; }
  .nav-item { justify-content: center; padding: 0; }
}
@media (max-width: 767px) {
  .sidebar {
    width: min(82vw, 300px);
    transform: translateX(-100%);
    box-shadow: 12px 0 30px rgba(15, 23, 42, 0.16);
  }
  .sidebar--collapsed { width: min(82vw, 300px); }
  .sidebar--mobile-open { transform: translateX(0); }
  .sidebar__brand { justify-content: flex-start; }
  .sidebar--collapsed .nav-item { justify-content: flex-start; padding: 0 14px; }
}
</style>
