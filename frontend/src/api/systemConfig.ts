import Axios from './https'
import type { ApiResponse } from './customer'

/** 配置值类型 */
export type ConfigValueType = 'text' | 'textarea' | 'switch' | 'image' | 'select' | 'number'

/** 配置分组 */
export type ConfigGroup = 'system' | 'auth' | 'mail'

/** 系统配置项 */
export interface SystemConfig {
  id: string
  config_key: string
  value: string
  value_type: ConfigValueType
  group: ConfigGroup
  options: string
  /** 是否可由未认证的公开接口读取 */
  public: boolean
  description: string
  sort_order: number
}

/** 分组 */
export interface SystemConfigGroup {
  group: ConfigGroup
  configs: SystemConfig[]
}

/** 分组响应 */
export interface SystemConfigGroupResponse {
  groups: SystemConfigGroup[]
}

/** 批量更新项 */
export interface ConfigBatchItem {
  config_key: string
  value: string
}

/**
 * 获取公开配置
 *
 * 未认证即可调用，登录页据此渲染站点名与 Logo。
 */
export function getPublicConfig(): Promise<ApiResponse<Record<string, string>>> {
  return Axios.get('/api/public/config')
}

/** 获取按分组组织的系统配置 */
export function getSystemConfigs(): Promise<ApiResponse<SystemConfigGroupResponse>> {
  return Axios.get('/api/v1/system-configs')
}

/**
 * 批量更新配置值
 *
 * 一次提交整份配置：逐项保存会产生中间态。
 */
export function batchUpdateConfigs(
  items: ConfigBatchItem[]
): Promise<ApiResponse> {
  return Axios.put('/api/v1/system-configs/batch', { items })
}

/** API Key 列表项 */
export interface ApiKeyItem {
  id: string
  name: string
  /** 前缀，形如 sk-xxxxxxxx；明文不返回 */
  api_key_prefix: string
  user_id: string
  user_name: string
  status: 'active' | 'inactive' | 'expired'
  expires_at: string | null
  allowed_ips: string[] | null
  last_used_at: string | null
  created_at: string
  created_by: string
}

/** API Key 列表响应 */
export interface ApiKeyListResponse {
  list: ApiKeyItem[]
  total: number
}

/** 创建响应 */
export interface ApiKeyCreateResponse {
  api_key: ApiKeyItem
  /** 明文密钥，仅此一次返回 */
  plaintext: string
}

/** 列表查询请求 */
export interface ApiKeyListRequest {
  page?: number
  page_size?: number
  keyword?: string
  status?: 'active' | 'inactive' | 'expired'
}

/** 查询 API Key 列表 */
export function getApiKeys(params?: ApiKeyListRequest): Promise<ApiResponse<ApiKeyListResponse>> {
  return Axios.get('/api/v1/api-keys', { params })
}

/** 创建 API Key */
export function createApiKey(data: {
  name: string
  user_id: string
  expires_at?: string | null
  allowed_ips?: string[]
}): Promise<ApiResponse<ApiKeyCreateResponse>> {
  return Axios.post('/api/v1/api-keys', data)
}

/** 更新 API Key */
export function updateApiKey(
  id: string,
  data: {
    name?: string
    status?: 'active' | 'inactive'
    expires_at?: string | null
    allowed_ips?: string[]
  }
): Promise<ApiResponse> {
  return Axios.put(`/api/v1/api-keys/${id}`, data)
}

/** 删除 API Key */
export function deleteApiKey(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/api-keys/${id}`)
}