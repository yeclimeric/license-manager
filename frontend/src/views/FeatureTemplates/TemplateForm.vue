<template>
  <Layout app-name="Cedar-V" :page-title="pageTitle">
    <div class="template-form data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ pageTitle }}</h1>
        </div>
      </header>

      <section class="data-list-workspace">
        <div class="form-card">
          <el-alert
            v-if="loadError"
            :title="loadError"
            type="error"
            show-icon
            :closable="false"
            class="form-alert"
          />

          <section class="form-section">
            <h2 class="form-section__title">{{ t('featureTemplates.form.basicInfo') }}</h2>
            <el-form ref="formRef" :model="form" :rules="rules" label-width="140px">
              <el-form-item :label="t('featureTemplates.form.name')" prop="name">
                <el-input
                  v-model="form.name"
                  :placeholder="t('featureTemplates.form.namePlaceholder')"
                  maxlength="200"
                  show-word-limit
                />
              </el-form-item>

              <!-- 标识符被客户端 SDK 引用，创建后不可修改 -->
              <el-form-item :label="t('featureTemplates.form.identifier')" prop="identifier">
                <el-input
                  v-model="form.identifier"
                  :placeholder="t('featureTemplates.form.identifierPlaceholder')"
                  :readonly="isEdit"
                  maxlength="100"
                >
                  <template v-if="isEdit" #suffix>
                    <el-tooltip :content="t('featureTemplates.view.identifierReadonly')" placement="top">
                      <el-icon class="form-lock"><Lock /></el-icon>
                    </el-tooltip>
                  </template>
                </el-input>
                <div class="form-hint">{{ t('featureTemplates.form.identifierHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('featureTemplates.form.category')" prop="category">
                <el-select v-model="form.category" class="form-select" :disabled="isSystemTemplate">
                  <el-option
                    v-for="option in categories"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </el-form-item>

              <el-form-item :label="t('featureTemplates.form.sortOrder')">
                <el-input-number v-model="form.sort_order" :min="0" />
                <div class="form-hint">{{ t('featureTemplates.form.sortOrderHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('featureTemplates.form.description')">
                <el-input
                  v-model="form.description"
                  type="textarea"
                  :rows="3"
                  :placeholder="t('featureTemplates.form.descriptionPlaceholder')"
                />
              </el-form-item>
            </el-form>
          </section>

          <section class="form-section">
            <div class="form-section__head">
              <div>
                <h2 class="form-section__title">{{ t('featureTemplates.form.featureItems') }}</h2>
                <p class="form-section__desc">{{ t('featureTemplates.form.featureItemsHint') }}</p>
              </div>
              <el-button :icon="Plus" :disabled="isSystemTemplate" @click="addFeature">
                {{ t('featureTemplates.actions.addFeature') }}
              </el-button>
            </div>

            <el-empty
              v-if="form.feature_config.length === 0"
              :description="t('featureTemplates.view.noFeatures')"
              :image-size="60"
            />

            <div
              v-for="(item, index) in form.feature_config"
              :key="index"
              class="feature-row"
            >
              <div class="feature-row__index">{{ index + 1 }}</div>
              <el-input
                v-model="item.key"
                :placeholder="t('featureTemplates.form.featureKeyPlaceholder')"
                class="feature-row__key"
              />
              <el-input
                v-model="item.name"
                :placeholder="t('featureTemplates.form.featureNamePlaceholder')"
                class="feature-row__name"
              />
              <el-input
                v-model="item.description"
                :placeholder="t('featureTemplates.form.featureDescriptionPlaceholder')"
                class="feature-row__desc"
              />
              <div class="feature-row__actions">
                <el-button
                  text
                  :icon="ArrowUp"
                  :disabled="index === 0 || isSystemTemplate"
                  @click="moveFeature(index, -1)"
                />
                <el-button
                  text
                  :icon="ArrowDown"
                  :disabled="index === form.feature_config.length - 1 || isSystemTemplate"
                  @click="moveFeature(index, 1)"
                />
                <el-button
                  text
                  type="danger"
                  :icon="Delete"
                  :disabled="isSystemTemplate"
                  @click="removeFeature(index, item.name)"
                />
              </div>
            </div>
          </section>

          <section class="form-section">
            <h2 class="form-section__title">{{ t('featureTemplates.form.advancedConfig') }}</h2>
            <el-form label-width="140px">
              <el-form-item :label="t('featureTemplates.form.usageLimits')">
                <el-input
                  v-model="form.usage_limits"
                  type="textarea"
                  :rows="3"
                  :placeholder="t('featureTemplates.form.usageLimitsHint')"
                  :disabled="isSystemTemplate"
                />
              </el-form-item>
              <el-form-item :label="t('featureTemplates.form.customParameters')">
                <el-input
                  v-model="form.custom_parameters"
                  type="textarea"
                  :rows="3"
                  :placeholder="t('featureTemplates.form.customParametersHint')"
                  :disabled="isSystemTemplate"
                />
              </el-form-item>
            </el-form>
          </section>

          <footer class="form-actions">
            <el-button @click="handleCancel">{{ t('featureTemplates.actions.cancel') }}</el-button>
            <el-button type="primary" :loading="saving" :disabled="isSystemTemplate" @click="handleSave">
              {{ t('featureTemplates.actions.save') }}
            </el-button>
          </footer>
        </div>
      </section>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Lock, Plus, ArrowUp, ArrowDown, Delete } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import {
  createFeatureTemplate,
  getFeatureTemplateDetail,
  updateFeatureTemplate,
  type FeatureItem,
  type FeatureTemplate,
  type FeatureTemplateCategory
} from '@/api/featureTemplate'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  templateId?: string
  isEdit?: boolean
  categories: FeatureTemplateCategory[]
}>()

const emit = defineEmits<{
  save: []
  cancel: []
}>()

const { t } = useI18n()

const saving = ref(false)
const loading = ref(false)
const loadError = ref('')
const isSystemTemplate = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  identifier: '',
  description: '',
  category: '',
  sort_order: 0,
  feature_config: [] as FeatureItem[],
  usage_limits: '',
  custom_parameters: ''
})

const pageTitle = computed(() =>
  props.isEdit ? t('featureTemplates.editTemplate') : t('featureTemplates.addTemplate')
)

const featureKeyPattern = /^[A-Za-z0-9_-]+$/

// 功能标识去重校验放在保存时统一做，避免逐项输入时反复报错
const collectFeatureErrors = (): string | null => {
  const seen = new Set<string>()
  for (const item of form.feature_config) {
    const key = (item.key || '').trim()
    if (!key) return t('featureTemplates.validation.featureKeyRequired')
    if (!featureKeyPattern.test(key)) return t('featureTemplates.validation.featureKeyInvalid')
    if (seen.has(key)) return t('featureTemplates.validation.featureKeyDuplicate', { key })
    seen.add(key)
    if (!(item.name || '').trim()) return t('featureTemplates.validation.featureNameRequired')
  }
  return null
}

const rules: FormRules = {
  name: [
    { required: true, message: t('featureTemplates.validation.nameRequired'), trigger: 'blur' },
    { max: 200, message: t('featureTemplates.validation.nameTooLong'), trigger: 'blur' }
  ],
  category: [
    { required: true, message: t('featureTemplates.validation.categoryRequired'), trigger: 'change' }
  ],
  identifier: [
    {
      validator: (_rule, value, callback) => {
        // 编辑模式由后端拒绝修改，前端不校验
        if (props.isEdit) {
          callback()
          return
        }
        const identifier = (value || '').trim()
        if (!identifier) {
          // 留空时服务端按名称生成
          callback()
          return
        }
        if (identifier.length > 100) {
          callback(new Error(t('featureTemplates.validation.identifierTooLong')))
          return
        }
        if (!/^[a-z][a-z0-9_]*$/.test(identifier)) {
          callback(new Error(t('featureTemplates.validation.identifierInvalid')))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

const loadTemplate = async () => {
  if (!props.templateId) return
  loading.value = true
  loadError.value = ''
  try {
    const res = await getFeatureTemplateDetail(props.templateId)
    const template = res.data as FeatureTemplate
    form.name = template.name
    form.identifier = template.identifier
    form.description = template.description || ''
    form.category = template.category
    form.sort_order = template.sort_order ?? 0
    form.feature_config = (template.feature_config || []).map((item) => ({
      key: item.key,
      name: item.name,
      description: item.description
    }))
    form.usage_limits = JSON.stringify(template.usage_limits || {}, null, 2)
    form.custom_parameters = JSON.stringify(template.custom_parameters || {}, null, 2)
    isSystemTemplate.value = template.is_system
  } catch (error) {
    loadError.value = getErrorMessage(error, t('featureTemplates.message.detailFailed'))
  } finally {
    loading.value = false
  }
}

watch(
  () => props.categories,
  (list) => {
    if (!form.category && list.length > 0) {
      const general = list.find((c) => c.value === 'general')
      form.category = general ? general.value : list[0].value
    }
  },
  { immediate: true }
)

const addFeature = () => {
  form.feature_config.push({ key: '', name: '', description: '' })
}

const removeFeature = async (index: number, name: string) => {
  try {
    await ElMessageBox.confirm(
      t('featureTemplates.confirm.removeFeatureMessage', { name }),
      t('featureTemplates.confirm.removeFeatureTitle'),
      {
        confirmButtonText: t('featureTemplates.actions.confirm'),
        cancelButtonText: t('featureTemplates.actions.cancel'),
        type: 'warning'
      }
    )
  } catch {
    return
  }
  form.feature_config.splice(index, 1)
}

const moveFeature = (index: number, delta: number) => {
  const target = index + delta
  if (target < 0 || target >= form.feature_config.length) return
  const items = form.feature_config
  const [item] = items.splice(index, 1)
  items.splice(target, 0, item)
}

const parseJsonField = (raw: string): Record<string, any> | undefined => {
  const text = raw.trim()
  if (!text) return undefined
  return JSON.parse(text)
}

const handleSave = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  const featureError = collectFeatureErrors()
  if (featureError) {
    ElMessage.error(featureError)
    return
  }

  let usageLimits: Record<string, any> | undefined
  let customParameters: Record<string, any> | undefined
  try {
    usageLimits = parseJsonField(form.usage_limits)
    customParameters = parseJsonField(form.custom_parameters)
  } catch {
    ElMessage.error(t('featureTemplates.validation.jsonInvalid'))
    return
  }

  const features = form.feature_config.map((item) => ({
    key: item.key.trim(),
    name: item.name.trim(),
    description: (item.description || '').trim()
  }))

  saving.value = true
  try {
    if (props.isEdit && props.templateId) {
      // 不传 identifier：标识符不可修改
      await updateFeatureTemplate(props.templateId, {
        name: form.name,
        description: form.description,
        category: form.category,
        feature_config: features,
        usage_limits: usageLimits,
        custom_parameters: customParameters,
        sort_order: form.sort_order
      })
      ElMessage.success(t('featureTemplates.message.updated'))
    } else {
      const identifier = form.identifier.trim()
      await createFeatureTemplate({
        name: form.name,
        // 留空交由服务端按名称生成
        ...(identifier ? { identifier } : {}),
        description: form.description,
        category: form.category,
        feature_config: features,
        usage_limits: usageLimits,
        custom_parameters: customParameters,
        sort_order: form.sort_order
      })
      ElMessage.success(t('featureTemplates.message.created'))
    }
    emit('save')
  } catch (error) {
    ElMessage.error(
      getErrorMessage(
        error,
        props.isEdit
          ? t('featureTemplates.message.updateFailed')
          : t('featureTemplates.message.createFailed')
      )
    )
  } finally {
    saving.value = false
  }
}

const handleCancel = () => emit('cancel')

onMounted(loadTemplate)
</script>

<style lang="scss" scoped>
.form-card {
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
  padding: 24px 28px;
}

.form-alert {
  margin-bottom: 16px;
}

.form-section {
  padding-bottom: 8px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--app-border-light);

  &:last-of-type {
    border-bottom: none;
  }

  &__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }

  &__title {
    margin: 0 0 4px;
    font-size: 16px;
    font-weight: 600;
    color: var(--app-text-primary);
  }

  &__desc {
    margin: 0 0 16px;
    font-size: 13px;
    color: var(--app-text-secondary);
  }
}

.form-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--app-text-secondary);
  line-height: 1.5;
}

.form-lock {
  color: var(--app-text-secondary);
  cursor: help;
}

.form-select {
  width: 240px;
}

.feature-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;

  &__index {
    width: 24px;
    text-align: center;
    font-size: 12px;
    color: var(--app-text-secondary);
    flex-shrink: 0;
  }

  &__key {
    width: 180px;
    flex-shrink: 0;
  }

  &__name {
    width: 200px;
    flex-shrink: 0;
  }

  &__desc {
    flex: 1;
    min-width: 0;
  }

  &__actions {
    display: flex;
    gap: 2px;
    flex-shrink: 0;
  }
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 20px;
  border-top: 1px solid var(--app-border-light);
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}
</style>
