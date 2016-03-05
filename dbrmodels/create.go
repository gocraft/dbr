package main

import (
	"os"
	"fmt"
)

func DoCreate(verbose bool) {
	p := new(Project)
	for {
		fmt.Printf("Enter project name: ")
		fmt.Scanf("%s", &p.Name)
		fmt.Printf("Enter files location path: ")
		fmt.Scanf("%s", &p.Path)
		fmt.Printf("Enter DB host:[localhost] ")
		fmt.Scanf("%s", &p.DBHost)
		if p.DBHost == "" {
			p.DBHost = "localhost"
		}
		fmt.Printf("Enter DB port:[3306] ")
		fmt.Scanf("%s", &p.DBPort)
		if p.DBPort == "" {
			p.DBPort = "3306"
		}
		fmt.Printf("Enter DB user:[root] ")
		fmt.Scanf("%s", &p.DBUser)
		if p.DBUser == "" {
			p.DBUser = "root"
		}
		fmt.Printf("Enter DB password:[] ")
		fmt.Scanf("%s", &p.DBPass)
		if p.DBPass == "" {
			p.DBPass = ""
		}
		fmt.Printf("Enter DB name:[test] ")
		fmt.Scanf("%s", &p.DBName)
		if p.DBName == "" {
			p.DBName = "test"
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
