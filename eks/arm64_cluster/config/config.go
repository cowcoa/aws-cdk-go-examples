package config

import (
	"strconv"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "MyEKSClusterStack"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, change EKS cluster name by 'cdk.json/context/clusterName'.
func ClusterName(scope constructs.Construct) string {
	clusterName := "MyEKSCluster"

	ctxValue := scope.Node().TryGetContext(jsii.String("clusterName"))
	if v, ok := ctxValue.(string); ok {
		clusterName = v
	}

	return clusterName
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

// VPC config
const vpcMask = 16
const vpcIpv4 = "192.168.0.0"

var VpcCidr = vpcIpv4 + "/" + strconv.Itoa(vpcMask)

const MaxAzs = 3
const SubnetMask = vpcMask + MaxAzs

// EKS config
// We are going to create K8S version 1.21, Graviton 2 cluster.
// NOTE: All listed IAM users MUST already exist.
var EksMasterUsers = [...]string{
	"Cow",
	"CowAdmin",
}
