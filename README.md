# gocraft/dbr (database records) [![GoDoc](https://godoc.org/github.com/gocraft/web?status.png)](https://godoc.org/github.com/gocraft/dbr)

gocraft/db is a data access library for Go with a focus on simplicity, performance, and ease of use.

## Installation
From your GOPATH:

```bash
go get github.com/gocraft/dbr
```

## Getting Started

```go
package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)

// Simple data model
type Suggestion struct {
	Id     int64
	Title  string
}

// Hold a single global connection (pooling provided by sql driver)
var connection *dbr.Connection

func main() {
	// Create the connection during application initialization
	db, _ := sql.Open("mysql", "root@unix(/tmp/mysqld.sock)/your_database")
	connection = dbr.NewConnection(db, nil)

	// Create a session for each business unit of execution (e.g. a web request or goworkers job)
	sess := connection.NewSession(nil)

	// Get a record
	var suggestion Suggestion
	err := sess.Select("id, title").From("suggestions").Where("id = ?", 13).LoadStruct(&suggestion)

	if err != nil {
		println("Record not found")
	} else {
		println("Title:", suggestion.Title)
	}
}
```

## Features
* **Simple reading and wrting** -  Structs, primitives, and slices thereof are supported. Results are mapped directly into your data structures without extra effort.
* **Composable query building** - Allow you to easily create queries in a dynamic and fluent manner.
* **Builtin session-aware logging** - A pluggable logging interface is available which supports events, errors, and timers.

## Performance
Performance is a main goal. There are many existing data access solutions/patterns that trade performance for ease of use. While dbr should be simple to learn and use, it should also be fast and memory effient.



TODO: (TS) babble about prepared statements, allocs, maybe reflection

## Usage Examples
The tests are a great place to see how to perform various queries, but here are a few highlights.

### Simple Insert/Select Record
```go
// Create new user
// Creating a record atomatically populates an int64 Id field if present
user := &User{DisplayName: "Tyler"}
_, err := sess.InsertInto("users").Columns("display_name").Record(user).Exec()
if err != nil {
	log.Fatalln("Error creating new user")
}

// Create a new suggestion for the user
suggestion := &Suggestion{UserId: user.Id, Title: "New Suggestion"}
_, err = sess.InsertInto("suggestions").Columns("user_id", "title").Record(suggestion).Exec()
if err != nil {
	log.Fatalln("Error creating new suggestion")
}

// Select all suggestions from the user
allSuggestions := []*Suggestion{}
count, err := sess.Select("*").From("suggestions").Where("user_id = ?", user.Id).LoadStructs(&allSuggestions)
if err != nil {
	log.Fatalln("Error selecting suggestions")
}
```

### Inserting multiple records
```go
// Start bulding an INSERT statement
createDevsBuilder := sess.InsertInto("developers").Columns("name", "language")

// Add some new developers
for i := 0; i < 3; i++ {
	createDevsBuilder.Record(&Developer{Name: "Gopher", Language: "Go"})
}

// Execute statment
_, err := createDevsBuilder.Exec()
if err != nil {
	log.Fatalln("Error creating developers", err)
}
```

### Update Records
```go
// Update any Rubyists to Gophers
response, err := sess.Update("developers").Set("name", "Gopher").Set("language", "Go").Where("language = ?", "Ruby").Exec()


// Alternatively use a map of attributes to update
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
response, err := sess.Update("developers").SetMap(attrsMap).Where("language = ?", "Ruby").Exec()
```



TODO: (TS) add some cool examples; e.g. primitive loading, inserting/updating records, complex conditions, txns, embedded structs, struct tags, json marshaling



## Thanks & Authors
Inspiration from these excellent libraries:
*  [sqlx](https://github.com/jmoiron/sqlx) - various useful tools and utils for interacting with database/sql.
*  [Squirrel](https://github.com/lann/squirrel) - simple fluent query builder.

Authors:
*  Jonathan Novak -- [https://github.com/cypriss](https://github.com/cypriss)
*  Tyler Smith -- [https://github.com/tyler-smith](https://github.com/tyler-smith)

---






## TODO: (TS) REMOVE OR CLEANUP
## Usage

// At app initialization or something:
db, err := sql.Open("mysql", "...")
connection := dbr.New(db) // global variable

// In a business unit of execution (web request, job):
sess := connection.NewSession()

// Load records directly into a record, a map, or a slice:
err := sess.Select("*").From("suggestions").Where("x = ?", x).Load(&suggestion)

// Get a raw SQL string back:
sqlString, err := sess.Select("*").From("suggestions").Where("x = ?", x).Sql()

sess.Select("*").From("suggestions").WhereEq(dbr.Eq{"deleted_at": nil})


// To be determined: given a type like type Suggestion struct {...},  how do we map from results -> record efficiently


// Additionally, logging/metrics. Ideas:
// tight integreation with Health
// or...
// option to log all sql queries by table name

txn := sess.MustBegin()
err := txn.InsertInto("suggestions", []string{"title", "user_id"}, &sugg)
rowsUpdated, err := txn.Update("suggestions", []string{"title", "user_id"}, &sugg)
txn.Commit()
