package util

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"net/url"
	"sync"
	"time"

	exifremove "github.com/scottleedavis/go-exif-remove"

	"github.com/ByronLiang/aws-gw-lambda/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const BackupBucketPrefix = "backup-"

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

func DownloadBatchFromS3WithBytes(sess *session.Session, bucket string, fullFileName string) (fileBytes []byte, err error) {
	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err = downloader.Download(buf, &s3.GetObjectInput{
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

func UploadBatch2S3ByBytes(sess *session.Session, bucket string, fullFileName string, fileBytes []byte) (err error) {
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fullFileName),
		Body:   bytes.NewReader(fileBytes),
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

// 列出bucket
func ListBuck(s3Obj *s3.S3) (map[string]struct{}, error) {
	output, err := s3Obj.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	nameSet := make(map[string]struct{})
	for _, buck := range output.Buckets {
		nameSet[*buck.Name] = struct{}{}
	}
	return nameSet, nil
}

// 创建备份bucket
func CreateBackupBucket(s3Obj *s3.S3, bucket string) error {
	backupBuckName := GetBackupBucketName(bucket)
	_, err := s3Obj.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(backupBuckName),
	})

	if err != nil {
		return err
	}

	err = s3Obj.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(backupBuckName),
	})

	if err != nil {
		return err
	}

	return nil
}

func CopyObjectToBackupBucket(s3Obj *s3.S3, targetBucket, fromBucket, item string) error {
	source := fmt.Sprintf("%s/%s", fromBucket, item)
	// Copy the item
	_, err := s3Obj.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(targetBucket),
		CopySource: aws.String(url.PathEscape(source)),
		Key:        aws.String(item),
	})

	if err != nil {
		return err
	}

	err = s3Obj.WaitUntilObjectExists(&s3.HeadObjectInput{
		Bucket: aws.String(targetBucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return err
	}

	return nil
}

func RefreshObject(sess *session.Session, bucket, path string) {
	backupBuckName := GetBackupBucketName(bucket)
	param := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}
	if path != "" {
		param.Prefix = aws.String(path)
	}
	s3Obj := s3.New(sess)
	// 获取 备份Bucket 的存量文件
	backupBucketObjectSet := GetAllObjectList(s3Obj, backupBuckName, path)
	s3Obj.ListObjectsPages(param, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, content := range output.Contents {
			objectName := *content.Key
			// 备份已有文件, 不进行压缩处理
			if _, ok := backupBucketObjectSet[objectName]; ok {
				log.Printf("object exist in backup bucket bucket: %s filename: %s", bucket, objectName)
				continue
			}
			fileFormat := FormatFromFilename(objectName)
			// 识别类型
			if fileFormat == JPEG || fileFormat == PNG {
				// 拉取源文件数据
				fileByte, err := DownloadBatchFromS3WithBytes(sess, bucket, objectName)
				if err != nil {
					log.Println("download file from s3 error", objectName)
					continue
				}
				// 对文件进行压缩等相关处理
				noExifImageBytes, err := exifremove.Remove(fileByte)
				if err != nil {
					log.Println("remove exif image error", objectName)
					continue
				}
				imageBuffer := bytes.NewBuffer(noExifImageBytes)
				img, _, err := image.Decode(imageBuffer)
				if err != nil {
					log.Printf("image decode error: bucket: %s filename: %s", bucket, objectName)
					continue
				}
				compressImageByte, err := GetImageCompressByte(img, fileFormat, 50)
				if err != nil {
					log.Printf("image compress error: bucket: %s filename: %s", bucket, objectName)
					continue
				}
				// 备份成功
				// copy object 到 备份 bucket
				err = CopyObjectToBackupBucket(s3Obj, backupBuckName, bucket, objectName)
				if err != nil {
					log.Printf("image CopyObjectToBackupBucket error: bucket: %s filename: %s", bucket, objectName)
					continue
				}
				// 将处理后的文件进行原路上传
				err = UploadBatch2S3ByBytes(sess, bucket, objectName, compressImageByte)
				if err != nil {
					log.Printf("image UploadBatch2S3ByBytes error: bucket: %s filename: %s", bucket, objectName)
					continue
				}
			}
		}
		return b
	})
}

func GetBackupBucketName(bucket string) string {
	return fmt.Sprintf("%s%s", BackupBucketPrefix, bucket)
}

func GetAllObjectList(s3Obj *s3.S3, bucket, path string) map[string]struct{} {
	param := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}
	// 指定目录
	if path != "" {
		param.Prefix = aws.String(path)
	}
	objectSet := make(map[string]struct{})
	s3Obj.ListObjectsPages(param, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, content := range output.Contents {
			objectSet[*content.Key] = struct{}{}
		}
		return b
	})
	return objectSet
}
