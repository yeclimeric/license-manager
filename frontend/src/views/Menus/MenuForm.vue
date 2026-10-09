<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="560px"
    @update:model-value="handleVisibleChange"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item :label="t('rbac.form.nameKey')" prop="name_key">
        <el-input v-model="form.name_key" :placeholder="t('rbac.form.nameKeyPlaceholder')" maxlength="100" />
      </el-form-item>

      <el-form-item :label="t('rbac.table.type')" prop="type">
        <el-radio-group v-model="form.type">
          <el-radio value="menu">{{ t('rbac.type.menu') }}</el-radio>
          <el-radio value="directory">{{ t('rbac.type.directory') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <!-- 分组目录无路径，仅作为组织容器 -->
      <el-form-item v-if="form.type === 'menu'" :label="t('rbac.table.path')" prop="path">
        <el-input v-model="form.path" placeholder="/example" maxlength="200" />
      </el-form-item>

      <el-form-item :label="t('rbac.form.icon')">
        <el-input v-model="form.icon" placeholder="Document" maxlength="64" />
      </el-form-item>

      <el-form-item :label="t('rbac.form.parentMenu')">
        <el-select v-model="form.parent_id" :placeholder="t('rbac.form.placeholder')" clearable>
          <el-option
            v-for="option in parents"
            :key="option.id"
            :label="option.label"
            :value="option.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('rbac.table.permission')">
        <el-input v-model="form.permission" placeholder="resource:action" maxlength="100" />
      </el-form-item>

      <el-form-item :label="t('rbac.table.sortOrder')">
        <el-input-number v-model="form.sort_order" :min="0" :max="999" />
      </el-form-item>

      <el-form-item :label="t('rbac.form.hidden')">
        <el-switch v-model="form.hidden" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleVisibleChange(false)">
        {{ t('featureTemplates.actions.cancel') }}
      </el-button>
      <el-button type="primary" :loading="submitting" @click="handleSave">
        {{ t('featureTemplates.actions.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  createMenu,
  updateMenu,
  type MenuItem,
  type MenuUpsertRequest
} from '@/api/rbac'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  visible: boolean
  menu: MenuItem | null
  isEdit: boolean
  parents: { id: string; label: string }[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const { t } = useI18n()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive<MenuUpsertRequest>({
  name_key: '',
  type: 'menu',
  path: '',
  icon: '',
  parent_id: null,
  permission: '',
  sort_order: 0,
  hidden: false
})

const dialogTitle = computed(() =>
  props.isEdit ? t('featureTemplates.actions.edit') : t('featureTemplates.actions.create')
)

const rules: FormRules = {
  name_key: [
    { required: true, message: t('rbac.validation.nameKeyRequired'), trigger: 'blur' }
  ],
  path: [
    {
      validator: (_rule, value, callback) => {
        // 分组目录无路径可省，菜单项必须填
        if (form.type === 'directory') {
          callback()
          return
        }
        if (!value) {
          callback(new Error(t('rbac.validation.pathRequired')))
          return
        }
        callback()
      },
      trigger: 'blur'
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

    form.name_key = ''
    form.type = 'menu'
    form.path = ''
    form.icon = ''
    form.parent_id = null
    form.permission = ''
    form.sort_order = 0
    form.hidden = false

    if (props.isEdit && props.menu) {
      form.name_key = props.menu.name_key
      form.type = props.menu.type
      form.path = props.menu.path || ''
      form.icon = props.menu.icon || ''
      form.parent_id = props.menu.parent_id ?? null
      form.permission = props.menu.permission || ''
      form.sort_order = props.menu.sort_order
      form.hidden = props.menu.hidden
    }
  }
)

const handleSave = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (props.isEdit && props.menu) {
      await updateMenu(props.menu.id, form)
      ElMessage.success(t('rbac.message.menuUpdated'))
    } else {
      await createMenu(form)
      ElMessage.success(t('rbac.message.menuCreated'))
    }
    emit('update:visible', false)
    emit('saved')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('rbac.message.menuLoadFailed')))
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
:deep(.el-form-item) {
  margin-bottom: 18px;
}
</style>
