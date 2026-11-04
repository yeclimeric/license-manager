import { describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useDataList, type PageRequest } from './useDataList'

interface Row {
  id: string
}

const makeServer = (total: number) => {
  const rows: Row[] = Array.from({ length: total }, (_, i) => ({ id: String(i + 1) }))
  return rows
}

describe('useDataList — 删除后回退不变量', () => {
  it('删除末页最后一条 → 自动回退到上一页，且只发一次请求', async () => {
    const calls: PageRequest<{ keyword: string }>[] = []
    let rows = makeServer(21) // 3 页 × 10

    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async (req) => {
        calls.push(req)
        const start = (req.page - 1) * req.pageSize
        return { list: rows.slice(start, start + req.pageSize), total: rows.length }
      }
    })

    list.page.value = 3
    list.pageSize.value = 10
    await list.reload()
    expect(list.items.value.length).toBe(1) // 第 3 页只有 1 条

    rows = rows.slice(0, 20) // 删掉末页那一条
    const before = calls.length
    await list.remove(async () => {
      /* 后端删除成功 */
    })

    // 断言 1：页码回退到 2
    expect(list.page.value).toBe(2)
    // 断言 2：回退后只补发一次请求，且请求的是第 2 页
    const newCalls = calls.slice(before)
    expect(newCalls).toHaveLength(1)
    expect(newCalls[0].page).toBe(2)
    // 断言 3：表格有数据，不停留在空页
    expect(list.items.value.length).toBe(10)
  })

  it('删除非末页最后一条 → 页码不变', async () => {
    const calls: PageRequest<{ keyword: string }>[] = []
    const rows = makeServer(30)

    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async (req) => {
        calls.push(req)
        const start = (req.page - 1) * req.pageSize
        return { list: rows.slice(start, start + req.pageSize), total: rows.length }
      }
    })

    list.page.value = 2
    list.pageSize.value = 10
    await list.reload()
    expect(list.items.value.length).toBe(10)

    await list.remove(async () => {})
    expect(list.page.value).toBe(2)
  })

  it('第一页只有一条时删除 → 不回退到第 0 页', async () => {
    const rows = makeServer(1)
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async (req) => ({
        list: rows.slice((req.page - 1) * req.pageSize, req.page * req.pageSize),
        total: rows.length
      })
    })

    list.pageSize.value = 10
    await list.reload()
    expect(list.items.value.length).toBe(1)

    await list.remove(async () => {})
    expect(list.page.value).toBe(1)
    expect(list.page.value).toBeGreaterThanOrEqual(1)
  })
})

describe('useDataList — 分页与筛选语义', () => {
  it('query() 重置到第 1 页并重查', async () => {
    const calls: PageRequest<{ keyword: string }>[] = []
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async (req) => {
        calls.push(req)
        return { list: [], total: 0 }
      }
    })

    list.page.value = 5
    await list.query()
    expect(list.page.value).toBe(1)
    expect(calls.at(-1)?.page).toBe(1)
  })

  it('reset() 恢复初始筛选条件', async () => {
    const calls: PageRequest<{ keyword: string }>[] = []
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async (req) => {
        calls.push(req)
        return { list: [], total: 0 }
      }
    })

    list.filters.keyword = 'abc'
    await list.reset()
    expect(list.filters.keyword).toBe('')
    expect(calls.at(-1)?.filters.keyword).toBe('')
  })

  it('setPageSize() 重置到第 1 页', async () => {
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async () => ({ list: [], total: 0 })
    })

    list.page.value = 4
    await list.setPageSize(50)
    expect(list.page.value).toBe(1)
    expect(list.pageSize.value).toBe(50)
  })

  it('后端 total 收缩导致页码越界时自动夹紧并补查', async () => {
    const calls: PageRequest<{ keyword: string }>[] = []
    let total = 100
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async (req) => {
        calls.push(req)
        const start = (req.page - 1) * req.pageSize
        const all = Array.from({ length: total }, (_, i) => ({ id: String(i + 1) }))
        return { list: all.slice(start, start + req.pageSize), total }
      }
    })

    list.pageSize.value = 10
    list.page.value = 10
    await list.reload()
    expect(list.page.value).toBe(10)

    total = 15 // 服务端数据大幅减少，page 10 已越界
    const before = calls.length
    await list.reload()

    expect(list.page.value).toBe(2)
    expect(calls.length - before).toBe(2) // 原请求 + 夹紧后的补查
  })
})

describe('useDataList — 错误与并发', () => {
  it('请求失败时写入 error 且不吞掉异常语义', async () => {
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: async () => {
        throw new Error('boom')
      },
      describeError: () => '加载失败'
    })

    await list.reload()
    expect(list.error.value).toBe('加载失败')
    expect(list.items.value).toEqual([])
  })

  it('并发相同请求被合并为一次', async () => {
    const fetcher = vi.fn(async () => ({ list: [], total: 0 }))
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: fetcher
    })

    await Promise.all([list.reload(), list.reload(), list.reload()])
    expect(fetcher).toHaveBeenCalledTimes(1)
  })

  it('后发请求覆盖先发请求（竞态保护）', async () => {
    const resolvers: Array<(v: { list: Row[]; total: number }) => void> = []
    const list = useDataList<Row, { keyword: string }>({
      initialFilters: { keyword: '' },
      fetchPage: () => new Promise((resolve) => resolvers.push(resolve))
    })

    const first = list.reload()
    list.filters.keyword = 'second'
    const second = list.reload()

    resolvers[1]?.({ list: [{ id: 'second' }], total: 1 })
    resolvers[0]?.({ list: [{ id: 'first' }], total: 1 })
    await Promise.all([first, second])
    await nextTick()

    expect(list.items.value).toEqual([{ id: 'second' }])
  })
})
