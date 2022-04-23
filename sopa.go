package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// GetSopaReportPages goes to the page that lists the SOPA reports
// for a particular forest
// https://www.fs.fed.us/sopa/forest-level.php?110801
func GetSopaReportPages(url string) ([]string, error) {
	res := get(url)

	// Load the HTML document into goquery document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Parse page to get the links to each SOPA report
	projectPages := []string{}
	doc.Find("table table tbody td a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "html") {
			val, exists := s.Attr("href")
			if !exists {
				val = ""
			}
			projectPages = append(projectPages, fmt.Sprintf("%s%s", baseUrl, val))
		}
	})

	return projectPages, nil
}

func getProjects(url string) ([]ProjectUpdate, error) {
	// get nav page html
	res := get(url)

	// load the HTML document into goquery document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var region []string // region and district
	var project ProjectUpdate
	projects := []ProjectUpdate{}
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		if trim(s.Text()) == "No Projects matching your search criteria found..." {
			return
		}
		// ignore first tree
		if i < 3 {
			return
		}

		// new region
		val, exists := s.Attr("title")
		if exists {
			switch val {
			case "Group of Projects":
				region = []string{}
				s.Find("td").Each(func(i int, s *goquery.Selection) {
					region = append(region, trim(s.Text()))
				})
			case "ProjectDescription":
				project.SetDescription(trim(s.Text()))
			case "ProjectLocation":
				project.Location = strings.Replace(trim(s.Text()), "Location:", "", 1)
				project.Location = project.Location[1:] // remove first char, which isn't an ascii space
				projects = append(projects, project)
			}
			return
		}

		if len(s.Has("td").Nodes) == 0 {
			return
		}

		// if no region assume national
		if region[1] == "" {
			region[1] = "National"
		}

		project = ProjectUpdate{
			Region:   region[1],
			District: region[0],
		}

		// Set sopa report date on each project update
		project.SetSopaReportDateFromURL(url)

		s.Find("td").Each(func(i int, s *goquery.Selection) {
			html, err := s.Html()
			if err != nil {
				log.Fatal(err.Error())
			}
			html = trim(html) // TODO we're parsing the space in between projects, probably shouldn't
			switch i {
			case 0:
				project.SetNameAndCode(html)
			case 1:
				project.SetPurposes(s.Text())
			case 2:
				project.SetStatus(html)
			case 3:
				project.Decision = trim(s.Text())
			case 4:
				project.ExpectedImplementation = trim(s.Text())
			case 5:
				project.SetContacts(html)
			}
		})
	})
	return projects, nil
}

var getDate = regexp.MustCompile(`(\d\d-\d\d-\d\d\d\d)`)

func getReportDocumentMeta(id string) ([]ProjectDocument, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(fmt.Sprintf(
		"https://www.fs.usda.gov/wps/PA_Nepa/neparssgetfile?project=%s",
		id,
	))
	if err != nil {
		return []ProjectDocument{}, err
	}

	docs := []ProjectDocument{}
	for _, item := range feed.Items {
		if len(item.Link) > 0 {
			split := strings.Split(item.Categories[len(item.Categories)-1], "/")
			doc := ProjectDocument{
				Name:     item.Title,
				Url:      item.Link,
				Category: split[len(split)-1],
			}

			if dateStrings := getDate.FindStringSubmatch(item.Description); len(dateStrings) == 2 {
				doc.DateString = dateStrings[1]
				date, err := time.Parse("01-02-2006", dateStrings[1])
				if err == nil {
					doc.Date = date
				}
			}

			docs = append(docs, doc)
		}
	}

	return docs, nil
}
