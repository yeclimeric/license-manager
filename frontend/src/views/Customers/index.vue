<template>
  <Layout app-name="Cedar-V" :page-title="pageTitle">
    <div v-if="viewMode === 'list'" class="customer-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('customers.title') }}</h1>
          <p>{{ t('customers.description') }}</p>
        </div>
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-button type="primary" :icon="Plus" @click="openCreate">
            {{ t('customers.actions.add') }}
          </el-button>
          <span class="data-list-toolbar__divider" aria-hidden="true"></span>
          <div class="customer-toolbar-content data-list-toolbar__filters">
            <el-alert v-if="enumError" :title="enumError" type="warning" show-icon :closable="false" class="enum-alert">
              <template #default>
                <el-button link type="primary" @click="loadEnums">
                  {{ t('customers.actions.retry') }}
                </el-button>
              </template>
            </el-alert>

            <div class="filter-grid data-list-filter__grid">
              <el-input
                v-model="filters.keyword"
                :placeholder="t('customers.search.placeholder')"
                clearable
                @keyup.enter="handleQuery"
              />
              <el-select v-model="filters.customerType" :placeholder="t('customers.filter.customerType')" clearable>
                <el-option v-for="option in customerTypeOptions" :key="option.key" :label="option.display" :value="option.key" />
              </el-select>
              <el-select v-model="filters.customerLevel" :placeholder="t('customers.filter.customerLevel')" clearable>
                <template #label="{ label, value }">
                  <el-rate v-if="value" :model-value="Number(label) || 0" disabled class="customer-level-rate" />
                  <span v-else>{{ label }}</span>
                </template>
                <el-option v-for="option in customerLevelOptions" :key="option.key" :label="option.display" :value="option.key">
                  <el-rate :model-value="Number(option.display) || 0" disabled class="customer-level-rate" />
                </el-option>
              </el-select>
              <el-select v-model="filters.status" :placeholder="t('customers.filter.status')" clearable>
                <el-option v-for="option in statusOptions" :key="option.key" :label="option.display" :value="option.key" />
              </el-select>
              <div class="filter-actions data-list-filter__actions">
                <el-button plain type="primary" :loading="isUpdating" @click="handleQuery">{{ t('customers.actions.query') }}</el-button>
                <el-button @click="handleReset">{{ t('customers.actions.reset') }}</el-button>
              </div>
            </div>
          </div>
        </div>

        <div class="table-card data-list-table">
        <el-alert v-if="listError" :title="listError" type="error" show-icon :closable="false" class="data-list-alert">
          <template #default>
            <el-button link type="primary" @click="loadCustomers">
              {{ t('customers.actions.retry') }}
            </el-button>
          </template>
        </el-alert>

        <div class="table-scroll data-list-table__scroll">
          <el-table
            v-loading="skeletonVisible"
            class="content-skeleton"
            :class="{ 'is-table-skeleton': skeletonVisible, 'is-table-updating': isUpdating }"
            :data="customers"
            :element-loading-text="t('customers.table.loading')"
            row-key="id"
          >
            <el-table-column prop="customer_code" :label="t('customers.table.customerCode')" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <el-button link type="primary" @click="openDetail(row)">{{ row.customer_code }}</el-button>
              </template>
            </el-table-column>
            <el-table-column prop="customer_name" :label="t('customers.table.customerName')" min-width="220" show-overflow-tooltip />
            <el-table-column prop="customer_type_display" :label="t('customers.table.customerType')" min-width="130" align="center" />
            <el-table-column prop="contact_person" :label="t('customers.table.contactPerson')" min-width="220" show-overflow-tooltip />
            <el-table-column prop="customer_level_display" :label="t('customers.table.customerLevel')" min-width="170" align="center">
              <template #default="{ row }">
                <el-rate :model-value="Number(row.customer_level_display) || 0" disabled class="customer-level-rate" />
              </template>
            </el-table-column>
            <el-table-column prop="status_display" :label="t('customers.table.status')" min-width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="light">{{ row.status_display }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('customers.table.createTime')" min-width="140" align="center">
              <template #default="{ row }">{{ formatDateShort(row.created_at) }}</template>
            </el-table-column>
            <el-table-column :label="t('customers.table.operation')" fixed="right" width="300" align="center" class-name="operation-column">
              <template #default="{ row }">
                <div class="data-list-actions">
                  <el-button plain size="small" type="primary" @click="openLicenses(row)">{{ t('customers.actions.viewLicense') }}</el-button>
                  <el-button plain size="small" type="primary" @click="openEdit(row)">{{ t('customers.actions.edit') }}</el-button>
                  <el-button
                    plain
                    size="small"
                    :type="row.status === 'active' ? 'warning' : 'success'"
                    :loading="actionId === row.id"
                    @click="toggleStatus(row)"
                  >
                    {{ row.status === 'active' ? t('customers.actions.disable') : t('customers.actions.enable') }}
                  </el-button>
                  <el-button plain size="small" type="danger" :loading="actionId === row.id" @click="removeCustomer(row)">
                    {{ t('customers.actions.delete') }}
                  </el-button>
                </div>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :description="t('customers.table.empty')" :image-size="72" />
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
            @current-change="loadCustomers"
          />
        </div>
        </div>
      </section>
    </div>

    <CustomerForm
      v-else-if="viewMode === 'form'"
      :customer-id="isEditMode ? currentCustomerId : undefined"
      :is-edit="isEditMode"
      @save="handleFormSaved"
      @cancel="showList"
    />
    <CustomerView v-else :customer-id="currentCustomerId" @back="showList" />
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Plus } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import CustomerForm from './CustomerForm.vue'
import CustomerView from './CustomerView.vue'
import {
  deleteCustomer,
  getCustomers,
  toggleCustomerStatus,
  type Customer,
  type CustomerQueryRequest
} from '@/api/customer'
import {
  getCustomerLevelEnums,
  getCustomerTypeEnums,
  getStatusEnums,
  type RawEnumItem
} from '@/api/enum'
import { formatDateShort } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

