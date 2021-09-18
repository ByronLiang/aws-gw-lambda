package handle

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

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
	if path, ok := parameters["path"]; ok {
		resizeConfig := &model.ResizeConfig{Bucket: os.Getenv("BUCKET")}
		paths := strings.Split(path, "/")
		if len(paths) == 2 {
			size := paths[0]
			// 请求文件名
			resizeConfig.FileName = paths[1]
			// 解析长宽
			sizeGroup := strings.Split(size, "x")
			if len(sizeGroup) == 2 {
				width := sizeGroup[0]
				height := sizeGroup[1]
				// 解析长度
				if widthInt, err := strconv.Atoi(width); err != nil {
					log.Println("image size width parse int error")
					return FailAPIGatewayProxyResponse("image size width parse int error"), sizeParseIntErr
				} else {
					resizeConfig.Width = widthInt
				}
				// 解析宽度
				if heightInt, err := strconv.Atoi(height); err != nil {
					log.Println("image size height parse int error")
					return FailAPIGatewayProxyResponse("image size height parse int error"), sizeParseIntErr
				} else {
					resizeConfig.Height = heightInt
				}
				// 下载图片二进制数据
				_, originImageSize, err := util.DownloadFromS3WithBytes(resizeConfig.Bucket, resizeConfig.FileName)
				if err != nil {
					// 文件名不存在
					log.Printf("download origin file error: %s", err.Error())
					return FailAPIGatewayProxyResponse("filename error"), fileNoExistErr
				}
				log.Println("download image ", resizeConfig.FileName)
				log.Println("image size: ", originImageSize)
				// TODO: 图片裁剪流程
				// TODO: 图片裁剪结束 将处理后的资源上传回S3
				// 最终跳转到已成功处理的图片资源
				return SuccessAPIGatewayProxyResponse("https://byronegg.s3.amazonaws.com/branches.png"), nil
			}
			log.Println("image size group query error ", size)
			return FailAPIGatewayProxyResponse("image size query error"), sizeGroupErr
		}
		log.Println("path split size error ", path)
		return FailAPIGatewayProxyResponse("path size error"), pathSizeErr
	}
	log.Println("path data error ")
	return FailAPIGatewayProxyResponse("parameters error"), parametersErr
}
