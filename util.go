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

func structMap(value reflect.Value) map[string]reflect.Value {
	m := make(map[string]reflect.Value)
	structValue(m, value)
	return m
}

var (
	typeValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

var nilablesIndex = make(map[int]bool)
var nilables = map[string]reflect.Type{
	"uint":      reflect.TypeOf(*new(NullInt64)),
	"uint8":     reflect.TypeOf(*new(NullInt64)),
	"uint16":    reflect.TypeOf(*new(NullInt64)),
	"uint32":    reflect.TypeOf(*new(NullInt64)),
	"uint64":    reflect.TypeOf(*new(NullInt64)),
	"int":       reflect.TypeOf(*new(NullInt64)),
	"int8":      reflect.TypeOf(*new(NullInt64)),
	"int16":     reflect.TypeOf(*new(NullInt64)),
	"int32":     reflect.TypeOf(*new(NullInt64)),
	"int64":     reflect.TypeOf(*new(NullInt64)),
	"float32":   reflect.TypeOf(*new(NullFloat64)),
	"float64":   reflect.TypeOf(*new(NullFloat64)),
	"string":    reflect.TypeOf(*new(NullString)),
	"time.Time": reflect.TypeOf(*new(NullTime)),
	"bool":      reflect.TypeOf(*new(NullBool)),
}

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
			if field.PkgPath != "" && !field.Anonymous {
				// unexported
				continue
			}
			tag := field.Tag.Get("db")
			if tag == "-" {
				// ignore
				continue
			}
			nilable := false
			if tag == "!" {
				nilable = true
				tag = ""
			}
			if tag == "" {
				// no tag, but we can record the field name
				tag = camelCaseToSnakeCase(field.Name)
			}
			fieldValue := value.Field(i)
			if _, ok := m[tag]; !ok {
				_, supportedNilable := nilables[field.Type.String()]
				if nilable && supportedNilable {
					m[tag] = reflect.New(nilables[field.Type.String()]).Elem()
					nilablesIndex[len(m)-1] = true
				} else {
					m[tag] = fieldValue
				}
			}
			structValue(m, fieldValue)
		}
	}
}
