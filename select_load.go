package dbr

import (
	"fmt"
	"reflect"
	"time"
)

func (b *SelectBuilder) LoadAll(dest interface{}) (int, error) {
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
		panic("invalid type passed to LoadAll. Need a map or addr of slice")
	}

	if !(kindOfDest == reflect.Map || kindOfDest == reflect.Slice) {
		panic("invalid type passed to AllBySql. Need a map or addr of slice")
	}

	recordType := valueOfDest.Type().Elem()
	if recordType.Kind() != reflect.Ptr {
		panic("Elements need to be pointers to structures")
	}

	recordType = recordType.Elem()
	if recordType.Kind() != reflect.Struct {
		panic("Elements need to be pointers to structures")
	}

	//
	// Get full SQL
	//
	fullSql, err := Interpolate(b.ToSql())
	if err != nil {
		return 0, err
	}

	numberOfRowsReturned := 0

	// Start the timer:
	startTime := time.Now()
	defer func() {
		b.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), map[string]string{"sql": fullSql})
	}()

	// Run the query:
	rows, err := b.cxn.Db.Query(fullSql)
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
			holder, err := b.holderFor(recordType, newRecord, rows)
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

// Returns ErrNotFound if nothing was found
func (b *SelectBuilder) LoadOne(dest interface{}) error {
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
	fullSql, err := Interpolate(b.ToSql())
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { b.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), kvs{"sql": fullSql}) }()

	// Run the query:
	rows, err := b.cxn.Db.Query(fullSql)
	if err != nil {
		b.EventErrKv("dbr.select_one.query.error", err, kvs{"sql": fullSql})
		return err
	}

	if rows.Next() {
		// Build a 'holder', which is an []interface{}. Each value will be the address of the field corresponding to our newly made record:
		holder, err := b.holderFor(recordType, indirectOfDest, rows)
		if err != nil {
			return err
		}

		// Load up our new structure with the row's values
		err = rows.Scan(holder...)
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return ErrNotFound
}

// Returns ErrNotFound if no value was found, and it was therefore not set.
func (b *SelectBuilder) LoadValue(dest interface{}) error {
	return nil
}
