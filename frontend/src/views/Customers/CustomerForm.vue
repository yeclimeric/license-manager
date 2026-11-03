<template>
  <div class="form-page">
    <SubpageHeader
      :title="isEdit ? t('customers.editCustomer') : t('customers.addCustomer')"
      :back-label="t('customers.actions.backToList')"
      :back-disabled="submitting"
      @back="emit('cancel')"
    />

    <el-alert v-if="enumError" :title="enumError" type="warning" show-icon :closable="false">
      <template #default>
        <el-button link type="primary" @click="loadEnums">{{ t('customers.actions.retry') }}</el-button>
      </template>
    </el-alert>

    <section v-if="detailError" class="state-card">
      <el-result icon="error" :title="t('customers.message.getDetailFailed')" :sub-title="detailError">
        <template #extra>
          <el-button type="primary" @click="loadCustomer">{{ t('customers.actions.retry') }}</el-button>
          <el-button @click="emit('cancel')">{{ t('customers.view.back') }}</el-button>
        </template>
      </el-result>
    </section>

    <el-form
      v-else
      ref="formRef"
      v-loading="detailLoading"
      :model="formData"
      :rules="formRules"
      label-position="top"
      class="form-sections"
    >
      <section class="form-card">
        <h2>{{ t('customers.form.basicInfo') }}</h2>
        <div class="form-grid">
          <el-form-item :label="t('customers.form.customerName')" prop="name">
            <el-input v-model="formData.name" :placeholder="t('customers.form.placeholder.enter')" />
          </el-form-item>
          <el-form-item :label="t('customers.form.customerType')" prop="type">
            <el-select v-model="formData.type" :placeholder="t('customers.form.placeholder.select')">
              <el-option v-for="option in customerTypeOptions" :key="option.key" :label="option.display" :value="option.key" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('customers.form.customerLevel')" prop="level">
            <el-select v-model="formData.level" :placeholder="t('customers.form.placeholder.select')">
              <template #label="{ label, value }">
                <el-rate v-if="value" :model-value="Number(label) || 0" disabled class="customer-level-rate" />
                <span v-else>{{ label }}</span>
              </template>
              <el-option v-for="option in customerLevelOptions" :key="option.key" :label="option.display" :value="option.key">
                <el-rate :model-value="Number(option.display) || 0" disabled class="customer-level-rate" />
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item :label="t('customers.form.status')" prop="status">
            <el-radio-group v-model="formData.status">
              <el-radio v-for="option in statusOptions" :key="option.key" :value="option.key">{{ option.display }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </div>
      </section>

      <section class="form-card">
        <h2>{{ t('customers.form.contactInfo') }}</h2>
        <div class="form-grid">
          <el-form-item :label="t('customers.form.contactPerson')" prop="contact">
            <el-input v-model="formData.contact" :placeholder="t('customers.form.placeholder.enter')" />
          </el-form-item>
          <el-form-item :label="t('customers.form.phone')" prop="phone">
            <el-input v-model="formData.phone" :placeholder="t('customers.form.placeholder.enter')" />
          </el-form-item>
          <el-form-item :label="t('customers.form.email')" prop="email">
            <el-input v-model="formData.email" :placeholder="t('customers.form.placeholder.enter')" />
          </el-form-item>
          <el-form-item :label="t('customers.form.address')" prop="address" class="full-width">
            <el-input v-model="formData.address" :placeholder="t('customers.form.placeholder.enter')" />
          </el-form-item>
        </div>
      </section>

      <section class="form-card">
        <h2>{{ t('customers.form.businessInfo') }}</h2>
        <div class="form-grid">
          <el-form-item :label="t('customers.form.companySize')" prop="companySize">
            <el-select v-model="formData.companySize" :placeholder="t('customers.form.placeholder.select')">
              <el-option v-for="option in companySizeOptions" :key="option.key" :label="option.display" :value="option.key" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('customers.form.description')" prop="description" class="full-width">
            <el-input
              v-model="formData.description"
              type="textarea"
              :rows="4"
              :placeholder="t('customers.form.placeholder.enter')"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
        </div>
      </section>
      <FormActionBar
        :cancel-label="t('customers.actions.cancel')"
        :submit-label="isEdit ? t('customers.actions.saveChanges') : t('customers.actions.createCustomer')"
        :loading="submitting"
        :disabled="detailLoading || Boolean(detailError)"
        @cancel="emit('cancel')"
        @submit="handleSave"
      />
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  createCustomer,
  getCustomerDetail,
  updateCustomer,
  type CustomerCreateRequest
} from '@/api/customer'
import {
  getCompanySizeEnums,
  getCustomerLevelEnums,
  getCustomerTypeEnums,
  getStatusEnums,
  type RawEnumItem
} from '@/api/enum'
import { getErrorMessage } from '@/utils/error'
import FormActionBar from '@/components/common/FormActionBar.vue'
import SubpageHeader from '@/components/common/SubpageHeader.vue'

