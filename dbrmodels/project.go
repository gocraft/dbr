package main

import (
	"path/filepath"
	"os/user"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
)

type Project struct {
	Name   string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
	Path   string
}

var dbrDir string

func init() {
	u, _ := user.Current()
	dbrDir = u.HomeDir + "/.dbrmodels/"
	err := os.MkdirAll(dbrDir, 0700)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func Store(p *Project) error {
	data, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dbrDir+p.Name, data, 0600)
	return err
}

func GetAllProjects() []*Project {
	projects := make([]*Project, 0)
	filepath.Walk(dbrDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if info.Size() == 0 {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		p := new(Project)
		err = json.Unmarshal(data, p)
		if err != nil {
			return err
		}
		projects = append(projects, p)
		return nil
	})
	return projects
}

func GetProject(name string) *Project {
	projects := GetAllProjects()
	for _, proj := range projects {
		if proj.Name == name {
			return proj
			break
		}
	}
	return nil
}
