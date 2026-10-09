import Axios from './https'
import type { ApiResponse } from './customer'

// 功能模板状态
export type FeatureTemplateStatus = 'active' | 'inactive'

// 功能项
export interface FeatureItem {
  key: string;
  name: string;
  description?: string;
  required?: boolean;
}

// 功能模板接口类型定义 - 字段与后端 models.FeatureTemplate 对齐
export interface FeatureTemplate {
  id: string;
  name: string;
  identifier: string;                   // 客户端 SDK 引用，创建后不可修改
  description: string;
  category: string;
  category_display: string;
  feature_config: FeatureItem[];
  usage_limits: Record<string, any>;
  custom_parameters: Record<string, any>;
  sort_order: number;
  status: FeatureTemplateStatus;
  status_display: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
  can_delete: boolean;
}

// 功能模板分类
export interface FeatureTemplateCategory {
  id: string;
  value: string;
  name: string;
  label: string;
  sort_order: number;
  is_system: boolean;
  template_count: number;
  can_delete: boolean;
}

// 查询请求参数
export interface FeatureTemplateQueryRequest {
  page?: number;
  page_size?: number;
  keyword?: string;
  category?: string;
  status?: FeatureTemplateStatus;
  sort_by?: 'created_at' | 'updated_at' | 'name' | 'identifier' | 'sort_order';
  sort_order?: 'asc' | 'desc';
}

// 创建请求参数
// identifier 可留空：留空时服务端按名称生成
export interface FeatureTemplateCreateRequest {
  name: string;
  identifier?: string;
  description?: string;
  category?: string;
  feature_config?: FeatureItem[];
  usage_limits?: Record<string, any>;
  custom_parameters?: Record<string, any>;
  sort_order?: number;
}

// 更新请求参数
// 不含 identifier：标识符不可修改
export interface FeatureTemplateUpdateRequest {
  name?: string;
  description?: string;
  category?: string;
  feature_config?: FeatureItem[];
  usage_limits?: Record<string, any>;
  custom_parameters?: Record<string, any>;
  sort_order?: number;
  status?: FeatureTemplateStatus;
}

// 列表响应
export interface FeatureTemplateQueryResponse {
  code: string;
  message: string;
  timestamp: string;
  data: {
    list: FeatureTemplate[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
  };
}

// 分类请求
export interface CategoryCreateRequest {
  value: string;
  label: string;
  sort_order?: number;
}

export interface CategoryUpdateRequest {
  label?: string;
  sort_order?: number;
}

/**
 * 查询功能模板列表
 */
export function getFeatureTemplates(
  params?: FeatureTemplateQueryRequest
): Promise<FeatureTemplateQueryResponse> {
  return Axios.get('/api/v1/feature-templates', { params })
}

/**
 * 获取功能模板详情
 */
export function getFeatureTemplateDetail(id: string): Promise<ApiResponse<FeatureTemplate>> {
  return Axios.get(`/api/v1/feature-templates/${id}`)
}

/**
 * 创建功能模板（标识符可留空自动生成）
 */
export function createFeatureTemplate(
  data: FeatureTemplateCreateRequest
): Promise<ApiResponse<FeatureTemplate>> {
  return Axios.post('/api/v1/feature-templates', data)
}

/**
 * 更新功能模板（标识符不可修改）
 */
export function updateFeatureTemplate(
  id: string,
  data: FeatureTemplateUpdateRequest
): Promise<ApiResponse<FeatureTemplate>> {
  return Axios.put(`/api/v1/feature-templates/${id}`, data)
}

/**
 * 删除功能模板
 */
export function deleteFeatureTemplate(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/feature-templates/${id}`)
}

/**
 * 查询分类列表
 */
export function getFeatureTemplateCategories(): Promise<
  ApiResponse<{ list: FeatureTemplateCategory[]; total: number }>
> {
  return Axios.get('/api/v1/feature-templates/categories')
}

/**
 * 新建自定义分类
 */
export function createFeatureTemplateCategory(
  data: CategoryCreateRequest
): Promise<ApiResponse<FeatureTemplateCategory>> {
  return Axios.post('/api/v1/feature-template-categories', data)
}

/**
 * 更新自定义分类
 */
export function updateFeatureTemplateCategory(
  id: string,
  data: CategoryUpdateRequest
): Promise<ApiResponse<FeatureTemplateCategory>> {
  return Axios.put(`/api/v1/feature-template-categories/${id}`, data)
}

/**
 * 删除自定义分类
 */
export function deleteFeatureTemplateCategory(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/feature-template-categories/${id}`)
}
