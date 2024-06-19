package dbr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

var (
	filledRecord = nullTypedRecord{
		StringVal:  NewNullString("wow"),
		Int64Val:   NewNullInt64(1483272000),
		Float64Val: NewNullFloat64(1.618),
		TimeVal:    NewNullTime(time.Date(2009, 1, 3, 18, 15, 5, 0, time.UTC)),
		BoolVal:    NewNullBool(true),
	}
)

func TestNullTypesScanning(t *testing.T) {
	for _, test := range []struct {
		in nullTypedRecord
	}{
		{},
		{
			in: filledRecord,
		},
	} {
		for _, sess := range testSession {
			reset(t, sess)

			tx, err := sess.Begin()
			require.NoError(t, err)

			if sess.Dialect == dialect.MSSQL {
				tx.UpdateBySql("SET IDENTITY_INSERT null_types ON;").Exec()
			}

			test.in.Id = 1
			_, err = tx.InsertInto("null_types").Columns("id", "string_val", "int64_val", "float64_val", "time_val", "bool_val").Record(test.in).Exec()
			require.NoError(t, err)

			err = tx.Commit()
			require.NoError(t, err)

			var record nullTypedRecord
			err = sess.Select("*").From("null_types").Where(Eq("id", test.in.Id)).LoadOne(&record)
			require.NoError(t, err)
			if sess.Dialect == dialect.PostgreSQL {
				// TODO: https://github.com/lib/pq/issues/329
				if !record.TimeVal.Time.IsZero() {
					record.TimeVal.Time = record.TimeVal.Time.UTC()
				}
			}
			require.Equal(t, test.in, record)
		}
	}
}

func TestNullInt64Unmarshal(t *testing.T) {
	var test struct {
		Num NullInt64
	}
	err := json.Unmarshal([]byte(`{"num":null}`), &test)
	require.NoError(t, err)
	require.Equal(t, int64(0), test.Num.Int64)
	require.False(t, test.Num.Valid)
}

func TestNullTypesActuallyNullJSON(t *testing.T) {
	var out struct {
		Bool   NullBool    `json:"b"`
		Float  NullFloat64 `json:"f"`
		String NullString  `json:"s"`
		Time   NullTime    `json:"t"`
		Int    NullInt64   `json:"i"`
	}
	jsonBs := []byte(`{"b":null,"f":null,"s":null,"t":null,"i":null}`)
	err := json.Unmarshal(jsonBs, &out)
	require.NoError(t, err)
	require.False(t, out.Bool.Valid)
	require.False(t, out.Float.Valid)
	require.False(t, out.String.Valid)
	require.False(t, out.Time.Valid)
	require.False(t, out.Int.Valid)
}

func TestNullTypesJSON(t *testing.T) {
	for _, test := range []struct {
		in   interface{}
		in2  interface{}
		out  interface{}
		want string
	}{
		{
			in:   &filledRecord.BoolVal,
			in2:  filledRecord.BoolVal,
			out:  new(NullBool),
			want: "true",
		},
		{
			in:   &filledRecord.Float64Val,
			in2:  filledRecord.Float64Val,
			out:  new(NullFloat64),
			want: "1.618",
		},
		{
			in:   &filledRecord.Int64Val,
			in2:  filledRecord.Int64Val,
			out:  new(NullInt64),
			want: "1483272000",
		},
		{
			in:   &filledRecord.StringVal,
			in2:  filledRecord.StringVal,
			out:  new(NullString),
			want: `"wow"`,
		},
		{
			in:   &filledRecord.TimeVal,
			in2:  filledRecord.TimeVal,
			out:  new(NullTime),
			want: `"2009-01-03T18:15:05Z"`,
		},
	} {
		// marshal ptr
		b, err := json.Marshal(test.in)
		require.NoError(t, err)
		require.Equal(t, test.want, string(b))

		// marshal value
		b, err = json.Marshal(test.in2)
		require.NoError(t, err)
		require.Equal(t, test.want, string(b))

		// unmarshal
		err = json.Unmarshal(b, test.out)
		require.NoError(t, err)
		require.Equal(t, test.in, test.out)
	}
}
