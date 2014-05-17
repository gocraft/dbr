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
