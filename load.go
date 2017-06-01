package dbr

import (
	"database/sql"
	"fmt"
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

type keyValueMap map[string]interface{}

type kvScanner struct {
	column string
	m      keyValueMap
}

func (kv *kvScanner) Scan(v interface{}) error {
	kv.m[kv.column] = v
	return nil
}

type pointersExtractor func(columns []string, value reflect.Value) []interface{}

var (
	dummyDest       sql.Scanner = dummyScanner{}
	typeScanner                 = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	typeKeyValueMap             = reflect.TypeOf(keyValueMap(nil))
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

func mapExtractor(columns []string, value reflect.Value) []interface{} {
	if value.IsNil() {
		value.Set(reflect.MakeMap(value.Type()))
	}
	m := value.Convert(typeKeyValueMap).Interface().(keyValueMap)
	var ptr []interface{}
	for _, c := range columns {
		ptr = append(ptr, &kvScanner{column: c, m: m})
	}
	return ptr
}

func dummyExtractor(columns []string, value reflect.Value) []interface{} {
	return []interface{}{value.Addr().Interface()}
}

func findExtractor(t reflect.Type) (pointersExtractor, error) {
	if reflect.PtrTo(t).Implements(typeScanner) {
		return dummyExtractor, nil
	}

	switch t.Kind() {
	case reflect.Map:
		if !t.ConvertibleTo(typeKeyValueMap) {
			return nil, fmt.Errorf("expected %v, got %v", typeKeyValueMap, t)
		}
		return mapExtractor, nil
	case reflect.Ptr:
		inner, err := findExtractor(t.Elem())
		if err != nil {
			return nil, err
		}
		return getIndirectExtractor(inner), nil
	case reflect.Struct:
		return getStructFieldsExtractor(t), nil
	}
	return dummyExtractor, nil
}
