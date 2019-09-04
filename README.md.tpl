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

{{ "ExampleOpen" | example }}

### Create and use Tx

{{ "ExampleTx" | example }}

### SelectStmt loads data into structs

{{ "ExampleSelectStmt_Load" | example }}

### SelectStmt with where-value interpolation

{{ "ExampleSelectStmt_Where" | example }}

### SelectStmt with joins

{{ "ExampleSelectStmt_Join" | example }}

### SelectStmt with raw SQL

{{ "ExampleSelectBySql" | example }}

### InsertStmt adds data from struct

{{ "ExampleInsertStmt_Record" | example }}

### InsertStmt adds data from value

{{ "ExampleInsertStmt_Pair" | example }}


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
