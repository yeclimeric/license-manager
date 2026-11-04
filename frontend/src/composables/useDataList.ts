import { reactive, ref, watch, type Ref } from 'vue'
import type { PageData } from '@/api/types'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from './useInitialSkeleton'

export interface PageRequest<F> {
  page: number
  pageSize: number
  filters: F
}

/** 分页响应统一复用契约层 PageData，避免局部再定义一份漂移类型。 */
export type PageFetcher<T, F> = (request: PageRequest<F>) => Promise<PageData<T>>

export interface DataListPaginationConfig {
  pageSizes: number[]
  layout: string
  pagerCount: number
}

export const DEFAULT_PAGINATION: DataListPaginationConfig = {
  pageSizes: [10, 20, 50, 100],
  layout: 'total, sizes, prev, pager, next, jumper',
  pagerCount: 5
}

export interface UseDataListOptions<T, F extends Record<string, unknown>> {
  fetchPage: PageFetcher<T, F>
  initialFilters: F
  describeError?: (error: unknown) => string
  pageSize?: number
  pagination?: Partial<DataListPaginationConfig>
}

const stableStringify = (value: unknown): string => {
  if (value === null || typeof value !== 'object') return JSON.stringify(value) ?? 'null'
  if (Array.isArray(value)) return `[${value.map(stableStringify).join(',')}]`
  const obj = value as Record<string, unknown>
  return `{${Object.keys(obj).sort().map(k => `${JSON.stringify(k)}:${stableStringify(obj[k])}`).join(',')}}`
}

export const useDataList = <T, F extends Record<string, unknown>>(options: UseDataListOptions<T, F>) => {
  const items = ref([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(options.pageSize ?? 10)
  const loading = ref(false)
  const error = ref('')
  const filters = reactive({ ...options.initialFilters }) as F
  const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)

  const inflight = new Map<string, Promise<void>>()
  let seq = 0

  const requestKey = (): string => `${page.value}|${pageSize.value}|${stableStringify(filters)}`

  const runFetch = async (mySeq: number): Promise<void> => {
    loading.value = true
    error.value = ''
    try {
      const res = await options.fetchPage({
        page: page.value,
        pageSize: pageSize.value,
        filters: { ...filters } as F
      })
      if (mySeq !== seq) return
      items.value = res.list ?? []
      total.value = res.total ?? 0
      markInitialized()
      const maxPage = Math.max(1, Math.ceil(total.value / pageSize.value))
      if (page.value > maxPage) {
        page.value = maxPage
        await reload()
      }
    } catch (e) {
      if (mySeq !== seq) return
      error.value = options.describeError
        ? options.describeError(e)
        : getErrorMessage(e, '')
    } finally {
      if (mySeq === seq) loading.value = false
    }
  }

  const reload = async (): Promise<void> => {
    const key = requestKey()
    const pending = inflight.get(key)
    if (pending) return pending
    const mySeq = ++seq
    const task = runFetch(mySeq).finally(() => { inflight.delete(key) })
    inflight.set(key, task)
    return task
  }

  const query = async (): Promise<void> => {
    page.value = 1
    await reload()
  }

  const reset = async (): Promise<void> => {
    Object.assign(filters, options.initialFilters)
    page.value = 1
    await reload()
  }

  const setPageSize = async (size: number): Promise<void> => {
    if (size === pageSize.value) return
    pageSize.value = size
    page.value = 1
    await reload()
  }

  const remove = async (deleteFn: () => Promise<unknown>): Promise<void> => {
    const isLastRowOfPage = items.value.length === 1 && page.value > 1
    await deleteFn()
    if (isLastRowOfPage) page.value -= 1
    await reload()
  }

  watch(page, () => { void reload() })

  return {
    items, total, page, pageSize, filters, loading, error,
    skeletonVisible, isUpdating,
    pagination: { ...DEFAULT_PAGINATION, ...options.pagination },
    reload, query, reset, setPageSize, remove
  }
}