type ViewMode = 'list' | 'form' | 'detail'

const { t } = useI18n()
const router = useRouter()
const viewMode = ref<ViewMode>('list')
const isEditMode = ref(false)
const currentCustomerId = ref('')
const customers = ref<Customer[]>([])
const loading = ref(false)
const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)
const listError = ref('')
const enumError = ref('')
const actionId = ref('')
const filters = reactive({ customerType: '', customerLevel: '', status: '', keyword: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const customerTypeOptions = ref<RawEnumItem[]>([])
const customerLevelOptions = ref<RawEnumItem[]>([])
const statusOptions = ref<RawEnumItem[]>([])

const pageTitle = computed(() => {
  if (viewMode.value === 'detail') return t('customers.viewCustomer')
  if (viewMode.value === 'form') return isEditMode.value ? t('customers.editCustomer') : t('customers.addCustomer')
  return t('customers.title')
})

const loadEnums = async () => {
  enumError.value = ''
  try {
    const [typeResponse, levelResponse, statusResponse] = await Promise.all([
      getCustomerTypeEnums(),
      getCustomerLevelEnums(),
      getStatusEnums()
    ])
    customerTypeOptions.value = typeResponse.data.items
    customerLevelOptions.value = levelResponse.data.items
    statusOptions.value = statusResponse.data.items
  } catch (error) {
    enumError.value = getErrorMessage(error, t('customers.message.loadEnumError'))
  }
}

const loadCustomers = async () => {
  loading.value = true
  listError.value = ''
  const params: CustomerQueryRequest = {
    page: pagination.page,
    page_size: pagination.pageSize,
    search: filters.keyword.trim() || undefined,
    customer_type: (filters.customerType || undefined) as CustomerQueryRequest['customer_type'],
    customer_level: filters.customerLevel || undefined,
    status: (filters.status || undefined) as CustomerQueryRequest['status']
  }
  try {
    const response = await getCustomers(params)
    customers.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    listError.value = getErrorMessage(error, t('customers.message.loadError'))
  } finally {
    loading.value = false
    markInitialized()
  }
}

const handleQuery = () => {
  pagination.page = 1
  loadCustomers()
}
const handleReset = () => {
  Object.assign(filters, { customerType: '', customerLevel: '', status: '', keyword: '' })
  pagination.page = 1
  loadCustomers()
}
const handleSizeChange = () => {
  pagination.page = 1
  loadCustomers()
}

const openCreate = () => {
  isEditMode.value = false
  currentCustomerId.value = ''
  viewMode.value = 'form'
}

const openEdit = (customer: Customer) => {
  isEditMode.value = true
  currentCustomerId.value = customer.id
  viewMode.value = 'form'
}

const openDetail = (customer: Customer) => {
  currentCustomerId.value = customer.id
  viewMode.value = 'detail'
}

const openLicenses = (customer: Customer) => {
  router.push({
    path: '/licenses/list',
    query: { customerId: customer.id, customerName: customer.customer_name }
  })
}

const showList = () => {
  viewMode.value = 'list'
}

const handleFormSaved = async () => {
  showList()
  await loadCustomers()
}

const toggleStatus = async (customer: Customer) => {
  const nextStatus = customer.status === 'active' ? 'disabled' : 'active'
  const isDisabling = nextStatus === 'disabled'
  try {
    await ElMessageBox.confirm(
      t(isDisabling ? 'customers.confirm.disableMessage' : 'customers.confirm.enableMessage', { name: customer.customer_name }),
      t(isDisabling ? 'customers.confirm.disableTitle' : 'customers.confirm.enableTitle'),
      {
        confirmButtonText: t('customers.confirm.confirm'),
        cancelButtonText: t('customers.confirm.cancel'),
        confirmButtonClass: 'dialog-confirm-danger',
        type: 'warning'
      }
    )
  } catch { return }

  actionId.value = customer.id
  try {
    const response = await toggleCustomerStatus(customer.id, nextStatus)
    ElMessage.success(response.message)
    await loadCustomers()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('customers.message.statusError')))
  } finally {
    actionId.value = ''
  }
}

