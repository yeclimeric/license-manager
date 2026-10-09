<template>
  <Layout app-name="Cedar-V" :page-title="t('rbac.page.menuTitle')">
    <div class="menu-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('rbac.page.menuTitle') }}</h1>
        </div>
      </header>

      <section class="data-list-workspace">
        <el-alert
          v-if="loadError"
          :title="loadError"
          type="error"
          show-icon
          :closable="false"
          class="menu-alert"
        />

        <div class="table-card data-list-table">
          <el-table
            v-loading="loading"
            :data="tree"
            row-key="id"
            :default-expand-all="true"
            :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
            class="menu-table"
          >
            <el-table-column :label="t('rbac.table.name')" min-width="200">
              <template #default="{ row }">
                <span class="menu-name">{{ t(row.name_key) }}</span>
                <el-tag v-if="row.is_system" size="small" type="info" effect="plain" class="menu-tag">
                  {{ t('rbac.form.systemMenu') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="name_key" :label="t('rbac.table.nameKey')" min-width="200">
              <template #default="{ row }">
                <code class="menu-key">{{ row.name_key }}</code>
              </template>
            </el-table-column>
            <el-table-column :label="t('rbac.table.type')" width="110" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="row.type === 'directory' ? 'warning' : ''" effect="plain">
                  {{ row.type === 'directory' ? t('rbac.type.directory') : t('rbac.type.menu') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="path" :label="t('rbac.table.path')" min-width="170">
              <template #default="{ row }">
                <code v-if="row.path" class="menu-path">{{ row.path }}</code>
                <span v-else class="menu-empty">—</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('rbac.table.permission')" min-width="180">
              <template #default="{ row }">
                <code v-if="row.permission" class="menu-perm">{{ row.permission }}</code>
                <span v-else class="menu-empty">—</span>
              </template>
            </el-table-column>
            <el-table-column
              prop="sort_order"
              :label="t('rbac.table.sortOrder')"
              width="90"
              align="center"
            />
            <el-table-column
              :label="t('rbac.table.actions')"
              width="160"
              align="center"
              fixed="right"
            >
              <template #default="{ row }">
                <el-button plain size="small" type="primary" @click="openEdit(row)">
                  {{ t('featureTemplates.actions.edit') }}
                </el-button>
                <el-button
                  plain
                  size="small"
                  type="danger"
                  :disabled="!row.can_delete"
                  @click="removeMenu(row)"
                >
                  {{ t('featureTemplates.actions.delete') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </section>
    </div>

    <!-- 菜单编辑弹框 -->
    <MenuForm
      v-model:visible="formVisible"
      :menu="currentMenu"
      :is-edit="isEdit"
      :parents="parentOptions"
      @saved="loadMenus"
    />
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import Layout from '@/components/common/layout/Layout.vue'
import MenuForm from './MenuForm.vue'
import { deleteMenu, getAllMenus, type MenuItem } from '@/api/rbac'
import { getErrorMessage } from '@/utils/error'

const { t } = useI18n()

const menus = ref<MenuItem[]>([])
const loading = ref(false)
const loadError = ref('')
const formVisible = ref(false)
const currentMenu = ref<MenuItem | null>(null)
const isEdit = ref(false)

// 直接以顶层菜单作为行：包一层虚拟根节点会多出一条空行
const tree = computed(() => menus.value)

// 上级菜单候选项：仅分组目录可作为父级
const parentOptions = computed(() => {
  const options: { id: string; label: string }[] = []
  const walk = (list: MenuItem[]) => {
    list.forEach((item) => {
      if (item.type === 'directory') {
        options.push({ id: item.id, label: t(item.name_key) })
      }
      if (item.children?.length) walk(item.children)
    })
  }
  walk(menus.value)
  return options
})

const loadMenus = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const res = await getAllMenus()
    menus.value = res.data?.list || []
  } catch (error) {
    loadError.value = getErrorMessage(error, t('rbac.message.menuLoadFailed'))
  } finally {
    loading.value = false
  }
}

const openEdit = (row: MenuItem) => {
  currentMenu.value = row
  isEdit.value = true
  formVisible.value = true
}

const removeMenu = async (row: MenuItem) => {
  if (!row.can_delete) return
  try {
    await ElMessageBox.confirm(
      t('rbac.confirm.deleteMenuMessage', { name: t(row.name_key) }),
      t('rbac.confirm.deleteMenu'),
      {
        confirmButtonText: t('featureTemplates.actions.confirm'),
        cancelButtonText: t('featureTemplates.actions.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }
  try {
    await deleteMenu(row.id)
    ElMessage.success(t('rbac.message.menuDeleted'))
    loadMenus()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('rbac.message.menuLoadFailed')))
  }
}

onMounted(loadMenus)
</script>

<style lang="scss" scoped>
.menu-alert {
  margin-bottom: 12px;
}

.menu-name {
  font-weight: 500;
}

.menu-tag {
  margin-left: 8px;
}

.menu-key,
.menu-path,
.menu-perm {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
  color: var(--app-text-secondary);
}

.menu-empty {
  color: var(--app-text-secondary);
}

// 内置树的根节点不渲染
.menu-table :deep(.el-table__row > td:first-child) {
  padding-left: 0;
}
</style>
