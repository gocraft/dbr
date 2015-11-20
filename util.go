package dbr

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"unicode"
)

func camelCaseToSnakeCase(name string) string {
	buf := new(bytes.Buffer)

	runes := []rune(name)

	for i := 0; i < len(runes); i++ {
		buf.WriteRune(unicode.ToLower(runes[i]))
		if !unicode.IsUpper(runes[i]) && i != len(runes)-1 && unicode.IsUpper(runes[i+1]) {
			buf.WriteRune('_')
		}
	}

	return buf.String()
}

func columnName(field reflect.StructField) string {
	tag := field.Tag.Get("db")
	if tag == "-" {
		// ignore
		return ""
	}
	if tag != "" {
		return tag
	}
	// no tag, but we can record the field name
	return camelCaseToSnakeCase(field.Name)
}

func structMap(value reflect.Value) map[string]reflect.Value {
	m := make(map[string]reflect.Value)
	structValue(m, value)
	return m
}

var (
	typeValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

func structValue(m map[string]reflect.Value, value reflect.Value) {
	if value.Type().Implements(typeValuer) {
		return
	}
	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return
		}
		structValue(m, value.Elem())
	case reflect.Struct:
		t := value.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				// unexported
				continue
			}
			col := columnName(field)
			if col == "" {
				continue
			}
			fieldValue := value.Field(i)
			if _, ok := m[col]; !ok {
				m[col] = fieldValue
			}
			structValue(m, fieldValue)
		}
	}
}
