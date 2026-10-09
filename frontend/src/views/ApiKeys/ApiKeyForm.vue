<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="520px"
    @update:model-value="handleVisibleChange"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item :label="t('apiKeys.table.name')" prop="name">
        <el-input v-model="form.name" maxlength="100" />
      </el-form-item>

      <!-- 所属用户决定密钥能做什么，创建后不可更改 -->
      <el-form-item :label="t('apiKeys.table.user')" prop="userId">
        <el-select v-model="form.userId" :disabled="isEdit" class="full-width">
          <el-option
            v-for="user in users"
            :key="user.id"
            :label="user.full_name || user.username"
            :value="user.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('apiKeys.status.active')">
        <el-switch
          :model-value="form.status === 'active'"
          :disabled="isEdit"
          @update:model-value="(v: boolean | string | number) => (form.status = v ? 'active' : 'inactive')"
        />
      </el-form-item>

      <el-form-item :label="t('apiKeys.table.expires')">
        <el-date-picker
          v-model="form.expiresAt"
          type="datetime"
          :placeholder="t('apiKeys.noExpiry')"
          class="full-width"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
        />
      </el-form-item>

      <el-form-item :label="t('apiKeys.table.ips')">
        <el-select
          v-model="form.allowedIps"
          multiple
          filterable
          allow-create
          default-first-option
          :placeholder="t('apiKeys.form.ipHint')"
          class="full-width"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleVisibleChange(false)">{{ t('apiKeys.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSave">
        {{ t('systemConfig.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  createApiKey,
  updateApiKey,
  type ApiKeyItem
} from '@/api/systemConfig'
import type { UserItem } from '@/api/user'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  visible: boolean
  keyItem: ApiKeyItem | null
  isEdit: boolean
  users: UserItem[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
  /** 明文密钥仅创建时返回，由父组件一次性展示 */
  plaintext: [value: string]
}>()

const { t } = useI18n()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive({
  name: '',
  userId: '',
  status: 'active',
  expiresAt: null as string | null,
  allowedIps: [] as string[]
})

const dialogTitle = computed(() =>
  props.isEdit ? t('apiKeys.form.edit') : t('apiKeys.form.create')
)

const rules: FormRules = {
  name: [{ required: true, message: t('apiKeys.validation.name'), trigger: 'blur' }],
  userId: [{ required: true, message: t('apiKeys.validation.user'), trigger: 'change' }]
}

const handleVisibleChange = (value: boolean) => {
  emit('update:visible', value)
}

watch(
  () => props.visible,
  (open) => {
    if (!open) return

    form.name = ''
    form.userId = ''
    form.status = 'active'
    form.expiresAt = null
    form.allowedIps = []

    if (props.isEdit && props.keyItem) {
      form.name = props.keyItem.name
      form.userId = props.keyItem.user_id
      form.status = props.keyItem.status === 'expired' ? 'active' : props.keyItem.status
      form.expiresAt = props.keyItem.expires_at
      form.allowedIps = props.keyItem.allowed_ips || []
    }
  }
)

const handleSave = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (props.isEdit && props.keyItem) {
      await updateApiKey(props.keyItem.id, {
        name: form.name,
        status: form.status as 'active' | 'inactive',
        expires_at: form.expiresAt,
        allowed_ips: form.allowedIps
      })
      ElMessage.success(t('apiKeys.message.updated'))
    } else {
      const res = await createApiKey({
        name: form.name,
        user_id: form.userId,
        expires_at: form.expiresAt,
        allowed_ips: form.allowedIps
      })
      ElMessage.success(t('apiKeys.message.created'))
      // 明文只此一次，趁父组件弹框还在时交给它展示
      if (res.data?.plaintext) {
        emit('plaintext', res.data.plaintext)
      }
    }
    emit('update:visible', false)
    emit('saved')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('apiKeys.message.loadFailed')))
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.full-width {
  width: 100%;
}
</style>