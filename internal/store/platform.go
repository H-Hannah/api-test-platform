package store

// AllProducts 列表查询时不按 product 过滤。
const AllProducts int64 = 0

// DefaultProductID 入库、测试数据等未指定 product 时的默认归属。
const DefaultProductID int64 = 1

// ResolveProductID 将 0 解析为默认 product。
func ResolveProductID(id int64) int64 {
	if id <= 0 {
		return DefaultProductID
	}
	return id
}
