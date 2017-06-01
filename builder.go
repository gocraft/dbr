package dbr

// Builder builds sql in one dialect like MySQL/PostgreSQL
// e.g. XxxBuilder
type Builder interface {
	Build(Dialect, Buffer) error
}

// BuildFunc is an adapter to allow the use of ordinary functions as Builder
type BuildFunc func(Dialect, Buffer) error

// Build implements Builder interface
func (b BuildFunc) Build(d Dialect, buf Buffer) error {
	return b(d, buf)
}
