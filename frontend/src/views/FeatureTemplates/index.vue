<template>
  <Layout app-name="Cedar-V" :page-title="pageTitle">
    <div v-if="viewMode === 'list'" class="template-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('featureTemplates.title') }}</h1>
          <p>{{ t('featureTemplates.description') }}</p>
        </div>
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-button type="primary" :icon="Plus" @click="openCreate">
            {{ t('featureTemplates.actions.create') }}
          </el-button>
          <el-button :icon="FolderOpened" @click="openCategories">
            {{ t('featureTemplates.actions.manageCategories') }}
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
                <el-button link type="primary" @click="loadTemplates">
                  {{ t('featureTemplates.actions.view') }}
                </el-button>
              </template>
            </el-alert>

            <div class="filter-grid data-list-filter__grid">
              <el-select v-model="filters.category" :placeholder="t('featureTemplates.filter.category')" clearable>
                <el-option
                  v-for="option in categories"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-select v-model="filters.status" :placeholder="t('featureTemplates.filter.status')" clearable>
                <el-option
                  v-for="option in statusOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-input
                v-model="filters.keyword"
                :placeholder="t('featureTemplates.filter.search')"
                clearable
                @keyup.enter="handleQuery"
              />
              <div class="filter-actions data-list-filter__actions">
                <el-button plain type="primary" :loading="loading" @click="handleQuery">
                  {{ t('featureTemplates.actions.confirm') }}
                </el-button>
                <el-button @click="handleReset">{{ t('featureTemplates.filter.reset') }}</el-button>
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
              :data="templates"
              row-key="id"
            >
              <el-table-column
                prop="name"
                :label="t('featureTemplates.table.name')"
                min-width="180"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <el-button link type="primary" class="template-link" @click="openEdit(row)">
                    {{ row.name }}
                  </el-button>
                  <el-tag v-if="row.is_system" size="small" type="info" effect="plain" class="template-system-tag">
                    {{ t('featureTemplates.category.isSystem') }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                prop="identifier"
                :label="t('featureTemplates.table.identifier')"
                min-width="170"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <code class="template-identifier">{{ row.identifier }}</code>
                </template>
              </el-table-column>
              <el-table-column
                prop="category_display"
                :label="t('featureTemplates.table.category')"
                min-width="120"
                align="center"
              />
              <el-table-column
                prop="status_display"
                :label="t('featureTemplates.table.status')"
                min-width="100"
                align="center"
              >
                <template #default="{ row }">
                  <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="light">
                    {{ row.status_display }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                :label="t('featureTemplates.table.createdAt')"
                min-width="140"
                align="center"
              >
                <template #default="{ row }">{{ formatDateShort(row.created_at) }}</template>
              </el-table-column>
              <el-table-column
                :label="t('featureTemplates.table.actions')"
                fixed="right"
                width="230"
                align="center"
                class-name="operation-column"
              >
                <template #default="{ row }">
                  <div class="data-list-actions">
                    <el-button plain size="small" type="primary" @click="openEdit(row)">
                      {{ t('featureTemplates.actions.edit') }}
                    </el-button>
                    <el-button
                      plain
                      size="small"
                      :type="row.status === 'active' ? 'warning' : 'success'"
                      :loading="actionId === row.id"
                      :disabled="row.is_system"
                      @click="toggleStatus(row)"
                    >
                      {{ row.status === 'active' ? t('featureTemplates.actions.disable') : t('featureTemplates.actions.enable') }}
                    </el-button>
                    <el-button
                      plain
                      size="small"
                      type="danger"
                      :loading="actionId === row.id"
                      :disabled="!row.can_delete"
                      @click="removeTemplate(row)"
                    >
                      {{ t('featureTemplates.actions.delete') }}
                    </el-button>
                  </div>
                </template>
              </el-table-column>
              <template #empty>
                <el-empty :description="t('featureTemplates.table.empty')" :image-size="72" />
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
              @current-change="loadTemplates"
            />
          </div>
        </div>
      </section>
    </div>

    <TemplateForm
      v-if="viewMode === 'form'"
      :template-id="isEditMode ? currentTemplateId : undefined"
      :is-edit="isEditMode"
      :categories="categories"
      @save="handleFormSaved"
      @cancel="showList"
    />

    <CategoryManager
      v-model:visible="categoryVisible"
      :categories="categories"
      @changed="loadCategories"
    />
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, FolderOpened } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import TemplateForm from './TemplateForm.vue'
import CategoryManager from './CategoryManager.vue'
import {
  deleteFeatureTemplate,
  getFeatureTemplateCategories,
  getFeatureTemplates,
  updateFeatureTemplate,
  type FeatureTemplate,
  type FeatureTemplateCategory,
  type FeatureTemplateQueryRequest,
  type FeatureTemplateStatus
} from '@/api/featureTemplate'
import { formatDateShort } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

type ViewMode = 'list' | 'form'

const { t } = useI18n()

const viewMode = ref<ViewMode>('list')
const isEditMode = ref(false)
const currentTemplateId = ref('')
const categoryVisible = ref(false)
const templates = ref<FeatureTemplate[]>([])
const categories = ref<FeatureTemplateCategory[]>([])
const loading = ref(false)
const listError = ref('')
const actionId = ref('')
const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)

