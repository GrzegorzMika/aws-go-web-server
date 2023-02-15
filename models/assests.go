package models

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
)

type S3Bucket struct {
	session *session.Session
	svc     *s3.S3
}

func (sb S3Bucket) ListS3Content(bucket string) ([]string, error) {
	var fileNames []string

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(100),
	}

	result, err := sb.svc.ListObjectsV2(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return fileNames, err
	}

	for _, obj := range result.Contents {
		fileNames = append(fileNames, *obj.Key)
	}
	return fileNames, nil
}

func (sb S3Bucket) DownloadS3File(bucket, key string) error {
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sb.session)

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(key)
	if err != nil {
		return fmt.Errorf("failed to create file %q, %v", key, err)
	}

	// Write the contents of S3 Object to the file
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}
	log.Printf("file downloaded, %d bytes\n", n)
	return nil
}

func (sb S3Bucket) RefreshAssets() error {
	return nil
}
