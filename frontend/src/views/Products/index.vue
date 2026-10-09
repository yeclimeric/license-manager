<template>
  <Layout app-name="Cedar-V" :page-title="pageTitle">
    <div v-if="viewMode === 'list'" class="product-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('products.title') }}</h1>
          <p>{{ t('products.description') }}</p>
        </div>
        <el-button :icon="Refresh" circle :loading="loading" @click="loadProducts" />
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-button type="primary" :icon="Plus" @click="openCreate">
            {{ t('products.actions.create') }}
          </el-button>
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
                <el-button link type="primary" @click="loadProducts">
                  {{ t('products.actions.view') }}
                </el-button>
              </template>
            </el-alert>

            <div class="filter-grid data-list-filter__grid">
              <el-select v-model="filters.status" :placeholder="t('products.filter.status')" clearable>
                <el-option
                  v-for="option in statusOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-input
                v-model="filters.keyword"
                :placeholder="t('products.filter.search')"
                clearable
                @keyup.enter="handleQuery"
              />
              <div class="filter-actions data-list-filter__actions">
                <el-button plain type="primary" :loading="loading" @click="handleQuery">
                  {{ t('products.actions.confirm') }}
                </el-button>
                <el-button @click="handleReset">{{ t('products.actions.cancel') }}</el-button>
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
              :data="products"
              :element-loading-text="t('products.table.actions')"
              row-key="id"
            >
              <el-table-column
                prop="name"
                :label="t('products.table.name')"
                min-width="160"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <el-button link type="primary" class="product-link" @click="openEdit(row)">
                    {{ displayName(row) }}
                  </el-button>
                </template>
              </el-table-column>
              <el-table-column
                prop="code"
                :label="t('products.table.code')"
                min-width="150"
                show-overflow-tooltip
              />
              <el-table-column
                prop="status_display"
                :label="t('products.table.status')"
                min-width="90"
                align="center"
              >
                <template #default="{ row }">
                  <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="light">
                    {{ row.status_display }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                prop="trial_display"
                :label="t('products.table.trial')"
                min-width="110"
                align="center"
              >
                <template #default="{ row }">
                  <span v-if="row.trial_display" class="product-trial">{{ row.trial_display }}</span>
                  <span v-else class="product-trial--none">{{ t('products.table.noTrial') }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="user_count"
                :label="t('products.table.userCount')"
                min-width="100"
                align="center"
              />
              <el-table-column
                prop="auth_code_count"
                :label="t('products.table.authCodeCount')"
                min-width="110"
                align="center"
              />
              <el-table-column
                :label="t('products.table.createdAt')"
                min-width="130"
                align="center"
              >
                <template #default="{ row }">{{ formatDateShort(row.created_at) }}</template>
              </el-table-column>
              <el-table-column
                :label="t('products.table.actions')"
                fixed="right"
                width="300"
                align="center"
                class-name="operation-column"
              >
                <template #default="{ row }">
                  <div class="data-list-actions">
                    <el-button plain size="small" type="primary" @click="openEdit(row)">
                      {{ t('products.actions.edit') }}
                    </el-button>
                    <el-button plain size="small" type="primary" @click="openCreateAuthCode(row)">
                      {{ t('products.actions.createAuthCode') }}
                    </el-button>
                    <el-button plain size="small" @click="viewPublicKey(row)">
                      {{ t('products.actions.viewPublicKey') }}
                    </el-button>
                    <el-button
                      plain
                      size="small"
                      :type="row.status === 'active' ? 'warning' : 'success'"
                      :loading="actionId === row.id"
                      @click="toggleStatus(row)"
                    >
                      {{ row.status === 'active' ? t('products.actions.disable') : t('products.actions.enable') }}
                    </el-button>
                    <el-button
                      plain
                      size="small"
                      type="danger"
                      :loading="actionId === row.id"
                      :disabled="!row.can_delete"
                      @click="removeProduct(row)"
                    >
                      {{ t('products.actions.delete') }}
                    </el-button>
                  </div>
                </template>
              </el-table-column>
              <template #empty>
                <el-empty :description="t('products.table.empty')" :image-size="72" />
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
              @current-change="loadProducts"
            />
          </div>
        </div>
      </section>
    </div>

    <PublicKeyDialog
      v-model:visible="publicKeyVisible"
      :product-id="publicKeyProductId"
    />

    <ProductForm
      v-if="viewMode === 'form'"
      :product-id="isEditMode ? currentProductId : undefined"
      :is-edit="isEditMode"
      @save="handleFormSaved"
      @cancel="showList"
    />
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, Refresh } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import ProductForm from './ProductForm.vue'
import PublicKeyDialog from './PublicKeyDialog.vue'
import {
  deleteProductOffering,
  getProductOfferings,
  updateProductOfferingStatus,
  type ProductOffering,
  type ProductOfferingQueryRequest,
  type ProductOfferingStatus
} from '@/api/productOffering'
import { formatDateShort } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

type ViewMode = 'list' | 'form'

const { t } = useI18n()

const viewMode = ref<ViewMode>('list')
const isEditMode = ref(false)
const currentProductId = ref('')
const publicKeyVisible = ref(false)
const publicKeyProductId = ref('')
const products = ref<ProductOffering[]>([])
const loading = ref(false)
const listError = ref('')
const actionId = ref('')
const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)

