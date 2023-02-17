package models

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"time"
)

type S3Bucket struct {
	session *session.Session
	svc     *s3.S3
}

func NewS3Bucket(region string) (*S3Bucket, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}

	return &S3Bucket{
		session: sess,
		svc:     s3.New(sess),
	}, nil
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

func (sb S3Bucket) GetURL(bucket, key string) (string, error) {
	req, _ := sb.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(SessionTimeout * time.Second)

	if err != nil {
		log.Println("Failed to sign request", err)
		return "", err
	}

	return urlStr, nil
}
