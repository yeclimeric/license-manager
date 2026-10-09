import Axios from './https'
import type { ApiResponse } from './customer'

/** 审计结果 */
export type AuditStatus = 'success' | 'failure'

/** 审计详情 */
export interface AuditDetail {
  path?: string
  method?: string
  req_body?: string
  resp_body?: string
  resp_id?: string
}

/** 审计日志条目 */
export interface AuditLog {
  id: string
  user_id: string
  username: string
  /** 形如 "POST /v1/authorization-codes" */
  action: string
  resource_type: string
  resource_id: string
  product_id: string | null
  detail: AuditDetail
  ip: string
  user_agent: string
  status: AuditStatus
  created_at: string
}

/** 列表查询请求 */
export interface AuditLogQueryRequest {
  page?: number
  page_size?: number
  resource_type?: string
  user_id?: string
  status?: AuditStatus
  action?: string
  product_id?: string
  start_date?: string
  end_date?: string
  sort_by?: 'created_at' | 'action' | 'resource_type' | 'status'
  sort_order?: 'asc' | 'desc'
}

/** 列表响应 */
export interface AuditLogQueryResponse {
  code: string
  message: string
  timestamp: string
  data: {
    list: AuditLog[]
    total: number
    page: number
    page_size: number
    total_pages: number
  }
}

/** 统计响应 */
export interface AuditStatistics {
  total: number
  success: number
  failure: number
  by_resource: Record<string, number>
  by_status: Record<string, number>
}

/**
 * 查询审计日志列表
 */
export function getAuditLogs(params?: AuditLogQueryRequest): Promise<AuditLogQueryResponse> {
  return Axios.get('/api/v1/audit-logs', { params })
}

/**
 * 查询审计统计
 */
export function getAuditStatistics(): Promise<ApiResponse<AuditStatistics>> {
  return Axios.get('/api/v1/audit-logs/statistics')
}