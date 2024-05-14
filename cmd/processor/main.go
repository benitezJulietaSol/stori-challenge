package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"stori-challenge/cmd/processor/handler"
)

func main() {
	lambda.Start(handler.LambdaEvent)
}
