package common

import (
	"circledigital.in/real-state-erp/utils/custom"
	"time"
)

// paginatedModel defines the interface for models that can be paginated
type paginatedModel interface {
	GetCreatedAt() time.Time
}

// CreatePaginatedResponse creates PaginatedData for the list of paginatedModel
func CreatePaginatedResponse[T paginatedModel](data *[]T) *custom.PaginatedData {
	result := *data
	pageInfo := custom.PageInfo{}

	if len(result) > custom.LIMIT {
		result = result[:custom.LIMIT]
		pageInfo = custom.PageInfo{
			NextPage: true,
			Cursor:   encodeCursor(result[custom.LIMIT-1].GetCreatedAt()),
		}
	}

	return &custom.PaginatedData{
		PageInfo: pageInfo,
		Items:    result,
	}
}