<template>
  <Layout app-name="Cedar-V" :page-title="t('dashboard.title')">
    <div class="dashboard-page">
      <section class="overview-panel">
        <header class="page-header">
          <div><h1>{{ t('dashboard.heading') }}</h1><p>{{ t('dashboard.description') }}</p></div>
          <div class="header-actions">
            <span v-if="lastUpdatedAt" class="updated-at">{{ t('dashboard.updatedAt', { time: lastUpdatedAt }) }}</span>
            <el-button :icon="Refresh" :loading="refreshing" @click="refreshAll">{{ t('dashboard.refresh') }}</el-button>
          </div>
        </header>
        <div v-if="homeError && !home" class="state-block state-block--compact">
          <el-icon><WarningFilled /></el-icon><span>{{ homeError }}</span>
          <el-button link type="primary" @click="loadHome">{{ t('dashboard.retry') }}</el-button>
        </div>
        <section v-else class="overview-grid" :aria-label="t('dashboard.overviewTitle')">
          <article v-for="card in overviewCards" :key="card.key" class="metric-card" :class="`metric-card--${card.tone}`">
            <div class="metric-card__header">
              <span class="metric-label">{{ card.label }}</span>
              <el-icon class="metric-icon"><component :is="card.icon" /></el-icon>
            </div>
            <el-skeleton v-if="homeSkeletonVisible" :rows="1" class="metric-skeleton" />
            <strong v-else class="metric-value">{{ home ? card.value : '--' }}</strong>
            <span class="metric-help">{{ home ? card.help : homeError || t('dashboard.noData') }}</span>
          </article>
        </section>
      </section>

      <section class="panel expiry-panel">
        <div class="panel-header panel-header--stacked"><div><h2>{{ t('dashboard.expiry.title') }}</h2><p>{{ t('dashboard.expiry.description') }}</p></div></div>
        <div class="expiry-grid">
          <article v-for="item in expiryCards" :key="item.key" class="expiry-card" :class="`expiry-card--${item.tone}`">
            <div class="expiry-card__top"><span>{{ item.label }}</span><el-skeleton v-if="homeSkeletonVisible" :rows="1" /><strong v-else>{{ home ? item.value : '--' }}</strong></div>
            <p>{{ home ? item.help : homeError || t('dashboard.noData') }}</p>
          </article>
        </div>
      </section>

      <section class="panel trend-panel">
        <div class="panel-header trend-header">
          <div><h2>{{ t('dashboard.trendTitle') }}</h2><p>{{ t('dashboard.trendDescription') }}</p></div>
          <div class="trend-controls">
            <el-radio-group v-model="trendPeriod" size="small" @change="handleTrendPeriodChange">
              <el-radio-button value="7d">{{ t('dashboard.period7d') }}</el-radio-button><el-radio-button value="30d">{{ t('dashboard.period30d') }}</el-radio-button><el-radio-button value="custom">{{ t('dashboard.customRange') }}</el-radio-button>
            </el-radio-group>
            <el-date-picker v-if="trendPeriod === 'custom'" v-model="customRange" type="daterange" value-format="YYYY-MM-DD" :range-separator="t('dashboard.to')" :start-placeholder="t('dashboard.startDate')" :end-placeholder="t('dashboard.endDate')" :disabled-date="disableFutureDate" @change="loadTrends" />
          </div>
        </div>
        <div v-loading="trendLoading" class="panel-body chart-body">
          <div v-if="trendError" class="state-block"><el-icon><WarningFilled /></el-icon><span>{{ trendError }}</span><el-button link type="primary" @click="loadTrends">{{ t('dashboard.retry') }}</el-button></div>
          <div v-else class="trend-grid">
            <article class="trend-card"><h3>{{ t('dashboard.trends.authorizationCreation') }}</h3><DashboardTrendChart :data="authorizationTrend" color="#019c7c" :empty-text="t('dashboard.noTrendData')" /></article>
            <article class="trend-card"><h3>{{ t('dashboard.trends.activation') }}</h3><DashboardTrendChart :data="activationTrend" color="#4c86f7" :empty-text="t('dashboard.noTrendData')" /></article>
          </div>
        </div>
      </section>

      <section class="recent-grid">
        <article class="panel recent-panel">
          <div class="panel-header"><h2>{{ t('dashboard.recentAuthorizations.title') }}</h2><el-button link type="primary" @click="router.push({ name: 'licenses-list' })">{{ t('dashboard.viewAll') }}</el-button></div>
          <div class="panel-body table-wrap">
            <el-table v-loading="homeSkeletonVisible" :data="home?.recent_authorizations || []" @row-click="openAuthorization">
              <el-table-column prop="customer_name" :label="t('dashboard.columns.customerName')" min-width="130" show-overflow-tooltip><template #default="{ row }">{{ row.customer_name || t('dashboard.noCustomer') }}</template></el-table-column>
              <el-table-column prop="product_offering_name" :label="t('dashboard.columns.productOffering')" min-width="150" show-overflow-tooltip><template #default="{ row }">{{ row.product_offering_name || t('dashboard.noProductOffering') }}</template></el-table-column>
              <el-table-column :label="t('dashboard.columns.activationUsage')" min-width="120"><template #default="{ row }">{{ row.current_activations }} / {{ row.max_activations }}</template></el-table-column>
              <el-table-column :label="t('dashboard.columns.expiryTime')" min-width="150">
                <template #default="{ row }">
                  <el-tooltip v-if="row.validity_mode === 'activation_based'" :content="t('dashboard.validity.activationBasedHint')" placement="top">
                    <span class="validity-marked">{{ formatDate(row.expires_at || row.end_date) }}<el-icon class="validity-icon"><InfoFilled /></el-icon></span>
                  </el-tooltip>
                  <span v-else>{{ formatDate(row.expires_at || row.end_date) }}</span>
                </template>
              </el-table-column>
              <el-table-column :label="t('dashboard.columns.createdAt')" min-width="125" align="center"><template #default="{ row }">{{ formatDate(row.created_at) }}</template></el-table-column>
              <template #empty><el-empty :description="t('dashboard.noRecentAuthorizations')" :image-size="58" /></template>
            </el-table>
          </div>
        </article>
        <article class="panel recent-panel">
          <div class="panel-header"><h2>{{ t('dashboard.recentActivations.title') }}</h2><el-button link type="primary" @click="router.push({ name: 'licenses-list' })">{{ t('dashboard.viewAll') }}</el-button></div>
          <div class="panel-body table-wrap">
            <el-table v-loading="homeSkeletonVisible" :data="home?.recent_activations || []" @row-click="openActivation">
              <el-table-column prop="customer_name" :label="t('dashboard.columns.customerName')" min-width="130" show-overflow-tooltip><template #default="{ row }">{{ row.customer_name || t('dashboard.noCustomer') }}</template></el-table-column>
              <el-table-column prop="hardware_fingerprint" :label="t('dashboard.columns.device')" min-width="160" show-overflow-tooltip />
              <el-table-column :label="t('dashboard.columns.activationTime')" min-width="125"><template #default="{ row }">{{ formatDate(row.activated_at) }}</template></el-table-column>
              <el-table-column :label="t('dashboard.columns.expiryTime')" min-width="125"><template #default="{ row }">{{ formatDate(row.end_date) }}</template></el-table-column>
              <template #empty><el-empty :description="t('dashboard.noRecentActivations')" :image-size="58" /></template>
            </el-table>
          </div>
        </article>
      </section>

    </div>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Coin, Document, InfoFilled, Monitor, Refresh, WarningFilled } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import DashboardTrendChart from '@/components/dashboard/DashboardTrendChart.vue'
