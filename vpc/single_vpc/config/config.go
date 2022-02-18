package config

import "strconv"

// Deployment stage config
type DeploymentStage string

const (
	DeploymentStage_DEV  DeploymentStage = "DEV"
	DeploymentStage_PROD DeploymentStage = "PROD"
)

const CurrentDeploymentStage = DeploymentStage_DEV

const (
	StackName  = "CdkGolangExample-SingleVpc"
	vpcMask    = 16
	vpcIpv4    = "192.168.0.0"
	MaxAzs     = 3
	SubnetMask = vpcMask + MaxAzs
)

var VpcCidr = vpcIpv4 + "/" + strconv.Itoa(vpcMask)
