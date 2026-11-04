import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

// store 模块顶层会读 localStorage，vitest 默认 node 环境没有它，
// 且 loadPermissions 依赖的 API 需在动态 import 之前 mock。
const mocks = vi.hoisted(() => ({ getMyPermissions: vi.fn() }))
vi.mock('@/api/rbac', () => ({ getMyPermissions: mocks.getMyPermissions }))
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
    mocks.getMyPermissions.mockReset()
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

  it('operator 无权限集合时一律拒绝', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator' })
    expect(store.isAdmin).toBe(false)
    expect(store.hasPermission('license:read')).toBe(false)
  })

  it('权限集合按码精确判断', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator' })
    store.permissions = ['license:read', 'customer:read']
    expect(store.hasPermission('license:read')).toBe(true)
    expect(store.hasPermission('license:delete')).toBe(false)
  })

  it('权限集合中的 * 通配放行', () => {
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator' })
    store.permissions = ['*']
    expect(store.hasPermission('anything:write')).toBe(true)
  })

  it('未登录时既不管理员也不持有权限', () => {
    const store = makeStore()
    expect(store.isAdmin).toBe(false)
    expect(store.hasPermission('dashboard:read')).toBe(false)
  })

  it('loadPermissions 从后端实时拉取权限集合', async () => {
    mocks.getMyPermissions.mockResolvedValue({
      code: '000000',
      message: 'ok',
      data: { permissions: ['license:read', 'authorization_code:create'] }
    })
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator' })

    await store.loadPermissions()

    expect(store.permissionsLoaded).toBe(true)
    expect(store.permissionsLoadFailed).toBe(false)
    expect(store.hasPermission('authorization_code:create')).toBe(true)
    expect(store.hasPermission('role:read')).toBe(false)
  })

  it('loadPermissions 失败时保持空集合并标记，供守卫 fail-open', async () => {
    mocks.getMyPermissions.mockRejectedValue(new Error('network down'))
    const store = makeStore()
    store.setUserInfo({ username: 'op', role: 'operator' })

    await store.loadPermissions()

    expect(store.permissionsLoaded).toBe(true)
    expect(store.permissionsLoadFailed).toBe(true)
    expect(store.permissions).toEqual([])
    expect(store.hasPermission('license:read')).toBe(false)
  })

  it('并发 loadPermissions 合并为一次请求', async () => {
    let resolveFn: (value: unknown) => void = () => {}
    mocks.getMyPermissions.mockReturnValue(new Promise(resolve => { resolveFn = resolve }))
    const store = makeStore()

    const first = store.loadPermissions()
    const second = store.loadPermissions()
    resolveFn({ code: '000000', message: 'ok', data: { permissions: ['license:read'] } })
    await Promise.all([first, second])

    expect(mocks.getMyPermissions).toHaveBeenCalledTimes(1)
  })

  it('logout 清空权限，避免换账号后残留', async () => {
    mocks.getMyPermissions.mockResolvedValue({
      code: '000000', message: 'ok', data: { permissions: ['license:read'] }
    })
    const store = makeStore()
    await store.loadPermissions()
    expect(store.hasPermission('license:read')).toBe(true)

    store.logout()

    expect(store.permissions).toEqual([])
    expect(store.permissionsLoaded).toBe(false)
    expect(store.hasPermission('license:read')).toBe(false)
  })
})