import { getDashboardBusinessTrends, getDashboardHome, type DashboardBusinessTrendsData, type DashboardHomeData, type DashboardTrendPoint, type RecentActivationItem, type RecentAuthorizationItem } from '@/api/dashboard'
import { formatDate } from '@/utils/date'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

const { t } = useI18n()
const router = useRouter()
const home = ref<DashboardHomeData | null>(null)
const authorizationTrend = ref<DashboardTrendPoint[]>([])
const activationTrend = ref<DashboardTrendPoint[]>([])
const homeLoading = ref(false)
const trendLoading = ref(false)
const { skeletonVisible: homeSkeletonVisible, markInitialized: markHomeInitialized } = useInitialSkeleton(homeLoading)
const refreshing = ref(false)
const homeError = ref('')
const trendError = ref('')
const lastUpdatedAt = ref('')
const trendPeriod = ref<'7d' | '30d' | 'custom'>('30d')
const customRange = ref<[string, string] | null>(null)
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const overviewCards = computed(() => {
  const overview = home.value?.overview
  return [
    { key: 'valid', label: t('dashboard.overview.validAuthorizations'), icon: Document, value: overview?.valid_authorizations.total ?? 0, help: t('dashboard.overview.validBreakdown', { activated: overview?.valid_authorizations.activated ?? 0, notActivated: overview?.valid_authorizations.not_activated ?? 0 }), tone: 'brand' },
    { key: 'devices', label: t('dashboard.overview.activatedDevices'), icon: Monitor, value: overview?.activated_devices.total ?? 0, help: t('dashboard.overview.activatedDevicesHelp', { manual: overview?.activated_devices.manual ?? 0, trial: overview?.activated_devices.trial ?? 0 }), tone: 'info' },
    { key: 'slots', label: t('dashboard.overview.remainingSlots'), icon: Coin, value: overview?.remaining_activation_slots ?? 0, help: t('dashboard.overview.remainingSlotsHelp'), tone: 'warning' }
  ]
})
const expiryCards = computed(() => {
  const reminders = home.value?.expiry_reminders
  return [
    { key: 'due7', label: t('dashboard.expiry.due7'), value: reminders?.due_7_days.authorization_count ?? 0, help: t('dashboard.expiry.affectedDevices', { n: reminders?.due_7_days.affected_device_count ?? 0 }), tone: 'danger' },
    { key: 'due30', label: t('dashboard.expiry.due8To30'), value: reminders?.due_8_to_30_days.authorization_count ?? 0, help: t('dashboard.expiry.affectedDevices', { n: reminders?.due_8_to_30_days.affected_device_count ?? 0 }), tone: 'warning' },
    { key: 'expired', label: t('dashboard.expiry.expired'), value: reminders?.expired.authorization_count ?? 0, help: t('dashboard.expiry.affectedDevices', { n: reminders?.expired.affected_device_count ?? 0 }), tone: 'expired' }
  ]
})
const errorText = (error: any) => error?.backendMessage || error?.response?.data?.message || t('dashboard.loadFailed')
const loadHome = async () => {
  homeLoading.value = true; homeError.value = ''
  try { home.value = (await getDashboardHome()).data } catch (error) { homeError.value = errorText(error) } finally { homeLoading.value = false; markHomeInitialized() }
}
const loadTrends = async () => {
  if (trendPeriod.value === 'custom' && !customRange.value) return
  trendLoading.value = true; trendError.value = ''
  try {
    const params: Parameters<typeof getDashboardBusinessTrends>[0] = { period: trendPeriod.value, timezone }
    if (trendPeriod.value === 'custom' && customRange.value) { const [start_date, end_date] = customRange.value; Object.assign(params, { start_date, end_date }) }
    const data: DashboardBusinessTrendsData = (await getDashboardBusinessTrends(params)).data
    authorizationTrend.value = data.authorization_creation_trend; activationTrend.value = data.activation_trend
  } catch (error) { trendError.value = errorText(error) } finally { trendLoading.value = false }
}
const refreshAll = async () => { refreshing.value = true; await Promise.allSettled([loadHome(), loadTrends()]); lastUpdatedAt.value = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }); refreshing.value = false }
const handleTrendPeriodChange = () => { if (trendPeriod.value !== 'custom' || customRange.value) loadTrends() }
const disableFutureDate = (date: Date) => date.getTime() > Date.now()
const openAuthorization = (row: RecentAuthorizationItem) => router.push({ name: 'licenses-view', params: { id: row.id } })
const openActivation = (row: RecentActivationItem) => router.push({ name: 'licenses-view', params: { id: row.authorization_code_id } })
onMounted(refreshAll)
</script>

