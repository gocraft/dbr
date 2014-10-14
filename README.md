# gocraft/dbr (database records) [![GoDoc](https://godoc.org/github.com/gocraft/web?status.png)](https://godoc.org/github.com/gocraft/dbr)

gocraft/dbr provides additions to Go's database/sql for super fast performance and convenience.

## Getting Started

```go
package main

import (
	"database/sql"
  "fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)

// Simple data model
type Suggestion struct {
	Id        int64
	Title     string
	CreatedAt dbr.NullTime
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
		fmt.Println(err.Error())
	} else {
		fmt.Println("Title:", suggestion.Title)
	}

	// JSON-ready, with dbr.Null* types serialized like you want
	recordJson, _ := json.Marshal(&suggestion)
	fmt.Println(string(recordJson))
}
```

## Feature highlights

### Automatically map results to structs
Querying is the heart of gocraft/dbr. Automatically map results to structs:
```go
var posts []*struct {
	Id int64
	Title string
	Body dbr.NullString
}
err := sess.Select("id, title, body").
	From("posts").Where("id = ?", id).LoadStruct(&post)
```

Additionally, easily query a single value or a slice of values:
```go
id, err := sess.SelectBySql("SELECT id FROM posts WHERE title=?", title).ReturnInt64()
ids, err := sess.SelectBySql("SELECT id FROM posts", title).ReturnInt64s()
```

See below for many more examples.

### Use a Sweet Query Builder or use Plain SQL
gocraft/dbr supports both.

Sweet Query Builder:
```go
builder := sess.Select("title", "body").
	From("posts").
	Where("created_at > ?", someTime).
	OrderBy("id ASC").
	Limit(10)

var posts []*Post
n, err := builder.LoadStructs(&posts)
```

Plain SQL:
```go
n, err := sess.SelectBySql(`SELECT title, body FROM posts WHERE created_at > ?
                              ORDER BY id ASC LIMIT 10`, someTime).LoadStructs(&post)
```

### IN queries that aren't horrible
Traditionally, database/sql uses prepared statements, which means each argument in an IN clause needs its own question mark. gocraft/dbr, on the other hand, handles interpolation itself so that you can easily use a single question mark paired with a dynamically sized slice.

```go
// Traditional database/sql way:
ids := []int64{1,2,3,4,5}
questionMarks := []string
for _, _ := range ids {
	questionMarks = append(questionMarks, "?")
}
query := fmt.Sprintf("SELECT * FROM posts WHERE id IN (%s)",
	strings.Join(questionMarks, ",") // lolwut
rows, err := db.Query(query, ids) 

// gocraft/dbr way:
ids := []int64{1,2,3,4,5}
n, err := sess.SelectBySql("SELECT * FROM posts WHERE id IN ?", ids) // yay
```

### Amazing instrumentation
Writing instrumented code is a first-class concern for gocraft/dbr. We instrument each query to emit to a gocraft/health-compatible EventReceiver interface. NOTE: we have not released gocraft/health yet. This allows you to instrument your app to easily connect gocraft/dbr to your metrics systems, such statsd.

### Faster performance than using using database/sql directly
Every time you call database/sql's db.Query("SELECT ...") method, under the hood, the mysql driver will create a prepared statement, execute it, and then throw it away. This has a big performance cost.

gocraft/dbr doesn't use prepared statements. We ported mysql's query escape functionality directly into our package, which means we interpolate all of those question marks with their arguments before they get to MySQL. The result of this is that it's way faster, and just as secure.

