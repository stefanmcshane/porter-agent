package utils

type Ordering string

const (
	OrderAsc  Ordering = "asc"
	OrderDesc Ordering = "desc"
)

type Query struct {
	Limit  int
	Offset int
	SortBy string
	Order  Ordering
}

type QueryOption interface {
	Apply(*Query)
}

func WithLimit(limit uint) QueryOption {
	return withLimit(limit)
}

type withLimit int

func (w withLimit) Apply(q *Query) {
	q.Limit = int(w)
}

func WithOffset(offset int64) QueryOption {
	return withOffset(offset)
}

type withOffset int

func (w withOffset) Apply(q *Query) {
	q.Offset = int(w)
}

func WithOrder(order Ordering) QueryOption {
	return withOrder(order)
}

type withOrder Ordering

func (w withOrder) Apply(q *Query) {
	q.Order = Ordering(w)
}

func WithSortBy(sortBy string) QueryOption {
	return withSortBy(sortBy)
}

type withSortBy string

func (w withSortBy) Apply(q *Query) {
	q.SortBy = string(w)
}
