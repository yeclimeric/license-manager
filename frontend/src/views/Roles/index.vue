<template>
  <Layout app-name="Cedar-V" :page-title="t('rbac.page.roleTitle')">
    <div class="role-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('rbac.page.roleTitle') }}</h1>
        </div>
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-input
            v-model="keyword"
            :placeholder="t('rbac.filter.keyword')"
            clearable
            class="filter-control"
            @keyup.enter="loadRoles"
          />
          <div class="filter-actions">
            <el-button plain type="primary" :loading="loading" @click="loadRoles">
              {{ t('rbac.filter.query') }}
            </el-button>
            <el-button @click="resetSearch">{{ t('rbac.filter.reset') }}</el-button>
          </div>
        </div>

        <div class="table-card data-list-table">
          <el-table v-loading="loading" :data="roles" row-key="id">
            <el-table-column :label="t('rbac.form.displayName')" min-width="180">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEdit(row)">{{ row.display_name }}</el-button>
                <el-tag v-if="row.is_system" size="small" type="info" effect="plain" class="role-tag">
                  {{ t('rbac.form.systemRole') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="name" :label="t('rbac.form.name')" min-width="160">
              <template #default="{ row }">
                <code class="role-key">{{ row.name }}</code>
              </template>
            </el-table-column>
            <el-table-column :label="t('rbac.form.permissions')" min-width="220">
              <template #default="{ row }">
                <span class="role-perm-count">{{ (row.permissions || []).length }}</span>
              </template>
            </el-table-column>
            <el-table-column
              prop="user_count"
              :label="t('rbac.table.users')"
              width="100"
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
                  @click="removeRole(row)"
                >
                  {{ t('featureTemplates.actions.delete') }}
                </el-button>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :description="t('rbac.filter.empty')" :image-size="72" />
            </template>
          </el-table>
        </div>
      </section>
    </div>

    <!-- 角色编辑弹框 -->
    <RoleForm
      v-model:visible="formVisible"
      :role="currentRole"
      :is-edit="isEdit"
      :permissions="matrix"
      :menus="menuTree"
      @saved="loadRoles"
    />
  </Layout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import Layout from '@/components/common/layout/Layout.vue'
import RoleForm from './RoleForm.vue'
import {
  deleteRole,
  getAllMenus,
  getPermissionMatrix,
  getRoles,
  type MenuItem,
  type PermissionResource,
  type Role
} from '@/api/rbac'
import { getErrorMessage } from '@/utils/error'

const { t, locale } = useI18n()

const roles = ref<Role[]>([])
const matrix = ref<PermissionResource[]>([])
const menuTree = ref<MenuItem[]>([])
const keyword = ref('')
const loading = ref(false)
const formVisible = ref(false)
const currentRole = ref<Role | null>(null)
const isEdit = ref(false)

const loadRoles = async () => {
  loading.value = true
  try {
    const res = await getRoles(keyword.value)
    roles.value = res.data?.list || []
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('rbac.message.loadFailed')))
  } finally {
    loading.value = false
  }
}

const loadMatrix = async () => {
  try {
    const res = await getPermissionMatrix()
    matrix.value = res.data?.resources || []
  } catch {
    matrix.value = []
  }
}

const loadMenus = async () => {
  try {
    const res = await getAllMenus()
    menuTree.value = res.data?.list || []
  } catch {
    menuTree.value = []
  }
}

const resetSearch = () => {
  keyword.value = ''
  loadRoles()
}

const openEdit = (row: Role) => {
  currentRole.value = row
  isEdit.value = true
  formVisible.value = true
}

const removeRole = async (row: Role) => {
  if (!row.can_delete) return
  try {
    await ElMessageBox.confirm(
      t('rbac.confirm.deleteRoleMessage', { name: row.display_name }),
      t('rbac.confirm.deleteRole'),
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
    await deleteRole(row.id)
    ElMessage.success(t('rbac.message.deleted'))
    loadRoles()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('rbac.message.loadFailed')))
  }
}

onMounted(async () => {
  // 语言变化时表格内的角色名需要重新渲染
  locale.value
  await Promise.all([loadRoles(), loadMatrix(), loadMenus()])
})
</script>

<style lang="scss" scoped>
.role-tag {
  margin-left: 8px;
}

.role-key {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
  color: var(--app-text-secondary);
}

.role-perm-count {
  display: inline-block;
  min-width: 28px;
  padding: 2px 8px;
  text-align: center;
  border-radius: 10px;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-light);
  font-size: 12px;
  color: var(--app-text-secondary);
}
</style>
