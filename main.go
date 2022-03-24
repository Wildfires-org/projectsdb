package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	aws_access_key_id     = "xxx"
	aws_secret_access_key = "xxx"
	aws_region_id         = "us-east-1"
)

func main() {
	// forests, err := GetForests()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Printf("Found {%d} forests\n", len(forests))

	// forests = GetProjects(forests)
	// err = saveProjectsCsv(forests)

	docs := getReportDocumentUrls("https://www.fs.usda.gov/project/?project=58124")
	fmt.Println(docs)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, ""),
	})
	if err != nil {
		log.Fatal(err)
	}
	uploader := s3manager.NewUploader(sess)

	for _, doc := range docs {
		fmt.Println("uploading reports")
		err = uploadReport(doc, 58124, uploader)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}

func getReportDocumentUrls(url string) []ProjectDocument {
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

	docs := []ProjectDocument{}
	doc.Find("main div#centercol div").Each(func(i int, s *goquery.Selection) {
		category, hasId := s.Attr("id")
		if hasId {
			s.Find("ul li ul li").Each((func(i int, s *goquery.Selection) {
				dateString := s.Find("div").Text()
				date, err := time.Parse("01-02-2006", s.Find("div").Text())
				if err != nil {
					fmt.Printf("Unable to parse {%s} date string", dateString)
				}

				link := s.Find("a")
				url, hasURL := link.Attr("href")
				if !hasURL {
					return
				}

				docs = append(docs, ProjectDocument{
					DateString: dateString,
					Date:       date,
					Category:   category,
					Name:       link.Text(),
					Url:        url,
				})
			}))
		}
	})
	return docs
}

type ProjectDocument struct {
	DateString string
	Date       time.Time
	Category   string
	Name       string
	Url        string
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
		"Project Web Link",
		"Project Region",
	})
	for _, forest := range forests {
		writer.WriteAll(forest.AsCsv())
	}
	writer.Flush()

	return nil
}

func uploadReport(doc ProjectDocument, projectId int, uploader *s3manager.Uploader) error {
	fmt.Println("geting %s", doc.Name)
	resp, err := http.Get(doc.Url)
	if err != nil {
		return err
	}

	fmt.Println("uploading %s", doc.Name)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   resp.Body,
		Bucket: aws.String("nepa-reports"),
		Key: aws.String(fmt.Sprintf(
			"%d/%s/%s.pdf",
			projectId,
			doc.Category,
			doc.Name,
		)),
	})
	if err != nil {
		return err
	}

	return nil
}

func printProject(p Project) {
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
