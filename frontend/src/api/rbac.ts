import Axios from './https'
import type { ApiResponse } from './customer'

/** 菜单类型 */
export type MenuType = 'menu' | 'directory'

/** 菜单节点 */
export interface MenuItem {
  id: string
  /** i18n 键，如 menu.product_management；不是翻译后的文本 */
  name_key: string
  type: MenuType
  path: string
  icon: string
  parent_id: string | null
  /** 所需权限码，如 product_offering:read */
  permission: string
  sort_order: number
  status: 'active' | 'inactive'
  is_system: boolean
  hidden: boolean
  description: string
  created_at?: string
  updated_at?: string
  children: MenuItem[] | null
  can_delete: boolean
}

/** 权限资源 */
export interface PermissionResource {
  id: string
  resource: string
  name_key: string
  name: string
  sort_order: number
  actions: PermissionAction[]
}

/** 权限动作 */
export interface PermissionAction {
  code: string
  name: string
}

/** 角色 */
export interface Role {
  id: string
  name: string
  display_name: string
  description: string
  is_system: boolean
  status: 'active' | 'inactive'
  created_at: string
  permissions?: string[]
  menu_ids?: string[]
  user_count: number
  can_delete: boolean
}

/** 创建角色请求 */
export interface RoleCreateRequest {
  name: string
  display_name: string
  description?: string
  permissions?: string[]
  menu_ids?: string[]
}

/** 更新角色请求 */
export interface RoleUpdateRequest {
  display_name?: string
  description?: string
  status?: 'active' | 'inactive'
  permissions?: string[]
  menu_ids?: string[]
}

/** 菜单创建/更新请求 */
export interface MenuUpsertRequest {
  name_key?: string
  type?: MenuType
  path?: string
  icon?: string
  parent_id?: string | null
  permission?: string
  sort_order?: number
  status?: 'active' | 'inactive'
  hidden?: boolean
  description?: string
}

/**
 * 获取当前用户可见菜单树
 *
 * 侧边栏据此渲染，用户之间的差异由后端按角色过滤。
 */
export function getUserMenus(): Promise<ApiResponse<{ list: MenuItem[]; total: number }>> {
  return Axios.get('/api/v1/menus')
}

/**
 * 获取全部菜单树（菜单管理界面用）
 */
export function getAllMenus(): Promise<ApiResponse<{ list: MenuItem[]; total: number }>> {
  return Axios.get('/api/v1/menus/all')
}

/**
 * 创建菜单
 */
export function createMenu(data: MenuUpsertRequest): Promise<ApiResponse<MenuItem>> {
  return Axios.post('/api/v1/menus', data)
}

/**
 * 更新菜单
 */
export function updateMenu(id: string, data: MenuUpsertRequest): Promise<ApiResponse<MenuItem>> {
  return Axios.put(`/api/v1/menus/${id}`, data)
}

/**
 * 删除菜单
 */
export function deleteMenu(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/menus/${id}`)
}

/**
 * 获取权限矩阵
 */
export function getPermissionMatrix(): Promise<
  ApiResponse<{ resources: PermissionResource[]; actions: PermissionAction[] }>
> {
  return Axios.get('/api/v1/permissions')
}

/**
 * 获取角色列表
 */
export function getRoles(keyword?: string): Promise<ApiResponse<{ list: Role[]; total: number }>> {
  return Axios.get('/api/v1/roles', { params: { keyword: keyword || undefined } })
}

/**
 * 获取角色详情
 */
export function getRoleDetail(id: string): Promise<ApiResponse<Role>> {
  return Axios.get(`/api/v1/roles/${id}`)
}

/**
 * 创建角色
 */
export function createRole(data: RoleCreateRequest): Promise<ApiResponse<Role>> {
  return Axios.post('/api/v1/roles', data)
}

/**
 * 更新角色
 */
export function updateRole(id: string, data: RoleUpdateRequest): Promise<ApiResponse<Role>> {
  return Axios.put(`/api/v1/roles/${id}`, data)
}

/**
 * 删除角色
 */
export function deleteRole(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/roles/${id}`)
}

/**
 * 获取用户绑定的角色ID
 */
export function getUserRoles(userId: string): Promise<ApiResponse<string[]>> {
  return Axios.get(`/api/v1/users/${userId}/roles`)
}

/**
 * 设置用户绑定的角色
 */
export function setUserRoles(userId: string, roleIds: string[]): Promise<ApiResponse> {
  return Axios.put(`/api/v1/users/${userId}/roles`, { role_ids: roleIds })
}
