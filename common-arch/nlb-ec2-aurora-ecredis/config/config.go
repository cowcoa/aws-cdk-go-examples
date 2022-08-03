package config

import (
	"fmt"
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

	stackName = fmt.Sprintf("%s-%s", stackName, string(TargetArch(scope)))

	return stackName
}

// DO NOT modify this function, change EC2 key pair name by 'cdk.json/context/keyPairName'.
func KeyPairName(scope constructs.Construct) string {
	keyPairName := "MyKeyPair"

	ctxValue := scope.Node().TryGetContext(jsii.String("keyPairName"))
	if v, ok := ctxValue.(string); ok {
		keyPairName = v
	}

	return keyPairName
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

// DO NOT modify this function, change EKS cluster nodegroup's archtecture type by 'cdk.json/context/targetArch'.
// Allowed values are: amd64, arm64.
type TargetArchType string

const (
	TargetArch_x86 TargetArchType = "amd64"
	TargetArch_arm TargetArchType = "arm64"
)

func TargetArch(scope constructs.Construct) TargetArchType {
	targetArch := TargetArch_x86

	ctxValue := scope.Node().TryGetContext(jsii.String("targetArch"))
	if v, ok := ctxValue.(string); ok {
		targetArch = TargetArchType(v)
	}

	return targetArch
}

// VPC config
const (
	vpcMask    = 16
	vpcIpv4    = "172.29.0.0"
	MaxAzs     = 3
	SubnetMask = vpcMask + MaxAzs
)

var VpcCidr = vpcIpv4 + "/" + strconv.Itoa(vpcMask)
