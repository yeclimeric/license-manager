<template>
  <div class="form-page">
    <SubpageHeader
      :title="t('pages.licenses.form.breadcrumb.createLicense')"
      :description="t('pages.licenses.form.description')"
      :back-label="t('pages.licenses.form.actions.backToList')"
      :back-disabled="submitting"
      @back="cancelCreate"
    />

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="form-content">
      <section class="form-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.basicInfo') }}</h2>
        </div>
        <div class="field-grid two-columns">
          <el-form-item :label="t('pages.licenses.form.fields.customerName')" prop="customer_id">
            <el-select
              v-model="form.customer_id"
              :placeholder="t('pages.licenses.form.placeholders.optionalCustomer')"
              :loading="customerLoading"
              clearable
              filterable
              remote
              :remote-method="searchCustomers"
              @change="selectCustomer"
            >
              <el-option v-for="customer in customers" :key="customer.id" :label="customer.name" :value="customer.id" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="form.customer_id" :label="t('pages.licenses.form.fields.customerId')">
            <el-input v-model="form.customer_code" disabled />
          </el-form-item>
        </div>
        <el-form-item :label="t('pages.licenses.form.fields.description')" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
            :placeholder="t('pages.licenses.form.placeholders.enterDescription')"
          />
        </el-form-item>
      </section>

      <section class="form-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.licenseConfig') }}</h2>
        </div>
        <div class="field-grid three-columns">
          <el-form-item :label="t('pages.licenses.form.fields.productOffering')" prop="product_offering_id">
            <el-select
              v-model="form.product_offering_id"
              :placeholder="t('pages.licenses.form.fields.productOfferingPlaceholder')"
              :loading="productLoading"
              clearable
              filterable
              @change="selectProductOffering"
            >
              <el-option
                v-for="item in productOfferings"
                :key="item.id"
                :label="item.alias || item.name"
                :value="item.id"
              >
                <span>{{ item.alias || item.name }}</span>
                <span class="product-option-code">{{ item.code }}</span>
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.deploymentType')" prop="deployment_type">
            <el-select v-model="form.deployment_type">
              <el-option
                v-for="(label, value) in deploymentTypeLabels"
                :key="value"
                :label="label"
                :value="value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.validityMode')" prop="validity_mode">
            <el-select v-model="form.validity_mode">
              <el-option
                v-for="(label, value) in validityModeLabels"
                :key="value"
                :label="label"
                :value="value"
              />
            </el-select>
          </el-form-item>
        </div>
        <el-alert
          v-if="inheritedHint"
          :title="inheritedHint"
          type="info"
          show-icon
          :closable="false"
          class="inherit-alert"
        />
        <div class="field-grid three-columns">
          <el-form-item :label="t('pages.licenses.form.fields.validityPeriod')" prop="validity_type">
            <el-select v-model="form.validity_type" @change="changeValidityType">
              <el-option :label="t('pages.licenses.form.validityTypes.limited')" value="limited" />
              <el-option :label="t('pages.licenses.form.validityTypes.permanent')" value="permanent" />
            </el-select>
          </el-form-item>
          <el-form-item
            v-if="form.validity_type === 'limited'"
            :label="t('pages.licenses.form.fields.dateRange')"
            prop="date_range"
            class="date-field"
          >
            <el-date-picker
              v-model="form.date_range"
              type="daterange"
              value-format="YYYY-MM-DD"
              format="YYYY-MM-DD"
              :start-placeholder="t('pages.licenses.form.placeholders.startDate')"
              :end-placeholder="t('pages.licenses.form.placeholders.endDate')"
              :disabled-date="disablePastDate"
              @change="calculateValidityDays"
            />
          </el-form-item>
          <el-form-item v-else :label="t('pages.licenses.form.fields.validityDays')" class="date-field">
            <el-input :model-value="t('pages.licenses.form.validityTypes.permanent')" disabled />
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.maxActivations')" prop="max_activations">
            <el-input-number v-model="form.max_activations" :min="1" :max="999999" controls-position="right" />
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.batchCount')" prop="batch_count">
            <el-input-number v-model="form.batch_count" :min="1" :max="100" controls-position="right" />
          </el-form-item>
        </div>
      </section>

      <section class="form-card config-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.featureConfig') }}</h2>
        </div>
        <JsonEditor ref="featureEditor" v-model="featureConfig" />
      </section>

      <section class="form-card config-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.usageLimits') }}</h2>
        </div>
        <JsonEditor ref="limitEditor" v-model="usageLimits" />
      </section>

      <section class="form-card config-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.customParameters') }}</h2>
        </div>
        <JsonEditor ref="parameterEditor" v-model="customParameters" />
      </section>
      <FormActionBar
        :cancel-label="t('pages.licenses.form.actions.cancel')"
        :submit-label="t('pages.licenses.form.actions.create')"
        :loading="submitting"
        @cancel="cancelCreate"
        @submit="submitForm"
      />
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { getCustomers } from '@/api/customer'
import { batchCreateLicenses, createLicense, type AuthorizationCodeCreateRequest } from '@/api/license'
import { getProductOfferings, type ProductOffering } from '@/api/productOffering'
import JsonEditor from '@/components/common/JsonEditor.vue'
import { getErrorMessage } from '@/utils/error'
import FormActionBar from '@/components/common/FormActionBar.vue'
import SubpageHeader from '@/components/common/SubpageHeader.vue'

