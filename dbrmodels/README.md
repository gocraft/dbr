# dbrmodels
Create projects for MySql database and generate table models structs for github.com/gocraft/dbr
## Getting started
Lets create MySql table
```sql
CREATE TABLE Persons
(
    PersonID    int(11) unsigned NOT NULL AUTO_INCREMENT,
    LastName    varchar(255),
    FirstName   varchar(255),
    Address     varchar(255) DEFAULT NULL,
    City        varchar(255) DEFAULT NULL,
	BirthDay 	date NOT NULL,
	PRIMARY KEY (`PersonID`)
);
```
and generate gocraft/dbr model
```go
package Persons

import "github.com/gocraft/dbr"

var fieldsNames = []string{"PersonID", "LastName", "FirstName", "Address", "City"}
var auto_increment_filed string = "PersonID"

type Persons struct {
	PersonID  int64          `db:"PersonID"`
	LastName  string         `db:"LastName"`
	FirstName string         `db:"FirstName"`
	Address   dbr.NullString `db:"Address"`
	City      dbr.NullString `db:"City"`
	BirthDay  dbr.NullTime   `db:"BirthDay"`
}

func New() *Persons {
	return new(Persons)
}

func NewSlice() []*Persons {
	return make([]*Persons, 0)
}

// return fields names
func FieldsNames() []string {
	return fieldsNames
}

// return fields names without auto_increment field for insert
func FieldsNamesWithOutAI() []string {
	var slice []string
	for _, iterator := range fieldsNames {
		if iterator == auto_increment_filed {
			continue
		}
		slice = append(slice, iterator)
	}
	return slice
}
```
## MySql types
| MySql | GO | NULL |
| ------------- | ------------- | ------------- |
| tinyint(1) | bool | dbr.NullBool |
| int | int64 | dbr.NullInt64 |
| float | float64 | dbr.NullFloat64 |
| date, datetime, timestamp | dbr.NullTime | dbr.NullTime |
| * | string | dbr.NullString |

# Install
```bash
go get github.com/gocraft/dbr/dbrmodels
```
## gocraft/dbr example
```go
// Get a record
var persons Persons.Persons
err := dbrSess.Select("*").From("Persons").Where("PersonID = ?", 1).Load(&persons)

// insert record
result, err := dbrSess.InsertInto("Persons").Columns(Persons.FieldsNamesWithOutAI()...).Record(&persons).Exec()
newAIID := result.LastInsertID()

// update record
_, err = dbrSess.Update("Persons").Set("Address", "far far away").Where("`PersonID`=?", newAIID).Exec()
```
# Projects

dbrmodels works with projects(databases) including next data:
* Project name
* DB Host
* DB Port
* DB User
* DB Password
* DB Name
* Path where to .go files located

### Example 
```json
{
        "Name": "test",
        "DBHost": "localhost",
        "DBPort": "3306",
        "DBUser": "root",
        "DBPass": "",
        "DBName": "test",
        "Path": "/home/finalist/go/src/github.com/finalist736/persons_dates_project/dbrmodels"
}
```

# Directory structure

```sql
SHOW TABLES;
- Persons
- TestDatesTable
```
converts into files:
```
- /home/finalist/go/src/github.com/finalist736/persons_dates_project/
---- dbrmodels/
--------- Persons/
-------------- model.go
--------- TestDatesTable/
-------------- model.go
```

# Using
start generate project
```bash
dbrmodels project_name
```
* list projects
```bash
dbrmodels ls
```
* create project
```bash
dbrmodels create
```
* edit project
```bash
dbrmodels edit project_name
```
* remove project
```bash
dbrmodels remove project_name
```
* view projects data
```bash
dbrmodels view project_name
```
