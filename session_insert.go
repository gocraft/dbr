package dbr

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Insert the record. Record should be a pointer to a structure.
// InsertInto("suggestions", []string{"title", "name", "etc"}, &suggestion)
func (sess *Session) InsertInto(table string, columns []string, record interface{}) error {
	if len(columns) < 1 {
		return errors.New("need at least one column")
	}

	valueOfRecord := reflect.ValueOf(record)
	indirectOfRecord := reflect.Indirect(valueOfRecord)
	kindOfRecord := valueOfRecord.Kind()

	if kindOfRecord != reflect.Ptr || indirectOfRecord.Kind() != reflect.Struct {
		panic("you need to pass in the address of a struct")
	}

	recordType := indirectOfRecord.Type()

	values, err := sess.valuesFor(recordType, indirectOfRecord, columns)
	if err != nil {
		return err
	}

	if len(values) != len(columns) {
		panic("that shouldn't have happened")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?%s)", table, strings.Join(columns, ", "), strings.Repeat(", ?", len(columns)-1))

	fmt.Println("values: ", values)
	fmt.Println("sql: ", sql)

	fullSql, err := Interpolate(sql, values)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("full sql: ", fullSql)

	result, err := sess.cxn.Db.Exec(fullSql)
	if err != nil {
		fmt.Println("got error on exec: ", err)
		return err
	}

	// If the structure has an "Id" field which is an int64, set it from the LastInsertId(). Otherwise, don't bother.
	idField := indirectOfRecord.FieldByName("Id")
	if idField.IsValid() && idField.Kind() == reflect.Int64 {
		lastId, err := result.LastInsertId()
		if err != nil {
			fmt.Println("got error on last insert id: ", err)
			return err
		}
		idField.Set(reflect.ValueOf(lastId))
	}

	return nil
}

func (sess *Session) valuesFor(recordType reflect.Type, record reflect.Value, columns []string) ([]interface{}, error) {
	fieldMap, err := sess.calculateFieldMap(recordType, columns, true)
	if err != nil {
		fmt.Println("err: calc field map")
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i, fieldIndex := range fieldMap {
		if fieldIndex == nil {
			panic("wtf bro")
		} else {
			field := record.FieldByIndex(fieldIndex)
			values[i] = field.Interface()
		}
	}

	return values, nil
}
