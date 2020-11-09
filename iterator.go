package dbr

import (
	"database/sql"
	"reflect"
)

// Iterator is an interface to iterate over the result of a sql query
// and scan each row one at a time instead of getting all into one slice.
// The principe is similar to the standard sql.Rows type.
type Iterator interface {
	Next() bool
	Scan(interface{}) error
	Close() error
	Err() error
}

// recordMeta holds target struct's reflect metadata cache
type recordMeta struct {
	ptr           []interface{}
	elemType      reflect.Type
	isSlice       bool
	isMap         bool
	isMapOfSlices bool
	ts            *tagStore
	columns       []string
}

func (m *recordMeta) scan(rows *sql.Rows, value interface{}) (err error) {
	var v, elem, keyElem reflect.Value

	if il, ok := value.(interfaceLoader); ok {
		v = reflect.ValueOf(il.v)
	} else {
		v = reflect.ValueOf(value)
	}

	if m.elemType != nil {
		elem = reflectAlloc(m.elemType)
	} else if m.isMapOfSlices {
		elem = reflectAlloc(v.Type().Elem().Elem())
	} else if m.isSlice || m.isMap {
		elem = reflectAlloc(v.Type().Elem())
	} else {
		elem = v
	}

	if m.isMap {
		err = m.ts.findPtr(elem, m.columns[1:], m.ptr[1:])
		if err != nil {
			return
		}
		keyElem = reflectAlloc(v.Type().Key())
		err = m.ts.findPtr(keyElem, m.columns[:1], m.ptr[:1])
		if err != nil {
			return
		}
	} else {
		err = m.ts.findPtr(elem, m.columns, m.ptr)
		if err != nil {
			return
		}
	}

	// Before scanning, set nil pointer to dummy dest.
	// After that, reset pointers to nil for the next batch.
	for i := range m.ptr {
		if m.ptr[i] == nil {
			m.ptr[i] = dummyDest
		}
	}
	err = rows.Scan(m.ptr...)
	if err != nil {
		return
	}
	for i := range m.ptr {
		m.ptr[i] = nil
	}

	if m.isSlice {
		v.Set(reflect.Append(v, elem))
	} else if m.isMapOfSlices {
		s := v.MapIndex(keyElem)
		if !s.IsValid() {
			s = reflect.Zero(v.Type().Elem())
		}
		v.SetMapIndex(keyElem, reflect.Append(s, elem))
	} else if m.isMap {
		v.SetMapIndex(keyElem, elem)
	}
	return
}

func newRecordMeta(column []string, value interface{}) (meta *recordMeta, err error) {
	ptr := make([]interface{}, len(column))

	var v reflect.Value
	var elemType reflect.Type

	if il, ok := value.(interfaceLoader); ok {
		v = reflect.ValueOf(il.v)
		elemType = il.typ
	} else {
		v = reflect.ValueOf(value)
	}

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, ErrInvalidPointer
	}
	v = v.Elem()
	isScanner := v.Addr().Type().Implements(typeScanner)
	isSlice := v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 && !isScanner
	isMap := v.Kind() == reflect.Map && !isScanner
	isMapOfSlices := isMap && v.Type().Elem().Kind() == reflect.Slice && v.Type().Elem().Elem().Kind() != reflect.Uint8
	if isMap {
		v.Set(reflect.MakeMap(v.Type()))
	}

	s := newTagStore()
	return &recordMeta{
		elemType:      elemType,
		isSlice:       isSlice,
		isMap:         isMap,
		isMapOfSlices: isMapOfSlices,
		ts:            s,
		ptr:           ptr,
		columns:       column,
	}, nil
}

// iteratorInternals is the Iterator implementation (hidden)
type iteratorInternals struct {
	rows       *sql.Rows
	recordMeta *recordMeta
	columns    []string
}

// Next prepares the next result row for reading with the Scan method.
// It returns false is case of error or if there are not more data to fetch.
// If the end of the resultset is reached, it automaticaly free resources like
// the Close method.
func (i *iteratorInternals) Next() bool {
	return i.rows.Next()
}

// Scan fill the given struct with the current row.
func (i *iteratorInternals) Scan(value interface{}) (err error) {
	// First scan
	if i.recordMeta == nil {
		i.recordMeta, err = newRecordMeta(i.columns, value)
		if err != nil {
			return err
		}
	}
	return i.recordMeta.scan(i.rows, value)

}

// Close frees ressources created by the request execution.
func (i *iteratorInternals) Close() error {
	if err := i.Err(); err != nil {
		i.rows.Close()
		return err
	}
	return i.rows.Close()
}

// Err returns the error that was encountered during iteration, or nil.
// Always check Err after an iteration, like with the standard sql.Err method.
func (i *iteratorInternals) Err() error {
	return i.rows.Err()
}
