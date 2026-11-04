// 列表页组件契约测试（happy-dom 环境）
//
// 目的：useDataList 迁移后，7 个列表页在组件层仍保持两条核心契约：
//   1) 挂载（onMounted）即触发初始加载（fetchPage → HTTP GET）
//   2) 翻页（el-pagination current-change）触发带 page=2 的新请求
//
// 统一拦截 HTTP 层（@/api/https），避免逐页 mock 各自 api 模块；
// 分页参数以宽松方式校验（字段名含 page/current 且值为 2），兼容各页差异。
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import i18n from '@/i18n'

import LinenseList from '@/views/Licenses/LinenseList.vue'
import Customers from '@/views/Customers/index.vue'
import Users from '@/views/Users/index.vue'
import ApiKeys from '@/views/ApiKeys/index.vue'
import AuditLogs from '@/views/AuditLogs/index.vue'
import Products from '@/views/Products/index.vue'
import FeatureTemplates from '@/views/FeatureTemplates/index.vue'

// ---- HTTP 层统一 stub ----
const httpMock = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  patch: vi.fn()
}))
vi.mock('@/api/https', () => ({ default: httpMock }))

// ---- vue-router stub（仅覆盖组件用到的 useRoute/useRouter）----
const mockRouter = {
  push: vi.fn(),
  replace: vi.fn(),
  back: vi.fn(),
  currentRoute: { value: { path: '/' } }
}
const mockRoute = { query: {}, params: {}, path: '/', name: undefined, fullPath: '/' }
vi.mock('vue-router', async (importOriginal) => ({
  ...(await importOriginal()),
  useRoute: () => mockRoute,
  useRouter: () => mockRouter
}))

// 万能成功响应：无论组件取 .data.list / .data.items / .data 都能容错
const okResponse = () => ({ code: '000000', data: { list: [], total: 0, items: [] }, message: 'ok' })

beforeEach(() => {
  httpMock.get.mockImplementation(async () => okResponse())
  httpMock.post.mockResolvedValue(okResponse())
  httpMock.put.mockResolvedValue(okResponse())
  httpMock.delete.mockResolvedValue(okResponse())
  httpMock.patch.mockResolvedValue(okResponse())
  mockRouter.push.mockClear()
  mockRouter.replace.mockClear()
})

// 翻页后查找是否出现 page 类字段等于 2 的请求参数（兼容 page/pageNum/current/current_page）
const hasPage2 = (calls: any[][]) =>
  calls.some((c) => {
    const params = c[1]?.params
    if (!params || typeof params !== 'object') return false
    return Object.entries(params).some(([k, v]) => /page|current/i.test(k) && v === 2)
  })

const pages = [
  { name: 'LinenseList', component: LinenseList },
  { name: 'Customers', component: Customers },
  { name: 'Users', component: Users },
  { name: 'ApiKeys', component: ApiKeys },
  { name: 'AuditLogs', component: AuditLogs },
  { name: 'Products', component: Products },
  { name: 'FeatureTemplates', component: FeatureTemplates }
]

const mountPage = (component: any) =>
  mount(component, {
    global: {
      plugins: [i18n],
      directives: { loading: { mounted() {}, updated() {} } }
    }
  })

describe.each(pages)('列表页组件契约', ({ component }) => {
  it('挂载（onMounted）即触发初始加载', async () => {
    mountPage(component)
    await flushPromises()
    expect(httpMock.get).toHaveBeenCalled()
  })

  it('翻页（current-change=2）触发带 page=2 的重新加载', async () => {
    const wrapper = mountPage(component)
    await flushPromises()
    expect(httpMock.get).toHaveBeenCalled()

    const pagination = wrapper.findComponent({ name: 'ElPagination' })
    expect(pagination.exists()).toBe(true)

    const before = httpMock.get.mock.calls.length
    await pagination.vm.$emit('update:current-page', 2)
    await pagination.vm.$emit('current-change', 2)
    await flushPromises()

    const after = httpMock.get.mock.calls.slice(before)
    expect(after.length).toBeGreaterThan(0)
    expect(hasPage2(after)).toBe(true)
  })
})
