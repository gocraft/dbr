package dbr

// I is quoted identifier
type I string

// Build quotes string with dialect.
func (i I) Build(d Dialect, buf Buffer) error {
	buf.WriteString(d.QuoteIdent(string(i)))
	return nil
}

// As creates an alias for expr.
func (i I) As(alias string) Builder {
	return as(i, alias)
}

func as(expr interface{}, alias string) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString(placeholder)
		buf.WriteValue(expr)
		buf.WriteString(" AS ")
		buf.WriteString(d.QuoteIdent(alias))
		return nil
	})
}
