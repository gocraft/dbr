package main

import (
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"dbr"
	"database/sql"
	"health"
	"os"
	"time"
)


type Suggestion struct {
	Id int64
	Title sql.NullString
}

func timing(log dbr.EventReceiver, start time.Time, name string) {
	log.Timing(name, time.Since(start).Nanoseconds())
}

func main() {
	fmt.Println("hi")
	
	stream := health.NewStream()
	stream.AddLogfileWriterSink(os.Stdout)
	
	
    db, err := sql.Open("mysql", "root:unprotected@unix(/tmp/mysql.sock)/uservoice_development?charset=utf8&parseTime=true")
    if err != nil {
      fmt.Println("Mysql error ", err)
      panic(err)
    }
	
	cxn := dbr.NewConnection(db, stream)
	
	// We're entering a web request yay
	sess := cxn.NewSession(stream.NewJob("api/v2/tickets/create"))
	//sess := cxn.NewSession(nil)
	
	var suggs []*Suggestion
	
	count, err := sess.SelectAll(&suggs, "SELECT id, title FROM suggestions order by id asc limit 100")
	fmt.Println("error = ", err, "count = ", count)
	fmt.Println("suggs = ", suggs[0])
	
	var oneSugg Suggestion
	found, err := sess.SelectOne(&oneSugg, "SELECT id, title FROM suggestions where id = ?", 0)
	fmt.Println("error = ", err, "found = ", found)
	fmt.Println("sugg = ", oneSugg)
}