package main

import (
	"bytes"
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

type Request events.APIGatewayProxyRequest

// UnzipRequest input request
type UnzipRequest struct {
	DownloadBucket string `json:"downloadBucket"`
	UploadBucket   string `json:"uploadBucket"`
	Item           string `json:"item"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, req Request) (Response, error) {
	var buf bytes.Buffer

	var r UnzipRequest
	fmt.Println(req.Body)
	json.Unmarshal([]byte(req.Body), &r)
	fmt.Println(r)

	if err := s3Unzip(r.DownloadBucket, r.UploadBucket, r.Item); err != nil {
		body, err := json.Marshal(map[string]interface{}{
			"message": err.Error(),
		})
		if err != nil {
			return Response{StatusCode: 404}, err
		}
		json.HTMLEscape(&buf, body)
		resp := Response{
			StatusCode:      400,
			IsBase64Encoded: false,
			Body:            buf.String(),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		return resp, nil
	}

	body, err := json.Marshal(map[string]interface{}{
		"message": "Uploaded Successfully!",
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
