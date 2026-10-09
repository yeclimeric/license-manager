/*
 * @Author: 13895237362 2205451508@qq.com
 * @Date: 2025-09-23 17:21:03
 * @LastEditors: 13895237362 2205451508@qq.com
 * @LastEditTime: 2025-09-24 16:18:40
 * @FilePath: \frontend\src\api\dashboard.ts
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import Axios from './https/index'

// 最近授权数据类型定义

// API响应类型
export interface RecentAuthorizationItem {
  id: string
  code: string
  customer_id?: string | null
  customer_name: string
  description: string
  status: 'normal' | 'locked' | 'expired'
  status_display: string
  start_date: string
  end_date: string
  max_activations: number
  current_activations: number
  /**
   * fixed 为固定期限；activation_based 表示从激活起算，
   * 此时 end_date 只是硬上限，任一设备的实际到期日以 expires_at 为准
   */
  validity_mode: 'fixed' | 'activation_based'
  expires_at: string
  product_offering_name: string
  created_at: string
  updated_at: string
}

// 请求格式
export interface ApiResponse<T> {
  code: string
  message: string
  data: T
}

// 最近授权响应类型
export interface RecentAuthorizationResponse {
  list: RecentAuthorizationItem[]
  total: number
}

// 获取最近授权列表 
export const getRecentAuthorizations = (params?: { limit?: number }): Promise<ApiResponse<RecentAuthorizationResponse>> => {
  return Axios.get('/api/v1/dashboard/recent-authorizations', { params })
}

// 授权趋势数据类型定义
export interface TrendDataItem {
  date: string
  total_authorizations: number
  new_authorizations: number
  expired_authorizations: number
}

export interface TrendPeriod {
  type: 'week' | 'month' | 'custom'
  start_date: string
  end_date: string
  description_display: string
}

export interface TrendSummary {
  total_count: number
  new_count: number
  expired_count: number
  growth_rate: number
}

export interface AuthorizationTrendResponse {
  period: TrendPeriod
  trend_data: TrendDataItem[]
  summary: TrendSummary
}

// 获取授权趋势数据
export const getAuthorizationTrend = (params: {
  type: 'week' | 'month' | 'custom'
  start_date?: string
  end_date?: string
  timezone?: string
}): Promise<ApiResponse<AuthorizationTrendResponse>> => {
  return Axios.get('/api/v1/dashboard/authorization-trend', { params })
}

// 获取仪表盘统计数据（新接口，返回卡片区数据）
export interface GrowthRate {
  auth_codes_mom: number
  licenses_mom: number
}

export interface StatsOverviewData {
  // stock metrics
  total_auth_codes: number
  active_licenses: number
  // flow metrics
  today_new_licenses: number
  yesterday_new_licenses: number
  month_new_auth_codes: number
  // risk metrics
  expiring_in_7days: number
  expiring_in_30days: number
  abnormal_alerts: number
  // growth rates (sub-text only)
  growth_rate: GrowthRate
}

export const getOverviewStats = (): Promise<ApiResponse<StatsOverviewData>> => {
  return Axios.get('/api/v1/stats/overview')
}

export interface DashboardCountBreakdown {
  total: number
  activated: number
  not_activated: number
}

export interface DashboardExpiryReminder {
  authorization_count: number
  affected_device_count: number
}

export interface RecentActivationItem {
  id: string
  authorization_code_id: string
  customer_name: string
  description: string
  hardware_fingerprint: string
  activated_at: string
  end_date: string
}

export interface DashboardHomeData {
  generated_at: string
  overview: {
    valid_authorizations: DashboardCountBreakdown
    // 手工与体验分开统计：体验型产品不占用客户的激活名额
    activated_devices: { total: number; manual: number; trial: number }
    remaining_activation_slots: number
  }
  expiry_reminders: {
    due_7_days: DashboardExpiryReminder
    due_8_to_30_days: DashboardExpiryReminder
    expired: DashboardExpiryReminder
  }
  recent_authorizations: RecentAuthorizationItem[]
  recent_activations: RecentActivationItem[]
}

export interface DashboardTrendPoint {
  date: string
  count: number
}

export interface DashboardBusinessTrendsData {
  period: {
    type: '7d' | '30d' | 'custom'
    start_date: string
    end_date: string
  }
  authorization_creation_trend: DashboardTrendPoint[]
  activation_trend: DashboardTrendPoint[]
}

export const getDashboardHome = (): Promise<ApiResponse<DashboardHomeData>> => {
  return Axios.get('/api/v1/dashboard/home')
}

export const getDashboardBusinessTrends = (params?: {
  period?: '7d' | '30d' | 'custom'
  start_date?: string
  end_date?: string
  timezone?: string
}): Promise<ApiResponse<DashboardBusinessTrendsData>> => {
  return Axios.get('/api/v1/dashboard/trends', { params })
}
