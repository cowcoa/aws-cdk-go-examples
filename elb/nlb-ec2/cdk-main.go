package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"

	//ec2ins "github.com/aws/aws-cdk-go/awscdk/v2/awsec2/Instance"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	elbv2tgt "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"nlb-ec2/config"
)

type NlbEc2StackProps struct {
	awscdk.StackProps
}

func NewNlbEc2Stack(scope constructs.Construct, id string, props *NlbEc2StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	// Import default VPC.
	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("DefaultVPC"), &awsec2.VpcLookupOptions{
		IsDefault: jsii.Bool(true),
	})

	// Create SecurityGroup for HTTP inbound request
	sg := awsec2.NewSecurityGroup(stack, jsii.String("EC2SG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(*stack.StackName() + "-EC2SG"),
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("EC2 communicate with external."),
	})
	sg.Connections().AllowFrom(sg, awsec2.Port_AllTraffic(),
		jsii.String("Allow all EC2 instance communicate each other with the this SG."))
	sg.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(80),
			ToPort:               jsii.Number(80),
			StringRepresentation: jsii.String("Receive HTTP requests."),
		}),
		jsii.String("Allow requests to HTTP server."),
		jsii.Bool(false),
	)

	// Get key-pair pointer.
	var keyPair *string = nil
	if len(config.KeyPairName(stack)) > 0 {
		keyPair = jsii.String(config.KeyPairName(stack))
	}

	// Setting up the launch script for Ec2 instance
	userData := awsec2.UserData_ForLinux(&awsec2.LinuxUserDataOptions{
		Shebang: jsii.String("#!/bin/bash"),
	})
	userData.AddCommands(
		jsii.String(`sudo su`),
		jsii.String(`yum install -y httpd`),
		jsii.String(`systemctl start httpd`),
		jsii.String(`systemctl enable httpd`),
		jsii.String(`echo "<h1>Hello World from $(hostname -f)</h1>" > /var/www/html/index.html`),
	)

	// Create EC2 role.
	ec2Role := awsiam.NewRole(stack, jsii.String("EC2Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		RoleName:  jsii.String(*stack.StackName() + "-" + *stack.Region() + "-EC2Role"),
	})

	ec2Instance := awsec2.NewInstance(stack, jsii.String("EC2Instance"), &awsec2.InstanceProps{
		InstanceName: jsii.String(*stack.StackName() + "-EC2Instance"),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_COMPUTE5, awsec2.InstanceSize_LARGE),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		AvailabilityZone: (*stack.AvailabilityZones())[0],
		BlockDevices: &[]*awsec2.BlockDevice{
			{
				DeviceName: jsii.String("/dev/xvda"),
				Volume: awsec2.BlockDeviceVolume_Ebs(jsii.Number(100), &awsec2.EbsDeviceOptions{
					DeleteOnTermination: jsii.Bool(true),
					VolumeType:          awsec2.EbsDeviceVolumeType_GP3,
					Iops:                jsii.Number(3600),
					Encrypted:           jsii.Bool(false),
				}),
			},
		},
		Role:                            ec2Role,
		KeyName:                         keyPair,
		SecurityGroup:                   sg,
		AllowAllOutbound:                jsii.Bool(true),
		PropagateTagsToVolumeOnCreation: jsii.Bool(true),
		RequireImdsv2:                   jsii.Bool(true),
		SourceDestCheck:                 jsii.Bool(false),
		UserData:                        userData,
		UserDataCausesReplacement:       jsii.Bool(false),
		// ResourceSignalTimeout:           nil,
		// PrivateIpAddress:                new(string),
		// Init:                            nil,
		// InitOptions:                     &awsec2.ApplyCloudFormationInitOptions{},
	})

	// Create Network Load Balancer
	nlb := elbv2.NewNetworkLoadBalancer(stack, jsii.String("NLB"), &elbv2.NetworkLoadBalancerProps{
		LoadBalancerName: jsii.String(*stack.StackName() + "-NLB"),
		Vpc:              vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		InternetFacing:     jsii.Bool(true),
		CrossZoneEnabled:   jsii.Bool(true),
		DeletionProtection: jsii.Bool(false),
	})

	listener := nlb.AddListener(jsii.String("Listener"), &elbv2.BaseNetworkListenerProps{
		Port:     jsii.Number(7891),
		Protocol: elbv2.Protocol_TCP,
		// AlpnPolicy:          "",
		// Certificates:        &[]elbv2.IListenerCertificate{},
		// DefaultAction:       nil,
		// DefaultTargetGroups: &[]elbv2.INetworkTargetGroup{},
		// SslPolicy: "",
	})

	listener.AddTargets(jsii.String("Target"), &elbv2.AddNetworkTargetsProps{
		TargetGroupName: jsii.String(*stack.StackName() + "Target"),
		Protocol:        elbv2.Protocol_TCP,
		Port:            jsii.Number(80),
		Targets: &[]elbv2.INetworkLoadBalancerTarget{
			elbv2tgt.NewInstanceTarget(ec2Instance, jsii.Number(80)),
		},
		PreserveClientIp: jsii.Bool(true),
		HealthCheck: &elbv2.HealthCheck{
			Enabled: jsii.Bool(true),
			// HealthyHttpCodes:      jsii.String("200"),
			HealthyThresholdCount: jsii.Number(3),
			Interval:              awscdk.Duration_Seconds(jsii.Number(30)),
			// Path:                    jsii.String("/"),
			Port:                    jsii.String("80"),
			Protocol:                elbv2.Protocol_TCP,
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(10)),
			UnhealthyThresholdCount: jsii.Number(3),
			// HealthyGrpcCodes:        new(string),
		},
		DeregistrationDelay: awscdk.Duration_Seconds(jsii.Number(10)),
		ProxyProtocolV2:     jsii.Bool(false),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewNlbEc2Stack(app, config.StackName(app), &NlbEc2StackProps{
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
