package model

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func FailRequestResponse(msg string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Body:       msg,
		StatusCode: http.StatusBadRequest,
	}
}

// 重定向 裁剪后图片地址
func SuccessRedirectResponse(data string) *events.APIGatewayProxyResponse {
	header := make(map[string]string)
	header["Location"] = data
	return &events.APIGatewayProxyResponse{
		Headers:    header,
		StatusCode: http.StatusMovedPermanently,
	}
}
