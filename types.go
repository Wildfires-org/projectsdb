package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var baseUrl = "https://www.fs.fed.us"

type Forest struct {
	State       string          `json:"state"`
	Name        string          `json:"name"`
	Url         string          `json:"url"`
	Id          int             `json:"id"`
	SopaReports []string        `json:"sopa_reports"`
	Projects    []ProjectUpdate `json:"projects"`
}

func (forest Forest) AsCsv() [][]string {
	rows := [][]string{}
	baseRow := []string{
		forest.Name,
		forest.State,
		forest.Url,
		fmt.Sprint(forest.Id),
	}
	for _, project := range forest.Projects {
		rows = append(rows, append(baseRow, []string{
			project.Name,
			strings.Join(project.Purposes, ", "),
			project.Status,
			project.Decision,
			project.ExpectedImplementation,
			project.Contact.Name,
			project.Contact.Email,
			project.Contact.Phone,
			project.Description,
			project.WebLink,
			project.Region,
			project.District,
			project.SopaReportDate,
		}...))
	}
	return rows
}

type ProjectUpdate struct {
	Name                   string   `json:"name"`
	Purposes               []string `json:"purpose"`
	Status                 string   `json:"status"`
	Decision               string   `json:"decision"`
	ExpectedImplementation string   `json:"expected_implementation"`
	Contact                Contact  `json:"contact"`
	Description            string   `json:"description"`
	WebLink                string   `json:"web_link"`
	Location               string   `json:"location"`
	Region                 string   `json:"region"`
	District               string   `json:"district"`
	SopaReportDate         string   `json:"sopa_report_date"`
	ProjectCode            string   `json:"project_code"`
}

func (project *ProjectUpdate) SetNameAndCode(html string) {
	nameSplit := strings.Split(html, "<br/>")
	project.Name = nameSplit[0] // TODO
	if len(nameSplit) > 1 {
		project.ProjectCode = nameSplit[1]
	}
}

func (project *ProjectUpdate) SetPurposes(text string) {
	purposes := strings.Split(text, " - ")
	for i, purpose := range purposes {
		purpose = strings.Trim(purpose, "-")
		purpose = strings.TrimSpace(purpose)
		purposes[i] = purpose
	}
	project.Purposes = purposes
}

func (project *ProjectUpdate) SetStatus(html string) {
	project.Status = strings.ReplaceAll(html, "<br/>", "\n")
}

func (project *ProjectUpdate) SetContacts(html string) {
	contacts := strings.Split(html, "<br/>")
	if len(contacts) != 3 {
		fmt.Println(html)
		log.Fatal("bad contacts")
	}
	project.Contact = Contact{
		Name:  contacts[0],
		Phone: contacts[1],
		Email: contacts[2],
	}
}

func (project *ProjectUpdate) SetSopaReportDateFromURL(url string) {
	project.SopaReportDate = url[len(url)-12 : len(url)-5] // example url: https://www.fs.fed.us/sopa/components/reports/sopa-110519-2021-07.html
}

var descriptionAndLink = regexp.MustCompile("Description:(.*)Web Link:(.*)")

func (project *ProjectUpdate) SetDescription(text string) {
	if matches := descriptionAndLink.FindStringSubmatch(text); len(matches) == 3 {
		project.Description = matches[1]
		project.WebLink = matches[2]
	} else {
		project.Description = strings.Replace(text, "Description:", "", 1)
		project.Description = project.Description[1:] // remove first char, which isn't an ascii space
	}
}

type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

var singleSpacePattern = regexp.MustCompile(`\s+`)

func trim(s string) string {
	s = strings.TrimSpace(s)
	s = singleSpacePattern.ReplaceAllString(s, " ")
	return s
}
