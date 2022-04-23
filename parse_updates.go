package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

/*
 * Parse just the updates
 */
type ParseUpdatesConfig struct {
	AccessKeyId     string `required help:"AWS Access Key ID" type:"string"`
	SecretAccessKey string `required help:"AWS Secret Access Key" type:"string"`
	RegionId        string `required help:"AWS Region ID" type:"string"`
	BucketName      string `required help:"S3 bucket that files will be uploaded to" type:"string"`
	SlackHookUrl    string `required`
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
	// uploader := s3manager.NewUploader(sess)

	forests, err := getMostRecentDataSet(s3Service, config.BucketName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Unable to get most recent data set")
		return err
	}

	for _, forest := range forests {
		hasNewData, err := checkForUpdates(forest)
		if err != nil {
			return err
		} else if !hasNewData {
			continue
		}
		http.Post(
			config.SlackHookUrl,
			"Content-type: application/json",
			strings.NewReader("{'text':'There is a new SOPA Report!'}"),
		)
		break

		// forests[i] = GetAllForestData(forest.Url) // TODO(gasparovic)
	}

	// TODO(gasparovic): if new data, upsert airtable & write blob to S3 to commit the transaction

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

func checkForUpdates(forest Forest) (bool, error) {
	// Get list of sopa reports from USFS
	pages, err := GetSopaReportPages(forest.Url)
	if err != nil {
		return false, err
	}

	sopaReportDatesFromUSFS := []string{}
	for _, page := range pages {
		sopaReportDatesFromUSFS = append(sopaReportDatesFromUSFS, GetSopaReportDateFromURL(page))
	}

	sopaReportDatesFromForest := []string{}
	for _, project := range forest.Projects {
		sopaReportDatesFromForest = insert(sopaReportDatesFromForest, project.SopaReportDate)
	}

	if len(sopaReportDatesFromUSFS) != len(sopaReportDatesFromForest) {
		return true, nil
	}

	for i := range sopaReportDatesFromForest {
		if sopaReportDatesFromUSFS[i] != sopaReportDatesFromForest[i] {
			return true, nil
		}
	}

	return false, nil
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
