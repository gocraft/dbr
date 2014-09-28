# gocraft/dbr (database records) [![GoDoc](https://godoc.org/github.com/gocraft/web?status.png)](https://godoc.org/github.com/gocraft/dbr)

gocraft/db is a data access library for Go with a focus on simplicity, performance, and ease of use.

## Installation
From your GOPATH:

```bash
go get github.com/gocraft/dbr
```

You'll also want a driver for database/sql. Currently only MySQL has been tested (specifically [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)). Postgres (and others) probably won't work yet, but I would like to support it. PRs welcome!

You can get the MySQL driver with:

```bash
go get github.com/go-sql-driver/mysql
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
	Id     int64 `json:"id"`
	Title  string `json:"title"`
	CreatedAt dbr.NullTime `json:"created_at"`
}

// Hold a single global connection (pooling provided by sql driver)
var connection *dbr.Connection

func main() {
	// Create the connection during application initialization
	db, _ := sql.Open("mysql", "root@unix(/tmp/mysqld.sock)/your_database")
	connection = dbr.NewConnection(db, nil)

	// Create a session for each business unit of execution (e.g. a web request or goworkers job)
	dbrSess := connection.NewSession(nil)

	// Get a record
	var suggestion Suggestion
	err := dbrSess.Select("id, title").From("suggestions").Where("id = ?", 13).LoadStruct(&suggestion)

	if err != nil {
		println(err.Error())
	} else {
		println("Title:", suggestion.Title)
	}

	// JSON-ready, with dbr.Null* types serialized like you want
	recordJson, _ := json.Marshal(&suggestion)
	println(string(recordJson))
}
```

## Feature highlights
* **Simple reading and wrting** -  Structs, primitives, and slices thereof are supported. Results are mapped directly into your data structures without extra effort.
* **Composable query building** - Allow you to easily create queries in a dynamic and fluent manner.
* **Session-aware logging** - A pluggable logging interface is available which supports events, errors, and timers.
* **Custom SQL interpolation** - Ability to support more advanced constructs like IN clauses, and avoid throwaway prepared statements.
* **JSON Friendly** - Null values encode like you want, hiding their implementation details

## Driver support
Currently only MySQL has been tested because that is what we use. I would like to support others if there is time, but contributions are gladly accepted. Feel free to make an issue if you're interested in adding support and we can discuss what it would take.

## Usage Examples
The tests are a great place to see how to perform various queries, but here are a few highlights.

### Simple Record CRUD
```go
// Create a new suggestion record
suggestion := &Suggestion{Title: "My Cool Suggestion", State: "open"}

// Insert; inserting a record automatically sets an int64 Id field if present
response, err := dbrSess.InsertInto("suggestions").Columns("title", "state").Record(suggestion).Exec()

// Update
response, err = dbrSess.Update("suggestions").Set("title", "My New Title").Where("id = ?", suggestion.Id).Exec()

// Select
var otherSuggestion Suggestion
err = dbrSess.Select("id, title").From("suggestions").Where("id = ?", suggestion.Id).LoadStruct(&otherSuggestion)

// Delete
response, err = dbrSess.DeleteFrom("suggestions").Where("id = ?", otherSuggestion.Id).Limit(1).Exec()
```

### Primitive Values
```go
// Load primitives into existing variables
var ids []int64
idCount, err := sess.Select("id").From("suggestions").LoadValues(&ids)

var titles []string
titleCount, err := sess.Select("title").From("suggestions").LoadValues(&titles)

// Or return them directly
ids, err = sess.Select("id").From("suggestions").ReturnInt64s()
titles, err = sess.Select("title").From("suggestions").ReturnStrings()
```

### Overriding Column Names With Struct Tags
```go
// By default dbr converts CamelCase property names to snake_case column_names
// You can override this with struct tags, just like with JSON tags
// This is especially helpful while migrating from legacy systems
type Suggestion struct {
	Id        int64          `json:"id"`
	Title     dbr.NullString `json:"title" db:"subject"` // subjects are called titles now
    CreatedAt dbr.NullTime `json:"created_at"`
}
```

