package main

import (
	"github.com/ByronLiang/aws-gw-lambda/handle"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handle.SimpleGwHandle)
}
