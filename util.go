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
		if i != len(runes)-1 && unicode.IsUpper(runes[i+1]) &&
			(unicode.IsLower(runes[i]) || unicode.IsDigit(runes[i]) ||
				(i != len(runes)-2 && unicode.IsLower(runes[i+2]))) {
			buf.WriteRune('_')
		}
	}

	return buf.String()
}

// structMap builds index to fast lookup fields in struct
func structMap(t reflect.Type) map[string][]int {
	m := make(map[string][]int)
	structTraverse(m, t, nil)
	return m
}

var (
	typeValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

func structTraverse(m map[string][]int, t reflect.Type, head []int) {
	if t.Implements(typeValuer) {
		return
	}
	switch t.Kind() {
	case reflect.Ptr:
		structTraverse(m, t.Elem(), head)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				// unexported
				continue
			}
			tag := field.Tag.Get("db")
			if tag == "-" {
				// ignore
				continue
			}
			if tag == "" {
				// no tag, but we can record the field name
				tag = camelCaseToSnakeCase(field.Name)
			}
			if _, ok := m[tag]; !ok {
				m[tag] = append(head, i)
			}
			structTraverse(m, field.Type, append(head, i))
		}
	}
}
