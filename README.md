# gocraft/dbr (database records) [![GoDoc](https://godoc.org/github.com/gocraft/web?status.png)](https://godoc.org/github.com/gocraft/dbr)

gocraft/dbr provides additions to Go's database/sql for super fast performance and convenience.

## Getting Started

```go
// create a connection
conn, _ := dbr.Open("postgres", "...")

// create a session for each business unit of execution (e.g. a web request or goworkers job)
sess := conn.NewSession(nil)

// get a record
var suggestion Suggestion
sess.Select("id", "title").From("suggestions").Where("id = ?", 1).LoadStruct(&suggestion)

// JSON-ready, with dbr.Null* types serialized like you want
json.Marshal(&suggestion)
```

## Feature highlights

### Join (new)
Join multiple tables.

```go
sess.Select("*").From("suggestions").Join("people", "people.suggestion_id = suggestions.id")
sess.Select("*").From("suggestions").LeftJoin("people", "people.suggestion_id = suggestions.id")
sess.Select("*").From("suggestions").RightJoin("people", "people.suggestion_id = suggestions.id")
sess.Select("*").From("suggestions").FullJoin("people", "people.suggestion_id = suggestions.id")
```

### Load (new)
The new `Load()` can replace LoadStruct, LoadStructs, LoadValue and LoadValues. In addition, the new `Load()` can load almost everything.

```go
var suggestions []Suggestion
sess.Select("*").From("suggestions").Load(&suggestions)
```

### Identity (new)

This can be used to quote database column or table.

```go
dbr.I("suggestions.id") // `suggestions`.`id`
```

### Subquery (new)

```go
sess.Select("count(id)").From(dbr.Select("*").From("suggestions").As("count"))
```

### Union (new)

```go
dbr.Union(
  dbr.Select("*"),
  dbr.Select("*"),
)

dbr.UnionAll(
  dbr.Select("*"),
  dbr.Select("*"),
)
```

Union can be used for `SelectStmt.From()`.

### Alias (new)

Also known as `AS`, and it is supported for:

* SelectStmt
* Identity
* Union

### Condition (new)

Building arbitrary condition with:

* And
* Or
* Eq
* Neq
* Gt
* Gte
* Lt
* Lte

```go
dbr.And(
  dbr.Or(
    dbr.Gt("created_at", "2015-09-10"),
    dbr.Lte("created_at", "2015-09-11"),
  ),
  dbr.Eq("title", "hello world"),
)
```

Building simple condition with:

* AndMap (previously EqMap)
* OrMap

```go
dbr.AndMap{
  "label": "testing",
  "age": 20,
}

dbr.OrMap{
  "label": "testing",
  "age": 20,
}
```

All these can be used where `Condition` is expected.

### Automatically map results to structs
Querying is the heart of gocraft/dbr. Automatically map results to structs:

```go
var suggestion Suggestion
sess.Select("id", "title", "body").From("suggestions").Where("id = ?", 1).LoadStruct(&suggestion)
```

Additionally, easily query a single value or a slice of values:

```go
var suggestions []Suggestion
sess.Select("id", "title", "body").From("suggestions").OrderBy("id").LoadStruct(&suggestions)
```

See below for many more examples.

### Use a Sweet Query Builder or use Plain SQL
gocraft/dbr supports both.

Sweet Query Builder:
```go
stmt := dbr.Select("title", "body").
	From("suggestions").
	OrderBy("id").
	Limit(10)
```

Plain SQL:

```go
builder := dbr.SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
```

### IN queries that aren't horrible
Traditionally, database/sql uses prepared statements, which means each argument in an IN clause needs its own question mark. gocraft/dbr, on the other hand, handles interpolation itself so that you can easily use a single question mark paired with a dynamically sized slice.

```go
ids := []int64{1, 2, 3, 4, 5}
builder.Where("id IN ?", ids) // `id` IN ?
```

### Amazing instrumentation
Writing instrumented code is a first-class concern for gocraft/dbr. We instrument each query to emit to a gocraft/health-compatible EventReceiver interface.

### Faster performance than using database/sql directly
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

### Making a session
All queries in gocraft/dbr are made in the context of a session. This is because when instrumenting your app, it's important to understand which business action the query took place in. See gocraft/health for more detail.

Here's an example web endpoint that makes a session:

### Simple Record CRUD

See `TestBasicCRUD`.

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

### Load from/to structs

```go
// columns are mapped by tag then by field
type Suggestion struct {
	ID int64  // id, will be autoloaded by last insert id
	Title string // title
	Url string `db:"-"` // ignored
	secret string // ignored
	Body dbr.NullString `db:"content"` // content
	User User
}

type User struct {
	Name string // name
}
```

### JSON encoding of Null* types
```go
// dbr.Null* types serialize to JSON like you want
suggestion := &Suggestion{Id: 1, Title: "Test Title"}
jsonBytes, err := json.Marshal(&suggestion)
fmt.Println(string(jsonBytes)) // {"id":1,"title":"Test Title","created_at":null}
```

### Inserting Multiple Records

```
sess.InsertInto("suggestions").Columns("title", "body")
	.Record(suggestion1)
	.Record(suggestion2)
```

### Updating Records

```go
sess.Update("suggestions").
	Set("title", "Gopher").
	Set("body", "I love go.").
	Where("id = ?", 1)
```

### Transactions

```go
tx, err := sess.Begin()
tx.Rollback()
```

## Driver support

* MySQL
* PostgreSQL

## gocraft

gocraft offers a toolkit for building web apps. Currently these packages are available:

* [gocraft/web](https://github.com/gocraft/web) - Go Router + Middleware. Your Contexts.
* [gocraft/dbr](https://github.com/gocraft/dbr) - Additions to Go's database/sql for super fast performance and convenience.
* [gocraft/health](https://github.com/gocraft/health) -  Instrument your web apps with logging and metrics.

These packages were developed by the [engineering team](https://eng.uservoice.com) at [UserVoice](https://www.uservoice.com) and currently power much of its infrastructure and tech stack.

## Thanks & Authors
Inspiration from these excellent libraries:
*  [sqlx](https://github.com/jmoiron/sqlx) - various useful tools and utils for interacting with database/sql.
*  [Squirrel](https://github.com/lann/squirrel) - simple fluent query builder.

Authors:
*  Jonathan Novak -- [https://github.com/cypriss](https://github.com/cypriss)
*  Tyler Smith -- [https://github.com/tyler-smith](https://github.com/tyler-smith)
*  Tai-Lin Chu -- [https://github.com/taylorchu](https://github.com/taylorchu)
*  Sponsored by [UserVoice](https://eng.uservoice.com)
