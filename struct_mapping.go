package dbr

import (
	"errors"
	"fmt"
	"reflect"
)

var destDummy interface{}

type fieldMapQueueElement struct {
	Type   reflect.Type
	Idxs   []int
	Prefix string
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

			for j := 0; j < lenFields; j++ {
				fieldStruct := curType.Field(j)

				// Skip unexported field
				if len(fieldStruct.PkgPath) != 0 {
					continue
				}

				name := fieldStruct.Tag.Get("dbr")
				if len(name) < 1 {
					name = fieldStruct.Tag.Get("db")
				}

				if name != "-" {
					if name == "" {
						name = NameMapping(fieldStruct.Name)
					}
					if curEntry.Prefix+name == col {
						fieldMap[i] = append(curIdxs, j)
						break QueueLoop
					}

					if fieldStruct.Type.Kind() == reflect.Struct {
						var idxs2 []int
						copy(idxs2, curIdxs)
						idxs2 = append(idxs2, j)
						queue = append(queue, fieldMapQueueElement{Type: fieldStruct.Type, Idxs: idxs2, Prefix: name + "__"})
					} else if fieldStruct.Type.Kind() == reflect.Ptr && fieldStruct.Type.Elem().Kind() == reflect.Struct {
						var idxs2 []int
						copy(idxs2, curIdxs)
						idxs2 = append(idxs2, j)
						queue = append(queue, fieldMapQueueElement{Type: fieldStruct.Type.Elem(), Idxs: idxs2, Prefix: name + "__"})
					}
				}
			}
		}

		if requireAllColumns && fieldMap[i] == nil {
			return nil, errors.New(fmt.Sprint("couldn't find match for column ", col))
		}
	}

	return fieldMap, nil
}

func (sess *Session) prepareHolderFor(record reflect.Value, fieldMap [][]int, holder []interface{}) ([]interface{}, error) {
	// Given a query and given a structure (field list), there's 2 sets of fields.
	// Take the intersection. We can fill those in. great.
	// For fields in the structure that aren't in the query, we'll let that slide if db:"-"
	// For fields in the structure that aren't in the query but without db:"-", return error
	// For fields in the query that aren't in the structure, we'll ignore them.

	for i, fieldIndex := range fieldMap {
		if fieldIndex == nil {
			holder[i] = &destDummy
		} else {
			field := fieldByIndexFor(record, fieldIndex)
			holder[i] = field.Addr().Interface()
		}
	}

	return holder, nil
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
			field := fieldByIndexFor(record, fieldIndex)
			values[i] = field.Interface()
		}
	}

	return values, nil
}

// Get or create field structure
func fieldByIndexFor(record reflect.Value, index []int) reflect.Value {
	if record.Kind() == reflect.Ptr {
		record = record.Elem()
	}

	if len(index) < 2 {
		return record.FieldByIndex(index)
	}

	field := record.FieldByIndex([]int{index[0]})
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}
	return fieldByIndexFor(field, index[1:])
}
