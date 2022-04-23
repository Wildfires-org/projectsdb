package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetForests goes to the naviation page of the USFS website
// and pulls the name, state, id, and url for each forest.
// This function only makes 1 web request
func GetForests() ([]Forest, error) {
	// get nav page html
	url := fmt.Sprintf("%s/sopa/nav-page.php", baseUrl)
	res := get(url)

	// load the HTML document into goquery document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// iterate through ul's while holding the context
	// of what state container we're in
	forests := []Forest{}
	state := ""
	doc.Find("table #content-table div").Last().Children().Each(func(i int, s *goquery.Selection) {
		if s.Is("h3") {
			state = s.Text()
		} else if s.Is("ul") {
			s.Find("a").Each(func(i int, s *goquery.Selection) {
				val, exists := s.Attr("href")
				if !exists {
					val = ""
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

// getIdFromUri is parsing a url that looks like this
// https://www.fs.fed.us/sopa/forest-level.php?110801
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
