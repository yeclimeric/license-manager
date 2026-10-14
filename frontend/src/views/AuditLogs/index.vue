<template>
  <Layout app-name="Cedar-V" :page-title="t('auditLogs.title')">
    <div class="audit-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('auditLogs.title') }}</h1>
          <p>{{ t('auditLogs.description') }}</p>
        </div>
      </header>

      <section class="data-list-workspace">
        <!-- 统计卡片 -->
        <div v-if="stats" class="stats-row">
          <div class="stats-card">
            <span class="stats-card__label">{{ t('auditLogs.stats.total') }}</span>
            <span class="stats-card__value">{{ stats.total }}</span>
          </div>
          <div class="stats-card stats-card--success">
            <span class="stats-card__label">{{ t('auditLogs.filter.success') }}</span>
            <span class="stats-card__value">{{ stats.success }}</span>
          </div>
          <div class="stats-card stats-card--failure">
            <span class="stats-card__label">{{ t('auditLogs.filter.failure') }}</span>
            <span class="stats-card__value">{{ stats.failure }}</span>
          </div>
        </div>

        <div class="data-list-toolbar">
          <span class="data-list-toolbar__divider" aria-hidden="true"></span>
          <div class="data-list-toolbar__filters">
            <el-alert
              v-if="listError"
              :title="listError"
              type="error"
              show-icon
              :closable="false"
              class="data-list-alert"
            >
              <template #default>
                <el-button link type="primary" @click="loadLogs">
                  {{ t('auditLogs.actions.retry') }}
                </el-button>
              </template>
            </el-alert>

            <div class="filter-grid data-list-filter__grid">
              <el-select
                v-model="filters.resourceType"
                :placeholder="t('auditLogs.filter.allTypes')"
                clearable
              >
                <el-option
                  v-for="option in resourceTypeOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-select
                v-model="filters.status"
                :placeholder="t('auditLogs.filter.allStatus')"
                clearable
              >
                <el-option value="success" :label="t('auditLogs.filter.success')" />
                <el-option value="failure" :label="t('auditLogs.filter.failure')" />
              </el-select>
              <el-input
                v-model="filters.action"
                :placeholder="t('auditLogs.table.action')"
                clearable
                @keyup.enter="handleQuery"
              />
              <div class="filter-actions data-list-filter__actions">
                <el-button plain type="primary" :loading="loading" @click="handleQuery">
                  {{ t('auditLogs.actions.query') }}
                </el-button>
                <el-button @click="handleReset">{{ t('auditLogs.actions.reset') }}</el-button>
              </div>
            </div>
          </div>
        </div>

        <div class="table-card data-list-table">
          <div class="table-scroll data-list-table__scroll">
            <el-table
              v-loading="skeletonVisible"
              class="content-skeleton"
              :class="{ 'is-table-skeleton': skeletonVisible, 'is-table-updating': isUpdating }"
              :data="logs"
              row-key="id"
            >
              <el-table-column
                prop="action"
                :label="t('auditLogs.table.action')"
                min-width="240"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <el-button link type="primary" class="audit-link" @click="openDetail(row)">
                    {{ row.action }}
                  </el-button>
                </template>
              </el-table-column>
              <el-table-column
                prop="resource_type"
                :label="t('auditLogs.table.resource')"
                min-width="150"
                show-overflow-tooltip
              />
              <el-table-column
                prop="status"
                :label="t('auditLogs.table.status')"
                width="100"
                align="center"
              >
                <template #default="{ row }">
                  <el-tag :type="row.status === 'success' ? 'success' : 'danger'" effect="light">
                    {{ row.status === 'success' ? t('auditLogs.filter.success') : t('auditLogs.filter.failure') }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                :label="t('auditLogs.table.operator')"
                width="130"
                align="center"
              >
                <template #default="{ row }">
                  {{ row.username || row.user_id || t('auditLogs.none') }}
                </template>
              </el-table-column>
              <el-table-column prop="ip" :label="t('auditLogs.table.ip')" width="140" align="center" />
              <el-table-column
                :label="t('auditLogs.table.time')"
                width="180"
                align="center"
              >
                <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
              </el-table-column>
              <template #empty>
                <el-empty :description="t('auditLogs.empty')" :image-size="72" />
              </template>
            </el-table>
          </div>

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
              @current-change="loadLogs"
            />
          </div>
        </div>
      </section>
    </div>

    <!-- 详情弹框 -->
    <el-dialog
      v-model="detailVisible"
      :title="t('auditLogs.detail.title')"
      width="720px"
    >
      <el-descriptions v-if="current" :column="1" border size="small">
        <el-descriptions-item :label="t('auditLogs.table.action')">
          {{ current.action }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('auditLogs.table.resource')">
          {{ current.resource_type || t('auditLogs.none') }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('auditLogs.table.resourceId')">
          {{ current.resource_id || t('auditLogs.none') }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('auditLogs.table.operator')">
          {{ current.username || current.user_id || t('auditLogs.none') }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('auditLogs.table.ip')">
          {{ current.ip || t('auditLogs.none') }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('auditLogs.table.time')">
          {{ formatDateTime(current.created_at) }}
        </el-descriptions-item>
      </el-descriptions>

      <section v-if="current?.detail?.req_body" class="detail-block">
        <h3 class="detail-block__title">{{ t('auditLogs.detail.reqBody') }}</h3>
        <pre class="detail-block__body">{{ prettyJson(current.detail.req_body) }}</pre>
      </section>

      <section v-if="current?.detail?.resp_id" class="detail-block">
        <h3 class="detail-block__title">{{ t('auditLogs.detail.respId') }}</h3>
        <code class="detail-block__inline">{{ current.detail.resp_id }}</code>
      </section>

      <section v-if="current?.detail?.resp_body" class="detail-block">
        <h3 class="detail-block__title">{{ t('auditLogs.detail.respBody') }}</h3>
        <pre class="detail-block__body">{{ prettyJson(current.detail.resp_body) }}</pre>
      </section>

      <template #footer>
        <el-button @click="detailVisible = false">{{ t('auditLogs.close') }}</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Layout from '@/components/common/layout/Layout.vue'
import {
  getAuditLogs,
  getAuditStatistics,
  type AuditLog,
  type AuditStatistics,
  type AuditStatus
} from '@/api/auditLog'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

const { t } = useI18n()

const logs = ref<AuditLog[]>([])
const stats = ref<AuditStatistics | null>(null)
const loading = ref(false)
const listError = ref('')
const detailVisible = ref(false)
const current = ref<AuditLog | null>(null)
const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)

const filters = reactive({ resourceType: '', status: '', action: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

// 资源类型选项：优先取统计数据里出现过的类型
const resourceTypeOptions = computed(() => {
  const fromStats = Object.keys(stats.value?.by_resource || {}).sort()
  if (fromStats.length > 0) {
    return fromStats.map((value) => ({ value, label: value }))
  }
  return [
    { value: 'authorization-codes', label: 'authorization-codes' },
    { value: 'licenses', label: 'licenses' },
    { value: 'customers', label: 'customers' },
    { value: 'product-offerings', label: 'product-offerings' },
    { value: 'feature-templates', label: 'feature-templates' }
  ]
})

// 请求/响应体是原始字符串，尝试格式化为易读 JSON
const prettyJson = (raw: string): string => {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

// 后端时间戳带时区偏移，前端需显式解析
const formatDateTime = (value: string): string => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const loadStats = async () => {
  try {
    const res = await getAuditStatistics()
    stats.value = res.data || null
  } catch {
    stats.value = null
  }
}

const loadLogs = async () => {
  loading.value = true
  listError.value = ''
  try {
    const res = await getAuditLogs({
      page: pagination.page,
      page_size: pagination.pageSize,
      resource_type: filters.resourceType || undefined,
      status: (filters.status || undefined) as AuditStatus | undefined,
      action: filters.action.trim() || undefined,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    logs.value = res.data.list
    pagination.total = res.data.total
    markInitialized()
  } catch (error) {
    listError.value = getErrorMessage(error, t('auditLogs.empty'))
  } finally {
    loading.value = false
  }
}

const handleQuery = () => {
  pagination.page = 1
  loadLogs()
}

const handleReset = () => {
  filters.resourceType = ''
  filters.status = ''
  filters.action = ''
  pagination.page = 1
  loadLogs()
}

const handleSizeChange = () => {
  pagination.page = 1
  loadLogs()
}

const openDetail = (row: AuditLog) => {
  current.value = row
  detailVisible.value = true
}

onMounted(async () => {
  await Promise.all([loadStats(), loadLogs()])
})
</script>

<style lang="scss" scoped>
/* 操作类型 + 状态 + 关键字 + 操作按钮，与共享默认列一致 */
.filter-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr)) auto;
}

.stats-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.stats-card {
  flex: 1;
  min-width: 140px;
  padding: 14px 18px;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
  display: flex;
  flex-direction: column;
  gap: 4px;

  &__label {
    font-size: 12px;
    color: var(--app-text-secondary);
  }

  &__value {
    font-size: 22px;
    font-weight: 600;
    color: var(--app-text-primary);
  }

  &--success {
    border-left: 3px solid var(--el-color-success);
  }

  &--failure {
    border-left: 3px solid var(--el-color-danger);
  }
}

.audit-link {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
}

.detail-block {
  margin-top: 16px;

  &__title {
    margin: 0 0 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--app-text-primary);
  }

  &__body {
    margin: 0;
    padding: 10px 12px;
    max-height: 200px;
    overflow: auto;
    background: var(--app-content-bg);
    border: 1px solid var(--app-border-light);
    border-radius: $border-radius-base;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 12px;
    line-height: 1.6;
    color: var(--app-text-primary);
    white-space: pre-wrap;
    word-break: break-all;
  }

  &__inline {
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 12px;
    color: var(--app-text-secondary);
  }
}
</style>