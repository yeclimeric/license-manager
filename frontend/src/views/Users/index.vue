<template>
  <Layout app-name="Cedar-V" :page-title="t('users.title')">
    <div class="user-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('users.title') }}</h1>
          <p>{{ t('users.description') }}</p>
        </div>
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('users.actions.create') }}
        </el-button>
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-input
            v-model="filters.search"
            :placeholder="t('users.filter.search')"
            clearable
            class="filter-control"
            @keyup.enter="loadUsers"
          />
          <el-select
            v-model="filters.status"
            :placeholder="t('users.filter.allStatus')"
            clearable
            class="filter-control status-filter"
          >
            <el-option value="active" :label="t('users.status.active')" />
            <el-option value="locked" :label="t('users.status.locked')" />
            <el-option value="inactive" :label="t('users.status.inactive')" />
          </el-select>
          <div class="filter-actions">
            <el-button plain type="primary" :loading="loading" @click="loadUsers">
              {{ t('users.filter.query') }}
            </el-button>
            <el-button @click="resetFilters">{{ t('users.filter.reset') }}</el-button>
          </div>
        </div>

        <div class="table-card data-list-table">
          <el-table v-loading="loading" :data="users" row-key="id">
            <el-table-column prop="username" :label="t('users.table.username')" min-width="130">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEdit(row)">{{ row.username }}</el-button>
              </template>
            </el-table-column>
            <el-table-column prop="full_name" :label="t('users.table.fullName')" min-width="120" show-overflow-tooltip />
            <el-table-column prop="email" :label="t('users.table.email')" min-width="200" show-overflow-tooltip />
            <el-table-column :label="t('users.table.roles')" min-width="200">
              <template #default="{ row }">
                <div v-if="row.roles?.length" class="role-tags">
                  <el-tag
                    v-for="role in row.roles"
                    :key="role.id"
                    size="small"
                    effect="plain"
                    type="info"
                  >
                    {{ role.display_name || role.name }}
                  </el-tag>
                </div>
                <span v-else class="user-empty">{{ t('users.empty.roles') }}</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('users.table.status')" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" effect="light" size="small">
                  {{ t(`users.status.${row.status}`) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('users.table.lastLogin')" min-width="160" align="center">
              <template #default="{ row }">
                {{ row.last_login_at ? formatDate(row.last_login_at) : t('users.empty.neverLogin') }}
              </template>
            </el-table-column>
            <el-table-column
              :label="t('users.table.actions')"
              width="230"
              align="center"
              fixed="right"
            >
              <template #default="{ row }">
                <el-button plain size="small" type="primary" @click="openEdit(row)">
                  {{ t('users.actions.edit') }}
                </el-button>
                <el-button plain size="small" @click="openReset(row)">
                  {{ t('users.actions.resetPassword') }}
                </el-button>
                <el-button plain size="small" type="danger" @click="removeUser(row)">
                  {{ t('users.actions.deleteUser') }}
                </el-button>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :description="t('users.message.empty')" :image-size="72" />
            </template>
          </el-table>

          <div class="pagination-row data-list-pagination">
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="pagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              :pager-count="5"
              :disabled="loading"
              @size-change="handleSizeChange"
              @current-change="loadUsers"
            />
          </div>
        </div>
      </section>
    </div>

    <!-- 用户编辑弹框 -->
    <UserForm
      v-model:visible="formVisible"
      :user="currentUser"
      :is-edit="isEdit"
      :roles="roleOptions"
      @saved="loadUsers"
    />

    <!-- 重置密码弹框 -->
    <el-dialog
      v-model="resetVisible"
      :title="t('users.actions.resetPassword')"
      width="420px"
    >
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="110px">
        <el-form-item :label="t('users.form.newPassword')" prop="newPassword">
          <el-input
            v-model="resetForm.newPassword"
            type="password"
            show-password
            :placeholder="t('users.form.passwordHint')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetVisible = false">{{ t('users.cancel') }}</el-button>
        <el-button type="primary" :loading="resetSubmitting" @click="submitReset">
          {{ t('users.save') }}
        </el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import UserForm from './UserForm.vue'
import {
  deleteUser,
  getUsers,
  resetUserPassword,
  type UserItem,
  type UserRoleItem,
  type UserStatus
} from '@/api/user'
import { getRoles } from '@/api/rbac'
import { getErrorMessage } from '@/utils/error'

const { t } = useI18n()

const users = ref<UserItem[]>([])
const roleOptions = ref<UserRoleItem[]>([])
const loading = ref(false)

const filters = reactive<{ search: string; status: UserStatus | '' }>({
  search: '',
  status: ''
})
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const formVisible = ref(false)
const currentUser = ref<UserItem | null>(null)
const isEdit = ref(false)

const resetVisible = ref(false)
const resetTarget = ref<UserItem | null>(null)
const resetSubmitting = ref(false)
const resetFormRef = ref<FormInstance>()
const resetForm = reactive({ newPassword: '' })

const statusTagType = (status: UserStatus) => {
  if (status === 'active') return 'success'
  if (status === 'locked') return 'warning'
  return 'info'
}

const formatDate = (value: string): string => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers({
      page: pagination.page,
      page_size: pagination.pageSize,
      search: filters.search || undefined,
      status: filters.status || undefined
    })
    users.value = res.data?.list || []
    pagination.total = res.data?.total ?? 0
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('users.message.loadFailed')))
  } finally {
    loading.value = false
  }
}

const loadRoles = async () => {
  try {
    const res = await getRoles()
    roleOptions.value = res.data?.list || []
  } catch {
    roleOptions.value = []
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  pagination.page = 1
  loadUsers()
}

const handleSizeChange = () => {
  pagination.page = 1
  loadUsers()
}

const openCreate = () => {
  currentUser.value = null
  isEdit.value = false
  formVisible.value = true
}

const openEdit = (row: UserItem) => {
  currentUser.value = row
  isEdit.value = true
  formVisible.value = true
}

const openReset = (row: UserItem) => {
  resetTarget.value = row
  resetForm.newPassword = ''
  resetVisible.value = true
}

const resetRules: FormRules = {
  newPassword: [
    { required: true, message: t('users.validation.newPassword'), trigger: 'blur' },
    { min: 8, message: t('users.validation.passwordLength'), trigger: 'blur' }
  ]
}

const submitReset = async () => {
  const valid = await resetFormRef.value?.validate().catch(() => false)
  if (!valid || !resetTarget.value) return

  resetSubmitting.value = true
  try {
    await resetUserPassword(resetTarget.value.id, resetForm.newPassword)
    ElMessage.success(t('users.message.passwordReset'))
    resetVisible.value = false
    loadUsers()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('users.message.loadFailed')))
  } finally {
    resetSubmitting.value = false
  }
}

const removeUser = async (row: UserItem) => {
  try {
    await ElMessageBox.confirm(
      t('users.confirm.deleteUserMsg', { name: row.username }),
      t('users.confirm.deleteUser'),
      {
        confirmButtonText: t('users.save'),
        cancelButtonText: t('users.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }
  try {
    await deleteUser(row.id)
    ElMessage.success(t('users.message.userDeleted'))
    loadUsers()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('users.message.loadFailed')))
  }
}

onMounted(async () => {
  await Promise.all([loadUsers(), loadRoles()])
})
</script>

<style lang="scss" scoped>
.role-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.user-empty {
  color: var(--app-text-secondary);
  font-size: 12px;
}
</style>