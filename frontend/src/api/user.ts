import Axios from './https'
import type { ApiResponse } from './customer'

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  code: string;
  message: string;
  data: {
    token: string;
    expires_in: number;
    user_info: {
      username: string;
      role: string;
    };
  };
}

export function Login(data: LoginRequest): Promise<LoginResponse> {
  return Axios.post('/api/v1/login', data)
}
/** 用户状态 */
export type UserStatus = 'active' | 'locked' | 'inactive'

/** 角色摘要 */
export interface UserRoleItem {
  id: string
  name: string
  display_name: string
}

/** 用户列表项 */
export interface UserItem {
  id: string
  username: string
  email: string
  full_name: string
  phone: string | null
  /** 旧单角色字段，保留以兼容既有调用 */
  role: string
  status: UserStatus
  remark: string | null
  last_login_at: string | null
  created_at: string
  roles: UserRoleItem[]
}

/** 用户详情 */
export interface UserDetail extends UserItem {
  login_attempts: number
  locked_until: string | null
  last_online_ip: string | null
  updated_at: string
}

/** 列表查询请求 */
export interface UserListRequest {
  page?: number
  page_size?: number
  search?: string
  status?: UserStatus
  role?: string
  sort?: 'created_at' | 'updated_at' | 'username' | 'full_name'
  order?: 'asc' | 'desc'
}

/** 列表响应 */
export interface UserListResponse {
  list: UserItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

/** 角色分配项 */
export interface RoleAssignment {
  role: string
}

/** 创建用户请求 */
export interface UserCreateRequest {
  username: string
  email: string
  password: string
  full_name: string
  phone?: string
  status: UserStatus
  remark?: string
  /** 至少一个角色，否则后端拒绝 */
  role_assignments: RoleAssignment[]
}

/** 更新用户请求 */
export interface UserUpdateRequest {
  email?: string
  full_name?: string
  phone?: string
  remark?: string
  status?: UserStatus
  password?: string
  role_assignments?: RoleAssignment[]
}

/** 查询用户列表 */
export function getUsers(params?: UserListRequest): Promise<ApiResponse<UserListResponse>> {
  return Axios.get('/api/v1/users', { params })
}

/** 查询用户详情 */
export function getUserDetail(id: string): Promise<ApiResponse<UserDetail>> {
  return Axios.get(`/api/v1/users/${id}`)
}

/** 创建用户 */
export function createUser(data: UserCreateRequest): Promise<ApiResponse<UserDetail>> {
  return Axios.post('/api/v1/users', data)
}

/** 更新用户 */
export function updateUser(id: string, data: UserUpdateRequest): Promise<ApiResponse<UserDetail>> {
  return Axios.put(`/api/v1/users/${id}`, data)
}

/** 删除用户 */
export function deleteUser(id: string): Promise<ApiResponse> {
  return Axios.delete(`/api/v1/users/${id}`)
}

/** 重置密码 */
export function resetUserPassword(id: string, newPassword: string): Promise<ApiResponse> {
  return Axios.post(`/api/v1/users/${id}/reset-password`, { new_password: newPassword })
}
