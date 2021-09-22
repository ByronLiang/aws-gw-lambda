package util

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ByronLiang/aws-gw-lambda/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var awsSession *session.Session
var awsSessionOnce sync.Once

// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, REGION 配置在环境变量里
func GetAwsSession() (*session.Session, error) {
	var err error
	region := config.ResizeImageLambdaConfig.Region
	awsSessionOnce.Do(func() {
		awsSession, err = session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: &region,
			},
		})
	})
	return awsSession, err
}

func DownloadFromS3WithBytes(bucket string, fullFileName string) (fileBytes []byte, fileSize int64, err error) {
	sess, err := GetAwsSession()
	if err != nil {
		err = errors.New("download data, create aws session error: " + err.Error())
		return
	}
	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	fileSize, err = downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fullFileName),
	})
	if err != nil {
		err = errors.New(fmt.Sprintf("download file from s3 error, bucket: %s, key: %s, region: %s, error: %s", bucket, fullFileName, getRegion(sess), err.Error()))
		return
	}
	fileBytes = buf.Bytes()
	return
}

func Upload2S3ByBytes(bucket string, fullFileName string, contentType string, fileBytes []byte) (err error) {
	sess, err := GetAwsSession()
	if err != nil {
		err = errors.New("upload data, create aws session error: " + err.Error())
		return
	}
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fullFileName),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		err = errors.New(fmt.Sprintf("upload file 2 s3 error, bucket: %s, key: %s, region: %s, error: %s", bucket, fullFileName, getRegion(sess), err.Error()))
	}
	return
}

func GetS3SignedURL(bucket string, fullFileName string) (signedUrl string, err error) {
	sess, err := GetAwsSession()
	if err != nil {
		err = errors.New("get signed url, create aws session error: " + err.Error())
		return
	}
	getreq, _ := s3.New(sess).GetObjectRequest(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    aws.String(fullFileName),
	})
	signedUrl, err = getreq.Presign(time.Second * 86400)
	if err != nil {
		err = errors.New(fmt.Sprintf("get signed url from s3 error, bucket: %s, key: %s, region: %s, error: %s", bucket, fullFileName, getRegion(sess), err.Error()))
	}
	return
}

func getRegion(sess *session.Session) string {
	region := ""
	if sess != nil && sess.Config != nil && sess.Config.Region != nil {
		region = *sess.Config.Region
	}
	return region
}
