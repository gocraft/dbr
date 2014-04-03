package dbr

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

type Session struct {
	cxn *Connection
	EventReceiver
}

func (cxn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = cxn.EventReceiver // Use parent instrumentation
	}
	return &Session{cxn: cxn, EventReceiver: log}
}



// Given a query and given a structure (field list), there's 2 sets of fields.
// Take the intersection. We can fill those in. great.
// For fields in the structure that aren't in the query, we'll let that slide if db:"-"
// For fields in the structure that aren't in the query but without db:"-", return error
// For fields in the query that aren't in the structure, we'll ignore them.

// dest can be:
// - addr of a structure
// - addr of slice of pointers to structures
// - map of pointers to structures (addr of map also ok)
// If it's a single structure, only the first record returned will be set.
// If it's a slice or map, the slice/map won't be emptied first. New records will be allocated for each found record.
// If its a map, there is the potential to overwrite values (keys are 'id')
// Returns the number of items found (which is not necessarily the # of items set)
func (sess *Session) SelectAll(dest interface{}, sql string, params ...interface{}) (int, error) {

	//
	// Validate the dest, and extract the reflection values we need.
	//
	valueOfDest := reflect.ValueOf(dest) // We want this to eventually be a map or slice
	kindOfDest := valueOfDest.Kind()     // And this eventually needs to be a map or slice as well

	if kindOfDest == reflect.Ptr {
		valueOfDest = reflect.Indirect(valueOfDest)
		kindOfDest = valueOfDest.Kind()
	} else if kindOfDest == reflect.Map {
		// we're good
	} else {
		panic("invalid type passed to AllBySql. Need a map or addr of slice")
	}

	if !(kindOfDest == reflect.Map || kindOfDest == reflect.Slice) {
		panic("invalid type passed to AllBySql. Need a map or addr of slice")
	}

	recordType := valueOfDest.Type().Elem()
	if recordType.Kind() != reflect.Ptr {
		panic("Elements need to be pointersto structures")
	}

	recordType = recordType.Elem()
	if recordType.Kind() != reflect.Struct {
		panic("Elements need to be pointers to structures")
	}

	//
	// Get full SQL
	//
	fullSql, err := Interpolate(sql, params)
	if err != nil {
		return 0, err
	}

	numberOfRowsReturned := 0
	
	// Start the timer:
	startTime := time.Now()
	defer func() { sess.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), map[string]string{"sql": sql}) }()

	// Run the query:
	rows, err := sess.cxn.Db.Query(fullSql)
	if err != nil {
		fmt.Println("dbr.error.query") // Kvs{"error": err.String(), "sql": fullSql}
		return 0, err
	}

	// Iterate over rows
	if kindOfDest == reflect.Slice {
		sliceValue := valueOfDest
		for rows.Next() {
			// Create a new record to store our row:
			pointerToNewRecord := reflect.New(recordType)
			newRecord := reflect.Indirect(pointerToNewRecord)

			// Build a 'holder', which is an []interface{}. Each value will be the address of the field corresponding to our newly made record:
			holder, err := sess.holderFor(recordType, newRecord, rows)
			if err != nil {
				return numberOfRowsReturned, err
			}

			// Load up our new structure with the row's values
			err = rows.Scan(holder...)
			if err != nil {
				return numberOfRowsReturned, err
			}

			fmt.Println(newRecord.Interface())

			// Append our new record to the slice:
			sliceValue = reflect.Append(sliceValue, pointerToNewRecord)

			numberOfRowsReturned += 1
		}
		valueOfDest.Set(sliceValue)
	} else { // Map

	}

	// Check for errors at the end. Supposedly these are error that can happen during iteration.
	if err = rows.Err(); err != nil {
		return numberOfRowsReturned, err
	}
	
	return numberOfRowsReturned, nil
}

var destDummy interface{}

func (sess *Session) holderFor(recordType reflect.Type, record reflect.Value, rows *sql.Rows) ([]interface{}, error) {

	// Parts of this could be cached by: {recordType, rows.Columns()}. Or, {recordType, sqlTempalte}
	// (i think. if sqlTemplate has mixed in values it might not be as efficient as it could be. It's almost like we want to parse the SQL query a bit to get Select and From and Join. Everything before Where.).

	// Given a query and given a structure (field list), there's 2 sets of fields.
	// Take the intersection. We can fill those in. great.
	// For fields in the structure that aren't in the query, we'll let that slide if db:"-"
	// For fields in the structure that aren't in the query but without db:"-", return error
	// For fields in the query that aren't in the structure, we'll ignore them.

	// Get the columns:
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("dbr.error.columns")
		return nil, err
	}
	lenColumns := len(columns)

	// compute fieldMap:
	// each value is either the field index in the record, or -1 if we don't want to map it to the structure.
	fieldMap := make([]int, lenColumns)
	fmt.Println(columns)

	for i, col := range columns {
		fieldMap[i] = -1
		lenFields := recordType.NumField()
		for j := 0; j < lenFields; j += 1 {
			fieldStruct := recordType.Field(j)
			name := fieldStruct.Tag.Get("db")
			if name != "-" {
				if name == "" {
					name = NameMapping(fieldStruct.Name)
				}
				if name == col {
					fieldMap[i] = j
				}
			}
		}
	}

	holder := make([]interface{}, lenColumns) // In the future, this should be passed into this function.
	for i, fieldIndex := range fieldMap {
		if fieldIndex == -1 {
			holder[i] = &destDummy
		} else {
			field := record.Field(fieldIndex)
			holder[i] = field.Addr().Interface()
		}

	}

	return holder, nil
}
