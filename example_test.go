package dbr

import (
	"fmt"
	"time"
)

func ExampleOpen() {
	// create a connection (e.g. "postgres", "mysql", or "sqlite3")
	conn, _ := Open("postgres", "...", nil)
	conn.SetMaxOpenConns(10)

	// create a session for each business unit of execution (e.g. a web request or goworkers job)
	sess := conn.NewSession(nil)

	// create a tx from sessions
	sess.Begin()
}

func ExampleSelect() {
	Select("title", "body").
		From("suggestions").
		OrderBy("id").
		Limit(10)
}

func ExampleSelectBySql() {
	SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
}

func ExampleSelectStmt_Load() {
	// columns are mapped by tag then by field
	type Suggestion struct {
		ID     int64      // id, will be autoloaded by last insert id
		Title  NullString `db:"subject"` // subjects are called titles now
		Url    string     `db:"-"`       // ignored
		secret string     // ignored
	}

	// By default gocraft/dbr converts CamelCase property names to snake_case column_names.
	// You can override this with struct tags, just like with JSON tags.
	// This is especially helpful while migrating from legacy systems.
	var suggestions []Suggestion
	sess := mysqlSession
	sess.Select("*").From("suggestions").Load(&suggestions)
}

func ExampleSelectStmt_Where() {
	// database/sql uses prepared statements, which means each argument
	// in an IN clause needs its own question mark.
	// gocraft/dbr, on the other hand, handles interpolation itself
	// so that you can easily use a single question mark paired with a
	// dynamically sized slice.

	sess := mysqlSession
	ids := []int64{1, 2, 3, 4, 5}
	sess.Select("*").From("suggestions").Where("id IN ?", ids)
}

func ExampleSelectStmt_Join() {
	sess := mysqlSession
	sess.Select("*").From("suggestions").
		Join("subdomains", "suggestions.subdomain_id = subdomains.id")

	sess.Select("*").From("suggestions").
		LeftJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

	// join multiple tables
	sess.Select("*").From("suggestions").
		Join("subdomains", "suggestions.subdomain_id = subdomains.id").
		Join("accounts", "subdomains.accounts_id = accounts.id")
}

func ExampleSelectStmt_As() {
	sess := mysqlSession
	sess.Select("count(id)").From(
		Select("*").From("suggestions").As("count"),
	)
}

func ExampleInsertStmt_Pair() {
	sess := mysqlSession
	sess.InsertInto("suggestions").
		Pair("title", "Gopher").
		Pair("body", "I love go.")
}

func ExampleInsertStmt_Record() {
	type Suggestion struct {
		ID        int64
		Title     NullString
		CreatedAt time.Time
	}
	sugg := &Suggestion{
		Title:     NewNullString("Gopher"),
		CreatedAt: time.Now(),
	}
	sess := mysqlSession
	sess.InsertInto("suggestions").
		Columns("title").
		Record(&sugg).
		Exec()

	// id is set automatically
	fmt.Println(sugg.ID)
}

func ExampleUpdateStmt() {
	sess := mysqlSession
	sess.Update("suggestions").
		Set("title", "Gopher").
		Set("body", "I love go.").
		Where("id = ?", 1)
}

func ExampleDeleteStmt() {
	sess := mysqlSession
	sess.DeleteFrom("suggestions").
		Where("id = ?", 1)
}

func ExampleTx() {
	sess := mysqlSession
	tx, err := sess.Begin()
	if err != nil {
		return
	}
	defer tx.RollbackUnlessCommitted()

	// do stuff...

	tx.Commit()
}

func ExampleAnd() {
	And(
		Or(
			Gt("created_at", "2015-09-10"),
			Lte("created_at", "2015-09-11"),
		),
		Eq("title", "hello world"),
	)
}

func ExampleI() {
	// I, identifier, can be used to quote.
	I("suggestions.id").As("id") // `suggestions`.`id`
}

func ExampleUnion() {
	Union(
		Select("*"),
		Select("*"),
	).As("subquery")
}

func ExampleUnionAll() {
	UnionAll(
		Select("*"),
		Select("*"),
	).As("subquery")
}
