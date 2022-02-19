package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

type CustomResEvent struct {
	RequestType        string `json:"requestType"`
	ServiceToken       string `json:"serviceToken"`
	ResponseURL        string `json:"responseURL"`
	LogicalResourceId  string `json:"logicalResourceId"`
	PhysicalResourceId string `json:"physicalResourceId"`
	ResourceType       string `json:"resourceType"`
	RequestId          string `json:"requestId"`
	StackId            string `json:"stackId"`
	ResourceProperties map[string]interface{}
}

type CustomResResponseData struct {
	MyNumber int
	MyName   string
}

type CustomResResponse struct {
	PhysicalResourceId string
	Data               CustomResResponseData
	NoEcho             bool
}

func HandleRequest(ctx context.Context, event CustomResEvent) (CustomResResponse, error) {
	fmt.Println("In Custom Function")
	fmt.Printf("%+v", event)
	log.Printf("hahahatest")

	buckName := fmt.Sprintf("%v", event.ResourceProperties["BuckName"])

	response := CustomResResponse{
		PhysicalResourceId: "aabbccdd",
		Data: CustomResResponseData{
			MyNumber: 987,
			MyName:   buckName,
		},
		NoEcho: false,
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
