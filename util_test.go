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
		{
			in:   "XMLName",
			want: "xml_name",
		},
	} {
		assert.Equal(t, test.want, camelCaseToSnakeCase(test.in))
	}
}

func TestStructMap(t *testing.T) {
	for _, test := range []struct {
		in       interface{}
		expected map[string][]int
	}{
		{
			in: struct {
				CreatedAt time.Time
			}{},
			expected: map[string][]int{"created_at": {0}},
		},
		{
			in: struct {
				intVal int
			}{},
			expected: map[string][]int{},
		},
		{
			in: struct {
				IntVal int `db:"test"`
			}{},
			expected: map[string][]int{"test": {0}},
		},
		{
			in: struct {
				IntVal int `db:"-"`
			}{},
			expected: map[string][]int{},
		},
		{
			in: struct {
				Test1 struct {
					Test2 int
				}
			}{},
			expected: map[string][]int{"test1": {0}, "test2": {0, 0}},
		},
	} {
		m := structMap(reflect.ValueOf(test.in).Type())
		assert.Equal(t, test.expected, m)
	}
}
