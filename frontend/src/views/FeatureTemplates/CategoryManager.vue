<template>
  <el-dialog
    :model-value="visible"
    :title="t('featureTemplates.category.title')"
    width="640px"
    @update:model-value="handleVisibleChange"
  >
    <el-alert
      v-if="loadError"
      :title="loadError"
      type="error"
      show-icon
      :closable="false"
      class="category-alert"
    />

    <div class="category-add">
      <el-input
        v-model="newCategory.value"
        :placeholder="t('featureTemplates.category.valuePlaceholder')"
        class="category-add__field"
        :disabled="editingId !== ''"
      />
      <el-input
        v-model="newCategory.label"
        :placeholder="t('featureTemplates.category.labelPlaceholder')"
        class="category-add__field"
      />
      <el-button type="primary" :loading="submitting" :disabled="editingId === ''" @click="submitEdit">
        {{ editingId ? t('featureTemplates.category.edit') : t('featureTemplates.category.add') }}
      </el-button>
      <el-button v-if="editingId" @click="cancelEdit">
        {{ t('featureTemplates.actions.cancel') }}
      </el-button>
    </div>
    <div class="category-hint">{{ t('featureTemplates.category.valueHint') }}</div>

    <el-table v-loading="loading" :data="categories" size="small" class="category-table">
      <el-table-column :label="t('featureTemplates.category.label')" min-width="140">
        <template #default="{ row }">
          <span>{{ row.label }}</span>
          <el-tag v-if="row.is_system" size="small" type="info" effect="plain" class="category-tag">
            {{ t('featureTemplates.category.isSystem') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="value" :label="t('featureTemplates.category.value')" min-width="130">
        <template #default="{ row }">
          <code class="category-value">{{ row.value }}</code>
        </template>
      </el-table-column>
      <el-table-column
        :label="t('featureTemplates.category.templateCount')"
        width="90"
        align="center"
      >
        <template #default="{ row }">{{ row.template_count }}</template>
      </el-table-column>
      <el-table-column :label="t('featureTemplates.category.sortOrder')" width="80" align="center">
        <template #default="{ row }">{{ row.sort_order }}</template>
      </el-table-column>
      <el-table-column :label="t('featureTemplates.table.actions')" width="130" align="center">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="startEdit(row)">
            {{ t('featureTemplates.category.edit') }}
          </el-button>
          <el-button
            text
            type="danger"
            size="small"
            :disabled="!row.can_delete"
            @click="removeCategory(row)"
          >
            {{ t('featureTemplates.category.delete') }}
          </el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty :description="t('featureTemplates.category.empty')" :image-size="60" />
      </template>
    </el-table>

    <template #footer>
      <el-button @click="handleVisibleChange(false)">
        {{ t('featureTemplates.actions.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  createFeatureTemplateCategory,
  deleteFeatureTemplateCategory,
  updateFeatureTemplateCategory,
  type FeatureTemplateCategory
} from '@/api/featureTemplate'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  visible: boolean
  categories: FeatureTemplateCategory[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  changed: []
}>()

const { t } = useI18n()

const submitting = ref(false)
const loading = ref(false)
const loadError = ref('')
const editingId = ref('')

const newCategory = reactive({ value: '', label: '' })

const handleVisibleChange = (value: boolean) => {
  emit('update:visible', value)
  if (value) {
    editingId.value = ''
    newCategory.value = ''
    newCategory.label = ''
    loadError.value = ''
  }
}

const categoryValuePattern = /^[a-z][a-z0-9_]*$/

const submitCreate = async () => {
  const value = newCategory.value.trim()
  const label = newCategory.label.trim()

  if (!value) {
    ElMessage.warning(t('featureTemplates.validation.categoryValueRequired'))
    return
  }
  if (!categoryValuePattern.test(value)) {
    ElMessage.warning(t('featureTemplates.validation.categoryValueInvalid'))
    return
  }
  if (!label) {
    ElMessage.warning(t('featureTemplates.validation.categoryLabelRequired'))
    return
  }

  submitting.value = true
  try {
    await createFeatureTemplateCategory({ value, label })
    ElMessage.success(t('featureTemplates.message.categoryCreated'))
    newCategory.value = ''
    newCategory.label = ''
    emit('changed')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('featureTemplates.message.categoryCreateFailed')))
  } finally {
    submitting.value = false
  }
}

const startEdit = (row: FeatureTemplateCategory) => {
  if (row.is_system) return
  editingId.value = row.id
  newCategory.value = row.value
  newCategory.label = row.label
}

const cancelEdit = () => {
  editingId.value = ''
  newCategory.value = ''
  newCategory.label = ''
}

const submitEdit = async () => {
  if (!editingId.value) {
    await submitCreate()
    return
  }

  const label = newCategory.label.trim()
  if (!label) {
    ElMessage.warning(t('featureTemplates.validation.categoryLabelRequired'))
    return
  }

  submitting.value = true
  try {
    await updateFeatureTemplateCategory(editingId.value, { label })
    ElMessage.success(t('featureTemplates.message.categoryUpdated'))
    cancelEdit()
    emit('changed')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('featureTemplates.message.categoryUpdateFailed')))
  } finally {
    submitting.value = false
  }
}

const removeCategory = async (row: FeatureTemplateCategory) => {
  if (!row.can_delete) {
    ElMessage.warning(
      t('featureTemplates.category.deleteBlocked', { count: row.template_count })
    )
    return
  }

  try {
    await ElMessageBox.confirm(
      t('featureTemplates.confirm.deleteCategoryMessage', { label: row.label }),
      t('featureTemplates.confirm.deleteCategoryTitle'),
      {
        confirmButtonText: t('featureTemplates.actions.confirm'),
        cancelButtonText: t('featureTemplates.actions.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }

  submitting.value = true
  try {
    await deleteFeatureTemplateCategory(row.id)
    ElMessage.success(t('featureTemplates.message.categoryDeleted'))
    emit('changed')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('featureTemplates.message.categoryDeleteFailed')))
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.visible,
  (value) => {
    if (value) {
      loading.value = false
      loadError.value = ''
    }
  }
)
</script>

<style lang="scss" scoped>
.category-alert {
  margin-bottom: 12px;
}

.category-add {
  display: flex;
  gap: 8px;
  margin-bottom: 6px;

  &__field {
    flex: 1;
    min-width: 0;
  }
}

.category-hint {
  margin-bottom: 14px;
  font-size: 12px;
  color: var(--app-text-secondary);
  line-height: 1.5;
}

.category-tag {
  margin-left: 8px;
}

.category-value {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
  color: var(--app-text-secondary);
}
</style>
