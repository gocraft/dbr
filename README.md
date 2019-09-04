# gocraft/dbr (database records)

[![GoDoc](https://godoc.org/github.com/gocraft/dbr?status.png)](https://godoc.org/github.com/gocraft/dbr)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgocraft%2Fdbr.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgocraft%2Fdbr?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/gocraft/dbr)](https://goreportcard.com/report/github.com/gocraft/dbr)
[![CircleCI](https://circleci.com/gh/gocraft/dbr.svg?style=svg)](https://circleci.com/gh/gocraft/dbr)

gocraft/dbr provides additions to Go's database/sql for super fast performance and convenience.

```
$ go get -u github.com/gocraft/dbr/v2
```

```go
import "github.com/gocraft/dbr/v2"
```

## Driver support

* MySQL
* PostgreSQL
* SQLite3

## Examples

See [godoc](https://godoc.org/github.com/gocraft/dbr) for more examples.

### Open connections

```go
// create a connection (e.g. "postgres", "mysql", or "sqlite3")
conn, _ := Open("postgres", "...", nil)
conn.SetMaxOpenConns(10)

// create a session for each business unit of execution (e.g. a web request or goworkers job)
sess := conn.NewSession(nil)

// create a tx from sessions
sess.Begin()
```

### Create and use Tx

```go
sess := mysqlSession
tx, err := sess.Begin()
if err != nil {
	return
}
defer tx.RollbackUnlessCommitted()

// do stuff...

tx.Commit()
```

### SelectStmt loads data into structs

```go
// columns are mapped by tag then by field
type Suggestion struct {
	ID	int64		// id, will be autoloaded by last insert id
	Title	NullString	`db:"subject"`	// subjects are called titles now
	Url	string		`db:"-"`	// ignored
	secret	string		// ignored
}

// By default gocraft/dbr converts CamelCase property names to snake_case column_names.
// You can override this with struct tags, just like with JSON tags.
// This is especially helpful while migrating from legacy systems.
var suggestions []Suggestion
sess := mysqlSession
sess.Select("*").From("suggestions").Load(&suggestions)
```

### SelectStmt with where-value interpolation

```go
// database/sql uses prepared statements, which means each argument
// in an IN clause needs its own question mark.
// gocraft/dbr, on the other hand, handles interpolation itself
// so that you can easily use a single question mark paired with a
// dynamically sized slice.

sess := mysqlSession
ids := []int64{1, 2, 3, 4, 5}
sess.Select("*").From("suggestions").Where("id IN ?", ids)
```

### SelectStmt with joins

```go
sess := mysqlSession
sess.Select("*").From("suggestions").
	Join("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
	LeftJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

// join multiple tables
sess.Select("*").From("suggestions").
	Join("subdomains", "suggestions.subdomain_id = subdomains.id").
	Join("accounts", "subdomains.accounts_id = accounts.id")
```

### SelectStmt with raw SQL

```go
SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
```

### InsertStmt adds data from struct

```go
type Suggestion struct {
	ID		int64
	Title		NullString
	CreatedAt	time.Time
}
sugg := &Suggestion{
	Title:		NewNullString("Gopher"),
	CreatedAt:	time.Now(),
}
sess := mysqlSession
sess.InsertInto("suggestions").
	Columns("title").
	Record(&sugg).
	Exec()

// id is set automatically
fmt.Println(sugg.ID)
```

### InsertStmt adds data from value

```go
sess := mysqlSession
sess.InsertInto("suggestions").
	Pair("title", "Gopher").
	Pair("body", "I love go.")
```


## Benchmark (2018-05-11)

```
BenchmarkLoadValues/sqlx_10-8         	    5000	    407318 ns/op	    3913 B/op	     164 allocs/op
BenchmarkLoadValues/dbr_10-8          	    5000	    372940 ns/op	    3874 B/op	     123 allocs/op
BenchmarkLoadValues/sqlx_100-8        	    2000	    584197 ns/op	   30195 B/op	    1428 allocs/op
BenchmarkLoadValues/dbr_100-8         	    3000	    558852 ns/op	   22965 B/op	     937 allocs/op
BenchmarkLoadValues/sqlx_1000-8       	    1000	   2319101 ns/op	  289339 B/op	   14031 allocs/op
BenchmarkLoadValues/dbr_1000-8        	    1000	   2310441 ns/op	  210092 B/op	    9040 allocs/op
BenchmarkLoadValues/sqlx_10000-8      	     100	  17004716 ns/op	 3193997 B/op	  140043 allocs/op
BenchmarkLoadValues/dbr_10000-8       	     100	  16150062 ns/op	 2394698 B/op	   90051 allocs/op
BenchmarkLoadValues/sqlx_100000-8     	      10	 170068209 ns/op	31679944 B/op	 1400053 allocs/op
BenchmarkLoadValues/dbr_100000-8      	      10	 147202536 ns/op	23680625 B/op	  900061 allocs/op
```

## Thanks & Authors
Inspiration from these excellent libraries:
* [sqlx](https://github.com/jmoiron/sqlx) - various useful tools and utils for interacting with database/sql.
* [Squirrel](https://github.com/lann/squirrel) - simple fluent query builder.

Authors:
* Jonathan Novak -- [https://github.com/cypriss](https://github.com/cypriss)
* Tai-Lin Chu -- [https://github.com/taylorchu](https://github.com/taylorchu)
* Sponsored by [UserVoice](https://eng.uservoice.com)

Contributors:
* Paul Bergeron -- [https://github.com/dinedal](https://github.com/dinedal) - SQLite dialect

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgocraft%2Fdbr.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgocraft%2Fdbr?ref=badge_large)
