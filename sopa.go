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
		fmt.Printf("looking into the {%s}\n", forest.Name)
		projectPages := getProjectPages(forest.Url)

		if len(projectPages) == 0 {
			continue
		}

		forests[i].Projects = getProjects(projectPages[0])
		fmt.Printf("found %d projects\n", len(forests[i].Projects))
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
					region = append(region, s.Text())
				})
			case "ProjectDescription":
				project.Description = s.Text()
			case "ProjectLocation":
				project.Location = s.Text()
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
			switch i {
			case 0:
				project.Name = s.Text()
			case 1:
				project.Purpose = s.Text()
			case 2:
				project.Status = s.Text()
			case 3:
				project.Decision = s.Text()
			case 4:
				project.ExpectedImplementation = s.Text()
			case 5:
				project.Contact = s.Text()
			}
		})
	})
	return projects
}
