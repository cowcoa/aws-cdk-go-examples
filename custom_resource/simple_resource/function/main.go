package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
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

type CustomResResponse struct {
	PhysicalResourceId string
	Data               map[string]interface{}
	NoEcho             bool
}

// SSMPutParameterAPI defines the interface for the PutParameter function.
// We use this interface to test the function using a mocked service.
type SSMPutParameterAPI interface {
	PutParameter(ctx context.Context,
		params *ssm.PutParameterInput,
		optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
}

// AddStringParameter creates an AWS Systems Manager string parameter
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a PutParameterOutput object containing the result of the service call and nil
//     Otherwise, nil and an error from the call to PutParameter
func AddStringParameter(c context.Context, api SSMPutParameterAPI, input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	return api.PutParameter(c, input)
}

/*
func OnCreate(event CustomResEvent) (CustomResResponse, error) {
	physicalResId := fmt.Sprintf("%v", event.ResourceProperties["PhysicalResourceId"])
	parameterName := fmt.Sprintf("%v", event.ResourceProperties["SSMParamName"])
	parameterValue := fmt.Sprintf("%v", event.ResourceProperties["SSMParamValue"])

	client := ssm.NewFromConfig(cfg)
	input := &ssm.PutParameterInput{
		Name:      &parameterName,
		Value:     &parameterValue,
		Type:      types.ParameterTypeString,
		Overwrite: true,
	}

	results, err := AddStringParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return response, err
	}

	fmt.Println("Parameter version:", results.Version)

	//buckName := fmt.Sprintf("%v", event.ResourceProperties["BuckName"])

	response = CustomResResponse{
		PhysicalResourceId: physicalResId,
		Data: map[string]interface{}{
			"SSMParamName":  parameterName,
			"SSMParamValue": parameterValue,
		},
		NoEcho: false,
	}

	return response, nil
}
*/

func HandleRequest(ctx context.Context, event CustomResEvent) (CustomResResponse, error) {
	var response CustomResResponse

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return response, err
	}

	physicalResId := fmt.Sprintf("%v", event.ResourceProperties["PhysicalResourceId"])
	parameterName := fmt.Sprintf("%v", event.ResourceProperties["SSMParamName"])
	parameterValue := fmt.Sprintf("%v", event.ResourceProperties["SSMParamValue"])

	client := ssm.NewFromConfig(cfg)
	input := &ssm.PutParameterInput{
		Name:      &parameterName,
		Value:     &parameterValue,
		Type:      types.ParameterTypeString,
		Overwrite: true,
	}

	results, err := AddStringParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return response, err
	}

	fmt.Println("Parameter version:", results.Version)

	//buckName := fmt.Sprintf("%v", event.ResourceProperties["BuckName"])

	response = CustomResResponse{
		PhysicalResourceId: physicalResId,
		Data: map[string]interface{}{
			"SSMParamName":  parameterName,
			"SSMParamValue": parameterValue,
		},
		NoEcho: false,
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
	/*
		var ctx context.Context
		var event CustomResEvent
		event.ResourceProperties = map[string]interface{}{
			"PhysicalResourceId": "CowPhyId",
			"SSMParamName":       "LocalRunParaName",
			"SSMParamValue":      "LocalRunParaValue",
		}
		HandleRequest(ctx, event)
	*/
}
