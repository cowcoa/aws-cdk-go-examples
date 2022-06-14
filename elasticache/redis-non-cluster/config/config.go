package config

import (
	"strconv"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "MyRedisNonClusterStack"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

const (
	vpcMask    = 16
	vpcIpv4    = "172.30.0.0"
	MaxAzs     = 3
	SubnetMask = vpcMask + MaxAzs
)

var VpcCidr = vpcIpv4 + "/" + strconv.Itoa(vpcMask)
