package dbr

import (
	"context"
)

// ReturnInt64 executes the SelectStmt and returns the value as an int64.
func (b *SelectStmt) ReturnInt64() (int64, error) {
	var v int64
	err := b.LoadOne(&v)
	return v, err
}

// ReturnInt64Context executes the SelectStmt and returns the value as an int64.
// The given context is passed into the query runner.
func (b *SelectStmt) ReturnInt64Context(ctx context.Context) (int64, error) {
	var v int64
	err := b.LoadOneContext(ctx, &v)
	return v, err
}

// ReturnInt64s executes the SelectStmt and returns the value as a slice of int64s.
func (b *SelectStmt) ReturnInt64s() ([]int64, error) {
	var v []int64
	_, err := b.Load(&v)
	return v, err
}

// ReturnInt64sContext executes the SelectStmt and returns the value as a slice of int64s.
// The given context is passed into the query runner.
func (b *SelectStmt) ReturnInt64sContext(ctx context.Context) ([]int64, error) {
	var v []int64
	_, err := b.LoadContext(ctx, &v)
	return v, err
}

// ReturnUint64 executes the SelectStmt and returns the value as an uint64.
func (b *SelectStmt) ReturnUint64() (uint64, error) {
	var v uint64
	err := b.LoadOne(&v)
	return v, err
}

// ReturnUint64Context executes the SelectStmt and returns the value as an uint64.
// The given context is passed into the query runner.
func (b *SelectStmt) ReturnUint64Context(ctx context.Context) (uint64, error) {
	var v uint64
	err := b.LoadOneContext(ctx, &v)
	return v, err
}

// ReturnUint64s executes the SelectStmt and returns the value as a slice of uint64s.
func (b *SelectStmt) ReturnUint64s() ([]uint64, error) {
	var v []uint64
	_, err := b.Load(&v)
	return v, err
}

// ReturnUint64sContext executes the SelectStmt and returns the value as a slice of uint64s.
// The given context is passed into the query runner.
func (b *SelectStmt) ReturnUint64sContext(ctx context.Context) ([]uint64, error) {
	var v []uint64
	_, err := b.LoadContext(ctx, &v)
	return v, err
}

// ReturnString executes the SelectStmt and returns the value as a string.
func (b *SelectStmt) ReturnString() (string, error) {
	var v string
	err := b.LoadOne(&v)
	return v, err
}

// ReturnStringContext executes the SelectStmt and returns the value as a string.
// The given context is passed into the query runner.
func (b *SelectStmt) ReturnStringContext(ctx context.Context) (string, error) {
	var v string
	err := b.LoadOneContext(ctx, &v)
	return v, err
}

// ReturnStrings executes the SelectStmt and returns the value as a slice of strings.
func (b *SelectStmt) ReturnStrings() ([]string, error) {
	var v []string
	_, err := b.Load(&v)
	return v, err
}

// ReturnStringsContext executes the SelectStmt and returns the value as a slice of strings.
// The given context is passed into the query runner.
func (b *SelectStmt) ReturnStringsContext(ctx context.Context) ([]string, error) {
	var v []string
	_, err := b.LoadContext(ctx, &v)
	return v, err
}
