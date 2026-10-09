import Axios from './https'
import type { ApiResponse } from './customer'

// 部署模式
export type DeploymentMode = 'online' | 'offline' | 'hybrid'

// 产品项状态
export type ProductOfferingStatus = 'active' | 'inactive'

// 产品项接口类型定义 - 字段与后端 models.ProductOffering 对齐
export interface ProductOffering {
  id: string;
  name: string;
  code: string;                       // 系统生成，不可修改
  alias?: string;                     // 别名，用于界面显示
  description: string;
  is_trial_product: boolean;
  trial_days?: number;
  feature_config: Record<string, any>;
  usage_limits: Record<string, any>;
  custom_parameters: Record<string, any>;
  max_activations?: number;
  duration_days?: number;
  deployment_mode: DeploymentMode;
  status: ProductOfferingStatus;
  status_display: string;
  created_at: string;
  updated_at: string;
  auth_code_count: number;
  user_count: number;
  trial_display?: string;             // 服务端组装的体验展示文案
  can_edit: boolean;
  can_delete: boolean;
}

// 产品项查询请求参数
export interface ProductOfferingQueryRequest {
  page?: number;
  page_size?: number;
  keyword?: string;                   // 模糊搜索名称/别名/编码
  status?: ProductOfferingStatus;
  sort_by?: 'created_at' | 'updated_at' | 'name' | 'code';
  sort_order?: 'asc' | 'desc';
}

// 创建产品项请求参数
// 不含 code：编码由服务端生成（D4 决策）
export interface ProductOfferingCreateRequest {
  name: string;
  alias?: string;
  description?: string;
  is_trial_product?: boolean;
  trial_days?: number;
  feature_config?: Record<string, any>;
  usage_limits?: Record<string, any>;
  custom_parameters?: Record<string, any>;
  max_activations?: number;
  duration_days?: number;
  deployment_mode?: DeploymentMode;
}

// 更新产品项请求参数
// 不含 code：编码不可修改（D4 决策）
export interface ProductOfferingUpdateRequest {
  name?: string;
  alias?: string;
  description?: string;
  is_trial_product?: boolean;
  trial_days?: number;
  feature_config?: Record<string, any>;
  usage_limits?: Record<string, any>;
  custom_parameters?: Record<string, any>;
  max_activations?: number;
  duration_days?: number;
  deployment_mode?: DeploymentMode;
  status?: ProductOfferingStatus;
}

// 产品项列表响应
export interface ProductOfferingQueryResponse {
  code: string;
  message: string;
  timestamp: string;
  data: {
    list: ProductOffering[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
  };
}

/**
 * 查询产品项列表
 */
export function getProductOfferings(
  params?: ProductOfferingQueryRequest
): Promise<ProductOfferingQueryResponse> {
  return Axios.get('/api/v1/product-offerings', { params })
}

/**
 * 获取产品项详情
 */
export function getProductOfferingDetail(id: string): Promise<ApiResponse<ProductOffering>> {
  return Axios.get(`/api/v1/product-offerings/${id}`)
}

/**
 * 创建产品项（编码由服务端生成）
 */
export function createProductOffering(
  data: ProductOfferingCreateRequest
): Promise<ApiResponse<ProductOffering>> {
  return Axios.post('/api/v1/product-offerings', data)
}

/**
 * 更新产品项（编码不可修改）
 */
export function updateProductOffering(
  id: string,
  data: ProductOfferingUpdateRequest
): Promise<ApiResponse<ProductOffering>> {
  return Axios.put(`/api/v1/product-offerings/${id}`, data)
}

/**
 * 删除产品项
 */
export function deleteProductOffering(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/product-offerings/${id}`)
}

/**
 * 启用/停用产品项
 */
export function updateProductOfferingStatus(
  id: string,
  status: ProductOfferingStatus
): Promise<ApiResponse<ProductOffering>> {
  return Axios.patch(`/api/v1/product-offerings/${id}/status`, { status })
}
