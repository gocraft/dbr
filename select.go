package dbr

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type SelectBuilder struct {
	*Session

	RawFullSql   string
	RawArguments []interface{}

	IsDistinct      bool
	Columns         []string
	FromTable       string
	WhereFragments  []whereFragment
	GroupBys        []string
	HavingFragments []whereFragment
	OrderBys        []string
	LimitCount      uint64
	LimitValid      bool
	OffsetCount     uint64
	OffsetValid     bool
}

type whereFragment struct {
	Condition   string
	Values      []interface{}
	EqualityMap map[string]interface{}
}

func (sess *Session) Select(cols ...string) *SelectBuilder {
	return &SelectBuilder{
		Session: sess,
		Columns: cols,
	}
}

func (sess *Session) SelectBySql(sql string, args ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		Session:      sess,
		RawFullSql:   sql,
		RawArguments: args,
	}
}

func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.IsDistinct = true
	return b
}

func (b *SelectBuilder) From(from string) *SelectBuilder {
	b.FromTable = from
	return b
}

func (b *SelectBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.WhereFragments = append(b.WhereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

func (b *SelectBuilder) GroupBy(group string) *SelectBuilder {
	b.GroupBys = append(b.GroupBys, group)
	return b
}

func (b *SelectBuilder) Having(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.HavingFragments = append(b.HavingFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

func (b *SelectBuilder) OrderBy(ord string) *SelectBuilder {
	b.OrderBys = append(b.OrderBys, ord)
	return b
}

func (b *SelectBuilder) OrderDir(ord string, isAsc bool) *SelectBuilder {
	if isAsc {
		b.OrderBys = append(b.OrderBys, ord+" ASC")
	} else {
		b.OrderBys = append(b.OrderBys, ord+" DESC")
	}
	return b
}

func (b *SelectBuilder) Limit(limit uint64) *SelectBuilder {
	b.LimitCount = limit
	b.LimitValid = true
	return b
}

func (b *SelectBuilder) Offset(offset uint64) *SelectBuilder {
	b.OffsetCount = offset
	b.OffsetValid = true
	return b
}

// Assumes page/perPage are valid. Page and perPage must be >= 1
func (b *SelectBuilder) Paginate(page, perPage uint64) *SelectBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

func (b *SelectBuilder) ToSql() (string, []interface{}) {
	if b.RawFullSql != "" {
		return b.RawFullSql, b.RawArguments
	}

	if len(b.Columns) == 0 {
		panic("no columns specified")
	}
	if len(b.FromTable) == 0 {
		panic("no table specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("SELECT ")

	if b.IsDistinct {
		sql.WriteString("DISTINCT ")
	}

	sql.WriteString(strings.Join(b.Columns, ", "))

	sql.WriteString(" FROM ")
	sql.WriteString(b.FromTable)

	if len(b.WhereFragments) > 0 {
		sql.WriteString(" WHERE ")
		var whereSql string
		var whereArgs []interface{}
		whereSql, whereArgs = whereFragmentsToSql(b.WhereFragments)
		sql.WriteString(whereSql)
		args = append(args, whereArgs...)
	}

	if len(b.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(b.GroupBys, ", "))
	}

	if len(b.HavingFragments) > 0 {
		sql.WriteString(" HAVING ")
		var havingSql string
		var havingArgs []interface{}
		havingSql, havingArgs = whereFragmentsToSql(b.HavingFragments)
		sql.WriteString(havingSql)
		args = append(args, havingArgs...)
	}

	if len(b.OrderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		sql.WriteString(strings.Join(b.OrderBys, ", "))
	}

	if b.LimitValid {
		sql.WriteString(" LIMIT ")
		fmt.Fprint(&sql, b.LimitCount)
	}

	if b.OffsetValid {
		sql.WriteString(" OFFSET ")
		fmt.Fprint(&sql, b.OffsetCount)
	}
	return sql.String(), args
}

//
// Where helpers:
//

func newWhereFragment(whereSqlOrMap interface{}, args []interface{}) whereFragment {
	switch pred := whereSqlOrMap.(type) {
	case string:
		return whereFragment{Condition: pred, Values: args}
	case map[string]interface{}:
		return whereFragment{EqualityMap: pred}
	case Eq:
		return whereFragment{EqualityMap: map[string]interface{}(pred)}
	default:
		panic("Invalid argument passed to Where. Pass a string or an Eq map.")
	}

	return whereFragment{}
}

// Invariant: only aclled when len(fragments) > 0
func whereFragmentsToSql(fragments []whereFragment) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	for _, f := range fragments {
		if f.Condition != "" {
			conditions = append(conditions, f.Condition)
			if len(f.Values) > 0 {
				args = append(args, f.Values...)
			}
		} else if f.EqualityMap != nil {
			eqSql, eqArgs := equalityMapToSql(f.EqualityMap)
			if eqSql != "" {
				conditions = append(conditions, eqSql)
				args = append(args, eqArgs...)
			}
		} else {
			panic("invalid equality map")
		}
	}
	return "(" + strings.Join(conditions, ") AND (") + ")", args
}

func equalityMapToSql(eq map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	for k, v := range eq {
		if v == nil {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", k))
		} else {
			vVal := reflect.ValueOf(v)
			if vVal.Kind() == reflect.Array || vVal.Kind() == reflect.Slice {
				vValLen := vVal.Len()
				if vValLen == 0 {
					if vVal.IsNil() {
						conditions = append(conditions, fmt.Sprintf("%s IS NULL", k))
					} else {
						conditions = append(conditions, "1=0")
					}
				} else if vValLen == 1 {
					conditions = append(conditions, fmt.Sprintf("%s = ?", k))
					args = append(args, vVal.Index(0).Interface())
				} else {
					conditions = append(conditions, fmt.Sprintf("%s IN ?", k))
					args = append(args, v)
				}
			} else {
				conditions = append(conditions, fmt.Sprintf("%s = ?", k))
				args = append(args, v)
			}
		}
	}

	if len(conditions) == 0 {
		return "", nil
	} else if len(conditions) == 1 {
		return conditions[0], args
	} else {
		return "(" + strings.Join(conditions, ") AND (") + ")", args
	}
}
