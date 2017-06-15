package dbr

import (
	"database/sql"
	"reflect"
)

// Load loads any value from sql.Rows
func Load(rows *sql.Rows, value interface{}) (int, error) {
	defer rows.Close()

	nilablesIndex = make(map[int]bool)
	columns, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, ErrInvalidPointer
	}
	v = v.Elem()
	isSlice := v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8
	count := 0
	for rows.Next() {
		var elem reflect.Value
		if isSlice {
			elem = reflect.New(v.Type().Elem()).Elem()
		} else {
			elem = v
		}
		ptr, err := findPtr(columns, elem)
		if err != nil {
			return 0, err
		}
		err = rows.Scan(ptr...)
		if err != nil {
			return 0, err
		}
		count++
		for n, field := range ptr {
			if _, nilable := nilablesIndex[n]; nilable {
				var valid bool
				var valueType = reflect.Indirect(reflect.ValueOf(field)).Field(0).Type().String()
				if valueType == "time.Time" {
					valid = reflect.Indirect(reflect.ValueOf(field)).Field(1).Interface().(bool)
				} else {
					valid = reflect.Indirect(reflect.ValueOf(field)).Field(0).Field(1).Interface().(bool)
				}
				if valid {
					switch valueType {
					case "sql.NullInt64":
						var __int64 int64 = 0
						reflect.ValueOf(&__int64).Elem().Set(reflect.Indirect(reflect.ValueOf(field)).Field(0).Field(0))
						switch sourceType := elem.Field(n).Type().String(); sourceType {
						case "int":
							elem.Field(n).Set(reflect.ValueOf(int(__int64)))
						case "int8":
							elem.Field(n).Set(reflect.ValueOf(int8(__int64)))
						case "int16":
							elem.Field(n).Set(reflect.ValueOf(int16(__int64)))
						case "int32":
							elem.Field(n).Set(reflect.ValueOf(int32(__int64)))
						case "uint":
							elem.Field(n).Set(reflect.ValueOf(uint(__int64)))
						case "uint8":
							elem.Field(n).Set(reflect.ValueOf(uint8(__int64)))
						case "uint16":
							elem.Field(n).Set(reflect.ValueOf(uint16(__int64)))
						case "uint32":
							elem.Field(n).Set(reflect.ValueOf(uint32(__int64)))
						case "uint64":
							elem.Field(n).Set(reflect.ValueOf(uint64(__int64)))
						}
					case "sql.NullFloat64":
						var __float64 float64 = 0
						reflect.ValueOf(&__float64).Elem().Set(reflect.Indirect(reflect.ValueOf(field)).Field(0).Field(0))
						if elem.Field(n).Type().String() == "float32" {
							elem.Field(n).Set(reflect.ValueOf(float32(__float64)))
						}
					default:
						if valueType == "time.Time" {
							elem.Field(n).Set(reflect.Indirect(reflect.ValueOf(field)).Field(0))
						} else {
							elem.Field(n).Set(reflect.Indirect(reflect.ValueOf(field)).Field(0).Field(0))
						}
					}
				}
			}
		}
		if isSlice {
			v.Set(reflect.Append(v, elem))
		} else {
			break
		}
	}
	return count, nil
}

type dummyScanner struct{}

func (dummyScanner) Scan(interface{}) error {
	return nil
}

var (
	dummyDest   sql.Scanner = dummyScanner{}
	typeScanner             = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

func findPtr(columns []string, value reflect.Value) ([]interface{}, error) {
	if value.Addr().Type().Implements(typeScanner) {
		return []interface{}{value.Addr().Interface()}, nil
	}
	switch value.Kind() {
	case reflect.Struct:
		var ptr []interface{}
		m := structMap(value)
		for _, key := range columns {
			if val, ok := m[key]; ok {
				ptr = append(ptr, val.Addr().Interface())
			} else {
				ptr = append(ptr, dummyDest)
			}
		}
		return ptr, nil
	case reflect.Ptr:
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return findPtr(columns, value.Elem())
	}
	return []interface{}{value.Addr().Interface()}, nil
}
