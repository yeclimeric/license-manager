<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="560px"
    @update:model-value="handleVisibleChange"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item :label="t('users.form.username')" prop="username">
        <el-input
          v-model="form.username"
          :readonly="isEdit"
          maxlength="50"
        />
      </el-form-item>

      <el-form-item :label="t('users.form.fullName')" prop="fullName">
        <el-input v-model="form.fullName" maxlength="100" />
      </el-form-item>

      <el-form-item :label="t('users.form.email')" prop="email">
        <el-input v-model="form.email" maxlength="255" />
      </el-form-item>

      <el-form-item :label="t('users.form.phone')">
        <el-input v-model="form.phone" maxlength="20" />
      </el-form-item>

      <!-- 仅创建时设置初始密码；编辑时改密码走重置入口 -->
      <el-form-item v-if="!isEdit" :label="t('users.form.newPassword')" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          :placeholder="t('users.form.passwordHint')"
        />
      </el-form-item>

      <el-form-item :label="t('users.form.status')" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio value="active">{{ t('users.status.active') }}</el-radio>
          <el-radio value="locked">{{ t('users.status.locked') }}</el-radio>
          <el-radio value="inactive">{{ t('users.status.inactive') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item :label="t('users.form.roles')" prop="roleNames">
        <el-select
          v-model="form.roleNames"
          multiple
          class="role-select"
          :placeholder="t('users.validation.rolesRequired')"
        >
          <el-option
            v-for="role in roles"
            :key="role.id"
            :label="role.display_name || role.name"
            :value="role.name"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleVisibleChange(false)">{{ t('users.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSave">
        {{ t('users.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  createUser,
  updateUser,
  type UserItem,
  type UserRoleItem,
  type UserStatus
} from '@/api/user'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  visible: boolean
  user: UserItem | null
  isEdit: boolean
  roles: UserRoleItem[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const { t } = useI18n()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive({
  username: '',
  fullName: '',
  email: '',
  phone: '',
  password: '',
  status: 'active' as UserStatus,
  roleNames: [] as string[]
})

const dialogTitle = computed(() =>
  props.isEdit ? t('users.form.edit') : t('users.form.create')
)

const rules: FormRules = {
  username: [
    { required: true, message: t('users.validation.usernameRequired'), trigger: 'blur' },
    { min: 3, max: 50, message: t('users.validation.usernameRequired'), trigger: 'blur' }
  ],
  fullName: [
    { required: true, message: t('users.validation.fullNameRequired'), trigger: 'blur' }
  ],
  email: [
    { required: true, message: t('users.validation.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('users.validation.emailInvalid'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('users.validation.passwordRequired'), trigger: 'blur' },
    { min: 8, message: t('users.validation.passwordLength'), trigger: 'blur' }
  ],
  roleNames: [
    {
      type: 'array',
      required: true,
      min: 1,
      message: t('users.validation.rolesRequired'),
      trigger: 'change'
    }
  ]
}

const handleVisibleChange = (value: boolean) => {
  emit('update:visible', value)
}

watch(
  () => props.visible,
  (open) => {
    if (!open) return

    form.username = ''
    form.fullName = ''
    form.email = ''
    form.phone = ''
    form.password = ''
    form.status = 'active'
    form.roleNames = []

    if (props.isEdit && props.user) {
      form.username = props.user.username
      form.fullName = props.user.full_name
      form.email = props.user.email
      form.phone = props.user.phone || ''
      form.status = props.user.status
      form.roleNames = (props.user.roles || []).map((role) => role.name)
    }
  }
)

const handleSave = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (props.isEdit && props.user) {
      await updateUser(props.user.id, {
        full_name: form.fullName,
        email: form.email,
        phone: form.phone,
        status: form.status,
        role_assignments: form.roleNames.map((role) => ({ role }))
      })
      ElMessage.success(t('users.message.userUpdated'))
    } else {
      await createUser({
        username: form.username,
        full_name: form.fullName,
        email: form.email,
        phone: form.phone,
        password: form.password,
        status: form.status,
        role_assignments: form.roleNames.map((role) => ({ role }))
      })
      ElMessage.success(t('users.message.userCreated'))
    }
    emit('update:visible', false)
    emit('saved')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('users.message.loadFailed')))
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.role-select {
  width: 100%;
}
</style>