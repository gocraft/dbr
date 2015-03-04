package dbr

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestNullTypeScanning(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Create and test scanning a completely NULL row
	nullRecordPrototype := &nullTypedRecord{}
	res, err := s.InsertInto("null_types").Columns("string_val", "int64_val", "float64_val", "time_val", "bool_val").Record(nullRecordPrototype).Exec()
	assert.NoError(t, err)
	nullID, err := res.LastInsertId()
	assert.NoError(t, err)

	nullTypeSet := &nullTypedRecord{}
	err = s.Select("*").From("null_types").Where("id = ?", nullID).LoadStruct(nullTypeSet)
	assert.NoError(t, err)
	assert.Equal(t, nullRecordPrototype, nullTypeSet)

	// Create and test scanning a completely NOT NULL row
	notNullRecordPrototype := newNullTypedRecordWithData()
	res, err = s.InsertInto("null_types").Columns("string_val", "int64_val", "float64_val", "time_val", "bool_val").Record(notNullRecordPrototype).Exec()
	assert.NoError(t, err)
	notNullID, err := res.LastInsertId()
	assert.NoError(t, err)

	notNullTypeSet := &nullTypedRecord{}
	err = s.Select("*").From("null_types").Where("id = ?", notNullID).LoadStruct(notNullTypeSet)
	assert.NoError(t, err)
	assert.Equal(t, notNullRecordPrototype, notNullTypeSet)

	notNullTypeSet.StringVal.String = "newString"
	assert.NotEqual(t, notNullRecordPrototype, notNullTypeSet)
}

func TestNullTypeJSONMarshal(t *testing.T) {
	type nullTypeJSONTest struct {
		record       *nullTypedRecord
		expectedJSON []byte
	}

	tests := []nullTypeJSONTest{
		nullTypeJSONTest{
			record:       &nullTypedRecord{},
			expectedJSON: []byte(`{"Id":0,"StringVal":null,"Int64Val":null,"Float64Val":null,"TimeVal":null,"BoolVal":null}`),
		},
		nullTypeJSONTest{
			record:       newNullTypedRecordWithData(),
			expectedJSON: []byte(`{"Id":0,"StringVal":"wow","Int64Val":42,"Float64Val":1.618,"TimeVal":"2009-01-03T18:15:05Z","BoolVal":true}`),
		},
	}

	for _, test := range tests {
		// Marshal the record
		rawJSON, err := json.Marshal(test.record)
		assert.NoError(t, err)
		assert.Equal(t, test.expectedJSON, rawJSON)

		// Umarshal it back
		newRecord := &nullTypedRecord{}
		err = json.Unmarshal([]byte(rawJSON), newRecord)
		assert.NoError(t, err)
		assert.Equal(t, test.record, newRecord)
	}
}

func newNullTypedRecordWithData() *nullTypedRecord {
	return &nullTypedRecord{
		StringVal:  NullString{sql.NullString{String: "wow", Valid: true}},
		Int64Val:   NullInt64{sql.NullInt64{Int64: 42, Valid: true}},
		Float64Val: NullFloat64{sql.NullFloat64{Float64: 1.618, Valid: true}},
		TimeVal:    NullTime{mysql.NullTime{Time: time.Date(2009, 1, 3, 18, 15, 5, 0, time.UTC), Valid: true}},
		BoolVal:    NullBool{sql.NullBool{Bool: true, Valid: true}},
	}
}