interface CustomerOption {
  id: string
  name: string
  code: string
}

interface LicenseFormModel {
  customer_id: string
  customer_code: string
  product_offering_id: string
  deployment_type: 'standalone' | 'cloud'
  validity_mode: 'fixed' | 'activation_based'
  batch_count: number
  max_validity_days: number | null
  description: string
  validity_type: 'limited' | 'permanent'
  validity_days: number
  date_range: [string, string] | null
  max_activations: number
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const featureEditor = ref<InstanceType<typeof JsonEditor>>()
const limitEditor = ref<InstanceType<typeof JsonEditor>>()
const parameterEditor = ref<InstanceType<typeof JsonEditor>>()
const submitting = ref(false)
const customerLoading = ref(false)
const customers = ref<CustomerOption[]>([])
const productLoading = ref(false)
const productOfferings = ref<ProductOffering[]>([])
const inheritedHint = ref('')

const deploymentTypeLabels = computed(() => ({
  standalone: t('pages.licenses.form.fields.deploymentTypes.standalone'),
  cloud: t('pages.licenses.form.fields.deploymentTypes.cloud')
}))

const validityModeLabels = computed(() => ({
  fixed: t('pages.licenses.form.fields.validityModes.fixed'),
  activation_based: t('pages.licenses.form.fields.validityModes.activation_based')
}))

// 加载可选产品项（仅启用状态）
const loadProductOfferings = async () => {
  productLoading.value = true
  try {
    const res = await getProductOfferings({ page: 1, page_size: 100, status: 'active' })
    productOfferings.value = res.data.list
  } catch {
    productOfferings.value = []
  } finally {
    productLoading.value = false
  }
}

// 选择产品项后继承其默认配置，继承值仍可被手工覆盖
const selectProductOffering = (id?: string | null) => {
  if (!id) {
    inheritedHint.value = ''
    return
  }
  const product = productOfferings.value.find(item => item.id === id)
  if (!product) {
    inheritedHint.value = ''
    return
  }

  const inherited: string[] = []
  if (product.duration_days && product.duration_days > 0) {
    form.validity_days = product.duration_days
    if (form.validity_type === 'limited') {
      form.date_range = buildDateRange(product.duration_days)
    }
    inherited.push('validity')
  }
  if (product.max_activations && product.max_activations > 0) {
    form.max_activations = product.max_activations
    inherited.push('maxActivations')
  }
  if (product.deployment_mode === 'offline') {
    form.deployment_type = 'standalone'
    inherited.push('deployment')
  }
  if (Object.keys(product.feature_config || {}).length > 0) {
    featureConfig.value = product.feature_config
    inherited.push('featureConfig')
  }
  if (Object.keys(product.usage_limits || {}).length > 0) {
    usageLimits.value = product.usage_limits
    inherited.push('usageLimits')
  }
  if (Object.keys(product.custom_parameters || {}).length > 0) {
    customParameters.value = product.custom_parameters
    inherited.push('customParameters')
  }

  // 体验型产品项的有效期上限提示
  if (product.is_trial_product && product.trial_days) {
    form.max_validity_days = product.trial_days
  } else {
    form.max_validity_days = null
  }

  inheritedHint.value = inherited.length > 0
    ? t('pages.licenses.form.messages.inheritedFromProduct')
    : ''
}

// 由天数生成起止日期范围（今天起 N 天）
const buildDateRange = (days: number): [string, string] => {
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  const end = new Date(start)
  end.setDate(end.getDate() + Math.max(days - 1, 0))
  const fmt = (d: Date) =>
    `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  return [fmt(start), fmt(end)]
}
const featureConfig = ref<Record<string, unknown> | null>({})
const usageLimits = ref<Record<string, unknown> | null>({})
const customParameters = ref<Record<string, unknown> | null>({})

const form = reactive<LicenseFormModel>({
  customer_id: String(route.query.customerId || ''),
  customer_code: String(route.query.customerId || ''),
  product_offering_id: '',
  deployment_type: 'cloud',
  validity_mode: 'fixed',
  batch_count: 1,
  max_validity_days: null,
  description: '',
  validity_type: 'limited',
  validity_days: 365,
  date_range: null,
  max_activations: 1
})

const validateDateRange = (_rule: unknown, value: [string, string] | null, callback: (error?: Error) => void) => {
  if (form.validity_type === 'permanent') return callback()
  if (!value?.[0] || !value?.[1]) return callback(new Error(t('pages.licenses.form.validation.dateRangeRequired')))

  const start = new Date(`${value[0]}T00:00:00`)
  const end = new Date(`${value[1]}T00:00:00`)
  if (end <= start) return callback(new Error(t('pages.licenses.form.validation.endDateAfterStart')))
  if ((end.getTime() - start.getTime()) / 86400000 > 3650) {
    return callback(new Error(t('pages.licenses.form.validation.validityPeriodTooLong')))
  }
  callback()
}

const rules: FormRules<LicenseFormModel> = {
  description: [{ max: 500, message: t('pages.licenses.form.validation.descriptionLength'), trigger: 'blur' }],
  validity_type: [{ required: true, message: t('pages.licenses.form.validation.validityTypeRequired'), trigger: 'change' }],
  date_range: [{ validator: validateDateRange, trigger: 'change' }],
  max_activations: [
    { required: true, message: t('pages.licenses.form.validation.maxActivationsRequired'), trigger: 'change' },
    { type: 'number', min: 1, max: 999999, message: t('pages.licenses.form.validation.maxActivationsRange'), trigger: 'change' }
  ]
}

const disablePastDate = (date: Date) => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return date.getTime() < today.getTime()
}

const calculateValidityDays = (range: [string, string] | null) => {
  if (!range?.[0] || !range?.[1]) return
  const start = new Date(`${range[0]}T00:00:00`)
  const end = new Date(`${range[1]}T00:00:00`)
  form.validity_days = Math.ceil((end.getTime() - start.getTime()) / 86400000)
}

const changeValidityType = () => {
  if (form.validity_type === 'permanent') {
    form.validity_days = 365000
    form.date_range = null
    formRef.value?.clearValidate('date_range')
  } else {
    form.validity_days = 365
  }
}

const selectCustomer = (id?: string | null) => {
  form.customer_code = id ? customers.value.find(customer => customer.id === id)?.code || id : ''
}

const ensureRouteCustomer = () => {
  const id = String(route.query.customerId || '')
  const name = String(route.query.customerName || '')
  if (id && name && !customers.value.some(customer => customer.id === id)) {
    customers.value.unshift({ id, name, code: id })
  }
}

const searchCustomers = async (query = '') => {
  customerLoading.value = true
  try {
    const response = await getCustomers({
      page: 1,
      page_size: 100,
      customer_name: query.trim() || undefined,
      status: 'active',
      sort: 'customer_name',
      order: 'asc'
    })
    customers.value = response.data.list.map(customer => ({
      id: customer.id,
      name: customer.customer_name,
      code: customer.customer_code || customer.id
    }))
    ensureRouteCustomer()
    if (form.customer_id) selectCustomer(form.customer_id)
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.form.messages.searchCustomerErrorRetry')))
  } finally {
    customerLoading.value = false
  }
}

const editorsAreValid = () => [featureEditor, limitEditor, parameterEditor]
  .every(editor => editor.value?.validate() ?? true)

const submitForm = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!editorsAreValid()) {
    ElMessage.error(t('pages.licenses.form.keyValue.validationFailed'))
    return
  }

  const payload: AuthorizationCodeCreateRequest = {
    customer_id: form.customer_id || undefined,
    description: form.description.trim() || undefined,
    validity_days: form.validity_days,
    max_activations: form.max_activations,
    feature_config: featureConfig.value || {},
    usage_limits: JSON.stringify(usageLimits.value || {}),
    custom_parameters: JSON.stringify(customParameters.value || {}),
    product_offering_id: form.product_offering_id || undefined,
    deployment_type: form.deployment_type,
    validity_mode: form.validity_mode
  }

  // 数量大于 1 时走批量创建
  const isBatch = form.batch_count > 1
  if (isBatch && (form.batch_count < 1 || form.batch_count > 100)) {
    ElMessage.error(t('pages.licenses.form.validation.batchCountRange'))
    return
  }

  submitting.value = true
  try {
    if (isBatch) {
      const batchRes = await batchCreateLicenses({ ...payload, count: form.batch_count })
      ElMessage.success(
        t('pages.licenses.form.messages.batchCreateSuccess', { count: batchRes.data?.count ?? form.batch_count })
      )
    } else {
      const response = await createLicense(payload)
      ElMessage.success(response.message || t('pages.licenses.form.messages.createSuccess'))
    }
    await router.push({
      name: 'licenses-list',
      query: form.customer_id
        ? { customerId: form.customer_id, customerName: customers.value.find(item => item.id === form.customer_id)?.name || '' }
        : undefined
    })
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.form.messages.submitErrorRetry')))
  } finally {
    submitting.value = false
  }
}

const goBack = () => router.back()

const cancelCreate = async () => {
  try {
    await ElMessageBox.confirm(
      t('pages.licenses.form.messages.cancelConfirm'),
      t('pages.licenses.form.messages.cancelTitle'),
      {
        confirmButtonText: t('pages.licenses.form.messages.cancelConfirmButton'),
        cancelButtonText: t('pages.licenses.form.messages.cancelCancelButton'),
        type: 'warning'
      }
    )
    goBack()
  } catch {
    // Continue editing.
  }
}

onMounted(async () => {
  ensureRouteCustomer()
  await Promise.all([searchCustomers(), loadProductOfferings()])
})
</script>

<style scoped lang="scss">
.form-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  padding: var(--layout-content-padding);
  box-sizing: border-box;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-card {
  padding: 20px;
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
  background: var(--app-content-bg);
  box-shadow: var(--app-card-shadow);
}

.section-heading {
  margin-bottom: 18px;

  h2 {
    margin: 0;
    color: var(--app-text-primary);
    font-size: 17px;
  }
}

.field-grid {
  display: grid;
  gap: 0 20px;
}

.two-columns {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.three-columns {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-card :deep(.el-select),
.form-card :deep(.el-input-number),
.form-card :deep(.el-date-editor) {
  width: 100%;
}

.date-field {
  grid-column: span 2;
}

.config-card {
  overflow: hidden;
}

@media (max-width: 900px) {
  .form-page {
    padding: 12px;
  }

  .three-columns {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {

  .two-columns,
  .three-columns {
    grid-template-columns: 1fr;
  }

  .date-field {
    grid-column: auto;
  }
}
.inherit-alert {
  margin-bottom: 16px;
}

.product-option-code {
  float: right;
  margin-left: 16px;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
  color: var(--app-text-secondary);
}
</style>
