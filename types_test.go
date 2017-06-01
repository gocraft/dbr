package dbr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mailru/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

var (
	filledRecord = nullTypedRecord{
		StringVal:  NewNullString("wow"),
		Int64Val:   NewNullInt64(42),
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
			test.in.ID = nextID()
			_, err := sess.InsertInto("null_types").Columns("id", "string_val", "int64_val", "float64_val", "time_val", "bool_val").Record(test.in).Exec()
			assert.NoError(t, err)

			var record nullTypedRecord
			err = sess.Select("*").From("null_types").Where(Eq("id", test.in.ID)).LoadStruct(&record)
			assert.NoError(t, err)
			if sess.Dialect == dialect.PostgreSQL {
				// TODO: https://github.com/lib/pq/issues/329
				if !record.TimeVal.Time.IsZero() {
					record.TimeVal.Time = record.TimeVal.Time.UTC()
				}
			}
			assert.Equal(t, test.in, record)
		}
	}
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
			want: "42",
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
		assert.NoError(t, err)
		assert.Equal(t, test.want, string(b))

		// marshal value
		b, err = json.Marshal(test.in2)
		assert.NoError(t, err)
		assert.Equal(t, test.want, string(b))

		// unmarshal
		err = json.Unmarshal(b, test.out)
		assert.NoError(t, err)
		assert.Equal(t, test.in, test.out)
	}
}
