<template>
  <el-dialog
    :model-value="visible"
    :title="t('products.publicKey.title')"
    width="640px"
    @update:model-value="handleVisibleChange"
  >
    <el-alert
      :title="t('products.publicKey.notice')"
      type="info"
      show-icon
      :closable="false"
      class="pk-alert"
    />

    <el-descriptions v-if="publicKey.product_code" :column="1" border size="small">
      <el-descriptions-item :label="t('products.publicKey.productCode')">
        <code class="pk-code">{{ publicKey.product_code }}</code>
      </el-descriptions-item>
    </el-descriptions>

    <el-empty
      v-if="loaded && !publicKey.public_key"
      :description="t('products.publicKey.empty')"
      :image-size="60"
    />

    <div v-else-if="publicKey.public_key" class="pk-body">
      <pre class="pk-content">{{ publicKey.public_key }}</pre>
    </div>

    <template #footer>
      <el-button @click="handleVisibleChange(false)">
        {{ t('products.actions.confirm') }}
      </el-button>
      <el-button
        v-if="publicKey.public_key"
        type="primary"
        :icon="DocumentCopy"
        @click="copyPublicKey"
      >
        {{ t('products.publicKey.copy') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { DocumentCopy } from '@element-plus/icons-vue'
import { getProductOfferingPublicKey } from '@/api/productSigning'
import { getErrorMessage } from '@/utils/error'

const props = defineProps<{
  visible: boolean
  productId: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const { t } = useI18n()

const loading = ref(false)
const loaded = ref(false)
const publicKey = reactive({
  product_id: '',
  product_code: '',
  public_key: ''
})

const handleVisibleChange = (value: boolean) => {
  emit('update:visible', value)
  if (!value) {
    loaded.value = false
  }
}

// 打开弹框时按需拉取，关闭后不缓存下次的结果
watch(
  () => [props.visible, props.productId] as const,
  async ([visible, productId]) => {
    if (!visible || !productId) return

    loading.value = true
    loaded.value = false
    publicKey.product_id = ''
    publicKey.product_code = ''
    publicKey.public_key = ''

    try {
      const res = await getProductOfferingPublicKey(productId)
      if (res.data) {
        publicKey.product_id = res.data.product_id
        publicKey.product_code = res.data.product_code
        publicKey.public_key = res.data.public_key || ''
      }
    } catch (error) {
      ElMessage.error(
        getErrorMessage(error, t('products.message.detailFailed'))
      )
    } finally {
      loading.value = false
      loaded.value = true
    }
  },
  { immediate: true }
)

const copyPublicKey = async () => {
  try {
    await navigator.clipboard.writeText(publicKey.public_key)
    ElMessage.success(t('products.publicKey.copied'))
  } catch {
    // 剪贴板不可用时（非安全上下文），提示用户手动选中
    ElMessage.info(t('products.publicKey.copy'))
  }
}
</script>

<style scoped lang="scss">
.pk-alert {
  margin-bottom: 16px;
}

.pk-body {
  margin-top: 12px;
}

.pk-code {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
  color: var(--app-text-secondary);
}

.pk-content {
  margin: 0;
  padding: 12px 14px;
  max-height: 260px;
  overflow: auto;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-light);
  border-radius: $border-radius-base;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--app-text-primary);
  white-space: pre-wrap;
  word-break: break-all;
  user-select: all;
}
</style>
