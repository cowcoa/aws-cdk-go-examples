package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
)

/*
type MyEvent struct {
	Name string `json:"name"`
}

type Book struct {
	ISBN   string `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//return fmt.Sprintf("Hello, World: %s!", name.Name), nil
	// Fetch a specific book record from the DynamoDB database. We'll
	// make this more dynamic in the next section.
	log.Printf("APIGatewayProxyRequest Event: %+v", req)

	bk := &Book{
		ISBN:   "978-1420931693",
		Title:  "The Republic",
		Author: "Plato",
	}

	js, err := json.Marshal(bk)
	if err != nil {
		//return serverError(err)
		fmt.Println(err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}
*/

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
	fmt.Printf("Processing request data for request %s.\n", request.RequestContext.RequestID)
	fmt.Printf("QueryStringParameters: %+v", request.QueryStringParameters)

	chatroom := request.QueryStringParameters["chatroom"]
	if len(chatroom) == 0 {
		return clientError(http.StatusBadRequest)
	}

	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		fmt.Printf("LoadDefaultConfig error: %s", err.Error())
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)

	now.Unix()

	// Create SSM client.
	ddb := dynamodb.NewFromConfig(cfg)

	queryResult, err := ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("ChatTable"),
		IndexName:              aws.String("ChatTableGSI"),
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

	var chatInfoArray []ChatInfo
	for _, item := range queryResult.Items {
		nameAtt := item["name"].(*types.AttributeValueMemberS)
		timeAtt := item["time"].(*types.AttributeValueMemberS)
		commentAtt := item["comment"].(*types.AttributeValueMemberS)
		fmt.Printf("name: %s, time: %s, comment: %s.\n", nameAtt.Value, timeAtt.Value, commentAtt.Value)

		/*
			chatInfo := ChatInfo{
				Name:    nameAtt.Value,
				Time:    timeAtt.Value,
				Comment: commentAtt.Value,
			}
		*/
		chatInfo := new(ChatInfo)

		attributevalue.UnmarshalMap(item, chatInfo)

		chatInfoArray = append(chatInfoArray, *chatInfo)

	}

	jsobj, err := json.Marshal(chatInfoArray)
	if err != nil {
		return serverError(err)
	}
	fmt.Println("output json object")
	log.Println(string(jsobj))

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsobj),
	}, nil
}
