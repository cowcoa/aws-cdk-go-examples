package main

import (
	"os"
	"single-vpc/config"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/constructs-go/constructs/v10"
)

type VpcCdkStackProps struct {
	awscdk.StackProps
}

func NewVpcStack(scope constructs.Construct, id string, props *VpcCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	ngwNum := 0
	subnetConfigs := []*awsec2.SubnetConfiguration{
		{
			Name:                jsii.String("PublicSubnet"),
			MapPublicIpOnLaunch: jsii.Bool(true),
			SubnetType:          awsec2.SubnetType_PUBLIC,
			CidrMask:            jsii.Number(float64(config.SubnetMask)),
		},
	}

	if config.CurrentDeploymentStage == config.DeploymentStage_PROD {
		ngwNum = config.MaxAzs
		privateSub := &awsec2.SubnetConfiguration{
			Name:       jsii.String("PrivateSubnet"),
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_NAT,
			CidrMask:   jsii.Number(float64(config.SubnetMask)),
		}
		subnetConfigs = append(subnetConfigs, privateSub)
	}

	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		VpcName:             jsii.String(*stack.StackName() + "-Vpc"),
		Cidr:                jsii.String(config.VpcCidr),
		EnableDnsHostnames:  jsii.Bool(true),
		EnableDnsSupport:    jsii.Bool(true),
		MaxAzs:              jsii.Number(float64(config.MaxAzs)),
		NatGateways:         jsii.Number(float64(ngwNum)),
		SubnetConfiguration: &subnetConfigs,
	})

	// Tagging subnets
	// Public subnets
	for index, subnet := range *vpc.PublicSubnets() {
		subnetName := *stack.StackName() + "-PublicSubnet0" + strconv.Itoa(index+1)
		awscdk.Tags_Of(subnet).Add(jsii.String("Name"), jsii.String(subnetName), &awscdk.TagProps{})
	}
	// Private subnets
	for index, subnet := range *vpc.PrivateSubnets() {
		subnetName := *stack.StackName() + "-PrivateSubnet0" + strconv.Itoa(index+1)
		awscdk.Tags_Of(subnet).Add(jsii.String("Name"), jsii.String(subnetName), &awscdk.TagProps{})
	}

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewVpcStack(app, config.StackName, &VpcCdkStackProps{
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
