package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type UploadDocumentsConfig struct {
	AccessKeyId     string `required help:"AWS Access Key ID" type:"string"`
	SecretAccessKey string `required help:"AWS Secret Access Key" type:"string"`
	RegionId        string `required help:"AWS Region ID" type:"string"`
	BucketName      string `required help:"S3 bucket that files will be uploaded to" type:"string"`
	ForestDataFile  string `required help:"Forest Data JSON File" type:"path"`
}

func UploadDocuments(config UploadDocumentsConfig) error {
	/*
		file, err := ioutil.ReadFile(config.ForestDataFile)
		if err != nil {
			log.WithFields(log.Fields{
				"file":  config.ForestDataFile,
				"error": err.Error(),
			}).Error("Unable to open forest data file")
			return err
		}

		forests := []Forest{}

		err = json.Unmarshal([]byte(file), &forests)
		if err != nil {
			log.WithFields(log.Fields{
				"file":  config.ForestDataFile,
				"error": err.Error(),
			}).Error("Unable to parse forest data file")
			return err
		}

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

		// TODO
		// for _, doc := range forests[i].Projects[j].ProjectDocuments {
		// 	err := uploadReport(doc, project.Id, uploader, s3Service)
		// 	if err != nil {
		// 		fmt.Printf("error uploading %s at %s for %s {%s}", doc.Name, doc.Url, forest.Name, err.Error())
		// 	}
		// }
	*/
	return nil
}

func uploadReport(
	bucketName string,
	doc ProjectDocument,
	projectId string,
	uploader *s3manager.Uploader,
	s3Service *s3.S3,
) error {
	key := aws.String(fmt.Sprintf(
		"%s/%s/%s.pdf",
		projectId,
		doc.Category,
		doc.Name,
	))
	bucket := aws.String(bucketName)

	output, err := s3Service.HeadObject(&s3.HeadObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			// There is no file in the bucket
		} else {
			return err
		}
	} else if *output.ContentLength > 0 {
		return nil
	}

	res := get(doc.Url)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   res.Body,
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		return err
	}

	return nil
}
