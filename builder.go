package dbr

// Builder builds SQL in Dialect like MySQL/PostgreSQL.
// The raw SQL and values are stored in Buffer.
type Builder interface {
	Build(Dialect, Buffer) error
}

// BuildFunc implements Builder.
type BuildFunc func(Dialect, Buffer) error

// Build calls itself to build SQL.
func (b BuildFunc) Build(d Dialect, buf Buffer) error {
	return b(d, buf)
}
