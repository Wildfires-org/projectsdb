package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetProjects(forests []Forest) []Forest {
	for i, forest := range forests {
		fmt.Printf("looking into the {%s} of {%s}\n", forest.Name, forest.State)
		projectPages := getProjectPages(forest.Url)

		for _, projectPage := range projectPages {
			forests[i].Projects = append(forests[i].Projects, getProjects(projectPage)...)

		}
	}
	return forests
}

func getProjectPages(url string) []string {
	// get nav page html
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// load the HTML document into goquery document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	projectPages := []string{}
	doc.Find("table table tbody td a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "html") {
			val, exists := s.Attr("href")
			if !exists {
				val = "brokenn"
			}
			projectPages = append(projectPages, fmt.Sprintf("%s%s", baseUrl, val))
		}
	})
	return projectPages
}

func getProjects(url string) []Project {
	// get nav page html
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// load the HTML document into goquery document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var region []string
	var project Project
	projects := []Project{}
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
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
				project.Location = strings.Replace(trim(s.Text()), "Location: ", "", 1)
				projects = append(projects, project)
			}
			return
		}

		if len(s.Has("td").Nodes) == 0 {
			return
		}

		project = Project{
			Region: region,
		}

		s.Find("td").Each(func(i int, s *goquery.Selection) {
			html, err := s.Html()
			if err != nil {
				log.Fatal(err.Error())
			}
			html = trim(html)
			switch i {
			case 0:
				project.SetName(html)
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
	return projects
}
