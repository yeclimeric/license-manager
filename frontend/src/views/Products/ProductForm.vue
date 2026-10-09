<template>
  <Layout app-name="Cedar-V" :page-title="pageTitle">
    <div class="product-form data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ pageTitle }}</h1>
        </div>
      </header>

      <section class="data-list-workspace">
        <div class="wizard-card">
          <div class="wizard-steps">
            <div
              v-for="(step, index) in steps"
              :key="step.key"
              class="wizard-step"
              :class="{
                'wizard-step--active': currentStep === index,
                'wizard-step--done': currentStep > index
              }"
            >
              <span class="wizard-step__index">{{ index + 1 }}</span>
              <span class="wizard-step__label">{{ step.label }}</span>
            </div>
          </div>

          <el-alert
            v-if="loadError"
            :title="loadError"
            type="error"
            show-icon
            :closable="false"
            class="wizard-alert"
          />

          <!-- 步骤 1：填写基本信息 -->
          <section v-show="currentStep === 0" class="wizard-section">
            <h2 class="wizard-section__title">{{ t('products.form.basicInfo') }}</h2>
            <p class="wizard-section__desc">{{ t('products.form.stepBasicDesc') }}</p>

            <el-form ref="basicFormRef" :model="basicForm" :rules="basicRules" label-width="140px">
              <el-form-item :label="t('products.form.name')" prop="name">
                <el-input
                  v-model="basicForm.name"
                  :placeholder="t('products.form.namePlaceholder')"
                  maxlength="200"
                  show-word-limit
                />
              </el-form-item>

              <!-- 编码由系统生成，创建后只读展示（D4 决策） -->
              <el-form-item :label="t('products.form.code')">
                <el-input
                  :model-value="generatedCode || t('products.form.codePlaceholder')"
                  readonly
                  :placeholder="t('products.form.codePlaceholder')"
                >
                  <template #suffix>
                    <el-tooltip :content="t('products.form.codeHint')" placement="top">
                      <el-icon class="wizard-lock"><Lock /></el-icon>
                    </el-tooltip>
                  </template>
                </el-input>
                <div class="wizard-hint">{{ t('products.form.codeHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('products.form.alias')">
                <el-input
                  v-model="basicForm.alias"
                  :placeholder="t('products.form.aliasPlaceholder')"
                  maxlength="200"
                />
                <div class="wizard-hint">{{ t('products.form.aliasHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('products.form.description')">
                <el-input
                  v-model="basicForm.description"
                  type="textarea"
                  :rows="3"
                  :placeholder="t('products.form.descriptionPlaceholder')"
                />
              </el-form-item>
            </el-form>
          </section>

          <!-- 步骤 2：设置默认授权配置 -->
          <section v-show="currentStep === 1" class="wizard-section">
            <h2 class="wizard-section__title">{{ t('products.form.defaultConfig') }}</h2>
            <p class="wizard-section__desc">{{ t('products.form.stepDefaultDesc') }}</p>

            <el-form ref="policyFormRef" :model="policyForm" :rules="policyRules" label-width="140px">
              <el-form-item :label="t('products.form.maxActivations')">
                <el-input-number
                  v-model="policyForm.max_activations"
                  :min="1"
                  :max="999999"
                  :placeholder="t('products.form.maxActivationsPlaceholder')"
                />
              </el-form-item>

              <el-form-item :label="t('products.form.durationDays')">
                <el-input-number
                  v-model="policyForm.duration_days"
                  :min="1"
                  :placeholder="t('products.form.durationDaysPlaceholder')"
                />
              </el-form-item>

              <el-form-item :label="t('products.form.deploymentMode')">
                <el-select v-model="policyForm.deployment_mode" class="wizard-select">
                  <el-option
                    v-for="mode in deploymentModes"
                    :key="mode.value"
                    :label="mode.label"
                    :value="mode.value"
                  />
                </el-select>
              </el-form-item>

              <el-form-item :label="t('products.form.enableTrial')">
                <div class="wizard-trial">
                  <el-switch v-model="policyForm.is_trial_product" />
                  <span class="wizard-trial__desc">{{ t('products.form.enableTrialDesc') }}</span>
                </div>
              </el-form-item>

              <el-form-item
                v-if="policyForm.is_trial_product"
                :label="t('products.form.trialDays')"
                prop="trial_days"
              >
                <el-input-number
                  v-model="policyForm.trial_days"
                  :min="1"
                  :max="3650"
                  :placeholder="t('products.form.trialDaysPlaceholder')"
                />
              </el-form-item>
            </el-form>
          </section>

          <!-- 步骤 3：高级配置（可选） -->
          <section v-show="currentStep === 2" class="wizard-section">
            <h2 class="wizard-section__title">{{ t('products.form.advancedConfig') }}</h2>
            <p class="wizard-section__desc">{{ t('products.form.stepAdvancedDesc') }}</p>

            <el-form label-width="140px">
              <el-form-item :label="t('products.form.featureConfig')">
                <el-input
                  v-model="advancedForm.feature_config"
                  type="textarea"
                  :rows="4"
                  :placeholder="t('products.form.featureConfigHint')"
                />
                <div class="wizard-hint">{{ t('products.form.featureConfigHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('products.form.usageLimits')">
                <el-input
                  v-model="advancedForm.usage_limits"
                  type="textarea"
                  :rows="4"
                  :placeholder="t('products.form.usageLimitsHint')"
                />
                <div class="wizard-hint">{{ t('products.form.usageLimitsHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('products.form.customParameters')">
                <el-input
                  v-model="advancedForm.custom_parameters"
                  type="textarea"
                  :rows="4"
                  :placeholder="t('products.form.customParametersHint')"
                />
                <div class="wizard-hint">{{ t('products.form.customParametersHint') }}</div>
              </el-form-item>
            </el-form>
          </section>

          <footer class="wizard-actions">
            <el-button @click="handleCancel">{{ t('products.actions.cancel') }}</el-button>
            <div class="wizard-actions__right">
              <el-button v-if="currentStep > 0" @click="currentStep -= 1">
                {{ t('products.actions.previous') }}
              </el-button>
              <el-button v-if="currentStep < steps.length - 1" type="primary" @click="goNext">
                {{ t('products.actions.next') }}
              </el-button>
              <el-button
                v-else
                type="primary"
                :loading="saving"
                @click="handleSave"
              >
                {{ t('products.actions.save') }}
              </el-button>
            </div>
          </footer>
        </div>
      </section>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Lock } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import {
  createProductOffering,
  getProductOfferingDetail,
  updateProductOffering,
  type DeploymentMode,
  type ProductOffering,
  type ProductOfferingCreateRequest,
  type ProductOfferingUpdateRequest
} from '@/api/productOffering'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  productId?: string
  isEdit?: boolean
}>()

const emit = defineEmits<{
  save: []
  cancel: []
}>()

const { t } = useI18n()

const currentStep = ref(0)
const saving = ref(false)
const loading = ref(false)
const loadError = ref('')
const generatedCode = ref('')
const basicFormRef = ref<FormInstance>()
const policyFormRef = ref<FormInstance>()

const basicForm = reactive({
  name: '',
  alias: '',
  description: ''
})

const policyForm = reactive({
  max_activations: 1 as number | undefined,
  duration_days: 365 as number | undefined,
  deployment_mode: 'online' as DeploymentMode,
  is_trial_product: false,
  trial_days: 15 as number | undefined
})

const advancedForm = reactive({
  feature_config: '',
  usage_limits: '',
  custom_parameters: ''
})

const steps = computed(() => [
  { key: 'basic', label: t('products.form.stepBasic') },
  { key: 'policy', label: t('products.form.stepDefault') },
  { key: 'advanced', label: t('products.form.stepAdvanced') }
])

const deploymentModes = computed(() => [
  { value: 'online', label: t('products.form.deploymentOnline') },
  { value: 'offline', label: t('products.form.deploymentOffline') },
  { value: 'hybrid', label: t('products.form.deploymentHybrid') }
])

const pageTitle = computed(() =>
  props.isEdit ? t('products.editProduct') : t('products.addProduct')
)

const basicRules: FormRules = {
  name: [
    { required: true, message: t('products.validation.nameRequired'), trigger: 'blur' },
    { max: 200, message: t('products.validation.nameTooLong'), trigger: 'blur' }
  ]
}

const policyRules: FormRules = {
  trial_days: [
    {
      validator: (_rule, value, callback) => {
        if (!policyForm.is_trial_product) {
          callback()
          return
        }
        if (value === undefined || value === null || value <= 0) {
          callback(new Error(t('products.validation.trialDaysInvalid')))
          return
        }
        if (value > 3650) {
          callback(new Error(t('products.validation.trialDaysMax')))
          return
        }
        callback()
      },
      trigger: 'change'
    }
  ]
}

// 编辑模式下加载现有数据
const loadProduct = async () => {
  if (!props.productId) return
  loading.value = true
  loadError.value = ''
  try {
    const res = await getProductOfferingDetail(props.productId)
    const p: ProductOffering = res.data as ProductOffering
    basicForm.name = p.name
    basicForm.alias = p.alias || ''
    basicForm.description = p.description || ''
    generatedCode.value = p.code
    policyForm.max_activations = p.max_activations ?? 1
    policyForm.duration_days = p.duration_days ?? 365
    policyForm.deployment_mode = p.deployment_mode || 'online'
    policyForm.is_trial_product = p.is_trial_product
    policyForm.trial_days = p.trial_days ?? 15
    advancedForm.feature_config = JSON.stringify(p.feature_config || {}, null, 2)
    advancedForm.usage_limits = JSON.stringify(p.usage_limits || {}, null, 2)
    advancedForm.custom_parameters = JSON.stringify(p.custom_parameters || {}, null, 2)
  } catch (error) {
    loadError.value = getErrorMessage(error, t('products.message.detailFailed'))
  } finally {
    loading.value = false
  }
}

// 解析 JSON 文本，空串返回 undefined 以便后端保留原值
const parseJsonField = (raw: string): Record<string, any> | undefined => {
  const text = raw.trim()
  if (!text) return undefined
  try {
    return JSON.parse(text)
  } catch {
    throw new Error(t('products.validation.jsonInvalid'))
  }
}

const goNext = async () => {
  if (currentStep.value === 0) {
    const valid = await basicFormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  if (currentStep.value === 1) {
    const valid = await policyFormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  currentStep.value += 1
}

const handleSave = async () => {
  // 保存前再校验一次全部步骤，避免用户在后续步骤回改导致基础信息失效
  const basicValid = await basicFormRef.value?.validate().catch(() => false)
  if (!basicValid) {
    currentStep.value = 0
    return
  }
  const policyValid = await policyFormRef.value?.validate().catch(() => false)
  if (!policyValid) {
    currentStep.value = 1
    return
  }

  let featureConfig: Record<string, any> | undefined
  let usageLimits: Record<string, any> | undefined
  let customParameters: Record<string, any> | undefined
  try {
    featureConfig = parseJsonField(advancedForm.feature_config)
    usageLimits = parseJsonField(advancedForm.usage_limits)
    customParameters = parseJsonField(advancedForm.custom_parameters)
  } catch (error) {
    ElMessage.error((error as Error).message)
    return
  }

  saving.value = true
  try {
    if (props.isEdit && props.productId) {
      // 不传 code：编码不可修改（D4 决策）
      const payload: ProductOfferingUpdateRequest = {
        name: basicForm.name,
        alias: basicForm.alias,
        description: basicForm.description,
        max_activations: policyForm.max_activations,
        duration_days: policyForm.duration_days,
        deployment_mode: policyForm.deployment_mode,
        is_trial_product: policyForm.is_trial_product,
        trial_days: policyForm.is_trial_product ? policyForm.trial_days : undefined,
        feature_config: featureConfig,
        usage_limits: usageLimits,
        custom_parameters: customParameters
      }
      await updateProductOffering(props.productId, payload)
      ElMessage.success(t('products.message.updated'))
    } else {
      const payload: ProductOfferingCreateRequest = {
        name: basicForm.name,
        alias: basicForm.alias,
        description: basicForm.description,
        max_activations: policyForm.max_activations,
        duration_days: policyForm.duration_days,
        deployment_mode: policyForm.deployment_mode,
        is_trial_product: policyForm.is_trial_product,
        trial_days: policyForm.is_trial_product ? policyForm.trial_days : undefined,
        feature_config: featureConfig,
        usage_limits: usageLimits,
        custom_parameters: customParameters
      }
      await createProductOffering(payload)
      ElMessage.success(t('products.message.created'))
    }
    emit('save')
  } catch (error) {
    ElMessage.error(
      getErrorMessage(error, props.isEdit ? t('products.message.updateFailed') : t('products.message.createFailed'))
    )
  } finally {
    saving.value = false
  }
}

const handleCancel = () => emit('cancel')

onMounted(loadProduct)
</script>

<style lang="scss" scoped>
.wizard-card {
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
  padding: 24px 28px;
}

.wizard-steps {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding-bottom: 20px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--app-border-light);
}

.wizard-step {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--app-text-secondary);
  font-size: 14px;

  &__index {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--app-border-light);
    color: var(--app-text-secondary);
    font-size: 13px;
    font-weight: 600;
  }

  &--active {
    color: var(--app-text-primary);
    font-weight: 500;

    .wizard-step__index {
      background: var(--el-color-primary);
      color: #fff;
    }
  }

  &--done .wizard-step__index {
    background: color-mix(in srgb, var(--el-color-primary) 20%, transparent);
    color: var(--el-color-primary);
  }
}

.wizard-section {
  &__title {
    margin: 0 0 4px;
    font-size: 16px;
    font-weight: 600;
    color: var(--app-text-primary);
  }

  &__desc {
    margin: 0 0 20px;
    font-size: 13px;
    color: var(--app-text-secondary);
  }
}

.wizard-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--app-text-secondary);
  line-height: 1.5;
}

.wizard-lock {
  color: var(--app-text-secondary);
  cursor: help;
}

.wizard-select {
  width: 220px;
}

.wizard-trial {
  display: flex;
  align-items: center;
  gap: 12px;

  &__desc {
    font-size: 13px;
    color: var(--app-text-secondary);
  }
}

.wizard-alert {
  margin-bottom: 16px;
}

.wizard-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px solid var(--app-border-light);

  &__right {
    display: flex;
    gap: 8px;
  }
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}
</style>
