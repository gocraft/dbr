package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const data = `package {{.PkgName}}
{{if .DbrUsed}}import "github.com/gocraft/dbr"{{end}}

var fieldsNames = []string{ {{.FieldsNames}} }
{{if .AutoInc}}var auto_increment_field string = "{{.AutoInc}}"{{end}}

type {{.Table}} struct {
	{{range .Fields}}{{.Name}}{{/*tab*/}} {{.Type}}{{/*tab*/}} {{.Tag}}
	{{end}}
}

func New() *{{.Table}} {
	return new({{.Table}})
}

func NewSlice() []*{{.Table}} {
	return make([]*{{.Table}}, 0)
}

func FieldsNames() []string {
	return fieldsNames
}
{{if .AutoInc}}
func FieldsNamesWithOutAI() []string {
	var slice []string
	for _, iterator := range fieldsNames {
		if iterator == auto_increment_field {
			continue
		}
		slice = append(slice, iterator)
	}
	return slice
}
{{end}}
`

type Field struct {
	Name string
	Type string
	Tag  string
}

type TplData struct {
	PkgName     string
	Table       string
	Fields      []Field
	FieldsNames string
	AutoInc     string
	DbrUsed     bool
}

func CreateTableModel(path, table string, db *sql.DB, verbose bool) {
	var (
		name  string
		typ   string
		null  string
		key   string
		def   sql.NullString
		extra string
	)

	template_data := TplData{}
	template_data.PkgName = table
	template_data.Table = strings.Title(table)

	// get table columns info
	q := fmt.Sprintf("SHOW COLUMNS FROM %s", table)
	if rows, err := db.Query(q); err == nil {
		if verbose {
			fmt.Println("\tfields:")
		}
		for rows.Next() {
			err := rows.Scan(&name, &typ, &null, &key, &def, &extra)
			if err != nil {
				fmt.Fprintf(os.Stderr, "rows.Scan: %s\n", err.Error())
				continue
			}
			if verbose {
				fmt.Printf("\t\tname: `%s` type: %s null: %s key: %s def: %s extra: %s\n", name, typ, null, key, def.String, extra)
			}
			titled_name := strings.Title(name)
			if extra == "auto_increment" {
				template_data.AutoInc = titled_name
			}

			if typ == "tinyint(1)" { // bool need be first because next `strings.Contains(typ, "int")`
				if null == "YES" {
					template_data.DbrUsed = true
					typ = "dbr.NullBool"
				} else {
					typ = "bool"
				}
			} else if strings.Contains(typ, "int") {
				if null == "YES" {
					template_data.DbrUsed = true
					typ = "dbr.NullInt64"
				} else {
					typ = "int64"
				}
			} else if strings.Contains(typ, "float") ||
				strings.Contains(typ, "decimal") ||
				strings.Contains(typ, "double") ||
				strings.Contains(typ, "real") {
				if null == "YES" {
					template_data.DbrUsed = true
					typ = "dbr.NullFloat64"
				} else {
					typ = "float64"
				}
			} else if strings.Contains(typ, "date") || strings.Contains(typ, "timestamp") {
				template_data.DbrUsed = true
				typ = "dbr.NullTime"
			} else {
				if null == "YES" {
					template_data.DbrUsed = true
					typ = "dbr.NullString"
				} else {
					typ = "string"
				}
			}

			tag := fmt.Sprintf("`db:\"%s\"`", name)
			if verbose {
				fmt.Printf("\t\t\t => %s %s %s\n", titled_name, typ, tag)
			}
			table_field := Field{titled_name, typ, tag}
			template_data.Fields = append(template_data.Fields, table_field)
			template_data.FieldsNames = fmt.Sprintf("%s, \"%s\"", template_data.FieldsNames, table_field.Name)
		}
		template_data.FieldsNames = strings.Trim(template_data.FieldsNames, ",")
	}
	t := template.Must(template.New("struct").Parse(data))

	fullPath := path + "/" + table
	fullFileName := fullPath + "/model.go"
	err := os.MkdirAll(fullPath, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file creating error: %s", err)
		return
	}

	file, err := os.Create(fullFileName)
	defer file.Close()
	if err == nil {
		if err := t.Execute(file, template_data); err != nil {
			fmt.Fprintf(os.Stderr, "template executing: %s", err)
			return
		}
		cmd := exec.Command("go", "fmt", fullFileName)
		err = cmd.Start()
		if err == nil {
			err = cmd.Wait()
		}
	}
}
