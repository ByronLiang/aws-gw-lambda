package handle

import (
	"strings"
	"testing"

	"github.com/ByronLiang/aws-gw-lambda/config"

	"github.com/ByronLiang/aws-gw-lambda/util"
)

func TestImageResizeHandle(t *testing.T) {
	basePath := "130x130/branches.png"
	paths := strings.Split(basePath, "/")
	conf, err := parsePath(paths)
	if err != nil {
		t.Log("parse path error")
		t.Error(err)
		return
	}
	sourceBytes, _, err := util.DownloadFromS3WithBytes(config.ResizeImageLambdaConfig.Bucket, conf.FileName)
	if err != nil {
		t.Log("download file from s3 error")
		t.Error(err)
		return
	}
	conf.ImageByte = sourceBytes
	resizedImageByte, err := util.GetResizedImage(&conf)
	if err != nil {
		t.Log("resize image error")
		t.Error(err)
		return
	}
	t.Log(len(resizedImageByte))
	fileFullPath := config.ResizeImageLambdaConfig.PathPrefix + basePath
	err = util.Upload2S3ByBytes(config.ResizeImageLambdaConfig.Bucket, fileFullPath, resizedImageByte)
	if err != nil {
		t.Log("upload s3 error")
		t.Error(err)
	}
	url := config.ResizeImageLambdaConfig.BucketUrl + "/" + fileFullPath
	t.Log("resource url: ", url)
}
