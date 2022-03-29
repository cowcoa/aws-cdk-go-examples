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

// Deployment stage config
type DeploymentStage string

const (
	DeploymentStage_DEV  DeploymentStage = "DEV"
	DeploymentStage_PROD DeploymentStage = "PROD"
)

const CurrentDeploymentStage = DeploymentStage_DEV

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

const KeyPairName = "EC2-Dev-Seoul-Key-Pair"