const filters = reactive({ category: '', status: '' as FeatureTemplateStatus | '', keyword: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const statusOptions = computed(() => [
  { value: 'active', label: t('featureTemplates.filter.statusActive') },
  { value: 'inactive', label: t('featureTemplates.filter.statusInactive') }
])

const pageTitle = computed(() =>
  viewMode.value === 'form'
    ? isEditMode.value
      ? t('featureTemplates.editTemplate')
      : t('featureTemplates.addTemplate')
    : t('featureTemplates.title')
)

const loadCategories = async () => {
  try {
    const res = await getFeatureTemplateCategories()
    categories.value = res.data?.list || []
  } catch {
    // 分类加载失败不阻断列表，降级为空下拉
    categories.value = []
  }
}

const buildQuery = (): FeatureTemplateQueryRequest => ({
  page: pagination.page,
  page_size: pagination.pageSize,
  keyword: filters.keyword || undefined,
  category: filters.category || undefined,
  status: filters.status || undefined,
  sort_by: 'created_at',
  sort_order: 'desc'
})

const loadTemplates = async () => {
  loading.value = true
  listError.value = ''
  try {
    const res = await getFeatureTemplates(buildQuery())
    templates.value = res.data.list
    pagination.total = res.data.total
    markInitialized()
  } catch (error) {
    listError.value = getErrorMessage(error, t('featureTemplates.message.loadFailed'))
  } finally {
    loading.value = false
  }
}

const handleQuery = () => {
  pagination.page = 1
  loadTemplates()
}

const handleReset = () => {
  filters.category = ''
  filters.status = ''
  filters.keyword = ''
  pagination.page = 1
  loadTemplates()
}

const handleSizeChange = () => {
  pagination.page = 1
  loadTemplates()
}

const openCreate = () => {
  isEditMode.value = false
  currentTemplateId.value = ''
  viewMode.value = 'form'
}

const openEdit = (row: FeatureTemplate) => {
  isEditMode.value = true
  currentTemplateId.value = row.id
  viewMode.value = 'form'
}

const openCategories = () => {
  categoryVisible.value = true
}

const showList = () => {
  viewMode.value = 'list'
  loadTemplates()
}

const handleFormSaved = () => {
  ElMessage.success(t('featureTemplates.message.created'))
  showList()
}

const toggleStatus = async (row: FeatureTemplate) => {
  const next: FeatureTemplateStatus = row.status === 'active' ? 'inactive' : 'active'
  const titleKey = next === 'inactive' ? 'featureTemplates.confirm.disableTitle' : 'featureTemplates.confirm.enableTitle'
  const messageKey = next === 'inactive' ? 'featureTemplates.confirm.disableMessage' : 'featureTemplates.confirm.enableMessage'

  try {
    await ElMessageBox.confirm(t(messageKey, { name: row.name }), t(titleKey), {
      confirmButtonText: t('featureTemplates.actions.confirm'),
      cancelButtonText: t('featureTemplates.actions.cancel'),
      type: 'warning'
    })
  } catch {
    return
  }

  actionId.value = row.id
  try {
    await updateFeatureTemplate(row.id, { status: next })
    ElMessage.success(t('featureTemplates.message.updated'))
    loadTemplates()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('featureTemplates.message.updateFailed')))
  } finally {
    actionId.value = ''
  }
}

const removeTemplate = async (row: FeatureTemplate) => {
  if (!row.can_delete) {
    ElMessage.warning(t('featureTemplates.view.systemTemplate'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('featureTemplates.confirm.deleteMessage', { name: row.name }),
      t('featureTemplates.confirm.deleteTitle'),
      {
        confirmButtonText: t('featureTemplates.actions.confirm'),
        cancelButtonText: t('featureTemplates.actions.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }

  actionId.value = row.id
  try {
    await deleteFeatureTemplate(row.id)
    ElMessage.success(t('featureTemplates.message.deleted'))
    if (templates.value.length === 1 && pagination.page > 1) {
      pagination.page -= 1
    }
    loadTemplates()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('featureTemplates.message.deleteFailed')))
  } finally {
    actionId.value = ''
  }
}

onMounted(async () => {
  await loadCategories()
  await loadTemplates()
})
</script>

<style lang="scss" scoped>
/* 分类 + 状态 + 关键字 + 操作按钮，与共享默认列一致 */
.filter-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr)) auto;
}

.template-link {
  font-weight: 500;
}

.template-system-tag {
  margin-left: 8px;
}

.template-identifier {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
  color: var(--app-text-secondary);
}
</style>
