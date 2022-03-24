package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	//forests, err := GetForests()
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//fmt.Printf("Found {%d} forests\n", len(forests))

	string_with_description := "Description: some text blah"
	new_str := strings.Replace(string_with_description, "Description: ", "", 1)
	fmt.Printf("%v \n", new_str)

	GetProjects([]Forest{{Name: "somethin", Url: "https://www.fs.fed.us/sopa/forest-level.php?110801"}})
	//err = saveProjectsCsv(forests)
}

func saveProjectsJson(forests []Forest) error {
	data, err := json.Marshal(forests)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("data/forests.json", data, 0644)
}

func saveProjectsCsv(forests []Forest) error {
	csvFile, err := os.Create("data/forest.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{
		"Forest Name",
		"Forest State",
		"Forest URL",
		"Forest ID",
		"Project Name",
		"Project Purposes",
		"Project Status",
		"Project Decision",
		"Project Expected Implementation",
		"Project Contact Name",
		"Project Contact Email",
		"Project Contact Phone",
		"Project Description",
		"Project Location",
		"Project Web Link",
		"Project Region",
		"Project SOPA Report Date",
		"Project Code",
	})
	for _, forest := range forests {
		writer.WriteAll(forest.AsCsv())
	}
	writer.Flush()

	return nil
}

func printProject(p ProjectUpdate) {
	fmt.Println()
	fmt.Printf("==== Name ==== \n%s \n", p.Name)
	// fmt.Printf("==== Purposes ==== \n")
	// for _, purpose := range p.Purposes {
	// 	fmt.Printf("{%s} ", purpose)
	// }

	// fmt.Printf("==== Status ==== \n%s \n", p.Status)
	// fmt.Printf("==== Decision ==== \n%s \n", p.Decision)
	// fmt.Printf("==== ExpectedImplementation ==== \n%s \n", p.ExpectedImplementation)
	// fmt.Printf("==== Contact ==== \n%s \n", p.Contact)
	// fmt.Printf("==== Description ==== \n%s \n", trim(p.Description))
	// fmt.Printf("==== WebLink ==== \n%s \n", trim(p.WebLink))
	fmt.Printf("==== Location ==== \n%s \n", p.Location)
	fmt.Printf("==== Region ==== \n%s \n", p.Region)
	// fmt.Println()
}
