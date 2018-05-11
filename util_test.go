package dbr

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		require.Equal(t, test.want, camelCaseToSnakeCase(test.in))
	}
}

func BenchmarkCamelCaseToSnakeCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		camelCaseToSnakeCase("getHTTPResponseCode")
	}
}

func TestFindValueByName(t *testing.T) {
	for _, test := range []struct {
		in   interface{}
		name []string
		want []string
	}{
		{
			in: struct {
				CreatedAt time.Time
			}{},
			name: []string{"created_at"},
			want: []string{"created_at"},
		},
		{
			in: struct {
				intVal int
			}{},
			name: []string{"int_val"},
		},
		{
			in: struct {
				IntVal int `db:"test"`
			}{},
			name: []string{"test"},
			want: []string{"test"},
		},
		{
			in: struct {
				IntVal int `db:"-"`
			}{},
			name: []string{"int_val"},
		},
		{
			in: struct {
				Test1 struct {
					Test2 int
				}
			}{},
			name: []string{"test2"},
			want: []string{"test2"},
		},
	} {
		found := make([]interface{}, len(test.name))
		s := newTagStore()
		s.findValueByName(reflect.ValueOf(test.in), test.name, found, false)

		var got []string
		for i, v := range found {
			if v != nil {
				got = append(got, test.name[i])
			}
		}

		require.Equal(t, test.want, got)
	}
}
