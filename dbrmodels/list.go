package main

import (
	"fmt"
	"os"
)

func DoList(verbose bool) {
	projects := GetAllProjects()
	if len(projects) > 0 {
		for _, proj := range projects {
			fmt.Println(proj.Name)
		}
	} else {
		if verbose {
			fmt.Fprintf(os.Stderr, "no projects yet. Type `dbrmodels create` for create one\n")
		}
	}
}
