package config

import (
	"strconv"
)

const StackName = "CdkGolangExample-EksArm64Cluster"

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