Check out these [benchmarks](https://github.com/tyler-smith/golang-sql-benchmark).

### JSON Friendly
Every try to JSON-encode a sql.NullString? You get:
```json
{
	"str1": {
		"Valid": true,
		"String": "Hi!"
	},
	"str2": {
		"Valid": false,
		"String": ""
  }
}
```

Not quite what you want. gocraft/dbr has dbr.NullString (and the rest of the Null* types) that encode correctly, giving you:

```json
{
	"str1": "Hi!",
	"str2": null
}
```

## Driver support
Currently only MySQL has been tested because that is what we use. Feel free to make an issue for Postgres if you're interested in adding support and we can discuss what it would take.

## Usage Examples

### Making a session
All queries in gocraft/dbr are made in the context of a session. This is because when instrumenting your app, it's important to understand which business action the query took place in. See gocraft/health for more detail.

Here's an example web endpoint that makes a session:
```go
// At app startup. If you have a gocraft/health stream, pass it here instead of nil.
dbrCxn = dbr.NewConnection(db, nil)

func SuggestionsIndex(rw http.ResponseWriter, r *http.Request) {
	// Make a session. If you have a gocraft/health job, pass it here instead of nil.
	dbrSess := connection.NewSession(nil)

	// Do queries with the session:
	var sugg Suggestion
	err := dbrSess.Select("id, title").From("suggestions").
		Where("id = ?", suggestion.Id).LoadStruct(&sugg)

	// Render stuff, etc. Nothing else needs to be done with dbr.
}
```

### Simple Record CRUD
```go
// Create a new suggestion record
suggestion := &Suggestion{Title: "My Cool Suggestion", State: "open"}

// Insert; inserting a record automatically sets an int64 Id field if present
response, err := dbrSess.InsertInto("suggestions").
	Columns("title", "state").Record(suggestion).Exec()

// Update
response, err = dbrSess.Update("suggestions").
	Set("title", "My New Title").Where("id = ?", suggestion.Id).Exec()

// Select
var otherSuggestion Suggestion
err = dbrSess.Select("id, title").From("suggestions").
	Where("id = ?", suggestion.Id).LoadStruct(&otherSuggestion)

// Delete
response, err = dbrSess.DeleteFrom("suggestions").
	Where("id = ?", otherSuggestion.Id).Limit(1).Exec()
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
	Id        int64
	Title     dbr.NullString `db:"subject"` // subjects are called titles now
	CreatedAt dbr.NullTime
}
```

### Embedded structs
```go
// Columns are mapped to fields breadth-first
type Suggestion struct {
    Id        int64
    Title     string
    User      *struct {
        Id int64 `db:"user_id"`
    }
}

var suggestion Suggestion
err := dbrSess.Select("id, title, user_id").From("suggestions").
	Limit(1).LoadStruct(&suggestion)
```

### JSON encoding of Null* types
```go
// dbr.Null* types serialize to JSON like you want
suggestion := &Suggestion{Id: 1, Title: "Test Title"}
jsonBytes, err := json.Marshal(&suggestion)
fmt.Println(string(jsonBytes)) // {"id":1,"title":"Test Title","created_at":null}
```

### Inserting Multiple Records
```go
// Start bulding an INSERT statement
createDevsBuilder := sess.InsertInto("developers").
	Columns("name", "language", "employee_number")

// Add some new developers
for i := 0; i < 3; i++ {
	createDevsBuilder.Record(&Dev{Name: "Gopher", Language: "Go", EmployeeNumber: i})
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
response, err := sess.Update("developers").
	Set("name", "Gopher").
	Set("language", "Go").
	Where("language = ?", "Ruby").Exec()


// Alternatively use a map of attributes to update
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
response, err := sess.Update("developers").
	SetMap(attrsMap).Where("language = ?", "Ruby").Exec()
```

### Transactions
```go
// Start txn
tx, err := c.Dbr.Begin()
if err != nil {
	return err
}

// Rollback unless we're successful. You can also manually call tx.Rollback() if you'd like.
defer tx.RollbackUnlessCommitted()

// Issue statements that might cause errors
res, err := tx.Update("suggestions").Set("state", "deleted").Where("deleted_at IS NOT NULL").Exec()
if err != nil {
	return err
}

// Commit the transaction
if err := tx.Commit(); err != nil {
	return err
}
```

### Generate SQL without executing
```go
// Create builder
builder := dbrSess.Select("*").From("suggestions").Where("subdomain_id = ?", 1)

// Get builder's SQL and arguments
sql, args := builder.ToSql()
fmt.Println(sql) // SELECT * FROM suggestions WHERE (subdomain_id = ?)
fmt.Println(args) // [1]

// Use raw database/sql for actual query
rows, err := db.Query(sql, args...)
if err != nil {
    log.Fatalln(err)
}
```

## Contributing
We gladly accept contributions. We want to keep dbr pretty light but I certainly don't mind discussing any changes or additions. Feel free to open an issue if you'd like to discus a potential change.

## Thanks & Authors
Inspiration from these excellent libraries:
*  [sqlx](https://github.com/jmoiron/sqlx) - various useful tools and utils for interacting with database/sql.
*  [Squirrel](https://github.com/lann/squirrel) - simple fluent query builder.

Authors:
*  Jonathan Novak -- [https://github.com/cypriss](https://github.com/cypriss)
*  Tyler Smith -- [https://github.com/tyler-smith](https://github.com/tyler-smith)
