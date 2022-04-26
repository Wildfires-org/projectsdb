package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

/*
 * Parse just the updates
 */
type ParseUpdatesConfig struct {
	AccessKeyId     string `env required help:"AWS Access Key ID" type:"string"`
	SecretAccessKey string `env required help:"AWS Secret Access Key" type:"string"`
	RegionId        string `env required help:"AWS Region ID" type:"string"`
	BucketName      string `help:"S3 bucket that files will be uploaded to" type:"string"`
	SlackHookUrl    string `env required`
}

func ParseUpdates(config ParseUpdatesConfig) error {
	// Setup AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.RegionId),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKeyId,
			config.SecretAccessKey,
			"",
		),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Unable to setup AWS Session")
		return err
	}

	s3Service := s3.New(sess, &aws.Config{})
	uploader := s3manager.NewUploader(sess)

	forests, err := getMostRecentDataSet(s3Service, config.BucketName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Unable to get most recent data set")
		return err
	}

	for i, forest := range forests {
		// Figure out if a new SOPA Report has been relaesed
		// if so this will return the link to it's projects page
		hasNewData, newSopaReportLink, err := checkForUpdates(forest)
		if err != nil {
			log.WithFields(log.Fields{
				"forest": forest.Name,
				"state":  forest.State,
			}).Info("Error getting list of SOPA reports for forest")
			return err
		} else if !hasNewData {
			log.WithFields(log.Fields{
				"forest": forest.Name,
				"state":  forest.State,
			}).Info("Forest data up to date")
			continue
		}

		log.WithFields(log.Fields{
			"forest": forest.Name,
			"state":  forest.State,
			"link":   newSopaReportLink,
		}).Info("Found new SOPA Report for forest")

		// Get projects
		newProjects, err := getProjects(newSopaReportLink)
		if err != nil {
			log.WithFields(log.Fields{
				"forest": forest.Name,
				"state":  forest.State,
				"page":   newSopaReportLink,
				"error":  err.Error(),
			}).Error("Issue getting projects")
			continue
		}

		log.WithFields(log.Fields{
			"forest": forest.Name,
			"state":  forest.State,
			"count":  len(newProjects),
			"link":   newSopaReportLink,
		}).Info("Parsed new projects")

		// Get all the documents for that new project
		for j := range newProjects {
			if len(newProjects[j].Id) > 0 {
				log.WithFields(log.Fields{
					"forest":  forest.Name,
					"state":   forest.State,
					"project": newProjects[j].Name,
				}).Info("Getting documents")

				docs, err := getReportDocumentMeta(newProjects[j].Id)
				if err != nil {
					log.WithFields(log.Fields{
						"forest":  forest.Name,
						"state":   forest.State,
						"project": newProjects[j].Name,
					}).Error("Issue getting documents")
					continue
				}

				newProjects[j].ProjectDocuments = append(
					newProjects[j].ProjectDocuments,
					docs...,
				)

			}
		}

		// Add new projects to forest
		forests[i].Projects = append(newProjects, forests[i].Projects...)
	}

	// TODO(gasparovic): if new data, upsert airtable before uploading to S3
	data, err := json.Marshal(forests)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Issue marshaling forest data json")
		return err
	}

	_, err = uploader.Upload(&s3manager.UploadInput{
		Body: bytes.NewReader(data),
		Key: aws.String(fmt.Sprintf(
			"%s.json",
			time.Now().Format("2006-01-02"),
		)),
		Bucket: aws.String(config.BucketName),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Issue writing json data to S3")
		return err
	}

	return nil
}

func getMostRecentDataSet(s3Service *s3.S3, bucketName string) ([]Forest, error) {
	// TODO(gasparovic): list bucket elements and get the one with the greatest time stamp
	output, err := s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("projectsdb.json"),
	})
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	data := []Forest{}
	err = json.Unmarshal(bytes, &data)
	return data, err

}

func checkForUpdates(forest Forest) (bool, string, error) {
	// Get list of sopa reports from USFS
	pages, err := GetSopaReportPages(forest.Url)
	if err != nil {
		return false, "", err
	}

	// get most recent report from USFS website
	sopaReportDatesFromUsfs := []string{}
	for _, page := range pages {
		sopaReportDatesFromUsfs = append(sopaReportDatesFromUsfs, GetSopaReportDateFromURL(page))
	}
	sort.Strings(sopaReportDatesFromUsfs)
	mostRecentSopaReportFromUsfs := sopaReportDatesFromUsfs[len(sopaReportDatesFromUsfs)-1]

	// get most recent report from saved data
	sopaReportDatesFromSavedData := []string{}
	for _, project := range forest.Projects {
		sopaReportDatesFromSavedData = insert(sopaReportDatesFromSavedData, project.SopaReportDate)
	}
	sort.Strings(sopaReportDatesFromSavedData)
	mostRecentSopaReportFromSavedData := sopaReportDatesFromSavedData[len(sopaReportDatesFromSavedData)-1]

	if mostRecentSopaReportFromUsfs <= mostRecentSopaReportFromSavedData {
		return false, "", nil
	}

	return true, pages[len(pages)-1], nil
}

func insert(arr []string, elm string) []string {
	i := sort.SearchStrings(arr, elm)
	for _, str := range arr {
		if str == elm {
			return arr
		}
	}
	arr = append(arr, "")
	copy(arr[i+1:], arr[i:])
	arr[i] = elm
	return arr
}
