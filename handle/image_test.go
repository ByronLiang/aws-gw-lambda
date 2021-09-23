package handle

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/disintegration/imaging"

	"github.com/ByronLiang/aws-gw-lambda/config"

	"github.com/ByronLiang/aws-gw-lambda/util"
)

func TestResizeImage(t *testing.T) {
	fileBt, err := ioutil.ReadFile("branches.png")
	if err != nil {
		t.Error(err)
		return
	}
	imageObj, err := imaging.Decode(bytes.NewReader(fileBt))
	// 针对参数为0是自适应裁剪
	newImageObj := imaging.Resize(imageObj, 150, 0, imaging.Lanczos)
	err = imaging.Save(newImageObj, "./newImage.png")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestImageResizeHandle(t *testing.T) {
	basePath := "130x130/branches.png"
	conf, _, err := parsePath(basePath)
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
	resizedImageByte, err := util.GetResizedImage(conf)
	if err != nil {
		t.Log("resize image error")
		t.Error(err)
		return
	}
	t.Log(len(resizedImageByte))
	fileFullPath := config.ResizeImageLambdaConfig.PathPrefix + basePath
	err = util.Upload2S3ByBytes(
		config.ResizeImageLambdaConfig.Bucket,
		fileFullPath,
		fileFormatContentType[conf.FileFormat],
		resizedImageByte)
	if err != nil {
		t.Log("upload s3 error")
		t.Error(err)
	}
	url := config.ResizeImageLambdaConfig.BucketUrl + "/" + fileFullPath
	t.Log("resource url: ", url)
}
