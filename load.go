package dbr

import (
	"database/sql"
	"reflect"
)

// Load loads any value from sql.Rows
func Load(rows *sql.Rows, value interface{}) (int, error) {
	defer rows.Close()

	column, err := rows.Columns()
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
	var elemType reflect.Type
	if isSlice {
		elemType = v.Type().Elem()
	} else {
		elemType = v.Type()
	}
	extractor, err := findExtractor(elemType)
	if err != nil {
		return count, err
	}
	for rows.Next() {
		var elem reflect.Value
		if isSlice {
			elem = reflect.New(v.Type().Elem()).Elem()
		} else {
			elem = v
		}
		ptr := extractor(column, elem)
		err = rows.Scan(ptr...)
		if err != nil {
			return count, err
		}
		count++
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

type pointersExtractor func(columns []string, value reflect.Value) []interface{}

var (
	dummyDest   sql.Scanner = dummyScanner{}
	typeScanner             = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

func getStructFieldsExtractor(t reflect.Type) pointersExtractor {
	mapping := structMap(t)
	return func(columns []string, value reflect.Value) []interface{} {
		var ptr []interface{}
		for _, key := range columns {
			if index, ok := mapping[key]; ok {
				ptr = append(ptr, value.FieldByIndex(index).Addr().Interface())
			} else {
				ptr = append(ptr, dummyDest)
			}
		}
		return ptr
	}
}

func getIndirectExtractor(extractor pointersExtractor) pointersExtractor {
	return func(columns []string, value reflect.Value) []interface{} {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return extractor(columns, value.Elem())
	}
}

func dummyExtractor(columns []string, value reflect.Value) []interface{} {
	return []interface{}{value.Addr().Interface()}
}

func findExtractor(t reflect.Type) (pointersExtractor, error) {
	if reflect.PtrTo(t).Implements(typeScanner) {
		return dummyExtractor, nil
	}

	switch t.Kind() {
	case reflect.Ptr:
		if inner, err := findExtractor(t.Elem()); err != nil {
			return nil, err
		} else {
			return getIndirectExtractor(inner), nil
		}
	case reflect.Struct:
		return getStructFieldsExtractor(t), nil
	}
	return dummyExtractor, nil
}