const removeCustomer = async (customer: Customer) => {
  try {
    await ElMessageBox.confirm(
      t('customers.confirm.deleteMessage', { name: customer.customer_name }),
      t('customers.confirm.deleteTitle'),
      { confirmButtonText: t('customers.confirm.confirm'), cancelButtonText: t('customers.confirm.cancel'), type: 'warning' }
    )
  } catch { return }

  actionId.value = customer.id
  try {
    const response = await deleteCustomer(customer.id)
    ElMessage.success(response.message)
    if (customers.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadCustomers()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('customers.message.deleteError')))
  } finally {
    actionId.value = ''
  }
}

onMounted(() => {
  loadEnums()
  loadCustomers()
})
</script>

<style scoped>
/* 筛选栅格用可收缩轨道（minmax(0px,1fr)）而非固定宽度。
   固定列宽在窄视口下会撑破工具栏，右侧按钮被截断。 */
.filter-grid { grid-template-columns:minmax(200px,1.6fr) repeat(3,minmax(0,1fr)) auto; }
.enum-alert { margin-bottom:12px; }
.customer-level-rate { height:22px; }
.customer-level-rate :deep(.el-rate__item) { margin-right:3px; }
.customer-level-rate :deep(.el-rate__icon) { margin-right:0; color:#f5a623; font-size:17px; }
.customer-level-rate :deep(.el-rate__icon:not(.is-active)) { color:#e8eeeb; }

@media (max-width:1200px) {
  .filter-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }
}

@media (max-width:768px) {
  .filter-grid { grid-template-columns:1fr; }
}
</style>
