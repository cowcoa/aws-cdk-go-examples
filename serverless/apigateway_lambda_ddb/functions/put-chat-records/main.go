package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
)

/*
type APIGatewayProxyRequest struct {
    Resource              string                        `json:"resource"` // The resource path defined in API Gateway
    Path                  string                        `json:"path"`     // The url path for the caller
    HTTPMethod            string                        `json:"httpMethod"`
    Headers               map[string]string             `json:"headers"`
    QueryStringParameters map[string]string             `json:"queryStringParameters"`
    PathParameters        map[string]string             `json:"pathParameters"`
    StageVariables        map[string]string             `json:"stageVariables"`
    RequestContext        APIGatewayProxyRequestContext `json:"requestContext"`
    Body                  string                        `json:"body"`
    IsBase64Encoded       bool                          `json:"isBase64Encoded,omitempty"`
}
type APIGatewayProxyResponse struct {
    StatusCode      int               `json:"statusCode"`
    Headers         map[string]string `json:"headers"`
    Body            string            `json:"body"`
    IsBase64Encoded bool              `json:"isBase64Encoded,omitempty"`
}
*/
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Body size = %d.\n", len(request.Body))
	log.Printf("Body string = %s.\n", request.Body)

	log.Printf("AWS_REGION: %s.\n", os.Getenv("AWS_REGION"))
	log.Printf("DYNAMODB_TABLE: %s.\n", os.Getenv("DYNAMODB_TABLE"))

	chatInfo, err := parseBodyStringToTypedObject(request.Body)
	if err != nil {
		return clientError(http.StatusBadRequest)
	}

	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return serverError(err)
	}

	// Get current time in UTC.
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)

	// Put chat records to DDB table.
	ddb := dynamodb.NewFromConfig(cfg)
	_, putErr := ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Item: map[string]types.AttributeValue{
			"name": &types.AttributeValueMemberS{
				Value: chatInfo.Name,
			},
			"time": &types.AttributeValueMemberS{
				Value: strconv.FormatInt(now.Unix(), 10),
			},
			"comment": &types.AttributeValueMemberS{
				Value: chatInfo.Comment,
			},
			"chat_room": &types.AttributeValueMemberS{
				Value: chatInfo.ChatRoom,
			},
		},
	})

	if putErr != nil {
		return serverError(putErr)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
	}, nil
}

type ChatInfo struct {
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	ChatRoom string `json:"chatRoom"`
}

func parseBodyStringToTypedObject(body string) (ChatInfo, error) {

	b := []byte(body)
	var chatInfo ChatInfo
	err := json.Unmarshal(b, &chatInfo)

	return chatInfo, err
}

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func main() {
	runtime.Start(handleRequest)
}
