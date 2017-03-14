package main

import (
	"fmt"
	"os"
)

func DoView(name string) {
	p := GetProject(name)
	if p == nil {
		fmt.Fprintf(os.Stderr, "no that project - %s\n", name)
		return
	}
	fmt.Printf("\nProject: %s\nDBName: %s\nHost: %s\nPort: %s\nUser: %s\nPass: %s\nPath: %s\n\n",
		p.Name, p.DBName, p.DBHost, p.DBPort, p.DBUser, p.DBPass, p.Path)

}
