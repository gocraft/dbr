package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("start working  - dbrmodels $project_name")
	fmt.Println("list projects  - dbrmodels list | ls")
	fmt.Println("create project - dbrmodels create")
	fmt.Println("edit project   - dbrmodels edit $project_name")
	fmt.Println("remove project - dbrmodels remove $project_name")
	fmt.Println("view project   - dbrmodels view $project_name")
}

func main() {

	verbose := false
	for index, v := range os.Args {
		if v == "-v" {
			verbose = true
			os.Args = append(os.Args[:index], os.Args[index+1:]...)
			break
		}
	}

	if len(os.Args) == 2 {
		arg := os.Args[1]
		if arg == "create" {
			DoCreate(verbose)
		} else if arg == "list" || arg == "ls" {
			DoList(verbose)
		} else {
			DoWork(arg, verbose)
		}
	} else if len(os.Args) == 3 {
		cmd := os.Args[1]
		name := os.Args[2]
		if cmd == "edit" {
			DoEdit(name, verbose)
		} else if cmd == "remove" {
			DoRemove(name, verbose)
		} else if cmd == "view" {
			DoView(name)
		} else {
			printUsage()
		}
	} else {
		printUsage()
	}
}
