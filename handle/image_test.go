package handle

import (
	"bytes"
	"image"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"os"
	"testing"

	"github.com/disintegration/imaging"

	"github.com/ByronLiang/aws-gw-lambda/config"

	"github.com/ByronLiang/aws-gw-lambda/util"
)

func TestResizeGif(t *testing.T) {
	fileBt, err := ioutil.ReadFile("cat.gif")
	if err != nil {
		t.Error(err)
		return
	}
	gifList, err := gif.DecodeAll(bytes.NewReader(fileBt))
	t.Log("gif pic count: ", len(gifList.Image))
	w := 100
	h := 100
	for i, gifData := range gifList.Image {
		newImg := imaging.Resize(gifData, w, h, imaging.Lanczos)
		gifList.Image[i] = image.NewPaletted(image.Rect(0, 0, w, h), gifList.Image[i].Palette)
		draw.Draw(gifList.Image[i], image.Rect(0, 0, w, h), newImg, image.Pt(0, 0), draw.Src)
	}
	outputFile, err := os.Create("cat2.gif")
	if err != nil {
		t.Error(err)
		return
	}
	defer outputFile.Close()

	err = gif.EncodeAll(outputFile, gifList)
	if err != nil {
		t.Error(err)
		return
	}
	// 返回 byte 数据
	//conf := &model.ResizeConfig{
	//	Width:      100,
	//	Height:     100,
	//	FileName:   "cat.gif",
	//	FileFormat: imaging.GIF,
	//	ImageByte:  fileBt,
	//}
	//resizeGifByte, err := util.GetResizeGif(conf)
	//t.Log("resizeGifByte len: ", len(resizeGifByte))
}

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
