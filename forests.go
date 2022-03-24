package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetForests() ([]Forest, error) {
	// get nav page html
	res, err := http.Get(fmt.Sprintf("%s/sopa/nav-page.php", baseUrl))
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

	forests := []Forest{}
	state := ""
	doc.Find("table #content-table div").Last().Children().Each(func(i int, s *goquery.Selection) {
		if s.Is("h3") {
			state = s.Text()
		} else if s.Is("ul") {
			s.Find("a").Each(func(i int, s *goquery.Selection) {
				val, exists := s.Attr("href")
				if !exists {
					val = "brokenn"
				}
				forests = append(forests, Forest{
					State: state,
					Name:  s.Text(),
					Url:   fmt.Sprintf("%s%s", baseUrl, val),
					Id:    getIdFromUri(val),
				})
			})
		}
	})
	return forests, nil
}

func getIdFromUri(uri string) int {
	split := strings.Split(uri, "?")
	if len(split) != 2 {
		return 0
	}
	id, err := strconv.Atoi(split[1])
	if err != nil {
		return 0
	}
	return id
}
