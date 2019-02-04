package dbr

type union struct {
	builder []Builder
	all     bool
}

func newUnion(builder ...Builder) *union {
	for i := range builder {
		// Suppress parentheses for SelectStmts.
		if s, ok := builder[i].(*SelectStmt); ok {
			builder[i] = (*selectStmtNoParens)(s)
		}
	}
	return &union{builder: builder}
}

// Union builds `... UNION ...`.
func Union(builder ...Builder) Builder {
	return newUnion(builder...)
}

// UnionAll builds `... UNION ALL ...`.
func UnionAll(builder ...Builder) Builder {
	u := newUnion(builder...)
	u.all = true
	return u
}

func (u *union) Build(_ Dialect, buf Buffer) error {
	for i, b := range u.builder {
		if i > 0 {
			buf.WriteString(" UNION ")
			if u.all {
				buf.WriteString("ALL ")
			}
		}
		buf.WriteString(placeholder)
		buf.WriteValue(b)
	}
	return nil
}

// Hide the SelectStmt type to avoid adding parentheses within a UNION.
type selectStmtNoParens SelectStmt

func (s *selectStmtNoParens) Build(d Dialect, buf Buffer) error {
	return (*SelectStmt)(s).Build(d, buf)
}
