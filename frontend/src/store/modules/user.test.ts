import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

// store 模块顶层会读 localStorage，vitest 默认 node 环境没有它，
// 必须在动态 import 之前 stub。
vi.stubGlobal('localStorage', {
  getItem: vi.fn(() => null),
  setItem: vi.fn(),
  removeItem: vi.fn()
})

const { useUserStore } = await import('./user')

const makeStore = () => {
  setActivePinia(createPinia())
  return useUserStore()
}

describe('useUserStore 权限判断', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('super_admin 是管理员且通配放行任意权限', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'root', role: 'super_admin' })
    expect(store.isAdmin).toBe(true)
    expect(store.hasPermission('license:delete')).toBe(true)
  })

  it('历史角色 admin 同样视为超级管理员', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'old', role: 'admin' })
    expect(store.isAdmin).toBe(true)
    expect(store.hasPermission('menu:create')).toBe(true)
  })

  it('operator / viewer 不是管理员，无权限集合时一律拒绝', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator' })
    expect(store.isAdmin).toBe(false)
    // 后端登录响应不含 permissions，此时不能误放行
    expect(store.hasPermission('license:read')).toBe(false)
  })

  it('后端补充 permissions 后按集合精确判断', () => {
    const store = makeStore()
    store.setUserInfo({
      username: 'op',
      role: 'operator',
      permissions: ['license:read', 'customer:read']
    })
    expect(store.hasPermission('license:read')).toBe(true)
    expect(store.hasPermission('license:delete')).toBe(false)
  })

  it('permissions 中的 * 通配放行', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator', permissions: ['*'] })
    expect(store.hasPermission('anything:write')).toBe(true)
  })

  it('未登录时既不管理员也不持有权限', () => {
    const store = makeStore()
    expect(store.isAdmin).toBe(false)
    expect(store.hasPermission('dashboard:read')).toBe(false)
  })
})
