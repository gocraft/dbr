package main

import (
	"fmt"
	"os"
)

func DoRemove(name string, verbose bool) {
	p := GetProject(name)
	if p == nil {
		fmt.Fprintf(os.Stderr, "not that project: %s\n", name)
		return
	}
	if err := os.Remove(dbrDir + p.Name); err != nil {
		fmt.Fprintf(os.Stderr, "project removing error: %s\n", err)
	} else {
		if verbose {
			fmt.Println("done")
		}
	}
}
