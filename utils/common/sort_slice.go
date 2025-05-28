package common

import (
	"sort"
	"time"
)

type dbModel interface {
	GetCreatedAt() time.Time
}

func SortDbModels[T dbModel](models []T) []T {
	sort.SliceStable(models, func(i, j int) bool {
		return models[i].GetCreatedAt().Before(models[j].GetCreatedAt())
	})
	return models
}
