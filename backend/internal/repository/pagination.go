package repository

// 分页契约常量：所有列表查询共用，避免各处硬编码不一致。
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// NormalizePage 归一化分页参数（单一真源）。
//
// 历史问题：repository 层把 page_size>100 截断为 100，而 4 个 service 层
// 只兜底 <=0、不截断，导致请求 page_size=500 时按 100 查询却按 500 计算
// total_pages 并返回 PageSize:500，响应与实际查询不一致。
// 现在 repository 与 service 都调用本函数，保证两处取值恒等。
func NormalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

// CalcTotalPages 计算总页数（单一真源）。
//
// 历史上存在 3 种算法：除法+余数补一、math.Ceil(float64)、
// (total+pageSize-1)/pageSize，其中前两种在 pageSize=0 时分别会
// 除零 panic 与 +Inf 转 int 未定义行为。统一为整数上取整并显式防零。
func CalcTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 || total <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}
