package utils

type Ordering string

const (
	OrderAsc  Ordering = "asc"
	OrderDesc Ordering = "desc"
)

type Query struct {
	Limit  int
	Offset int
	Order  Ordering
}

type QueryOption interface {
	Apply(*Query)
}

func WithLimit(limit uint) QueryOption {
	return withLimit(limit)
}

type withLimit uint

func (w withLimit) Apply(q *Query) {
	q.Limit = int(w)
}

func WithOffset(offset uint) QueryOption {
	return withOffset(offset)
}

type withOffset uint

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