<style lang="scss" scoped>
.dashboard-page { --dashboard-border:#dce9e4; --dashboard-surface:#fff; --dashboard-text:var(--app-text-primary); --dashboard-regular:#40574e; --dashboard-secondary:#6f827a; min-height:calc(100vh - var(--layout-header-height)); padding:24px; display:flex; flex-direction:column; gap:24px; background:var(--app-bg-color); }
.overview-panel,.panel { overflow:hidden; background:var(--dashboard-surface); border:0; border-radius: $border-radius-base; box-shadow:0 8px 24px rgba(18,67,52,.06); }
.page-header,.panel-header { min-height:64px; padding:14px 20px; display:flex; align-items:center; justify-content:space-between; gap:16px; border-bottom:1px solid var(--dashboard-border); }
.page-header h1,.panel-header h2 { margin:0; color:var(--dashboard-text); line-height:1.35; }.page-header h1 { font-size:22px; font-weight:700; }.panel-header h2 { font-size:18px; font-weight:650; }
.page-header p,.panel-header p { margin:5px 0 0; color:var(--dashboard-secondary); font-size:13px; line-height:1.55; }.header-actions { display:flex; align-items:center; gap:12px; flex-shrink:0; }.updated-at { color:var(--dashboard-secondary); font-size:13px; }
.overview-grid { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); }.metric-card { --metric:#087765; min-width:0; min-height:142px; padding:18px 22px 20px; display:flex; flex-direction:column; border-right:1px solid var(--dashboard-border); }.metric-card:last-child { border-right:0; }
.metric-card--brand,.metric-card--info,.metric-card--warning { --metric:#087765; }
.metric-card__header { min-height:36px; display:flex; align-items:center; justify-content:space-between; gap:16px; }.metric-icon { width:36px; height:36px; display:inline-flex; align-items:center; justify-content:center; flex:0 0 36px; background:#edf8f4; color:var(--metric); font-size:20px; }.metric-label { color:var(--dashboard-regular); font-size:14px; font-weight:600; }.metric-value { margin-top:8px; color:var(--metric); font-size:34px; font-weight:700; line-height:1; font-variant-numeric:tabular-nums; }.metric-help { margin-top:9px; overflow:hidden; color:var(--dashboard-secondary); font-size:13px; line-height:1.5; text-overflow:ellipsis; white-space:nowrap; }.metric-skeleton { width:48%; margin-top:8px; }
.panel-header--stacked { justify-content:flex-start; }.expiry-grid { padding:14px 16px 16px; display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:12px; }.expiry-card { --expiry:#d84d4d; --expiry-bg:#fff7f5; min-width:0; min-height:96px; padding:15px 18px; display:flex; justify-content:center; flex-direction:column; box-sizing:border-box; background:var(--expiry-bg); }.expiry-card--danger { --expiry:#d84d4d; --expiry-bg:#fff7f5; }.expiry-card--warning { --expiry:#bd7600; --expiry-bg:#fffaf0; }.expiry-card--expired { --expiry:#b83232; --expiry-bg:#feeeeb; }.expiry-card__top { display:flex; align-items:center; justify-content:space-between; gap:16px; color:var(--dashboard-regular); font-size:14px; }.expiry-card__top>span { font-weight:600; }.expiry-card__top strong { color:var(--expiry); font-size:28px; font-weight:700; line-height:1; font-variant-numeric:tabular-nums; }.expiry-card--expired .expiry-card__top strong { min-width:46px; height:46px; display:inline-flex; align-items:center; justify-content:center; background:var(--expiry); color:#fff; font-size:24px; }.expiry-card p { margin:9px 0 0; color:var(--dashboard-secondary); font-size:13px; line-height:1.5; }
.trend-header { align-items:flex-start; }.trend-controls { display:flex; align-items:center; justify-content:flex-end; gap:10px; flex-wrap:wrap; }.panel-body { min-height:120px; }.chart-body { min-height:285px; }.trend-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); }.trend-card { min-width:0; padding:16px 18px 8px; }.trend-card+.trend-card { border-left:1px solid var(--dashboard-border); }.trend-card h3 { margin:0 0 6px; color:var(--dashboard-regular); font-size:15px; font-weight:600; line-height:1.45; }
.recent-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:18px; }.recent-panel { min-width:0; }.table-wrap { padding:0 14px 10px; overflow-x:auto; }.table-wrap :deep(.el-table) { font-size:14px; }.table-wrap :deep(.el-table__row) { cursor:pointer; }.table-wrap :deep(.el-table__header th.el-table__cell) { background:color-mix(in srgb,var(--dashboard-surface) 82%,#cdeee1); color:var(--dashboard-regular); font-weight:600; }.state-block { min-height:260px; display:flex; align-items:center; justify-content:center; gap:8px; color:var(--dashboard-secondary); }.state-block--compact { min-height:118px; }
.dashboard-page :deep(.el-button),.dashboard-page :deep(.el-input__wrapper),.dashboard-page :deep(.el-radio-button__inner),.dashboard-page :deep(.el-loading-mask) { border-radius: $border-radius-base; }:global([data-theme="dark"]) .dashboard-page { --dashboard-border:var(--app-border-color); --dashboard-surface:var(--app-content-bg); --dashboard-regular:#c1d0ca; --dashboard-secondary:#93a69e; }:global([data-theme="dark"]) .metric-icon { background:color-mix(in srgb,var(--app-content-bg) 76%,#087765); }:global([data-theme="dark"]) .expiry-card--danger { --expiry-bg:color-mix(in srgb,var(--app-content-bg) 88%,#d84d4d); }:global([data-theme="dark"]) .expiry-card--warning { --expiry-bg:color-mix(in srgb,var(--app-content-bg) 88%,#bd7600); }:global([data-theme="dark"]) .expiry-card--expired { --expiry-bg:color-mix(in srgb,var(--app-content-bg) 82%,#b83232); }
@media (max-width:1023px) { .recent-grid,.trend-grid { grid-template-columns:1fr; }.trend-card+.trend-card { border-top:1px solid var(--dashboard-border); border-left:0; } }
@media (max-width:767px) { .dashboard-page { padding:14px; }.page-header,.trend-header { align-items:flex-start; flex-direction:column; }.header-actions,.trend-controls { width:100%; justify-content:space-between; }.overview-grid,.expiry-grid { grid-template-columns:1fr; }.metric-card { border-right:0; border-bottom:1px solid var(--dashboard-border); }.metric-card:last-child { border-bottom:0; }.trend-controls :deep(.el-date-editor) { width:100%; }.recent-grid { gap:14px; } }
.validity-marked {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.validity-icon {
  font-size: 13px;
  color: var(--app-text-secondary);
  cursor: help;
}
</style>
