package dbr

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	filledRecord = nullTypedRecord{
		StringVal:  NullString{sql.NullString{String: "wow", Valid: true}},
		Int64Val:   NullInt64{sql.NullInt64{Int64: 42, Valid: true}},
		Float64Val: NullFloat64{sql.NullFloat64{Float64: 1.618, Valid: true}},
		TimeVal:    NullTime{Time: time.Date(2009, 1, 3, 18, 15, 5, 0, time.UTC), Valid: true},
		BoolVal:    NullBool{sql.NullBool{Bool: true, Valid: true}},
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
		for _, sess := range []*Session{mysqlSession, postgresSession} {
			test.in.Id = nextID()
			_, err := sess.InsertInto("null_types").Columns("id", "string_val", "int64_val", "float64_val", "time_val", "bool_val").Record(test.in).Exec()
			assert.NoError(t, err)

			var record nullTypedRecord
			err = sess.Select("*").From("null_types").Where(Eq("id", test.in.Id)).LoadStruct(&record)
			assert.NoError(t, err)
			// TODO: https://github.com/lib/pq/issues/329
			record.TimeVal.Time = test.in.TimeVal.Time
			assert.Equal(t, test.in, record)
		}
	}
}

func TestNullTypesJSON(t *testing.T) {
	for _, test := range []struct {
		in   nullTypedRecord
		want []byte
	}{
		{
			want: []byte(`{"Id":0,"StringVal":null,"Int64Val":null,"Float64Val":null,"TimeVal":null,"BoolVal":null}`),
		},
		{
			in:   filledRecord,
			want: []byte(`{"Id":0,"StringVal":"wow","Int64Val":42,"Float64Val":1.618,"TimeVal":"2009-01-03T18:15:05Z","BoolVal":true}`),
		},
	} {
		b, err := json.Marshal(&test.in)
		assert.NoError(t, err)
		assert.Equal(t, test.want, b)

		var record nullTypedRecord
		err = json.Unmarshal(test.want, &record)
		assert.NoError(t, err)
		assert.Equal(t, test.in, record)
	}
}
