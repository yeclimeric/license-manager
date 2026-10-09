<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="720px"
    @update:model-value="handleVisibleChange"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="130px">
      <el-form-item :label="t('rbac.form.displayName')" prop="display_name">
        <el-input v-model="form.display_name" maxlength="100" />
      </el-form-item>

      <!-- 角色标识被用户绑定关系引用，创建后不可更改 -->
      <el-form-item :label="t('rbac.form.name')" prop="name">
        <el-input
          v-model="form.name"
          :placeholder="t('rbac.form.namePlaceholder')"
          :readonly="isEdit"
          maxlength="64"
        >
          <template v-if="isEdit" #suffix>
            <el-tooltip :content="t('rbac.form.name')" placement="top">
              <el-icon class="form-lock"><Lock /></el-icon>
            </el-tooltip>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item :label="t('rbac.form.description')">
        <el-input v-model="form.description" type="textarea" :rows="2" maxlength="255" />
      </el-form-item>
    </el-form>

    <el-divider content-position="left">{{ t('rbac.form.permissions') }}</el-divider>
    <div class="perm-grid">
      <div v-for="resource in permissions" :key="resource.resource" class="perm-row">
        <span class="perm-row__label">{{ t(`permission.${resource.resource}`) }}</span>
        <el-checkbox-group v-model="selected" class="perm-row__actions">
          <el-checkbox
            v-for="action in resource.actions"
            :key="action.code"
            :value="`${resource.resource}:${action.code}`"
            size="small"
          >
            {{ t(`permissionAction.${action.code}`) }}
          </el-checkbox>
        </el-checkbox-group>
      </div>
    </div>

    <el-divider content-position="left">{{ t('rbac.form.menus') }}</el-divider>
    <div class="menu-tree">
      <el-tree
        ref="treeRef"
        :data="menuTree"
        :props="{ label: 'name_key', children: 'children' }"
        node-key="id"
        show-checkbox
        default-expand-all
      >
        <template #default="{ data }">
          <span class="menu-node">
            <span>{{ t(data.name_key) }}</span>
            <code class="menu-node__perm">{{ data.permission }}</code>
          </span>
        </template>
      </el-tree>
    </div>

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
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage, ElTree, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Lock } from '@element-plus/icons-vue'
import {
  createRole,
  updateRole,
  type MenuItem,
  type PermissionResource,
  type Role
} from '@/api/rbac'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  visible: boolean
  role: Role | null
  isEdit: boolean
  permissions: PermissionResource[]
  menus: MenuItem[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const { t } = useI18n()

const formRef = ref<FormInstance>()
const treeRef = ref<InstanceType<typeof ElTree>>()
const submitting = ref(false)

const form = reactive({
  name: '',
  display_name: '',
  description: ''
})

// 权限勾选值为 "resource:action" 字符串数组
const selected = ref<string[]>([])

const dialogTitle = computed(() =>
  props.isEdit ? t('featureTemplates.actions.edit') : t('rbac.page.create')
)

const menuTree = computed(() => props.menus)

const rules: FormRules = {
  name: [
    { required: true, message: t('rbac.validation.nameRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        // 编辑模式下标识不可改，跳过格式校验
        if (props.isEdit) {
          callback()
          return
        }
        if (!/^[a-z][a-z0-9_]*$/.test(value || '')) {
          callback(new Error(t('rbac.validation.namePattern')))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ],
  display_name: [
    { required: true, message: t('rbac.validation.displayNameRequired'), trigger: 'blur' }
  ]
}

const handleVisibleChange = (value: boolean) => {
  emit('update:visible', value)
}

const resetState = () => {
  form.name = ''
  form.display_name = ''
  form.description = ''
  selected.value = []
  nextTick(() => treeRef.value?.setCheckedKeys([]))
}

// 打开时回填角色数据
watch(
  () => props.visible,
  (open) => {
    if (!open) return

    resetState()
    if (props.isEdit && props.role) {
      form.name = props.role.name
      form.display_name = props.role.display_name
      form.description = props.role.description || ''
      // 通配权限不作为勾选项展示
      selected.value = (props.role.permissions || []).filter((p) => p !== '*')
      const menuIds = props.role.menu_ids || []
      nextTick(() => treeRef.value?.setCheckedKeys(menuIds))
    }
  }
)

const handleSave = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  // 半选的分父节点也要绑定，否则子菜单会因为父不可见而被剔除
  const checked = treeRef.value?.getCheckedKeys() as string[]
  const halfChecked = treeRef.value?.getHalfCheckedKeys() as string[]
  const menuIds = [...(checked || []), ...(halfChecked || [])]

  submitting.value = true
  try {
    if (props.isEdit && props.role) {
      await updateRole(props.role.id, {
        display_name: form.display_name,
        description: form.description,
        permissions: selected.value,
        menu_ids: menuIds
      })
      ElMessage.success(t('rbac.message.updated'))
    } else {
      await createRole({
        name: form.name,
        display_name: form.display_name,
        description: form.description,
        permissions: selected.value,
        menu_ids: menuIds
      })
      ElMessage.success(t('rbac.message.created'))
    }
    emit('update:visible', false)
    emit('saved')
  } catch (error) {
    ElMessage.error(
      getErrorMessage(error, props.isEdit ? t('rbac.message.loadFailed') : t('rbac.message.loadFailed'))
    )
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.form-lock {
  color: var(--app-text-secondary);
  cursor: help;
}

.perm-grid {
  max-height: 220px;
  overflow: auto;
  padding-right: 4px;
}

.perm-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 6px 0;
  border-bottom: 1px solid var(--app-border-lighter);

  &:last-child {
    border-bottom: none;
  }

  &__label {
    width: 140px;
    flex-shrink: 0;
    font-size: 13px;
    color: var(--app-text-regular);
  }

  &__actions {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }
}

.menu-tree {
  max-height: 240px;
  overflow: auto;
  border: 1px solid var(--app-border-lighter);
  border-radius: $border-radius-base;
  padding: 8px;
}

.menu-node {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;

  &__perm {
    margin-left: auto;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 11px;
    color: var(--app-text-secondary);
  }
}
</style>
