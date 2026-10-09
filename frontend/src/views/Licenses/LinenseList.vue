<template>
  <div class="license-page data-list-page">
    <header class="data-list-header">
      <div>
        <button v-if="hasCustomerContext" type="button" class="context-back" @click="backToCustomers">
          <el-icon><ArrowLeft /></el-icon>
          <span>{{ t('pages.licenses.list.actions.backToCustomers') }}</span>
        </button>
        <h1>{{ pageTitle }}</h1>
        <p>{{ t('pages.licenses.list.description') }}</p>
      </div>
    </header>

    <section class="data-list-workspace">
      <div class="data-list-toolbar">
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('pages.licenses.list.createLicense') }}
        </el-button>
        <span class="data-list-toolbar__divider" aria-hidden="true"></span>
        <div class="filter-grid data-list-filter__grid data-list-toolbar__filters">
          <el-input
            v-model="filters.code"
            class="filter-control code-filter"
            :placeholder="t('pages.licenses.list.filter.codePlaceholder')"
            clearable
            @keyup.enter="queryLicenses"
          />
          <el-select
            v-model="filters.customerId"
            class="filter-control customer-filter"
            :placeholder="t('pages.licenses.list.filter.customerPlaceholder')"
            :loading="customerLoading"
            clearable
            filterable
          >
            <el-option
              v-for="customer in customerOptions"
              :key="customer.id"
              :label="customer.name"
              :value="customer.id"
            />
          </el-select>
          <el-select
            v-model="filters.status"
            class="filter-control status-filter"
            :placeholder="t('pages.licenses.list.filter.statusPlaceholder')"
            clearable
          >
            <el-option
              v-for="option in statusOptions"
              :key="option.key"
              :label="option.display"
              :value="option.key"
            />
          </el-select>
          <el-select
            v-model="filters.productOfferingId"
            class="filter-control product-filter"
            :placeholder="t('pages.licenses.filter.productOffering')"
            :loading="productLoading"
            clearable
          >
            <el-option
              :label="t('pages.licenses.filter.noProductOffering')"
              value="none"
            />
            <el-option
              v-for="item in productOptions"
              :key="item.id"
              :label="item.alias || item.name"
              :value="item.id"
            />
          </el-select>
          <el-select
            v-model="filters.deploymentType"
            class="filter-control deployment-filter"
            :placeholder="t('pages.licenses.filter.deploymentType')"
            clearable
          >
            <el-option
              v-for="option in deploymentTypeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
          <div class="filter-actions data-list-filter__actions">
            <el-button plain type="primary" :loading="isUpdating" @click="queryLicenses">
              {{ t('pages.licenses.list.filter.query') }}
            </el-button>
            <el-button @click="resetFilters">
              {{ t('pages.licenses.list.filter.reset') }}
            </el-button>
          </div>
        </div>
      </div>

      <div class="list-card data-list-table">
      <div v-if="loadError" class="load-error data-list-alert">
        <span>{{ loadError }}</span>
        <el-button link type="primary" @click="loadLicenses">
          {{ t('pages.licenses.list.actions.retry') }}
        </el-button>
      </div>

      <div class="data-list-table__scroll">
        <el-table
        :data="licenses"
        v-loading="skeletonVisible"
        class="content-skeleton"
        :class="{ 'is-table-skeleton': skeletonVisible, 'is-table-updating': isUpdating }"
        :element-loading-text="t('pages.licenses.list.table.loading')"
        empty-text=" "
      >
        <el-table-column prop="code" :label="t('pages.licenses.list.table.code')" width="280">
          <template #default="{ row }">
            <div class="license-code-cell">
              <button class="code-link" type="button" @click="openDetail(row)">
                {{ row.code }}
              </button>
              <el-tooltip :content="t('pages.licenses.detail.actions.copyCode')" placement="top">
                <el-button
                  link
                  type="primary"
                  :icon="CopyDocument"
                  class="copy-code-button"
                  :aria-label="t('pages.licenses.detail.actions.copyCode')"
                  @click.stop="copyAuthorizationCode(row.code)"
                />
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('pages.licenses.table.productOffering')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span v-if="row.product_offering_name">{{ row.product_offering_name }}</span>
            <span v-else class="list-empty-cell">{{ t('pages.licenses.table.noProductOffering') }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('pages.licenses.table.deploymentType')"
          width="110"
          align="center"
        >
          <template #default="{ row }">
            <el-tag v-if="row.deployment_type" size="small" effect="plain" :type="row.deployment_type === 'standalone' ? 'warning' : 'primary'">
              {{ row.deployment_type_display || row.deployment_type }}
            </el-tag>
            <span v-else class="list-empty-cell">{{ t('pages.licenses.table.noProductOffering') }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.status')" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="light">
              {{ statusText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.activationProgress')" width="260">
          <template #default="{ row }">
            <div class="activation-cell">
              <span>{{ activatedCount(row) }}/{{ row.max_activations }}</span>
              <el-progress :percentage="activationPercent(row)" :show-text="false" :stroke-width="6" />
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.endDate')" width="176">
          <template #default="{ row }">
            {{ row.end_date ? formatDate(row.end_date) : '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.customer')" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.customer_name || row.customer_info?.customer_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="description" :label="t('pages.licenses.list.table.description')" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.operation')" width="220" fixed="right" align="center">
          <template #default="{ row }">
            <div class="row-actions data-list-actions">
              <el-button plain size="small" type="primary" @click="openDetail(row)">
                {{ t('pages.licenses.list.actions.detail') }}
              </el-button>
              <el-button
                plain
                size="small"
                :type="isLocked(row) ? 'success' : 'warning'"
                :loading="actionId === row.id"
                @click="toggleLock(row)"
              >
                {{ t(isLocked(row) ? 'pages.licenses.list.actions.unlock' : 'pages.licenses.list.actions.lock') }}
              </el-button>
              <el-button plain size="small" type="danger" :disabled="actionId === row.id" @click="removeLicense(row)">
                {{ t('pages.licenses.list.actions.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="t('pages.licenses.list.table.empty')" />
        </template>
        </el-table>
      </div>

      <div class="pagination data-list-pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          :disabled="loading"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="changePageSize"
          @current-change="loadLicenses"
        />
      </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, CopyDocument, Plus } from '@element-plus/icons-vue'
import { getCustomers } from '@/api/customer'
import { getProductOfferings, type ProductOffering } from '@/api/productOffering'
import { getAuthorizationStatusEnums, type RawEnumItem } from '@/api/enum'
import {
  deleteLicense,
  getLicenses,
  lockAuthorizationCode,
  type AuthorizationCode,
  type LicenseQueryRequest
} from '@/api/license'
import { formatDate } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

interface CustomerOption {
  id: string
  name: string
}


const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const routeCustomerId = String(route.query.customerId || '')
const routeCustomerName = String(route.query.customerName || '')
const hasCustomerContext = Boolean(routeCustomerId)
const filters = reactive({
  customerId: routeCustomerId,
  code: '',
  status: '',
  productOfferingId: '',
  deploymentType: ''
})
const productOptions = ref<ProductOffering[]>([])
const productLoading = ref(false)
const deploymentTypeOptions = computed(() => [
  { value: 'standalone', label: t('pages.licenses.form.fields.deploymentTypes.standalone') },
  { value: 'cloud', label: t('pages.licenses.form.fields.deploymentTypes.cloud') }
])
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const licenses = ref<AuthorizationCode[]>([])
const customerOptions = ref<CustomerOption[]>([])
const statusOptions = ref<RawEnumItem[]>([])
const customerLoading = ref(false)
const loading = ref(false)
const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)
const loadError = ref('')
const actionId = ref('')

const selectedCustomerName = computed(() =>
  customerOptions.value.find(customer => customer.id === filters.customerId)?.name || routeCustomerName
)
const pageTitle = computed(() =>
  selectedCustomerName.value
    ? t('pages.licenses.list.customerTitle', { name: selectedCustomerName.value })
    : t('pages.licenses.list.title')
)

const activatedCount = (row: AuthorizationCode) =>
  row.current_activations ?? row.activated_licenses_count ?? 0

const activationPercent = (row: AuthorizationCode) => {
  if (!row.max_activations) return 0
  return Math.min(100, Math.round((activatedCount(row) / row.max_activations) * 100))
}

const isLocked = (row: AuthorizationCode) => row.is_locked === true || row.status === 'locked'

const statusText = (row: AuthorizationCode) =>
  row.status_display || t(`pages.licenses.list.status.${row.status}`)

const statusTagType = (status: AuthorizationCode['status']) => {
  if (status === 'normal') return 'success'
  if (status === 'locked') return 'warning'
  return 'danger'
}

const syncRouteCustomer = async () => {
  const query = { ...route.query }
  if (filters.customerId) {
    query.customerId = filters.customerId
    query.customerName = selectedCustomerName.value
  } else {
    delete query.customerId
    delete query.customerName
  }
  await router.replace({ query })
}

const openCreate = () => {
  router.push({
    name: 'licenses-create',
    query: filters.customerId
      ? { customerId: filters.customerId, customerName: selectedCustomerName.value }
      : undefined
  })
}

const openDetail = (row: AuthorizationCode) => {
  router.push({
    name: 'licenses-view',
    params: { id: row.id },
    query: filters.customerId
      ? { customerId: filters.customerId, customerName: selectedCustomerName.value }
      : undefined
  })
}

const queryLicenses = async () => {
  pagination.page = 1
  await syncRouteCustomer()
  await loadLicenses()
}

const resetFilters = async () => {
  filters.customerId = ''
  filters.code = ''
  filters.status = ''
  filters.productOfferingId = ''
  filters.deploymentType = ''
  await queryLicenses()
}

const changePageSize = () => {
  pagination.page = 1
  loadLicenses()
}

const loadProductOptions = async () => {
  productLoading.value = true
  try {
    const res = await getProductOfferings({ page: 1, page_size: 100 })
    productOptions.value = res.data.list
  } catch {
    productOptions.value = []
  } finally {
    productLoading.value = false
  }
}

const loadCustomers = async () => {
  customerLoading.value = true
  try {
    const response = await getCustomers({ page: 1, page_size: 100, status: 'active', sort: 'customer_name', order: 'asc' })
    customerOptions.value = response.data.list.map(customer => ({ id: customer.id, name: customer.customer_name }))
    if (routeCustomerId && routeCustomerName && !customerOptions.value.some(item => item.id === routeCustomerId)) {
      customerOptions.value.unshift({ id: routeCustomerId, name: routeCustomerName })
    }
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.list.message.customerLoadError')))
  } finally {
    customerLoading.value = false
  }
}

const loadStatuses = async () => {
  try {
    const response = await getAuthorizationStatusEnums()
    statusOptions.value = response.data.items
  } catch {
    statusOptions.value = (['normal', 'locked', 'expired'] as const).map(key => ({
      key,
      display: t(`pages.licenses.list.status.${key}`)
    }))
  }
}

const loadLicenses = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const params: LicenseQueryRequest = {
      page: pagination.page,
      page_size: pagination.pageSize,
      sort: 'created_at',
      order: 'desc'
    }
    if (filters.customerId) params.customer_id = filters.customerId
    if (filters.code.trim()) params.code = filters.code.trim()
    if (filters.status) params.status = filters.status as LicenseQueryRequest['status']
    if (filters.productOfferingId) params.product_offering_id = filters.productOfferingId
    if (filters.deploymentType) {
      params.deployment_type = filters.deploymentType as LicenseQueryRequest['deployment_type']
    }

    const response = await getLicenses(params)
    licenses.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    loadError.value = getErrorMessage(error, t('pages.licenses.list.message.loadError'))
  } finally {
    loading.value = false
    markInitialized()
  }
}

const isDialogCancel = (error: unknown) => error === 'cancel' || error === 'close'

const toggleLock = async (row: AuthorizationCode) => {
  const unlocking = isLocked(row)
  try {
    await ElMessageBox.confirm(
      t(unlocking ? 'pages.licenses.list.confirm.unlockMessage' : 'pages.licenses.list.confirm.lockMessage'),
      t(unlocking ? 'pages.licenses.list.confirm.unlockTitle' : 'pages.licenses.list.confirm.lockTitle'),
      {
        confirmButtonText: t('pages.licenses.list.confirm.confirm'),
        cancelButtonText: t('pages.licenses.list.confirm.cancel'),
        type: 'warning'
      }
    )
    actionId.value = row.id
    await lockAuthorizationCode(row.id, { is_locked: !unlocking })
    ElMessage.success(t(unlocking ? 'pages.licenses.list.message.unlockSuccess' : 'pages.licenses.list.message.lockSuccess'))
    await loadLicenses()
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(getErrorMessage(error, t(unlocking ? 'pages.licenses.list.message.unlockError' : 'pages.licenses.list.message.lockError')))
    }
  } finally {
    actionId.value = ''
  }
}

const removeLicense = async (row: AuthorizationCode) => {
  try {
    await ElMessageBox.confirm(
      t('pages.licenses.list.confirm.deleteMessage'),
      t('pages.licenses.list.confirm.deleteTitle'),
      {
        confirmButtonText: t('pages.licenses.list.confirm.deleteConfirm'),
        cancelButtonText: t('pages.licenses.list.confirm.cancel'),
        confirmButtonClass: 'dialog-confirm-danger',
        type: 'warning'
      }
    )
    actionId.value = row.id
    await deleteLicense(row.id)
    ElMessage.success(t('pages.licenses.list.message.deleteSuccess'))
    if (licenses.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadLicenses()
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(getErrorMessage(error, t('pages.licenses.list.message.deleteError')))
    }
  } finally {
    actionId.value = ''
  }
}

const backToCustomers = () => router.push('/customers')

const copyAuthorizationCode = async (code: string) => {
  try {
    await navigator.clipboard.writeText(code)
    ElMessage.success(t('pages.licenses.detail.messages.copySuccess'))
  } catch {
    ElMessage.error(t('pages.licenses.detail.messages.copyError'))
  }
}

onMounted(async () => {
  await Promise.all([loadCustomers(), loadStatuses(), loadProductOptions()])
  await loadLicenses()
})
</script>

<style scoped lang="scss">
.filter-grid { grid-template-columns:minmax(200px,1.4fr) minmax(200px,1.2fr) minmax(0,1fr) auto; }

.context-back { margin:0 0 6px; padding:0; display:inline-flex; align-items:center; gap:5px; border:0; background:transparent; color:#526a60; font:inherit; font-size:13px; cursor:pointer; }
.context-back:hover { color:var(--el-color-primary); }

.filter-control,.customer-filter,.code-filter { width:100%; }

.load-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 10px 12px;
  border-radius: $border-radius-base;
  background: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.code-link {
  max-width: 100%;
  padding: 0;
  overflow: hidden;
  border: 0;
  background: none;
  color: var(--el-color-primary);
  font: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.license-code-cell {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: ui-monospace, SFMono-Regular, Consolas, 'Liberation Mono', monospace;
  font-variant-numeric: tabular-nums;
}

.copy-code-button {
  width: 26px;
  height: 26px;
  padding: 0;
  flex-shrink: 0;
  opacity: 0.72;
}

.copy-code-button:hover,
.copy-code-button:focus-visible {
  opacity: 1;
}

.activation-cell {
  display: grid;
  grid-template-columns: 52px minmax(72px, 1fr);
  align-items: center;
  gap: 8px;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 1100px) {
  .filter-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }
}

@media (max-width: 768px) {
  .filter-grid { grid-template-columns:1fr; }
}
.list-empty-cell {
  color: var(--app-text-secondary);
}

.filter-control.product-filter {
  min-width: 160px;
}

.filter-control.deployment-filter {
  min-width: 140px;
}
</style>
