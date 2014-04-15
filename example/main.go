package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql/driver"
	"fmt"
	"dbr"
	"database/sql"
	"health"
	"os"
	"time"
	"reflect"
)


type Suggestion struct {
	Id int64
	Title sql.NullString
	
	Links struct {
		CreatedBy dbr.NullInt64 `db:"user_id"`
		ForumId dbr.NullInt64
	}
}

func timing(log dbr.EventReceiver, start time.Time, name string) {
	log.Timing(name, time.Since(start).Nanoseconds())
}

func main() {
	fmt.Println("hi")
	
	var poop Suggestion
	
	r := reflect.Indirect(reflect.ValueOf(&poop))
	
	f := r.FieldByIndex([]int{2, 1, 0, 0})
	fmt.Println(f)
	f.Set(reflect.ValueOf(int64(9)))
	
	
	fmt.Println("poop:")
	fmt.Println(poop)
	
	
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
	
	count, err := sess.SelectAll(&suggs, "SELECT id, title, user_id FROM suggestions order by id desc limit 5")
	fmt.Println("error = ", err, "count = ", count)
	fmt.Println("suggs = ", suggs[0])
	
	sess.InsertInto("suggestions", []string{"title", "user_id"}, suggs[0])
	
	suggs[0].Title.Valid = false
	var xxx interface{} = suggs[0].Title
	terd, ok := xxx.(driver.Valuer)
	if ok {
		zz, err := terd.Value()
		fmt.Println("its a valuer ", zz, err)
	}
}