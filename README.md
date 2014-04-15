# dbr (database records)

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
err := txn.Insert(&sugg)
txn.Commit()