const filters = reactive({ status: '' as ProductOfferingStatus | '', keyword: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const statusOptions = computed(() => [
  { value: 'active', label: t('products.filter.statusActive') },
  { value: 'inactive', label: t('products.filter.statusInactive') }
])

const pageTitle = computed(() => {
  if (viewMode.value === 'form') {
    return isEditMode.value ? t('products.editProduct') : t('products.addProduct')
  }
  return t('products.title')
})

// 名称列优先展示别名，为空回落编码
const displayName = (row: ProductOffering) => row.alias || row.code

const buildQuery = (): ProductOfferingQueryRequest => ({
  page: pagination.page,
  page_size: pagination.pageSize,
  keyword: filters.keyword || undefined,
  status: filters.status || undefined,
  sort_by: 'created_at',
  sort_order: 'desc'
})

const loadProducts = async () => {
  loading.value = true
  listError.value = ''
  try {
    const res = await getProductOfferings(buildQuery())
    products.value = res.data.list
    pagination.total = res.data.total
    markInitialized()
  } catch (error) {
    listError.value = getErrorMessage(error, t('products.message.loadFailed'))
  } finally {
    loading.value = false
  }
}

const handleQuery = () => {
  pagination.page = 1
  loadProducts()
}

const handleReset = () => {
  filters.status = ''
  filters.keyword = ''
  pagination.page = 1
  loadProducts()
}

const handleSizeChange = () => {
  pagination.page = 1
  loadProducts()
}

const openCreate = () => {
  isEditMode.value = false
  currentProductId.value = ''
  viewMode.value = 'form'
}

const openEdit = (row: ProductOffering) => {
  isEditMode.value = true
  currentProductId.value = row.id
  viewMode.value = 'form'
}

const showList = () => {
  viewMode.value = 'list'
  loadProducts()
}

const handleFormSaved = () => {
  ElMessage.success(t('products.message.updated'))
  showList()
}

// 创建授权在 S3 打通产品项与授权码后启用
const openCreateAuthCode = (_row: ProductOffering) => {
  ElMessage.info(t('products.actions.createAuthCode'))
}

const viewPublicKey = (row: ProductOffering) => {
  publicKeyProductId.value = row.id
  publicKeyVisible.value = true
}

const toggleStatus = async (row: ProductOffering) => {
  const next: ProductOfferingStatus = row.status === 'active' ? 'inactive' : 'active'
  const titleKey = next === 'inactive' ? 'products.confirm.disableTitle' : 'products.confirm.enableTitle'
  const messageKey = next === 'inactive' ? 'products.confirm.disableMessage' : 'products.confirm.enableMessage'

  try {
    await ElMessageBox.confirm(t(messageKey, { name: displayName(row) }), t(titleKey), {
      confirmButtonText: t('products.actions.confirm'),
      cancelButtonText: t('products.actions.cancel'),
      type: 'warning'
    })
  } catch {
    return
  }

  actionId.value = row.id
  try {
    await updateProductOfferingStatus(row.id, next)
    ElMessage.success(t('products.message.statusUpdated'))
    loadProducts()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('products.message.updateFailed')))
  } finally {
    actionId.value = ''
  }
}

const removeProduct = async (row: ProductOffering) => {
  // 存在授权码时后端会拒绝，前端提前拦截以给出更明确的原因
  if (!row.can_delete) {
    ElMessage.warning(
      t('products.confirm.deleteWithAuthCodes', { count: row.auth_code_count })
    )
    return
  }

  try {
    await ElMessageBox.confirm(
      t('products.confirm.deleteMessage', { name: displayName(row) }),
      t('products.confirm.deleteTitle'),
      {
        confirmButtonText: t('products.actions.confirm'),
        cancelButtonText: t('products.actions.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }

  actionId.value = row.id
  try {
    await deleteProductOffering(row.id)
    ElMessage.success(t('products.message.deleted'))
    // 删除当前页最后一条时回退一页
    if (products.value.length === 1 && pagination.page > 1) {
      pagination.page -= 1
    }
    loadProducts()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('products.message.deleteFailed')))
  } finally {
    actionId.value = ''
  }
}

onMounted(loadProducts)
</script>

<style lang="scss" scoped>
.product-link {
  font-weight: 500;
}

.product-trial {
  color: var(--el-color-primary);
  font-weight: 500;
}

.product-trial--none {
  color: var(--app-text-secondary);
}
</style>
