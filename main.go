package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	forests, err := GetForests()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Found {%d} forests\n", len(forests))

	forests = GetProjects(forests)
	data, err := json.Marshal(forests)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = ioutil.WriteFile("forests.json", data, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}

}

var baseUrl = "https://www.fs.fed.us"

type Forest struct {
	State       string    `json:"state"`
	Name        string    `json:"name"`
	Url         string    `json:"url"`
	Id          string    `json:"id"`
	SopaReports []string  `json:"sopa_reports"`
	Projects    []Project `json:"projects"`
}

type Project struct {
	Name                   string   `json:"name"`
	Purpose                string   `json:"purpose"`
	Status                 string   `json:"status"`
	Decision               string   `json:"decision"`
	ExpectedImplementation string   `json:"expected_implementation"`
	Contact                string   `json:"contact"`
	Description            string   `json:"description"`
	Location               string   `json:"location"`
	Region                 []string `json:"region"`
}

func printProject(p Project) {
	fmt.Println()
	fmt.Printf("Name %d - ", len(p.Name))
	fmt.Printf("Purpose %d - ", len(p.Purpose))
	fmt.Printf("Status %d - ", len(p.Status))
	fmt.Printf("Decision %d - ", len(p.Decision))
	fmt.Printf("ExpectedImplementation %d - ", len(p.ExpectedImplementation))
	fmt.Printf("Contact %d - ", len(p.Contact))
	fmt.Printf("Description %d - ", len(p.Description))
	fmt.Printf("Location %d - ", len(p.Location))
	fmt.Printf("Region %d - ", len(p.Region))
	fmt.Println()
}
