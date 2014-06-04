package dbr

import (
	"bytes"
	"reflect"
)

type whereFragment struct {
	Condition   string
	Values      []interface{}
	EqualityMap map[string]interface{}
}

func newWhereFragment(whereSqlOrMap interface{}, args []interface{}) *whereFragment {
	switch pred := whereSqlOrMap.(type) {
	case string:
		return &whereFragment{Condition: pred, Values: args}
	case map[string]interface{}:
		return &whereFragment{EqualityMap: pred}
	case Eq:
		return &whereFragment{EqualityMap: map[string]interface{}(pred)}
	default:
		panic("Invalid argument passed to Where. Pass a string or an Eq map.")
	}

	return nil
}

// Invariant: only aclled when len(fragments) > 0
func writeWhereFragmentsToSql(fragments []*whereFragment, sql *bytes.Buffer, args *[]interface{}) {
	anyConditions := false
	for _, f := range fragments {
		if f.Condition != "" {
			if anyConditions {
				sql.WriteString(" AND (")
			} else {
				sql.WriteRune('(')
				anyConditions = true
			}
			sql.WriteString(f.Condition)
			sql.WriteRune(')')
			if len(f.Values) > 0 {
				*args = append(*args, f.Values...)
			}
		} else if f.EqualityMap != nil {
			anyConditions = writeEqualityMapToSql(f.EqualityMap, sql, args, anyConditions)
		} else {
			panic("invalid equality map")
		}
	}
}

func writeEqualityMapToSql(eq map[string]interface{}, sql *bytes.Buffer, args *[]interface{}, anyConditions bool) bool {
	for k, v := range eq {
		if v == nil {
			anyConditions = writeWhereCondition(sql, k, " IS NULL", anyConditions)
		} else {
			vVal := reflect.ValueOf(v)

			if vVal.Kind() == reflect.Array || vVal.Kind() == reflect.Slice {
				vValLen := vVal.Len()
				if vValLen == 0 {
					if vVal.IsNil() {
						anyConditions = writeWhereCondition(sql, k, " IS NULL", anyConditions)
					} else {
						if anyConditions {
							sql.WriteString(" AND (1=0)")
						} else {
							sql.WriteString("(1=0)")
						}
					}
				} else if vValLen == 1 {
					anyConditions = writeWhereCondition(sql, k, " = ?", anyConditions)
					*args = append(*args, vVal.Index(0).Interface())
				} else {
					anyConditions = writeWhereCondition(sql, k, " IN ?", anyConditions)
					*args = append(*args, v)
				}
			} else {
				anyConditions = writeWhereCondition(sql, k, " = ?", anyConditions)
				*args = append(*args, v)
			}
		}
	}

	return anyConditions
}

func writeWhereCondition(sql *bytes.Buffer, k string, pred string, anyConditions bool) bool {
	if anyConditions {
		sql.WriteString(" AND (")
	} else {
		sql.WriteRune('(')
		anyConditions = true
	}
	sql.WriteString(k)
	Quoter.writeQuotedColumn(k, sql)
	sql.WriteString(pred)
	sql.WriteRune(')')

	return anyConditions
}
