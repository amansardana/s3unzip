package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type Request events.S3Event

// UnzipRequest input request
type UnzipRequest struct {
	DownloadBucket string `json:"downloadBucket"`
	UploadBucket   string `json:"uploadBucket"`
	Item           string `json:"item"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, req Request) {

	x, _ := json.MarshalIndent(req, "", "  ")
	fmt.Print("\n",string(x),"\n")

	unzipLambdaName := "unzip-test-dev-unzip"

	for _, record := range req.Records {
		r := &UnzipRequest{
			UploadBucket:   "unzip-files.amansardana.com",
			DownloadBucket: record.S3.Bucket.Name,
			Item:           record.S3.Object.Key,
		}
		rString, _ := json.Marshal(r)
		invokeLambda(unzipLambdaName, string(rString))
	}
}

func main() {
	lambda.Start(Handler)
}