### Embedded structs
```go
// Columns are mapped to fields breadth-first
type Suggestion struct {
    Id        int64        `json:"id"`
    Title     string       `json:"title"`
    User      *struct {
        Id int64 `json:"user_id" db:"user_id"`
    }
}

var suggestion Suggestion
err := dbrSess.Select("id, title, user_id").From("suggestions").Limit(1).LoadStruct(&suggestion)
```

### JSON encoding of Null* types
```go
// dbr.Null* types serialize to JSON like you want
suggestion := &Suggestion{Id: 1, Title: "Test Title"}
jsonBytes, err := json.Marshal(&suggestion)
println(string(jsonBytes)) // {"id":1,"title":"Test Title","created_at":null}
```

### Inserting Multiple Records
```go
// Start bulding an INSERT statement
createDevsBuilder := sess.InsertInto("developers").Columns("name", "language", "employee_number")

// Add some new developers
for i := 0; i < 3; i++ {
	createDevsBuilder.Record(&Developer{Name: "Gopher", Language: "Go", EmployeeNumber: i})
}

// Execute statment
_, err := createDevsBuilder.Exec()
if err != nil {
	log.Fatalln("Error creating developers", err)
}
```

### Updating Records
```go
// Update any rubyists to gophers
response, err := sess.Update("developers").Set("name", "Gopher").Set("language", "Go").Where("language = ?", "Ruby").Exec()


// Alternatively use a map of attributes to update
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
response, err := sess.Update("developers").SetMap(attrsMap).Where("language = ?", "Ruby").Exec()
```

### Transactions
```go
// Basic transaction usage

// Start transaction
tx, err := dbrSess.Begin()
if err != nil {
    log.Fatalln(err.Error())
}

// Issue some statements
tx.Update("suggestions").Set("state", "deleted").Where("deleted_at IS NOT NULL").Exec()
tx.Update("comments").Set("state", "deleted").Where("deleted_at IS NOT NULL").Exec()

// Commit the transaction
err = tx.Commit()
```

### Generate SQL without executing
If you're only interested in building queries or want the built SQL for logging, you can generate it without executing
```go
// Create builder
builder := dbrSess.Select("*").From("suggestions").Where("subdomain_id = ?", 1)

// Get builder's SQL and arguments
sql, args := builder.ToSql()
fmt.Println(sql) // SELECT * FROM suggestions WHERE (subdomain_id = ?)
fmt.Println(args) // [1]
```

If you're only interested interested in dbr's query building and logging you could do something like this
```go
func main() {
	// Create the connection during application initialization
	db, _ := sql.Open("mysql", "root@unix(/tmp/mysqld.sock)/your_database")
	connection := dbr.NewConnection(db, nil)

	// Create a session for each business unit of execution (e.g. a web request or goworkers job)
	dbrSess := connection.NewSession(nil)

	// Create builder
	builder := dbrSess.Select("*").From("suggestions").Where("subdomain_id = ?", 1)

	// Get builder's SQL and arguments
	sql, args := builder.ToSql()

    // Use raw database/sql for actual query
	rows, err := db.Query(sql, args...)
	if err != nil {
		log.Fatalln(err)
	}
}
```

## Performance
Performance is a primary goal of dbr. I have a [collection of dbr related benchmarks](https://github.com/tyler-smith/golang-sql-benchmark) that should probably be checked when making performance-impacting changes.

## Contributing
We gladly accept contributions. We want to keep dbr pretty light but I certainly don't mind discussing any changes or additions. Feel free to open an issue if you'd like to discus a potential change.

## Thanks & Authors
Inspiration from these excellent libraries:
*  [sqlx](https://github.com/jmoiron/sqlx) - various useful tools and utils for interacting with database/sql.
*  [Squirrel](https://github.com/lann/squirrel) - simple fluent query builder.

Authors:
*  Jonathan Novak -- [https://github.com/cypriss](https://github.com/cypriss)
*  Tyler Smith -- [https://github.com/tyler-smith](https://github.com/tyler-smith)
