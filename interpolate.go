package dbr

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocraft/dbr/dialect"
)

// Don't break the API
// FIXME: This will be removed in the future
func Interpolate(query string, value []interface{}) (string, error) {
	return InterpolateForDialect(query, value, dialect.MySQL)
}

// InterpolateForDialect replaces placeholder in query with corresponding value in dialect
func InterpolateForDialect(query string, value []interface{}, d Dialect) (string, error) {
	buf := new(bytes.Buffer)
	err := interpolate(query, value, d, buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

const (
	placeholder = "?"
)

func convertDollarPlaceholder(query string, n int, w StringWriter) {
	index := strings.Index(query, placeholder)
	if index == -1 {
		w.WriteString(query)
		return
	}
	w.WriteString(query[:index])
	w.WriteString(fmt.Sprintf("$%d", n))
	convertDollarPlaceholder(query[index+len(placeholder):], n+1, w)
}

var (
	dollarPlaceholderRegexp = regexp.MustCompile(`\$[0-9]+`)
)

func interpolate(query string, value []interface{}, d Dialect, w StringWriter) error {
	used := make([]bool, len(value))

	buf := new(bytes.Buffer)
	convertDollarPlaceholder(query, 1, buf)
	query = buf.String()

	for {
		index := dollarPlaceholderRegexp.FindStringIndex(query)
		if index == nil {
			break
		}
		w.WriteString(query[:index[0]])

		// after $
		n, _ := strconv.Atoi(query[index[0]+1 : index[1]])
		if n < 1 || n > len(value) {
			return ErrPlaceholderCount
		}
		err := encodePlaceholder(value[n-1], d, w)
		if err != nil {
			return err
		}
		used[n-1] = true
		query = query[index[1]:]
	}

	// placeholder not found; write remaining query
	w.WriteString(query)

	for _, u := range used {
		if !u {
			// unused value
			return ErrPlaceholderCount
		}
	}

	return nil
}

func encodePlaceholder(value interface{}, d Dialect, w StringWriter) error {
	if builder, ok := value.(Builder); ok {
		buf := NewBuffer()
		err := builder.Build(d, buf)
		if err != nil {
			return err
		}
		paren := true
		switch value.(type) {
		case *SelectStmt:
		case *union:
		default:
			paren = false
		}
		if paren {
			w.WriteString("(")
		}
		err = interpolate(buf.String(), buf.Value(), d, w)
		if err != nil {
			return err
		}
		if paren {
			w.WriteString(")")
		}
		return nil
	}

	if valuer, ok := value.(driver.Valuer); ok {
		// get driver.Valuer's data
		var err error
		value, err = valuer.Value()
		if err != nil {
			return err
		}
	}

	if value == nil {
		w.WriteString("NULL")
		return nil
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		w.WriteString(d.EncodeString(v.String()))
		return nil
	case reflect.Bool:
		w.WriteString(d.EncodeBool(v.Bool()))
		return nil
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		w.WriteString(strconv.FormatInt(v.Int(), 10))
		return nil
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		w.WriteString(strconv.FormatUint(v.Uint(), 10))
		return nil
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		w.WriteString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
		return nil
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			w.WriteString(d.EncodeTime(v.Interface().(time.Time).UTC()))
			return nil
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			w.WriteString(d.EncodeBytes(v.Bytes()))
			return nil
		}
		if v.Len() == 0 {
			// FIXME: support zero-length slice
			return ErrInvalidSliceLength
		}
		w.WriteString("(")
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w.WriteString(",")
			}
			err := encodePlaceholder(v.Index(i).Interface(), d, w)
			if err != nil {
				return err
			}
		}
		w.WriteString(")")
		return nil
	case reflect.Ptr:
		return encodePlaceholder(v.Elem().Interface(), d, w)
	}
	return ErrNotSupported
}
