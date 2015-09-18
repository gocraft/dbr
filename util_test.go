package dbr

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSnakeCase(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		{
			in:   "",
			want: "",
		},
		{
			in:   "IsDigit",
			want: "is_digit",
		},
		{
			in:   "Is",
			want: "is",
		},
		{
			in:   "IsID",
			want: "is_id",
		},
		{
			in:   "IsSQL",
			want: "is_sql",
		},
		{
			in:   "LongSQL",
			want: "long_sql",
		},
		{
			in:   "Float64Val",
			want: "float64_val",
		},
	} {
		assert.Equal(t, test.want, camelCaseToSnakeCase(test.in))
	}
}

func TestStructMap(t *testing.T) {
	for _, test := range []struct {
		in  interface{}
		ok  []string
		bad []string
	}{
		{
			in: struct {
				CreatedAt time.Time
			}{},
			ok: []string{"created_at"},
		},
		{
			in: struct {
				intVal int
			}{},
			bad: []string{"int_val"},
		},
		{
			in: struct {
				IntVal int `db:"test"`
			}{},
			ok:  []string{"test"},
			bad: []string{"int_val"},
		},
		{
			in: struct {
				IntVal int `db:"-"`
			}{},
			bad: []string{"int_val"},
		},
		{
			in: struct {
				Test1 struct {
					Test2 int
				}
			}{},
			ok: []string{"test2"},
		},
	} {
		m := structMap(reflect.ValueOf(test.in))
		for _, c := range test.ok {
			_, ok := m[c]
			assert.True(t, ok)
		}
		for _, c := range test.bad {
			_, ok := m[c]
			assert.False(t, ok)
		}
	}
}
