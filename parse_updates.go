package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	anyUpdates := false
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
		anyUpdates = hasNewData || anyUpdates

		// if error, log, but do not fail
		message := fmt.Sprintf(
			"Found new <%s|SOPA Report> for *%s* forest in *%s*",
			newSopaReportLink,
			forest.Name,
			forest.State,
		)
		err = sendSlackUpdate(config.SlackHookUrl, message)
		if err != nil {
			log.WithFields(log.Fields{
				"message": message,
				"link":    config.SlackHookUrl,
			}).Error("Issue posting update to slack")
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
				uploadProjectToAirtable(forest, newProjects)

			}
		}

		// Add new projects to forest
		forests[i].Projects = append(newProjects, forests[i].Projects...)
	}

	if !anyUpdates {
		message := "No new SOPA Reports found"
		err = sendSlackUpdate(config.SlackHookUrl, message)
		if err != nil {
			log.WithFields(log.Fields{
				"message": message,
				"link":    config.SlackHookUrl,
			}).Error("Issue posting update to slack")
			return err
		}
		return nil
	}

	data, err := json.Marshal(forests)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Issue marshaling forest data json")
		return err
	}

	file := fmt.Sprintf(
		"%s.json",
		time.Now().Format("2006-01-02"),
	)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   bytes.NewReader(data),
		Key:    aws.String(file),
		Bucket: aws.String(config.BucketName),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Issue writing new forest json data to S3")
		return err
	}

	log.WithFields(log.Fields{
		"file": file,
	}).Info("File with new forest data successfully written to S3")

	return nil
}

func getMostRecentDataSet(s3Service *s3.S3, bucketName string) ([]Forest, error) {
	// TODO(gasparovic): list bucket elements and get the one with the greatest time stamp
	listObjectsOutput, err := s3Service.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucketName,
		Prefix: aws.String(""),
	})
	if err != nil {
		return nil, err
	}

	mostRecentFile := *listObjectsOutput.Contents[0].Key
	for _, content := range listObjectsOutput.Contents[1:] {
		if mostRecentFile < *content.Key {
			mostRecentFile = *content.Key
		}

	}

	log.WithFields(log.Fields{
		"file":   mostRecentFile,
		"bucket": bucketName,
	}).Info("Found most recent forest data file")

	getObjectOutput, err := s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(mostRecentFile),
	})
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(getObjectOutput.Body)
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

	return true, pages[0], nil
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

func uploadProjectToAirtable(forest Forest, newProjects []ProjectUpdate) {
	// TODO(hank)
	// bundle into format ready for airtable?

	// send to airtable?

	// return error code if failed?

}

type SlackMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func sendSlackUpdate(slackHookUrl string, message string) error {
	data, err := json.Marshal(SlackMessage{
		Type: "mrkdwn",
		Text: message,
	})
	if err != nil {
		return err
	}

	_, err = http.Post(
		slackHookUrl,
		"Content-type: application/json",
		bytes.NewReader(data),
	)

	return err
}
