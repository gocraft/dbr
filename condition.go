package dbr

import (
	"reflect"
)

func buildCond(d Dialect, buf Buffer, pred string, cond ...Builder) error {
	var err error
	for i, c := range cond {
		if i > 0 {
			err = buf.WriteString(" ")
			if err != nil {
				return err
			}
			err = buf.WriteString(pred)
			if err != nil {
				return err
			}
			err = buf.WriteString(" ")
			if err != nil {
				return err
			}
		}
		err = buf.WriteString("(")
		if err != nil {
			return err
		}
		err = c.Build(d, buf)
		if err != nil {
			return err
		}
		err = buf.WriteString(")")
		if err != nil {
			return err
		}
	}
	return nil
}

// And creates AND from a list of conditions.
func And(cond ...Builder) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCond(d, buf, "AND", cond...)
	})
}

// Or creates OR from a list of conditions.
func Or(cond ...Builder) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCond(d, buf, "OR", cond...)
	})
}

func buildCmp(d Dialect, buf Buffer, pred string, column string, value interface{}) error {
	err := buf.WriteString(d.QuoteIdent(column))
	if err != nil {
		return err
	}
	err = buf.WriteString(" ")
	if err != nil {
		return err
	}
	err = buf.WriteString(pred)
	if err != nil {
		return err
	}
	err = buf.WriteString(" ")
	if err != nil {
		return err
	}
	err = buf.WriteString(placeholder)
	if err != nil {
		return err
	}

	err = buf.WriteValue(value)
	if err != nil {
		return err
	}
	return nil
}

// Eq is `=`.
// When value is nil, it will be translated to `IS NULL`.
// When value is a slice, it will be translated to `IN`.
// Otherwise it will be translated to `=`.
func Eq(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		var err error
		if value == nil {
			err = buf.WriteString(d.QuoteIdent(column))
			if err != nil {
				return err
			}
			err = buf.WriteString(" IS NULL")
			if err != nil {
				return err
			}
			return nil
		}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			if v.Len() == 0 {
				err = buf.WriteString(d.EncodeBool(false))
				if err != nil {
					return err
				}
				return nil
			}
			return buildCmp(d, buf, "IN", column, value)
		}
		return buildCmp(d, buf, "=", column, value)
	})
}

// Neq is `!=`.
// When value is nil, it will be translated to `IS NOT NULL`.
// When value is a slice, it will be translated to `NOT IN`.
// Otherwise it will be translated to `!=`.
func Neq(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		var err error

		if value == nil {
			err = buf.WriteString(d.QuoteIdent(column))
			if err != nil {
				return err
			}
			err = buf.WriteString(" IS NOT NULL")
			if err != nil {
				return err
			}
			return nil
		}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			if v.Len() == 0 {
				err = buf.WriteString(d.EncodeBool(true))
				if err != nil {
					return err
				}
				return nil
			}
			return buildCmp(d, buf, "NOT IN", column, value)
		}
		return buildCmp(d, buf, "!=", column, value)
	})
}

// Gt is `>`.
func Gt(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, ">", column, value)
	})
}

// Gte is '>='.
func Gte(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, ">=", column, value)
	})
}

// Lt is '<'.
func Lt(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, "<", column, value)
	})
}

// Lte is `<=`.
func Lte(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, "<=", column, value)
	})
}

func buildLike(d Dialect, buf Buffer, column, pattern string, isNot bool, escape []string) error {
	buf.WriteString(d.QuoteIdent(column))
	if isNot {
		buf.WriteString(" NOT LIKE ")
	} else {
		buf.WriteString(" LIKE ")
	}
	buf.WriteString(d.EncodeString(pattern))
	if len(escape) > 0 {
		buf.WriteString(" ESCAPE ")
		buf.WriteString(d.EncodeString(escape[0]))
	}
	return nil
}

// Like is `LIKE`, with an optional `ESCAPE` clause
func Like(column, value string, escape ...string) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildLike(d, buf, column, value, false, escape)
	})
}

// NotLike is `NOT LIKE`, with an optional `ESCAPE` clause
func NotLike(column, value string, escape ...string) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildLike(d, buf, column, value, true, escape)
	})
}
