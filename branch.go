package dbr

// When creates a WHEN statement given a condition and a value that's evaluated
// if the condition is true.
func When(cond Builder, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString("WHEN (")
		if err := cond.Build(d, buf); err != nil {
			return err
		}
		buf.WriteString(") THEN ")

		if builder, ok := value.(Builder); ok {
			return builder.Build(d, buf)
		}

		buf.WriteString(placeholder)
		buf.WriteValue(value)
		return nil
	})
}

// Else creates an ELSE statement given a value or a condition.
func Else(value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString("ELSE ")
		if builder, ok := value.(Builder); ok {
			return builder.Build(d, buf)
		}
		buf.WriteString(placeholder)
		buf.WriteValue(value)
		return nil
	})
}

// CaseBuilder interface that includes AS for aliasing CASE statements.
type CaseBuilder interface {
	Builder
	As(name string) Builder
}

// CaseBuildFunc implements Builder.
type CaseBuildFunc func(Dialect, Buffer) error

// Case creates a CASE statement from a list of conditions.
// If there are more than 1 conditions, the last one will be an else statement.
func Case(conds ...Builder) CaseBuilder {
	return CaseBuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString("CASE ")
		l := len(conds)
		for i, cond := range conds {
			if err := cond.Build(d, buf); err != nil {
				return err
			}
			if i < l-1 {
				buf.WriteString(" ")
			}
		}
		buf.WriteString(" END")
		return nil
	})
}

// As adds an alias to the CASE statement.
func (cb CaseBuildFunc) As(alias string) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		if err := cb(d, buf); err != nil {
			return err
		}
		buf.WriteString(" AS ")
		buf.WriteString(d.QuoteIdent(alias))
		return nil
	})
}

// Build calls itself to build SQL.
func (cb CaseBuildFunc) Build(d Dialect, buf Buffer) error {
	return cb(d, buf)
}
