package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

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

// SSMDeleteParameterAPI defines the interface for the DeleteParameter function.
// We use this interface to test the function using a mocked service.
type SSMDeleteParameterAPI interface {
	DeleteParameter(ctx context.Context,
		params *ssm.DeleteParameterInput,
		optFns ...func(*ssm.Options)) (*ssm.DeleteParameterOutput, error)
}

// RemoveParameter deletes an AWS Systems Manager string parameter.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a DeleteParameterOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to DeleteParameter.
func RemoveParameter(c context.Context, api SSMDeleteParameterAPI, input *ssm.DeleteParameterInput) (*ssm.DeleteParameterOutput, error) {
	return api.DeleteParameter(c, input)
}

type CustomResEvent struct {
	RequestType           string
	ServiceToken          string
	ResponseURL           string
	LogicalResourceId     string
	PhysicalResourceId    string
	ResourceType          string
	RequestId             string
	StackId               string
	ResourceProperties    map[string]interface{}
	OldResourceProperties map[string]interface{}
}

type CustomResResponse struct {
	PhysicalResourceId string
	Data               map[string]interface{}
	NoEcho             bool
}

func OnCreate(client *ssm.Client, ssmParamName string, ssmParamValue string, overwrite bool) error {
	input := &ssm.PutParameterInput{
		Name:      &ssmParamName,
		Value:     &ssmParamValue,
		Type:      types.ParameterTypeString,
		Overwrite: overwrite,
	}
	results, err := AddStringParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("Parameter version: ", results.Version)

	return nil
}

func OnDelete(client *ssm.Client, ssmParamName string) error {
	input := &ssm.DeleteParameterInput{
		Name: &ssmParamName,
	}
	_, err := RemoveParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("Deleted parameter: " + ssmParamName)

	return nil
}

func HandleRequest(ctx context.Context, event CustomResEvent) (CustomResResponse, error) {
	var response CustomResResponse

	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return response, err
	}

	// Create SSM client.
	client := ssm.NewFromConfig(cfg)

	// Extract input parameters.
	physicalResId := fmt.Sprintf("%v", event.ResourceProperties["PhysicalResourceId"])
	ssmParamName := fmt.Sprintf("%v", event.ResourceProperties["SSMParamName"])
	ssmParamValue := fmt.Sprintf("%v", event.ResourceProperties["SSMParamValue"])
	ssmParamNameOld := fmt.Sprintf("%v", event.OldResourceProperties["SSMParamName"])

	switch event.RequestType {
	case "Create":
		fmt.Printf("OnCreate, PhysicalResourceId: %s, SSMParamName: %s, SSMParamValue: %s\n", physicalResId, ssmParamName, ssmParamValue)
		OnCreate(client, ssmParamName, ssmParamValue, false)
	case "Update":
		fmt.Printf("OnUpdate, PhysicalResourceId: %s, SSMParamName: %s, SSMParamValue: %s\n", physicalResId, ssmParamName, ssmParamValue)
		if ssmParamName == ssmParamNameOld {
			OnCreate(client, ssmParamName, ssmParamValue, true)
		} else {
			OnDelete(client, ssmParamNameOld)
			OnCreate(client, ssmParamName, ssmParamValue, false)
		}
	case "Delete":
		fmt.Printf("OnDelete, PhysicalResourceId: %s, SSMParamName: %s, SSMParamValue: %s\n", physicalResId, ssmParamName, ssmParamValue)
		OnDelete(client, ssmParamName)
	}

	response = CustomResResponse{
		PhysicalResourceId: physicalResId,
		Data: map[string]interface{}{
			"SSMParamName":  ssmParamName,
			"SSMParamValue": ssmParamValue,
		},
		NoEcho: false,
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
