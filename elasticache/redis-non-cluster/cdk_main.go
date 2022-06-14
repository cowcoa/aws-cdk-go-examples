package main

import (
	"os"
	"redis-non-cluster/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type RedisNonClusterStackProps struct {
	awscdk.StackProps
}

func NewRedisNonClusterStack(scope constructs.Construct, id string, props *RedisNonClusterStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	// Create VPC
	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		VpcName:            jsii.String(*stack.StackName() + "-Vpc"),
		Cidr:               jsii.String(config.VpcCidr),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		MaxAzs:             jsii.Number(float64(config.MaxAzs)),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:                jsii.String("PublicSubnet"),
				MapPublicIpOnLaunch: jsii.Bool(true),
				SubnetType:          awsec2.SubnetType_PUBLIC,
				CidrMask:            jsii.Number(float64(config.SubnetMask)),
			},
			{
				Name:       jsii.String("PrivateSubnet"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(float64(config.SubnetMask)),
			},
		},
	})

	subnets := vpc.SelectSubnets(&awsec2.SubnetSelection{
		SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
	})

	// Create Redis Cluster
	subnetGroup := awselasticache.NewCfnSubnetGroup(stack, jsii.String("SubnetGroup"), &awselasticache.CfnSubnetGroupProps{
		CacheSubnetGroupName: jsii.String("private"),
		SubnetIds:            subnets.SubnetIds,
		Description:          jsii.String("try to describe something..."),
	})

	replicaGroup := awselasticache.NewCfnReplicationGroup(stack, jsii.String("RedisReplicationGroup"), &awselasticache.CfnReplicationGroupProps{
		CacheNodeType:            jsii.String("cache.m6g.large"),
		Engine:                   jsii.String("redis"),
		CacheSubnetGroupName:     subnetGroup.CacheSubnetGroupName(),
		MultiAzEnabled:           jsii.Bool(false),
		AutomaticFailoverEnabled: jsii.Bool(false),
		//NumCacheClusters:         jsii.Number(1),
		ReplicasPerNodeGroup:        jsii.Number(2),
		NumNodeGroups:               jsii.Number(1),
		ReplicationGroupDescription: jsii.String("This field is required."),
		ReplicationGroupId:          jsii.String("MyReplicaGroup"),
		SecurityGroupIds: &[]*string{
			vpc.VpcDefaultSecurityGroup(),
		},
	})
	replicaGroup.AddDependsOn(subnetGroup)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewRedisNonClusterStack(app, config.StackName(app), &RedisNonClusterStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEPLOY_ACCOUNT")
	region := os.Getenv("CDK_DEPLOY_REGION")

	if len(account) == 0 || len(region) == 0 {
		account = os.Getenv("CDK_DEFAULT_ACCOUNT")
		region = os.Getenv("CDK_DEFAULT_REGION")
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
