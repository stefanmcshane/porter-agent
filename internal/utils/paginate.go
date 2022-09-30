package utils

import (
	"fmt"
	"math"

	"gorm.io/gorm"
)

type PaginatedResult struct {
	NumPages    int64
	CurrentPage int64
	NextPage    int64
}

func Paginate(opts []QueryOption, db *gorm.DB, pagination *PaginatedResult) func(db *gorm.DB) *gorm.DB {
	q := Query{
		Limit:  0,
		Offset: 0,
		Order:  OrderAsc,
		SortBy: "id",
	}

	for _, opt := range opts {
		opt.Apply(&q)
	}

	if q.Limit == 0 {
		// default to returing 50 results per page
		q.Limit = 50
	}

	var totalRows int64

	db.Count(&totalRows)

	pagination.NumPages = int64(math.Ceil(float64(totalRows) / float64(q.Limit)))

	if q.Offset > 0 {
		pagination.CurrentPage = int64(q.Offset / q.Limit)

		if pagination.CurrentPage < pagination.NumPages {
			pagination.NextPage = pagination.CurrentPage + 1
		} else {
			pagination.NextPage = pagination.CurrentPage
		}
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.
			Offset(q.Offset).
			Limit(q.Limit).
			Order(fmt.Sprintf("%s %s", q.SortBy, q.Order))
	}
}
