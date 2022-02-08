package main

import (
	"fmt"

	forests "github.com/wildfires_org/projectsdb/get_forests"
	sopa "github.com/wildfires_org/projectsdb/sopa"
)

func main() {
	forests, err := forests.GetForests()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	sopa.GetProjects(forests)
}
