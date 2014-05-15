package dbr

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
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

func (sess *Session) QuerySelect(selectSql string) *Query {
	return &Query{Session: sess, SelectSql: selectSql}
}

func (sess *Session) SelectValue(dest interface{}, sql string, params ...interface{}) (bool, error) {
	// TODO: make sure dest is a ptr to something

	//
	// Get full SQL
	//
	fullSql, err := Interpolate(sql, params)
	if err != nil {
		return false, err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() {
		sess.TimingKv("dbr.select_value", time.Since(startTime).Nanoseconds(), map[string]string{"sql": fullSql})
	}()

	// Run the query:
	rows, err := sess.cxn.Db.Query(fullSql)
	if err != nil {
		fmt.Println("dbr.error.query") // Kvs{"error": err.String(), "sql": fullSql}
		return false, err
	}

	if rows.Next() {
		err = rows.Scan(dest)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (sess *Session) SelectUint64(sql string, params ...interface{}) (uint64, error) {
	var val uint64
	_, err := sess.SelectValue(&val, sql, params...)
	return val, err
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
	defer func() {
		sess.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), map[string]string{"sql": fullSql})
	}()

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

func (sess *Session) SelectOne(dest interface{}, sql string, params ...interface{}) (bool, error) {
	//
	// Validate the dest, and extract the reflection values we need.
	//
	valueOfDest := reflect.ValueOf(dest)
	indirectOfDest := reflect.Indirect(valueOfDest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr || indirectOfDest.Kind() != reflect.Struct {
		panic("you need to pass in the address of a struct")
	}

	recordType := indirectOfDest.Type()

	//
	// Get full SQL
	//
	fullSql, err := Interpolate(sql, params)
	if err != nil {
		return false, err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { sess.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), kvs{"sql": sql}) }()

	// Run the query:
	rows, err := sess.cxn.Db.Query(fullSql)
	if err != nil {
		sess.EventErrKv("dbr.select_one.query.error", err, kvs{"sql": fullSql})
		return false, err
	}

	if rows.Next() {
		// Build a 'holder', which is an []interface{}. Each value will be the address of the field corresponding to our newly made record:
		holder, err := sess.holderFor(recordType, indirectOfDest, rows)
		if err != nil {
			return false, err
		}

		// Load up our new structure with the row's values
		err = rows.Scan(holder...)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

var destDummy interface{}

type fieldMapQueueElement struct {
	Type reflect.Type
	Idxs []int
}

// recordType is the type of a structure
func (sess *Session) calculateFieldMap(recordType reflect.Type, columns []string, requireAllColumns bool) ([][]int, error) {
	// each value is either the slice to get to the field via FieldByIndex(index []int) in the record, or nil if we don't want to map it to the structure.
	lenColumns := len(columns)
	fieldMap := make([][]int, lenColumns)

	for i, col := range columns {
		fieldMap[i] = nil

		queue := []fieldMapQueueElement{fieldMapQueueElement{Type: recordType, Idxs: nil}}

	QueueLoop:
		for len(queue) > 0 {
			curEntry := queue[0]
			queue = queue[1:]

			curType := curEntry.Type
			curIdxs := curEntry.Idxs
			lenFields := curType.NumField()

			for j := 0; j < lenFields; j += 1 {
				fieldStruct := curType.Field(j)

				// Skip unexported field
				if len(fieldStruct.PkgPath) != 0 {
					continue
				}

				name := fieldStruct.Tag.Get("db")
				if name != "-" {
					if name == "" {
						name = NameMapping(fieldStruct.Name)
					}
					if name == col {
						fieldMap[i] = append(curIdxs, j)
						break QueueLoop
					}
				}

				if fieldStruct.Type.Kind() == reflect.Struct {
					var idxs2 []int
					copy(idxs2, curIdxs)
					idxs2 = append(idxs2, j)
					queue = append(queue, fieldMapQueueElement{Type: fieldStruct.Type, Idxs: idxs2})
				}
			}
		}

		if requireAllColumns && fieldMap[i] == nil {
			return nil, errors.New(fmt.Sprint("couldn't find match for column ", col))
		}
	}

	return fieldMap, nil
}

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
		return nil, err
	}
	lenColumns := len(columns)

	fieldMap, err := sess.calculateFieldMap(recordType, columns, false)
	if err != nil {
		return nil, err
	}

	holder := make([]interface{}, lenColumns) // In the future, this should be passed into this function.
	for i, fieldIndex := range fieldMap {
		if fieldIndex == nil {
			holder[i] = &destDummy
		} else {
			field := record.FieldByIndex(fieldIndex)
			holder[i] = field.Addr().Interface()
		}
	}

	return holder, nil
}
