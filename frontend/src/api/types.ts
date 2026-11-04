/**
 * 前后端数据契约唯一真源。
 *
 * 历史上 ApiResponse 在 customer.ts / license.ts / dashboard.ts 各定义一份，
 * 三份互不一致（dashboard 版缺 timestamp 且 data 必填），其余 8 个 api 文件
 * 又从 customer.ts 借用类型，形成隐性耦合。任何视图/组件不再从具体业务 api
 * 文件导入响应类型，一律从本文件导入。
 */

/**
 * 统一响应信封。后端业务码为字符串，成功为 '000000'。
 *
 * data 设为必填：错误响应由 axios 拦截器 reject，能走到业务代码的
 * 只有 code='000000' 的成功响应，其 data 必然存在（可为 null）。
 * 历史 dashboard 局部定义正是 data 必填，另两份却写成可选，此处统一。
 */
export interface ApiResponse<T = unknown> {
  code: string
  message: string
  /** 后端保留字段，UI 不消费；保留以对齐线上响应形状。 */
  timestamp?: string
  data: T
}

/**
 * 统一分页数据。后端 snake_case，前端只在契约层做一次映射，
 * 视图层统一使用 camelCase 的 page / pageSize / totalPages。
 */
export interface PageData<T> {
  list: T[]
  total: number
  page?: number
  page_size?: number
  total_pages?: number
}
