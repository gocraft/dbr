package main

import (
	"fmt"
	"os"
)

func DoEdit(name string, verbose bool) {
	p := GetProject(name)
	if p == nil {
		fmt.Fprintf(os.Stderr, "not that project: %s\n", name)
		return
	}
	for {
		var tmp string
		fmt.Printf("Enter files location path:[%s] ", p.Path)
		fmt.Scanf("%s", &tmp)
		if tmp != "" {
			p.Path = tmp
		}
		fmt.Printf("Enter DB host:[%s] ", p.DBHost)
		fmt.Scanf("%s", &tmp)
		if tmp != "" {
			p.DBHost = tmp
		}
		fmt.Printf("Enter DB port:[%s] ", p.DBPort)
		fmt.Scanf("%s", &tmp)
		if tmp != "" {
			p.DBPort = tmp
		}
		fmt.Printf("Enter DB user:[%s] ", p.DBUser)
		fmt.Scanf("%s", &tmp)
		if tmp != "" {
			p.DBUser = tmp
		}
		fmt.Printf("Enter DB password:[%s] ", p.DBPass)
		fmt.Scanf("%s", &tmp)
		if tmp != "" {
			p.DBPass = tmp
		}
		fmt.Printf("Enter DB name:[%s] ", p.DBName)
		fmt.Scanf("%s", &tmp)
		if tmp != "" {
			p.DBName = tmp
		}
		fmt.Printf("\nProject: %s\nDBName: %s\nHost: %s\nPort: %s\nUser: %s\nPass: %s\nPath: %s\n",
			p.Name, p.DBName, p.DBHost, p.DBPort, p.DBUser, p.DBPass, p.Path)
		var yesno string
		fmt.Printf("All correct:[Y/n] ")
		fmt.Scanf("%s", &yesno)
		if yesno == "" || yesno == "yes" || yesno == "y" || yesno == "Y" {
			break
		}
	}
	err := Store(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "project storing error: %s\n", err)
	} else {
		if verbose {
			fmt.Println("done")
		}
	}
}
