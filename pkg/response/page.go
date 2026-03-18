package response

type Page[T any] struct {
	PageSize int   `json:"pageSize"`
	PageNum  int   `json:"pageNum"`
	Total    int64 `json:"total"`
	Data     T     `json:"data"`
}

func PageOffset(pageNum, pageSize int) (int, int, int, int) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize
	limit := pageSize
	return pageNum, pageSize, offset, limit
}

func ParsePage[T any](data T, pageNum, pageSize int, total int64) *Page[T] {
	return &Page[T]{
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
		Data:     data,
	}
}
