package dbr

import (
	// "fmt"
	"bytes"
	"database/sql/driver"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Need to turn \x00, \n, \r, \, ', " and \x1a
// Returns an escaped, quoted string. eg, "hello 'world'" -> "'hello \'world\''"
func escapeAndQuoteString(val string) string {
	buf := bytes.Buffer{}

	buf.WriteRune('\'')

	for _, char := range val {
		if char == '\'' { // single quote: ' -> \'
			buf.WriteString("\\'")
		} else if char == '"' { // double quote: " -> \"
			buf.WriteString("\\\"")
		} else if char == '\\' { // slash: \ -> "\\"
			buf.WriteString("\\\\")
		} else if char == '\n' { // control: newline: \n -> "\n"
			buf.WriteString("\\n")
		} else if char == '\r' { // control: return: \r -> "\r"
			buf.WriteString("\\r")
		} else if char == 0 { // control: NUL: 0 -> "\x00"
			buf.WriteString("\\x00")
		} else if char == 0x1a { // control: \x1a -> "\x1a"
			buf.WriteString("\\x1a")
		} else {
			buf.WriteRune(char)
		}
	}

	buf.WriteRune('\'')

	return buf.String()
}

func isUint(k reflect.Kind) bool {
	return (k == reflect.Uint) ||
		(k == reflect.Uint8) ||
		(k == reflect.Uint16) ||
		(k == reflect.Uint32) ||
		(k == reflect.Uint64)
}

func isInt(k reflect.Kind) bool {
	return (k == reflect.Int) ||
		(k == reflect.Int8) ||
		(k == reflect.Int16) ||
		(k == reflect.Int32) ||
		(k == reflect.Int64)
}

func isFloat(k reflect.Kind) bool {
	return (k == reflect.Float32) ||
		(k == reflect.Float64)
}

// sql is like "id = ? OR username = ?"
// vals is like []interface{}{4, "bob"}
// NOTE that vals can only have values of certain types:
//   - Integers (signed and unsigned)
//   - floats
//   - strings (that are valid utf-8)
//   - booleans
//   - dates (TODO)
func Interpolate(sql string, vals []interface{}) (string, error) {
	if sql == ""  && len(vals) == 0 {
		return "", nil
	}

	curVal := 0
	maxVals := len(vals)

	buf := bytes.Buffer{}

	for _, r := range sql {
		if r != '?' {
			buf.WriteRune(r)
		} else if r == '?' && curVal < maxVals {
			v := vals[curVal]

			valuer, ok := v.(driver.Valuer)
			if ok {
				val, err := valuer.Value()
				if err != nil {
					return "", err
				} else {
					v = val
				}
			}

			valueOfV := reflect.ValueOf(v)
			kindOfV := valueOfV.Kind()

			if v == nil {
				buf.WriteString("NULL")
			} else if isInt(kindOfV) {
				var ival int64 = valueOfV.Int()

				buf.WriteString(strconv.FormatInt(ival, 10))
			} else if isUint(kindOfV) {
				var uival uint64 = valueOfV.Uint()

				buf.WriteString(strconv.FormatUint(uival, 10))
			} else if kindOfV == reflect.String {
				var str string = valueOfV.String()

				if !utf8.ValidString(str) {
					return "", ErrNotUTF8
				}

				buf.WriteString(escapeAndQuoteString(str))
			} else if isFloat(kindOfV) {
				var fval float64 = valueOfV.Float()

				buf.WriteString(strconv.FormatFloat(fval, 'f', -1, 64))
			} else if kindOfV == reflect.Bool {
				var bval bool = valueOfV.Bool()

				if bval {
					buf.WriteRune('1')
				} else {
					buf.WriteRune('0')
				}
			} else if kindOfV == reflect.Slice {
				typeOfV := reflect.TypeOf(v)
				subtype := typeOfV.Elem()
				kindOfSubtype := subtype.Kind()

				sliceLen := valueOfV.Len()
				stringSlice := make([]string, 0, sliceLen)

				if sliceLen == 0 {
					return "", ErrInvalidSliceLength
				} else if isInt(kindOfSubtype) {
					for i := 0; i < sliceLen; i += 1 {
						var ival int64 = valueOfV.Index(i).Int()
						stringSlice = append(stringSlice, strconv.FormatInt(ival, 10))
					}
				} else if isUint(kindOfSubtype) {
					for i := 0; i < sliceLen; i += 1 {
						var uival uint64 = valueOfV.Index(i).Uint()
						stringSlice = append(stringSlice, strconv.FormatUint(uival, 10))
					}
				} else if kindOfSubtype == reflect.String {
					for i := 0; i < sliceLen; i += 1 {
						var str string = valueOfV.Index(i).String()
						if !utf8.ValidString(str) {
							return "", ErrNotUTF8
						}
						stringSlice = append(stringSlice, escapeAndQuoteString(str))
					}
				} else {
					return "", ErrInvalidSliceValue
				}
				buf.WriteRune('(')
				buf.WriteString(strings.Join(stringSlice, ","))
				buf.WriteRune(')')
			} else {
				return "", ErrInvalidValue
			}

			curVal += 1
		} else {
			return "", ErrArgumentMismatch
		}
	}
	
	if curVal != maxVals {
		return "", ErrArgumentMismatch
	}

	return buf.String(), nil
}