interface CustomerFormData {
  name: string
  contact: string
  email: string
  phone: string
  address: string
  type: string
  level: string
  status: string
  companySize: string
  description: string
}

const props = defineProps<{ customerId?: string; isEdit?: boolean }>()
const emit = defineEmits<{ save: []; cancel: [] }>()
const { t } = useI18n()
const formRef = ref<FormInstance>()
const detailLoading = ref(false)
const submitting = ref(false)
const detailError = ref('')
const enumError = ref('')
const customerTypeOptions = ref<RawEnumItem[]>([])
const customerLevelOptions = ref<RawEnumItem[]>([])
const statusOptions = ref<RawEnumItem[]>([])
const companySizeOptions = ref<RawEnumItem[]>([])
const formData = reactive<CustomerFormData>({
  name: '', contact: '', email: '', phone: '', address: '',
  type: '', level: '', status: '', companySize: '', description: ''
})

const formRules: FormRules = {
  name: [{ required: true, message: t('customers.validation.nameRequired'), trigger: 'blur' }],
  contact: [{ required: true, message: t('customers.validation.contactRequired'), trigger: 'blur' }],
  type: [{ required: true, message: t('customers.validation.typeRequired'), trigger: 'change' }],
  level: [{ required: true, message: t('customers.validation.levelRequired'), trigger: 'change' }],
  status: [{ required: true, message: t('customers.validation.statusRequired'), trigger: 'change' }],
  companySize: [{ required: true, message: t('customers.validation.companySizeRequired'), trigger: 'change' }],
  email: [{ type: 'email', message: t('customers.validation.emailFormat'), trigger: 'blur' }],
  phone: [{ pattern: /^1[3-9]\d{9}$/, message: t('customers.validation.phoneFormat'), trigger: 'blur' }]
}

const loadEnums = async () => {
  enumError.value = ''
  try {
    const [typeResponse, levelResponse, statusResponse, sizeResponse] = await Promise.all([
      getCustomerTypeEnums(), getCustomerLevelEnums(), getStatusEnums(), getCompanySizeEnums()
    ])
    customerTypeOptions.value = typeResponse.data.items
    customerLevelOptions.value = levelResponse.data.items
    statusOptions.value = statusResponse.data.items
    companySizeOptions.value = sizeResponse.data.items
    if (!props.isEdit && !formData.status) formData.status = statusResponse.data.items[0]?.key || 'active'
  } catch (error) {
    enumError.value = getErrorMessage(error, t('customers.message.loadEnumError'))
  }
}

const loadCustomer = async () => {
  if (!props.isEdit || !props.customerId) return
  detailLoading.value = true
  detailError.value = ''
  try {
    const response = await getCustomerDetail(props.customerId)
    if (!response.data) throw new Error(t('customers.view.customerNotExist'))
    const customer = response.data
    Object.assign(formData, {
      name: customer.customer_name,
      contact: customer.contact_person,
      email: customer.email || '',
      phone: customer.phone || '',
      address: customer.address || '',
      type: customer.customer_type,
      level: customer.customer_level,
      status: customer.status,
      companySize: customer.company_size || '',
      description: customer.description || ''
    })
  } catch (error) {
    detailError.value = getErrorMessage(error, t('customers.message.getDetailFailed'))
  } finally {
    detailLoading.value = false
  }
}

const buildRequest = (): CustomerCreateRequest => ({
  customer_name: formData.name.trim(),
  customer_type: formData.type as CustomerCreateRequest['customer_type'],
  contact_person: formData.contact.trim(),
  email: formData.email.trim() || undefined,
  phone: formData.phone.trim() || undefined,
  address: formData.address.trim() || undefined,
  customer_level: formData.level,
  status: formData.status as CustomerCreateRequest['status'],
  company_size: formData.companySize || undefined,
  description: formData.description.trim() || undefined
})

const handleSave = async () => {
  if (!formRef.value || submitting.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const request = buildRequest()
    const response = props.isEdit && props.customerId
      ? await updateCustomer(props.customerId, request)
      : await createCustomer(request)
    ElMessage.success(response.message)
    emit('save')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('customers.message.saveError')))
  } finally {
    submitting.value = false
  }
}

onMounted(() => { loadEnums(); loadCustomer() })
</script>

<style scoped>
/* 外壳类（form-page/form-sections/form-card/form-grid/state-card/full-width）
   已提升为全局样式（assets/styles/form-page.scss），此处只保留本页特有样式。 */
.form-card h2 { margin:0 0 18px; color:var(--app-text-primary); font-size:16px; }
:deep(.el-select) { width:100%; }
.customer-level-rate { height:22px; }
.customer-level-rate :deep(.el-rate__item) { margin-right:3px; }
.customer-level-rate :deep(.el-rate__icon) { margin-right:0; color:#f5a623; font-size:17px; }
.customer-level-rate :deep(.el-rate__icon:not(.is-active)) { color:#e8eeeb; }

@media (max-width:768px) {
  .form-page { padding:12px; }
  .form-card { padding:16px; }
}
</style>
