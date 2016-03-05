package main

import (
	"database/sql"
	"fmt"
	"os"
)

func DoWork(name string, verbose bool) {
	p := GetProject(name)
	if p == nil {
		fmt.Fprintf(os.Stderr, "no that project - %s\n", name)
		fmt.Fprintf(os.Stderr, "type `dbrmodels create %s` to create one\n", name)
		return
	}
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", p.DBUser, p.DBPass, p.DBHost, p.DBPort, p.DBName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection error: %s\n", err)
		return
	}
	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection error: %s\n", err)
		return
	}
	defer db.Close()

	var tabl string
	var ty string
	if rows, err := db.Query("SHOW FULL TABLES WHERE Table_Type != 'VIEW'"); err == nil {
		for rows.Next() {
			rows.Scan(&tabl, &ty)
			if verbose {
				fmt.Printf("\ngenerate `%s` table\n", tabl)
			}
			CreateTableModel(p.Path, tabl, db, verbose)
		}
	}
}
