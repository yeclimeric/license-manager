<template>
  <div class="license-search-page">
    <header class="page-header">
      <div>
        <h1>{{ t('pages.licenses.search.title') }}</h1>
        <p>{{ t('pages.licenses.search.description') }}</p>
      </div>
      <el-button :icon="Back" @click="handleBack">{{ t('pages.licenses.actions.back') }}</el-button>
    </header>

    <section class="search-card">
      <el-alert v-if="loadError" :title="loadError" type="error" show-icon :closable="false">
        <template #default>
          <el-button link type="primary" @click="loadCustomers">{{ t('pages.licenses.search.retry') }}</el-button>
        </template>
      </el-alert>

      <div class="search-content">
        <div class="field-block">
          <label>{{ t('pages.licenses.search.customerLabel') }}</label>
          <el-select
            v-model="selectedCustomer"
            :placeholder="t('pages.licenses.search.selectCustomer')"
            :loading="customerLoading"
            clearable
            filterable
            class="customer-select"
          >
            <el-option
              v-for="customer in customers"
              :key="customer.id"
              :label="customer.customer_name"
              :value="customer.id"
            />
          </el-select>
          <span>{{ t('pages.licenses.search.customerHint') }}</span>
        </div>

        <div class="action-buttons">
          <el-button type="primary" :icon="Search" :loading="querying" @click="handleQuery">
            {{ t('pages.licenses.actions.query') }}
          </el-button>
          <el-button :icon="Plus" @click="handleCreateLicense">
            {{ t('pages.licenses.actions.createLicense') }}
          </el-button>
        </div>
      </div>

      <div v-if="selectedCustomerInfo" class="selected-customer">
        <span>{{ t('pages.licenses.search.selectedCustomer') }}</span>
        <strong>{{ selectedCustomerInfo.customer_name }}</strong>
      </div>
    </section>

    <section class="guide-grid">
      <article class="guide-card">
        <div class="guide-icon"><el-icon><Search /></el-icon></div>
        <div>
          <h2>{{ t('pages.licenses.search.queryTitle') }}</h2>
          <p>{{ t('pages.licenses.search.queryDescription') }}</p>
        </div>
      </article>
      <article class="guide-card">
        <div class="guide-icon"><el-icon><Plus /></el-icon></div>
        <div>
          <h2>{{ t('pages.licenses.search.createTitle') }}</h2>
          <p>{{ t('pages.licenses.search.createDescription') }}</p>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Back, Plus, Search } from '@element-plus/icons-vue'
import { getCustomers, type Customer } from '@/api/customer'
import { getLicenses } from '@/api/license'

const { t } = useI18n()
const router = useRouter()

const selectedCustomer = ref<string>('')
const customers = ref<Customer[]>([])
const customerLoading = ref(false)
const querying = ref(false)
const loadError = ref('')

const selectedCustomerInfo = computed(() => {
  if (!selectedCustomer.value) return null
  return customers.value.find(c => c.id === selectedCustomer.value) || null
})

// 获取所有客户列表
const loadCustomers = async () => {
  customerLoading.value = true
  loadError.value = ''
  try {
    const response = await getCustomers({ status: 'active', page_size: 100 })
    customers.value = response.data.list || []
  } catch (error: any) {
    console.error('Failed to load customers:', error)
    loadError.value = error.backendMessage || t('pages.licenses.search.loadError')
  } finally {
    customerLoading.value = false
  }
}

// 点击查询按钮时，检查是否有授权数据
const handleQuery = async () => {
  if (!selectedCustomer.value) {
    ElMessage.warning(t('pages.licenses.message.selectCustomerFirst'))
    return
  }

  querying.value = true
  try {
    const response = await getLicenses({
      customer_id: selectedCustomer.value,
      page_size: 1
    })

    if (response.data.list.length === 0) {
      ElMessage.warning(t('pages.licenses.message.noLicenseWarning'))
      return
    }

    router.push({
      name: 'licenses-list',
      query: {
        customerId: selectedCustomer.value,
        customerName: selectedCustomerInfo.value?.customer_name || ''
      }
    })
  } catch (error) {
    console.error('Query licenses failed:', error)
    ElMessage.error(t('pages.licenses.message.queryLicenseError'))
  } finally {
    querying.value = false
  }
}

const handleCreateLicense = () => {
  router.push({
    name: 'licenses-create',
    query: {
      customerId: selectedCustomer.value || '',
      customerName: selectedCustomerInfo.value?.customer_name || ''
    }
  })
}

const handleBack = () => router.push({ name: 'licenses-list' })

onMounted(() => {
  loadCustomers()
})
</script>

<style scoped>
.license-search-page {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 16px;
  padding: var(--layout-content-padding);
  box-sizing: border-box;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
}

.page-header h1 {
  margin: 0;
  color: var(--app-text-primary);
  font-size: 24px;
  line-height: 1.35;
}

.page-header p {
  margin: 4px 0 0;
  color: var(--app-text-secondary);
  font-size: 13px;
}

.search-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  background:
    radial-gradient(circle at 100% 0, color-mix(in srgb, var(--el-color-primary) 10%, transparent), transparent 36%),
    var(--app-content-bg);
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
  box-shadow: var(--app-card-shadow);
}

.search-content {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) auto;
  gap: 20px;
  align-items: end;
}

.field-block {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.field-block label {
  color: var(--app-text-primary);
  font-size: 14px;
  font-weight: 600;
}

.field-block > span {
  color: var(--app-text-secondary);
  font-size: 12px;
}

.customer-select {
  width: 100%;
}

.action-buttons {
  display: flex;
  gap: 8px;
  padding-bottom: 20px;
}

.selected-customer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 14px;
  border-top: 1px solid var(--app-border-color);
  color: var(--app-text-secondary);
  font-size: 13px;
}

.selected-customer strong {
  color: var(--el-color-primary);
  font-weight: 600;
}

.guide-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.guide-card {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 14px;
  padding: 18px;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
  box-shadow: var(--app-card-shadow);
}

.guide-icon {
  display: flex;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: $border-radius-base;
  background: color-mix(in srgb, var(--el-color-primary) 12%, transparent);
  color: var(--el-color-primary);
  font-size: 20px;
}

.guide-card h2 {
  margin: 0 0 5px;
  color: var(--app-text-primary);
  font-size: 15px;
}

.guide-card p {
  margin: 0;
  color: var(--app-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .license-search-page {
    padding: 12px;
  }

  .page-header {
    align-items: stretch;
    flex-direction: column;
  }

  .page-header .el-button {
    width: 100%;
  }

  .search-content,
  .guide-grid {
    grid-template-columns: 1fr;
  }

  .action-buttons {
    padding-bottom: 0;
  }

  .action-buttons .el-button {
    flex: 1;
  }
}
</style>
