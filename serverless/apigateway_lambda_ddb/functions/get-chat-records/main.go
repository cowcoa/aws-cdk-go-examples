package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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
	fmt.Printf("QueryStringParameters: %+v", request.QueryStringParameters)

	log.Printf("AWS_REGION: %s.\n", os.Getenv("AWS_REGION"))
	log.Printf("DYNAMODB_TABLE: %s.\n", os.Getenv("DYNAMODB_TABLE"))
	log.Printf("DYNAMODB_GSI: %s.\n", os.Getenv("DYNAMODB_GSI"))

	chatroom := request.QueryStringParameters["chatroom"]
	if len(chatroom) == 0 {
		return clientError(http.StatusBadRequest)
	}

	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return serverError(err)
	}

	// Query chat records from DDB GSI.
	ddb := dynamodb.NewFromConfig(cfg)
	queryResult, err := ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("DYNAMODB_TABLE")),
		IndexName:              aws.String(os.Getenv("DYNAMODB_GSI")),
		KeyConditionExpression: aws.String("chat_room = :chat_room"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":chat_room": &types.AttributeValueMemberS{
				Value: chatroom,
			},
		},
		Limit:            aws.Int32(10),
		ScanIndexForward: aws.Bool(false),
	})

	if err != nil {
		return serverError(err)
	}

	// Translate DDB Items to JSON string.
	var chatInfoArray []ChatInfo
	for _, item := range queryResult.Items {
		// Just for debug.
		nameAtt := item["name"].(*types.AttributeValueMemberS)
		timeAtt := item["time"].(*types.AttributeValueMemberS)
		commentAtt := item["comment"].(*types.AttributeValueMemberS)
		fmt.Printf("name: %s, time: %s, comment: %s.\n", nameAtt.Value, timeAtt.Value, commentAtt.Value)

		chatInfo := new(ChatInfo)
		attributevalue.UnmarshalMap(item, chatInfo)
		chatInfoArray = append(chatInfoArray, *chatInfo)
	}

	chatInfoArrayJson, err := json.Marshal(chatInfoArray)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(chatInfoArrayJson),
	}, nil
}

type ChatInfo struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Time    string `json:"time"`
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
