package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	// Lambda function config
	RoleName    = "CRLambdaRole"
	FuncionName = "ApigLambdaFunction"
	MemorySize  = 128
	MaxDuration = 60
	CodePath    = "function/."
	Handler     = "main"
	// Provider function config
	ProviderName = "CRProvider"
	// Custom resource config
	ResourceName = "SSMParamCustomRes"
	ResourceType = "Custom::SSMParamCustomRes"
	// SSM Config
	PhysicalIdKey    = "PhysicalResourceId"
	SsmParamNameKey  = "SSMParamName"
	SsmParamValueKey = "SSMParamValue"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "ApigLambdaDdb"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}
