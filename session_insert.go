package dbr

import (
	"fmt"
	"reflect"
)

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
