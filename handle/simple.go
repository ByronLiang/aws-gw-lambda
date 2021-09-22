package handle

import (
	"context"
	"log"

	"github.com/ByronLiang/aws-gw-lambda/model"

	"github.com/aws/aws-lambda-go/events"
)

func SimpleGwHandle(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	parameters := request.QueryStringParameters
	if path, ok := parameters["path"]; ok {
		log.Println("path: ", path)
		return model.SuccessRedirectResponse("https://byronegg.s3.amazonaws.com/branches.png"), nil
	}
	return model.FailRequestResponse("parameters error"), nil
}
