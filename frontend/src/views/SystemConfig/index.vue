<template>
  <Layout app-name="Cedar-V" :page-title="t('systemConfig.title')">
    <div class="config-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('systemConfig.title') }}</h1>
          <p>{{ t('systemConfig.description') }}</p>
        </div>
      </header>

      <section class="data-list-workspace">
        <el-alert
          v-if="loadError"
          :title="loadError"
          type="error"
          show-icon
          :closable="false"
          class="config-alert"
        />

        <div v-loading="loading" class="config-groups">
          <article v-for="group in groups" :key="group.group" class="config-card">
            <div class="config-card__header">
              <h2>{{ groupTitle(group.group) }}</h2>
            </div>
            <div class="config-card__body">
              <div v-for="item in group.configs" :key="item.config_key" class="config-row">
                <div class="config-row__label">
                  <span class="config-row__name">{{ labelFor(item.config_key) }}</span>
                  <code class="config-row__key">{{ item.config_key }}</code>
                  <el-tag v-if="item.public" size="small" type="success" effect="plain">
                    {{ t('systemConfig.field.public') }}
                  </el-tag>
                </div>
                <div class="config-row__control">
                  <el-switch
                    v-if="item.value_type === 'switch'"
                    :model-value="draft[item.config_key] === 'true'"
                    @update:model-value="(v: boolean | string | number) => setDraft(item.config_key, v ? 'true' : 'false')"
                  />
                  <el-input
                    v-else
                    :model-value="draft[item.config_key]"
                    class="config-input"
                    @update:model-value="(v: string) => setDraft(item.config_key, v)"
                  />
                </div>
              </div>
            </div>
          </article>
        </div>

        <div class="config-actions">
          <el-button
            type="primary"
            :loading="saving"
            :disabled="loading || !hasChanges"
            @click="save"
          >
            {{ saving ? t('systemConfig.saving') : t('systemConfig.save') }}
          </el-button>
          <el-button :disabled="saving || !hasChanges" @click="resetDraft">
            {{ t('apiKeys.cancel') }}
          </el-button>
        </div>
      </section>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import Layout from '@/components/common/layout/Layout.vue'
import {
  batchUpdateConfigs,
  getSystemConfigs,
  type ConfigBatchItem,
  type ConfigGroup,
  type SystemConfigGroup
} from '@/api/systemConfig'
import { getErrorMessage } from '@/utils/error'

const { t } = useI18n()

const groups = ref<SystemConfigGroup[]>([])
const loading = ref(false)
const saving = ref(false)
const loadError = ref('')

/** draft 保存当前编辑值，与 original 对比得出是否改动 */
const draft = reactive<Record<string, string>>({})
const original = reactive<Record<string, string>>({})

// 配置键 → 可读标签；未登记的键退回描述文字
const labelKeys: Record<string, string> = {
  'system.site_name': 'siteName',
  'system.logo_url': 'logo',
  'system.favicon_url': 'favicon',
  'system.show_copyright': 'copyright',
  'auth.show_username_password': 'showPwdLogin',
  'auth.session_timeout_hours': 'sessionTimeout',
  'auth.max_login_attempts': 'maxAttempts',
  'auth.lock_duration_minutes': 'lockDuration',
  'mail.enabled': 'mailEnabled',
  'mail.host': 'smtpHost',
  'mail.port': 'smtpPort',
  'mail.username': 'smtpUser',
  'mail.from': 'mailFrom'
}

const groupTitle = (group: ConfigGroup) => {
  const key = { system: 'system', auth: 'auth', mail: 'mail' }[group]
  return key ? t(`systemConfig.group.${key}`) : group
}

const labelFor = (configKey: string): string => {
  const labelKey = labelKeys[configKey]
  if (labelKey) return t(`systemConfig.label.${labelKey}`)
  return configKey
}

const setDraft = (key: string, value: string) => {
  draft[key] = value
}

const hasChanges = computed(() =>
  Object.keys(draft).some((key) => draft[key] !== original[key])
)

const loadConfigs = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const res = await getSystemConfigs()
    groups.value = res.data?.groups || []
    groups.value.forEach((group) => {
      group.configs.forEach((item) => {
        draft[item.config_key] = item.value
        original[item.config_key] = item.value
      })
    })
  } catch (error) {
    loadError.value = getErrorMessage(error, t('systemConfig.loadFailed'))
  } finally {
    loading.value = false
  }
}

const resetDraft = () => {
  Object.keys(draft).forEach((key) => {
    draft[key] = original[key]
  })
}

const save = async () => {
  const items: ConfigBatchItem[] = Object.keys(draft)
    .filter((key) => draft[key] !== original[key])
    .map((key) => ({ config_key: key, value: draft[key] }))

  if (items.length === 0) return

  saving.value = true
  try {
    await batchUpdateConfigs(items)
    // 保存成功后同步基线，避免再次点击时重复提交
    items.forEach((item) => {
      original[item.config_key] = item.value
    })
    ElMessage.success(t('systemConfig.saveSuccess'))
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('systemConfig.saveFailed')))
  } finally {
    saving.value = false
  }
}

onMounted(loadConfigs)
</script>

<style lang="scss" scoped>
.config-alert {
  margin-bottom: 12px;
}

.config-groups {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.config-card {
  background: var(--app-card-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
  overflow: hidden;

  &__header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--app-border-lighter);

    h2 {
      margin: 0;
      font-size: 14px;
      font-weight: 600;
      color: var(--app-text-primary);
    }
  }

  &__body {
    padding: 4px 16px;
  }
}

.config-row {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 12px 0;
  border-bottom: 1px solid var(--app-border-lighter);

  &:last-child {
    border-bottom: none;
  }

  &__label {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  &__name {
    font-size: 13px;
    color: var(--app-text-primary);
  }

  &__key {
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 11px;
    color: var(--app-text-secondary);
  }

  &__control {
    width: 280px;
    flex-shrink: 0;
    display: flex;
    justify-content: flex-end;
  }
}

.config-input {
  width: 100%;
}

.config-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--app-card-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
}
</style>