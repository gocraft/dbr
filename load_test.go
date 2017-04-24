package dbr

import (
	"database/sql/driver"
	"reflect"
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestLoad(t *testing.T) {
	type testStruct struct {
		A string
	}

	testcases := []struct {
		columns  []string
		expected interface{}
	}{
		{[]string{"a"}, "a"},
		{[]string{"a"}, []string{"a"}},
		{[]string{"a"}, testStruct{"a"}},
		{[]string{"a"}, &testStruct{"a"}},
		{[]string{"a"}, []testStruct{{"a"}}},
		{[]string{"a"}, []*testStruct{{"a"}}},
		{[]string{"a", "b"}, map[string]interface{}{"a": "a", "b": "b"}},
		{[]string{"a", "b"}, &map[string]interface{}{"a": "a", "b": "b"}},
		{[]string{"a", "b"}, []map[string]interface{}{{"a": "a", "b": "b"}}},
	}

	for _, tc := range testcases {
		session, dbmock := newSessionMock()
		values := []driver.Value{}
		for _, c := range tc.columns {
			values = append(values, c)
		}
		rows := sqlmock.NewRows(tc.columns).AddRow(values...)
		dbmock.ExpectQuery("SELECT .+").WillReturnRows(rows)
		v := reflect.New(reflect.TypeOf(tc.expected)).Elem().Addr().Interface()
		session.Select(tc.columns...).From("table").Load(v)
		assert.Equal(t, tc.expected, reflect.Indirect(reflect.ValueOf(v)).Interface())
	}
}

func BenchmarkLoad(b *testing.B) {
	session, dbmock := newSessionMock()
	rows := sqlmock.NewRows([]string{"a", "b", "c"})
	for i := 0; i < 100; i++ {
		rows = rows.AddRow(1, 2, 3)
	}
	dbmock.ExpectQuery("SELECT a, b, c FROM table").WillReturnRows(rows)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := make([]struct {
			A int `db:"a"`
			B int `db:"b"`
			C int
			D int `db:"-"`
			e int
			F int `db:"f"`
			G int `db:"g"`
			H int `db:"h"`
			i int
			j int
		}, 0, 100)
		session.Select("a", "b", "c").From("table").LoadStructs(&r)
	}
}

func newSessionMock() (SessionRunner, sqlmock.Sqlmock) {
	db, m, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	conn := Connection{DB: db, Dialect: dialect.MySQL, EventReceiver: nullReceiver}
	return conn.NewSession(nil), m
}
