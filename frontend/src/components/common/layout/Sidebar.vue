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
      <template v-for="entry in navItems" :key="entry.id">
        <!-- 分组目录：仅作标题，展开时才显示子菜单 -->
        <div v-if="entry.children?.length" :key="entry.id" class="nav-group">
          <button
            v-if="!isCollapsed || appStore.isMobile"
            type="button"
            class="nav-group__title"
            :aria-expanded="isGroupOpen(entry.id)"
            @click="toggleGroup(entry.id)"
          >
            <span class="nav-icon">
              <SidebarIcon v-if="entry.icon" :name="entry.icon" />
            </span>
            <span class="nav-group__label">{{ entry.label }}</span>
            <el-icon class="nav-group__arrow" :class="{ 'is-open': isGroupOpen(entry.id) }">
              <ArrowDown />
            </el-icon>
          </button>
          <el-collapse-transition>
            <div v-show="isGroupOpen(entry.id)">
              <el-tooltip
                v-for="child in entry.children"
                :key="child.id"
                :content="child.label"
                placement="right"
                :disabled="!isCollapsed || appStore.isMobile"
              >
                <button
                  type="button"
                  class="nav-item nav-item--nested"
                  :class="{ 'nav-item--active': child.active }"
                  :aria-current="child.active ? 'page' : undefined"
                  @click="$emit('navClick', child, $event)"
                >
                  <span class="nav-icon">
                    <SidebarIcon v-if="child.icon" :name="child.icon" :active="child.active" />
                  </span>
                  <span v-if="!isCollapsed || appStore.isMobile" class="nav-label">{{ child.label }}</span>
                </button>
              </el-tooltip>
            </div>
          </el-collapse-transition>
        </div>

        <el-tooltip
          v-else
          :key="entry.id"
          :content="entry.label"
          placement="right"
          :disabled="!isCollapsed || appStore.isMobile"
        >
          <button
            type="button"
            class="nav-item"
            :class="{ 'nav-item--active': entry.active }"
            :aria-current="entry.active ? 'page' : undefined"
            @click="$emit('navClick', entry, $event)"
          >
            <span class="nav-icon">
              <SidebarIcon v-if="entry.icon" :name="entry.icon" :active="entry.active" />
            </span>
            <span v-if="!isCollapsed || appStore.isMobile" class="nav-label">{{ entry.label }}</span>
          </button>
        </el-tooltip>
      </template>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/store/modules/app'
import SidebarIcon from '@/components/common/icons/SidebarIcon.vue'

interface NavItem {
  id: string
  label: string
  href: string
  icon?: string
  active?: boolean
  /** 分组目录下的子菜单 */
  children?: NavItem[]
}

const props = defineProps<{ navItems: NavItem[] }>()
defineEmits<{ navClick: [item: NavItem, event: Event] }>()
const { t } = useI18n()
const appStore = useAppStore()
const isCollapsed = computed(() => appStore.sidebarCollapsed)
const isMobileOpen = computed(() => appStore.isMobile && !appStore.sidebarCollapsed)

// 分组展开状态。默认展开：子菜单少时常驻可见，少一次点击。
const openGroups = ref<Record<string, boolean>>({})

const isGroupOpen = (id: string) => openGroups.value[id] !== false

// 当前路由所在分组始终展开，否则子菜单不可见
watch(
  () => props.navItems,
  (items) => {
    const next: Record<string, boolean> = { ...openGroups.value }
    items.forEach((item) => {
      if (!item.children?.length) return
      const active = item.children.some((child) => child.active)
      next[item.id] = next[item.id] ?? true
      if (active) next[item.id] = true
    })
    openGroups.value = next
  },
  { immediate: true, deep: true }
)

const toggleGroup = (id: string) => {
  openGroups.value = { ...openGroups.value, [id]: !isGroupOpen(id) }
}
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
/* 分组标题是目录名，不是可点击项。
   缩进与菜单项左侧对齐（同样 15px），否则比子项还浅，看上去像错了一级；
   字号沿用菜单项的 13px，11px 太小。uppercase 对中文无效却会让英文菜单名变形，故去掉。 */
.nav-group__title {
  width: 100%;
  min-height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 15px 4px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: normal;
  color: var(--app-text-secondary);

  &:hover {
    color: var(--app-text-primary);
  }

  &__label {
    /* 占满图标与文字之间的剩余宽度，把箭头推到最右 */
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__arrow {
    font-size: 12px;
    transition: transform 0.2s ease;

    &.is-open {
      transform: rotate(180deg);
    }
  }
}

.nav-item--nested {
  padding-left: 30px;
}
</style>
