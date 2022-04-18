package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "MyStack"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, change ECR repository name by 'cdk.json/context/imageRepoName'.
func EcrRepoName(scope constructs.Construct) string {
	ecrRepoName := "MyRepository"

	ctxValue := scope.Node().TryGetContext(jsii.String("imageRepoName"))
	if v, ok := ctxValue.(string); ok {
		ecrRepoName = v
	}

	return ecrRepoName
}
