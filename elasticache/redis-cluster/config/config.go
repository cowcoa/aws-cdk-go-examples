package config

import (
	"strconv"

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

// DO NOT modify this function, change stack name by 'cdk.json/context/password'.
// Password constraints:
// Must be only printable ASCII characters.
// Must be at least 16 characters and no more than 128 characters in length.
// Nonalphanumeric characters are restricted to (!, &, #, $, ^, <, >, -, ).
func Password(scope constructs.Construct) string {
	password := ""

	ctxValue := scope.Node().TryGetContext(jsii.String("password"))
	if v, ok := ctxValue.(string); ok {
		password = v
	}

	return password
}

// DO NOT modify this function, change stack name by 'cdk.json/context/port'.
func Port(scope constructs.Construct) float64 {
	password := float64(6379)

	ctxValue := scope.Node().TryGetContext(jsii.String("port"))
	if v, ok := ctxValue.(float64); ok {
		password = v
	}

	return password
}

// Deployment stage config
type DeploymentStageType string

const (
	DeploymentStage_DEV  DeploymentStageType = "DEV"
	DeploymentStage_PROD DeploymentStageType = "PROD"
)

// DO NOT modify this function, change EKS cluster name by 'cdk-cli-wrapper-dev.sh/--context deploymentStage='.
func DeploymentStage(scope constructs.Construct) DeploymentStageType {
	deploymentStage := DeploymentStage_PROD

	ctxValue := scope.Node().TryGetContext(jsii.String("deploymentStage"))
	if v, ok := ctxValue.(string); ok {
		deploymentStage = DeploymentStageType(v)
	}

	return deploymentStage
}

const (
	vpcMask    = 16
	vpcIpv4    = "172.33.0.0"
	MaxAzs     = 2
	SubnetMask = vpcMask + MaxAzs
)

var VpcCidr = vpcIpv4 + "/" + strconv.Itoa(vpcMask)
