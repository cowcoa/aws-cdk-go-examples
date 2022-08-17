package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"asg-ec2/config"
)

type AsgEc2StackProps struct {
	awscdk.StackProps
}

func NewAsgEc2Stack(scope constructs.Construct, id string, props *AsgEc2StackProps) awscdk.Stack {
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

	// Create EC2 role.
	ec2Role := awsiam.NewRole(stack, jsii.String("EC2Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchLogsReadOnlyAccess")),
		},
		RoleName: jsii.String(*stack.StackName() + "-" + *stack.Region() + "-EC2Role"),
	})

	// Create Auto Scaling Group
	asg := awsautoscaling.NewAutoScalingGroup(stack, jsii.String("ASG"), &awsautoscaling.AutoScalingGroupProps{
		AutoScalingGroupName: jsii.String(*stack.StackName() + "-ASG"),
		Vpc:                  vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		SecurityGroup: sg,
		InstanceType:  awsec2.InstanceType_Of(awsec2.InstanceClass_COMPUTE5, awsec2.InstanceSize_LARGE),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		BlockDevices: &[]*awsautoscaling.BlockDevice{
			{
				DeviceName: jsii.String("/dev/xvda"),
				Volume: awsautoscaling.BlockDeviceVolume_Ebs(jsii.Number(100), &awsautoscaling.EbsDeviceOptions{
					DeleteOnTermination: jsii.Bool(true),
					VolumeType:          awsautoscaling.EbsDeviceVolumeType_GP3,
					Iops:                jsii.Number(3600),
					Encrypted:           jsii.Bool(false),
				}),
			},
		},
		AssociatePublicIpAddress: jsii.Bool(true),
		KeyName:                  keyPair,
		RequireImdsv2:            jsii.Bool(true),
		UserData:                 userData,
		Role:                     ec2Role,
		MaxCapacity:              jsii.Number(5),
		DesiredCapacity:          jsii.Number(2),
		MinCapacity:              jsii.Number(1),
		GroupMetrics: &[]awsautoscaling.GroupMetrics{
			awsautoscaling.GroupMetrics_All(),
		},
		HealthCheck: awsautoscaling.HealthCheck_Elb(&awsautoscaling.ElbHealthCheckOptions{
			Grace: awscdk.Duration_Seconds(jsii.Number(180)),
		}),
		NewInstancesProtectedFromScaleIn: jsii.Bool(false),
		TerminationPolicies: &[]awsautoscaling.TerminationPolicy{
			awsautoscaling.TerminationPolicy_OLDEST_LAUNCH_CONFIGURATION,
			awsautoscaling.TerminationPolicy_OLDEST_LAUNCH_TEMPLATE,
			awsautoscaling.TerminationPolicy_OLDEST_INSTANCE,
			awsautoscaling.TerminationPolicy_DEFAULT,
		},
		UpdatePolicy: awsautoscaling.UpdatePolicy_RollingUpdate(&awsautoscaling.RollingUpdateOptions{
			MaxBatchSize:          jsii.Number(1),
			MinInstancesInService: jsii.Number(1),
		}),
		// AllowAllOutbound:                 jsii.Bool(false), // controll this by SecurityGroup
		// Notifications:                    &[]*awsautoscaling.NotificationConfiguration{}, // send notification to SNS when fleet change.
		// Init:                             nil, // only available in CloudFormation
		// Signals:                          nil, // only available in CloudFormation
		// SpotPrice:                        new(string), // don't need spot instances
		// Cooldown:                         nil, // only for simple scaling policy
		// IgnoreUnmodifiedSizeProperties:   new(bool),
		// InstanceMonitoring:               "", // enable this if need detailed monitoring
		// MaxInstanceLifetime:              nil, // no need to set this value, never replace instances on a schedule
	})

	// Setup autoscaling policy
	asg.ScaleOnCpuUtilization(jsii.String("CpuUtilizationPolicy"), &awsautoscaling.CpuUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(65),
		DisableScaleIn:           jsii.Bool(false),
		EstimatedInstanceWarmup:  awscdk.Duration_Seconds(jsii.Number(180)),
	})

	awscdk.NewCfnOutput(stack, jsii.String("Auto Scaling group name"), &awscdk.CfnOutputProps{
		Value: jsii.String(*asg.AutoScalingGroupName()),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewAsgEc2Stack(app, config.StackName(app), &AsgEc2StackProps{
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
