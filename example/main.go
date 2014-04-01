package main

import (
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"dbr"
	"database/sql"
	"health"
	"os"
)




func main() {
	fmt.Println("hi")
	
	stream := health.NewStream()
	stream.AddLogfileWriterSink(os.Stdout)
	
	
    db, err := sql.Open("mysql", "root:unprotected@unix(/tmp/mysql.sock)/uservoice_dev?charset=utf8&parseTime=true")
    if err != nil {
      fmt.Println("Mysql error ", err)
      panic(err)
    }
	
	cxn := dbr.NewConnection(db, stream.Job("_"))
	
	// We're entering a web request yay
	sess := cxn.NewSession(stream.Job("api/v2/tickets/create"))
	
	fmt.Println(sess)
	
}