package utils

import (
	"fmt"

	"gorm.io/gorm"
)

func Paginate(opts []QueryOption) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := Query{
			Limit:  0,
			Offset: 0,
			Order:  OrderAsc,
			SortBy: "id",
		}

		for _, opt := range opts {
			opt.Apply(&q)
		}

		return db.
			Offset(q.Offset).
			Limit(q.Limit).
			Order(fmt.Sprintf("%s %s", q.SortBy, q.Order))
	}
}
