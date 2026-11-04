/*
 * @Author: 13895237362 2205451508@qq.com
 * @Date: 2025-08-26 13:23:45
 * @LastEditors: 13895237362 2205451508@qq.com
 * @LastEditTime: 2025-10-17 10:35:22
 * @FilePath: \license-manager\frontend\src\store\modules\user.ts
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getMyPermissions } from '@/api/rbac'

interface UserInfo {
  id?: string
  username: string
  email?: string
  avatar?: string
  role: string
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string | null>(localStorage.getItem('token'))
  const userInfo = ref<UserInfo | null>(null)
  const isLoggedIn = ref(false)
  // 权限码集合由 /api/v1/me/permissions 实时拉取，不写进 JWT、不缓存到 localStorage
  const permissions = ref<string[]>([])
  const permissionsLoaded = ref(false)
  const permissionsLoadFailed = ref(false)
  let permissionsRequest: Promise<void> | null = null

  // 计算属性
  // 后端内置角色为 super_admin / operator / viewer（见 database/rbac_seed.go）；
  // 历史 users.role 可能仍是 admin（迁移时 backfill 把 admin 映射为 super_admin），
  // 两个名字都视为超级管理员。旧实现只认 admin，对 super_admin 恒 false。
  const isAdmin = computed(() => {
    const role = userInfo.value?.role
    return role === 'super_admin' || role === 'admin'
  })

  /**
   * 权限判断单一真源：permissions ref（由 loadPermissions 从后端实时拉取）。
   *
   * 超级管理员保留快速通道（role 来自登录响应），其权限集合本身也含 "*"。
   * 不用菜单树判断权限：/v1/menus 返回的是「可见菜单所需权限」而非「用户
   * 拥有的权限集合」，例如 operator 有 authorization_code:create 但菜单树无该码。
   */
  const hasPermission = (permission: string): boolean => {
    if (isAdmin.value) return true
    if (permissions.value.includes('*')) return true
    return permissions.value.includes(permission)
  }

  /**
   * 拉取当前用户权限码集合。并发调用合并为一次请求。
   *
   * 失败时保持空集合并置 permissionsLoadFailed：后端每个端点都有
   * RequirePermission 兜底，前端守卫此时 fail-open 不会造成越权写入，
   * 但也不会因为后端抖动导致整站不可用。
   */
  const loadPermissions = (): Promise<void> => {
    if (permissionsLoaded.value) return Promise.resolve()
    if (permissionsRequest) return permissionsRequest

    permissionsRequest = (async () => {
      try {
        const res = await getMyPermissions()
        permissions.value = Array.isArray(res.data?.permissions) ? res.data.permissions : []
      } catch {
        permissions.value = []
        permissionsLoadFailed.value = true
      } finally {
        permissionsLoaded.value = true
        permissionsRequest = null
      }
    })()
    return permissionsRequest
  }

  // 操作
  const setToken = (newToken: string | null) => {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
  }

  const setUserInfo = (info: UserInfo | null) => {
    userInfo.value = info
    isLoggedIn.value = !!info
    
    // 同时保存到localStorage
    if (info) {
      localStorage.setItem('userInfo', JSON.stringify(info))
    } else {
      localStorage.removeItem('userInfo')
    }
  }

  const setLoginData = (token: string, userInfo: UserInfo) => {
    setToken(token)
    setUserInfo(userInfo)
  }

  const logout = () => {
    setToken(null)
    setUserInfo(null)
    // 换账号后不能残留上一个账号的权限
    permissions.value = []
    permissionsLoaded.value = false
    permissionsLoadFailed.value = false
    permissionsRequest = null
  }

  const updateProfile = (profile: Partial<UserInfo>) => {
    if (userInfo.value) {
      userInfo.value = { ...userInfo.value, ...profile }
    }
  }

  // 初始化时检查登录状态
  const initAuth = () => {
    if (token.value) {
      // 从localStorage恢复用户信息
      const savedUserInfo = localStorage.getItem('userInfo')
      if (savedUserInfo) {
        try {
          const userInfoData = JSON.parse(savedUserInfo)
          setUserInfo(userInfoData)
        } catch (error) {
          console.error('恢复用户信息失败:', error)
          // 如果解析失败，清除无效数据
          localStorage.removeItem('userInfo')
          setToken(null)
        }
      } else {
        // 如果没有保存的用户信息，但有token，设置基本登录状态
        isLoggedIn.value = true
      }
    }
  }

  return {
    // 状态
    token,
    userInfo,
    isLoggedIn,
    permissions,
    permissionsLoaded,
    permissionsLoadFailed,
    // 计算属性
    isAdmin,
    hasPermission,
    // 操作
    setToken,
    setUserInfo,
    setLoginData,
    logout,
    updateProfile,
    initAuth,
    loadPermissions
  }
})