<template>
  <Layout app-name="Cedar-V" :page-title="t('apiKeys.title')">
    <div class="key-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('apiKeys.title') }}</h1>
          <p>{{ t('apiKeys.description') }}</p>
        </div>
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('apiKeys.form.create') }}
        </el-button>
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-input
            v-model="filters.keyword"
            :placeholder="t('apiKeys.keyword')"
            clearable
            class="filter-control"
            @keyup.enter="loadKeys"
          />
          <el-select
            v-model="filters.status"
            :placeholder="t('apiKeys.statusAll')"
            clearable
            class="filter-control status-filter"
          >
            <el-option value="active" :label="t('apiKeys.status.active')" />
            <el-option value="inactive" :label="t('apiKeys.status.inactive')" />
            <el-option value="expired" :label="t('apiKeys.status.expired')" />
          </el-select>
          <div class="filter-actions">
            <el-button plain type="primary" :loading="loading" @click="loadKeys">
              {{ t('rbac.filter.query') }}
            </el-button>
            <el-button @click="resetFilters">{{ t('rbac.filter.reset') }}</el-button>
          </div>
        </div>

        <div class="table-card data-list-table">
          <el-table v-loading="loading" :data="keys" row-key="id">
            <el-table-column prop="name" :label="t('apiKeys.table.name')" min-width="140">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEdit(row)">{{ row.name }}</el-button>
              </template>
            </el-table-column>
            <el-table-column :label="t('apiKeys.table.prefix')" min-width="140">
              <template #default="{ row }">
                <code class="key-prefix">{{ row.api_key_prefix }}…</code>
              </template>
            </el-table-column>
            <el-table-column prop="user_name" :label="t('apiKeys.table.user')" min-width="110" show-overflow-tooltip />
            <el-table-column :label="t('apiKeys.table.status')" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" effect="light" size="small">
                  {{ t(`apiKeys.status.${row.status}`) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('apiKeys.table.expires')" width="150" align="center">
              <template #default="{ row }">
                {{ row.expires_at ? formatDate(row.expires_at) : t('apiKeys.noExpiry') }}
              </template>
            </el-table-column>
            <el-table-column :label="t('apiKeys.table.ips')" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">
                <span v-if="row.allowed_ips?.length">{{ row.allowed_ips.join(', ') }}</span>
                <span v-else class="key-empty">{{ t('apiKeys.noIps') }}</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('apiKeys.table.lastUsed')" width="150" align="center">
              <template #default="{ row }">
                {{ row.last_used_at ? formatDate(row.last_used_at) : t('apiKeys.message.neverUsed') }}
              </template>
            </el-table-column>
            <el-table-column
              :label="t('apiKeys.table.actions')"
              width="160"
              align="center"
              fixed="right"
            >
              <template #default="{ row }">
                <el-button plain size="small" type="primary" @click="openEdit(row)">
                  {{ t('apiKeys.form.editLabel') }}
                </el-button>
                <el-button plain size="small" type="danger" @click="removeKey(row)">
                  {{ t('apiKeys.form.deleteKey') }}
                </el-button>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :description="t('apiKeys.message.empty')" :image-size="72" />
            </template>
          </el-table>

          <div class="pagination-row data-list-pagination">
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.pageSize"
              :page-sizes="[10, 20, 50]"
              :total="pagination.total"
              layout="total, sizes, prev, pager, next"
              :pager-count="5"
              :disabled="loading"
              @size-change="handleSizeChange"
              @current-change="loadKeys"
            />
          </div>
        </div>
      </section>
    </div>

    <!-- 创建/编辑弹框 -->
    <ApiKeyForm
      v-model:visible="formVisible"
      :key-item="currentKey"
      :is-edit="isEdit"
      :users="userOptions"
      @saved="loadKeys"
      @plaintext="showPlaintext"
    />

    <!-- 明文一次性展示 -->
    <el-dialog
      v-model="plaintextVisible"
      :title="t('apiKeys.plaintext.title')"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-alert :title="t('apiKeys.plaintext.warning')" type="warning" show-icon :closable="false" />
      <div class="plaintext-box">
        <code>{{ plaintext }}</code>
      </div>
      <template #footer>
        <el-button @click="plaintextVisible = false">{{ t('apiKeys.cancel') }}</el-button>
        <el-button type="primary" @click="copyPlaintext">
          {{ copied ? t('apiKeys.plaintext.copied') : t('apiKeys.plaintext.copy') }}
        </el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import ApiKeyForm from './ApiKeyForm.vue'
import {
  deleteApiKey,
  getApiKeys,
  type ApiKeyItem
} from '@/api/systemConfig'
import { getUsers, type UserItem } from '@/api/user'
import { getErrorMessage } from '@/utils/error'

const { t } = useI18n()

const keys = ref<ApiKeyItem[]>([])
const userOptions = ref<UserItem[]>([])
const loading = ref(false)

const filters = reactive<{ keyword: string; status: string }>({ keyword: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const formVisible = ref(false)
const currentKey = ref<ApiKeyItem | null>(null)
const isEdit = ref(false)

const plaintextVisible = ref(false)
const plaintext = ref('')
const copied = ref(false)

const statusTagType = (status: string) => {
  if (status === 'active') return 'success'
  if (status === 'expired') return 'danger'
  return 'info'
}

const formatDate = (value: string): string => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const loadKeys = async () => {
  loading.value = true
  try {
    const res = await getApiKeys({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: filters.keyword || undefined,
      status: (filters.status || undefined) as 'active' | 'inactive' | 'expired' | undefined
    })
    keys.value = res.data?.list || []
    pagination.total = res.data?.total ?? 0
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('apiKeys.message.loadFailed')))
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try {
    const res = await getUsers({ page: 1, page_size: 100 })
    userOptions.value = res.data?.list || []
  } catch {
    userOptions.value = []
  }
}

const resetFilters = () => {
  filters.keyword = ''
  filters.status = ''
  pagination.page = 1
  loadKeys()
}

const handleSizeChange = () => {
  pagination.page = 1
  loadKeys()
}

const openCreate = () => {
  currentKey.value = null
  isEdit.value = false
  formVisible.value = true
}

const openEdit = (row: ApiKeyItem) => {
  currentKey.value = row
  isEdit.value = true
  formVisible.value = true
}

/** 明文密钥仅创建时返回，由子组件 emit 触发 */
const showPlaintext = (value: string) => {
  plaintext.value = value
  copied.value = false
  plaintextVisible.value = true
}

const copyPlaintext = async () => {
  try {
    await navigator.clipboard.writeText(plaintext.value)
    copied.value = true
  } catch {
    ElMessage.warning(t('apiKeys.message.loadFailed'))
  }
}

const removeKey = async (row: ApiKeyItem) => {
  try {
    await ElMessageBox.confirm(
      t('apiKeys.confirm.deleteMsg', { name: row.name }),
      t('apiKeys.confirm.delete'),
      {
        confirmButtonText: t('apiKeys.confirm'),
        cancelButtonText: t('apiKeys.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }
  try {
    await deleteApiKey(row.id)
    ElMessage.success(t('apiKeys.message.deleted'))
    loadKeys()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('apiKeys.message.loadFailed')))
  }
}

onMounted(async () => {
  await Promise.all([loadKeys(), loadUsers()])
})
</script>

<style lang="scss" scoped>
.key-prefix {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
  color: var(--app-text-secondary);
}

.key-empty {
  color: var(--app-text-secondary);
  font-size: 12px;
}

.plaintext-box {
  margin-top: 16px;
  padding: 12px;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
  word-break: break-all;

  code {
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 13px;
    color: var(--app-text-primary);
  }
}
</style>