package handle

import (
	"errors"
	"log"
	"strconv"
	"strings"

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

func ImageResizeHandle(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	parameters := request.QueryStringParameters
	path, ok := parameters["path"]
	if !ok {
		log.Println("query parameter path error")
		return nil, parametersErr
	}
	paths := strings.Split(path, "/")
	if len(paths) != 2 {
		log.Printf("path split size error: %s", path)
		return nil, pathSizeErr
	}
	resizeConfig, err := parsePath(paths)
	// 解析失败, 返回原文件资源链接
	if err != nil {
		filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
		return SuccessAPIGatewayProxyResponse(filePath), nil
	}
	// 解析成功
	// 下载图片二进制数据
	_, originImageSize, err := util.DownloadFromS3WithBytes(
		config.ResizeImageLambdaConfig.Bucket,
		resizeConfig.FileName)
	if err != nil {
		// 文件名不存在
		log.Printf("download origin file error: %s", err.Error())
		filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
		return SuccessAPIGatewayProxyResponse(filePath), nil
	}
	log.Println("download image ", resizeConfig.FileName)
	log.Println("image size: ", originImageSize)
	// TODO: 图片裁剪流程
	// TODO: 图片裁剪结束 将处理后的资源上传回S3
	// 最终跳转到已成功处理的图片资源
	filePath := config.ResizeImageLambdaConfig.BucketUrl + "/" + resizeConfig.FileName
	return SuccessAPIGatewayProxyResponse(filePath), nil
	//return SuccessAPIGatewayProxyResponse("https://byronegg.s3.amazonaws.com/branches.png"), nil
}

func parsePath(pathList []string) (model.ResizeConfig, error) {
	resizeConfig := model.ResizeConfig{
		FileName: pathList[1], // 请求文件名
	}
	// TODO: 校验文件后缀
	size := pathList[0]
	// 解析长宽
	sizeGroup := strings.Split(size, "x")
	if len(sizeGroup) != 2 {
		log.Println("image size group query error ", size)
		return resizeConfig, sizeGroupErr
	}
	width := sizeGroup[0]
	height := sizeGroup[1]
	// 解析长度
	widthInt, err := strconv.Atoi(width)
	if err != nil {
		log.Println("image size width parse int error")
		return resizeConfig, sizeParseIntErr
	}
	resizeConfig.Width = widthInt
	// 解析宽度
	heightInt, err := strconv.Atoi(height)
	if err != nil {
		log.Println("image size height parse int error")
		return resizeConfig, sizeParseIntErr
	}
	resizeConfig.Height = heightInt
	return resizeConfig, nil
}
