package repository

import "strings"

// SafeOrderClause 构造经白名单校验的 ORDER BY 子句。
//
// Repository 是导出接口，不能假设调用方一定经过 handler 层的
// binding:"oneof" 校验——service 层或未来的调用方可以直接构造请求对象。
// 因此所有进入 SQL 的排序列名与方向都在此收敛，非法值静默回退到 allowed[0]。
//
// 参数：
//   - column / direction: 来自请求的排序字段与方向
//   - prefix: 表别名前缀（如 "licenses."），为空表示不带前缀
//   - allowed: 允许的列名白名单，第一项为缺省列
//
// 返回值恒为 "<prefix><allowed 中的列> ASC|DESC"，
// column 与 direction 中的任何字符都不会原样进入 SQL。
func SafeOrderClause(column, direction, prefix string, allowed ...string) string {
	if len(allowed) == 0 {
		return ""
	}

	col := allowed[0]
	for _, candidate := range allowed[1:] {
		if candidate == column {
			col = candidate
			break
		}
	}

	dir := "DESC"
	if strings.EqualFold(direction, "asc") {
		dir = "ASC"
	}

	return prefix + col + " " + dir
}
