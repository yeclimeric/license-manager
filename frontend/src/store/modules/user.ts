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

interface UserInfo {
  id?: string
  username: string
  email?: string
  avatar?: string
  role: string
  permissions?: string[]
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string | null>(localStorage.getItem('token'))
  const userInfo = ref<UserInfo | null>(null)
  const isLoggedIn = ref(false)

  // 计算属性
  // 后端内置角色为 super_admin / operator / viewer（见 database/rbac_seed.go）；
  // 历史 users.role 可能仍是 admin（迁移时 backfill 把 admin 映射为 super_admin），
  // 两个名字都视为超级管理员。旧实现只认 admin，对 super_admin 恒 false。
  const isAdmin = computed(() => {
    const role = userInfo.value?.role
    return role === 'super_admin' || role === 'admin'
  })

  /**
   * 权限判断单一真源。
   *
   * 注意：后端登录响应（models.LoginData.UserInfo）目前只含 username + role，
   * 不含 permissions，且 /v1/menus 返回的是「可见菜单所需权限」而非「用户
   * 拥有的权限集合」（例如 operator 有 authorization_code:create，但菜单树上
   * 没有该码）。因此这里不能作为路由守卫的严格依据——会把合法用户误挡。
   * 当前语义：超级管理员通配放行；其余角色仅在后端将来补充 permissions 后生效。
   */
  const hasPermission = (permission: string): boolean => {
    if (isAdmin.value) return true
    const permissions = userInfo.value?.permissions
    if (!permissions || permissions.length === 0) return false
    return permissions.includes('*') || permissions.includes(permission)
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
    // 计算属性
    isAdmin,
    hasPermission,
    // 操作
    setToken,
    setUserInfo,
    setLoginData,
    logout,
    updateProfile,
    initAuth
  }
})