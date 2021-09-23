package handle

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/ByronLiang/aws-gw-lambda/config"

	"github.com/ByronLiang/aws-gw-lambda/util"

	"github.com/ByronLiang/aws-gw-lambda/model"

	"github.com/aws/aws-lambda-go/events"
)

var (
	parametersErr   = errors.New("parameters error")
	pathSizeErr     = errors.New("path size error")
	sizeGroupErr    = errors.New("image size query error")
	sizeParseIntErr = errors.New("image size parse int error")
	fileNoExistErr  = errors.New("fileName No Exist error")
)

var fileFormatContentType = map[imaging.Format]string{
	imaging.JPEG: "image/jpeg",
	imaging.PNG:  "image/png",
	imaging.GIF:  "image/gif",
	imaging.BMP:  "image/bmp",
	imaging.TIFF: "image/tiff",
}

func ImageResizeHandle(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	parameters := request.QueryStringParameters
	path, ok := parameters["path"]
	if !ok {
		log.Println("query parameter path error")
		return nil, parametersErr
	}
	resizeConfig, isRedirectSource, err := parsePath(path)
	// 解析失败, 返回原文件资源链接
	if err != nil {
		if isRedirectSource {
			filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
			return model.SuccessRedirectResponse(filePath), nil
		} else {
			return nil, err
		}
	}
	// 解析成功
	// 下载图片二进制数据
	originImageByte, _, err := util.DownloadFromS3WithBytes(
		config.ResizeImageLambdaConfig.Bucket,
		resizeConfig.FileName)
	if err != nil {
		// 文件名不存在
		log.Printf("download origin file error: %s", err.Error())
		filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
		return model.SuccessRedirectResponse(filePath), nil
	}
	resizeConfig.ImageByte = originImageByte
	log.Printf("start to resize image: %s", resizeConfig.FileName)
	// 图片裁剪流程
	resizedImageByte, err := util.GetResizedImage(resizeConfig)
	if err != nil {
		log.Printf("get resized image error: %s", err.Error())
		filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
		return model.SuccessRedirectResponse(filePath), nil
	}
	// 图片裁剪结束 将处理后的资源上传回S3
	fileFullPath := config.ResizeImageLambdaConfig.PathPrefix + path
	err = util.Upload2S3ByBytes(
		config.ResizeImageLambdaConfig.Bucket,
		fileFullPath,
		fileFormatContentType[resizeConfig.FileFormat],
		resizedImageByte)
	if err != nil {
		log.Printf("upload resize image to s3 error: %s", err.Error())
		filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
		return model.SuccessRedirectResponse(filePath), nil
	}
	// 最终跳转到已成功处理的图片资源
	filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + fileFullPath
	return model.SuccessRedirectResponse(filePath), nil
}

func parsePath(path string) (*model.ResizeConfig, bool, error) {
	pathList := strings.Split(path, "/")
	if len(pathList) < 2 {
		// 路径异常, 直接响应异常
		log.Printf("path split size error: %s", path)
		return nil, false, pathSizeErr
	}
	// 针对 100x100/d1/d2/pic.png 兼容多级文件目录
	size := pathList[0]
	// 提取文件名
	fileNameList := make([]string, len(pathList)-1)
	copy(fileNameList, pathList[1:])
	fileName := strings.Join(fileNameList, "/")
	resizeConfig := &model.ResizeConfig{
		FileName: fileName, // 请求文件名
	}
	// 校验文件后缀
	fileFormat, err := imaging.FormatFromFilename(resizeConfig.FileName)
	if err != nil {
		log.Printf("file format unsupport error: %s", err.Error())
		return resizeConfig, true, err
	}
	resizeConfig.FileFormat = fileFormat
	// 解析长宽
	sizeGroup := strings.Split(size, "x")
	if len(sizeGroup) != 2 {
		log.Println("image size group query error ", size)
		return resizeConfig, true, sizeGroupErr
	}
	width := sizeGroup[0]
	height := sizeGroup[1]
	// 解析长度
	widthInt, err := strconv.Atoi(width)
	if err != nil {
		log.Println("image size width parse int error")
		return resizeConfig, true, sizeParseIntErr
	}
	resizeConfig.Width = widthInt
	// 解析宽度
	heightInt, err := strconv.Atoi(height)
	if err != nil {
		log.Println("image size height parse int error")
		return resizeConfig, true, sizeParseIntErr
	}
	resizeConfig.Height = heightInt
	return resizeConfig, false, nil
}
