package response

import (
	"gorm.io/gorm"
	"math"
)

type ResponsePage[T any] struct {
	List       []T   `json:"list"`
	PageNo     int64 `json:"pageNo"`
	PageSize   int64 `json:"pageSize"`
	TotalCount int64 `json:"totalCount"`
	TotalPage  int64 `json:"totalPage"`
}

func Paginate[T any](db *gorm.DB, pageNo int, pageSize int, result *ResponsePage[T]) error {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Query totalCount count
	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return err
	}

	// Calculate totalCount pages
	totalPage := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	// Query data with pagination
	offset := (pageNo - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&result.List).Error; err != nil {
		return err
	}

	// Set response data
	result.TotalCount = totalCount
	result.TotalPage = int64(totalPage)
	result.PageSize = int64(pageSize)
	result.PageNo = int64(pageNo)

	return nil
}
