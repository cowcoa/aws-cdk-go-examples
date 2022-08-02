package main

import (
	"os"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"nlb-ec2-aurora-ecredis/config"
)

type NEARStackProps struct {
	awscdk.StackProps
}

func NewVpc(stack awscdk.Stack) awsec2.Vpc {
	subnetConfigs := []*awsec2.SubnetConfiguration{
		{
			Name:                jsii.String("PublicSubnet"),
			MapPublicIpOnLaunch: jsii.Bool(true),
			SubnetType:          awsec2.SubnetType_PUBLIC,
			CidrMask:            jsii.Number(float64(config.SubnetMask)),
		},
	}

	if config.DeploymentStage(stack) == config.DeploymentStage_PROD {
		privateSub := &awsec2.SubnetConfiguration{
			Name:       jsii.String("PrivateSubnet"),
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
			CidrMask:   jsii.Number(float64(config.SubnetMask)),
		}
		subnetConfigs = append(subnetConfigs, privateSub)
	}

	vpc := awsec2.NewVpc(stack, jsii.String("VPC"), &awsec2.VpcProps{
		VpcName:             jsii.String(*stack.StackName() + "/VPC"),
		Cidr:                jsii.String(config.VpcCidr),
		EnableDnsHostnames:  jsii.Bool(true),
		EnableDnsSupport:    jsii.Bool(true),
		MaxAzs:              jsii.Number(float64(config.MaxAzs)),
		NatGateways:         jsii.Number(float64(0)),
		SubnetConfiguration: &subnetConfigs,
	})

	// Tagging subnets
	// Public subnets
	for index, subnet := range *vpc.PublicSubnets() {
		subnetName := *stack.StackName() + "/PublicSubnet0" + strconv.Itoa(index+1)
		awscdk.Tags_Of(subnet).Add(jsii.String("Name"), jsii.String(subnetName), &awscdk.TagProps{})
	}
	// Private subnets
	for index, subnet := range *vpc.PrivateSubnets() {
		subnetName := *stack.StackName() + "/PrivateSubnet0" + strconv.Itoa(index+1)
		awscdk.Tags_Of(subnet).Add(jsii.String("Name"), jsii.String(subnetName), &awscdk.TagProps{})
	}

	awscdk.NewCfnOutput(stack, jsii.String("vpcId"), &awscdk.CfnOutputProps{
		Value: vpc.VpcId(),
	})

	return vpc
}

func NewAsg(stack awscdk.Stack, vpc awsec2.Vpc) awsautoscaling.AutoScalingGroup {
	// Get key-pair pointer.
	var keyPair *string = nil
	if len(config.KeyPairName(stack)) > 0 {
		keyPair = jsii.String(config.KeyPairName(stack))
	}

	subnetType := awsec2.SubnetType_PUBLIC
	if config.DeploymentStage(stack) == config.DeploymentStage_PROD {
		subnetType = awsec2.SubnetType_PRIVATE_ISOLATED
	}

	// Create cluster node role.
	clusterNodeRole := awsiam.NewRole(stack, jsii.String("ClusterNodeRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEKSWorkerNodePolicy")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEC2ContainerRegistryReadOnly")),
		},
		RoleName: jsii.String(*stack.StackName() + "-ClusterNodeRole"),
	})

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

	nodeSG := awsec2.NewSecurityGroup(stack, jsii.String("SG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("NEAR PoC communicate with external."),
	})
	nodeSG.Connections().AllowFrom(nodeSG, awsec2.Port_AllTraffic(),
		jsii.String("Allow all nodes communicate each other with the this SG."))
	nodeSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(80),
			ToPort:               jsii.Number(80),
			StringRepresentation: jsii.String("Receive HTTP requests."),
		}),
		jsii.String("Allow requests to HTTP server."),
		jsii.Bool(false))

	asg := awsautoscaling.NewAutoScalingGroup(stack, jsii.String("ASG"), &awsautoscaling.AutoScalingGroupProps{
		//AllowAllOutbound:                 new(bool),
		//AssociatePublicIpAddress:         new(bool),
		AutoScalingGroupName: jsii.String(*stack.StackName() + "-ASG"),
		//BlockDevices:                     &[]*awsautoscaling.BlockDevice{},
		//Cooldown:                         nil,
		DesiredCapacity: jsii.Number(2),
		//GroupMetrics:                     &[]awsautoscaling.GroupMetrics{},
		HealthCheck: awsautoscaling.HealthCheck_Elb(&awsautoscaling.ElbHealthCheckOptions{
			Grace: awscdk.Duration_Seconds(jsii.Number(30)),
		}),
		//IgnoreUnmodifiedSizeProperties:   new(bool),
		//InstanceMonitoring:               "",
		KeyName:     keyPair,
		MaxCapacity: jsii.Number(3),
		//MaxInstanceLifetime:              nil,
		MinCapacity: jsii.Number(1),
		//NewInstancesProtectedFromScaleIn: new(bool),
		//Notifications:                    &[]*awsautoscaling.NotificationConfiguration{},
		//Signals:                          nil,
		//SpotPrice:                        new(string),
		//TerminationPolicies:              &[]awsautoscaling.TerminationPolicy{},
		UpdatePolicy: awsautoscaling.UpdatePolicy_RollingUpdate(&awsautoscaling.RollingUpdateOptions{
			MaxBatchSize:          jsii.Number(1),
			MinInstancesInService: jsii.Number(1),
		}),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: subnetType,
		},
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_COMPUTE5, awsec2.InstanceSize_LARGE),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		Vpc:           vpc,
		Init:          nil,
		RequireImdsv2: jsii.Bool(true),
		Role:          clusterNodeRole,
		SecurityGroup: nodeSG,
		UserData:      userData,
	})

	return asg
}

func NewNEARStack(scope constructs.Construct, id string, props *NEARStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	// Create VPC
	vpc := NewVpc(stack)

	nlb := elbv2.NewNetworkLoadBalancer(stack, jsii.String("NLB"), &elbv2.NetworkLoadBalancerProps{
		LoadBalancerName: jsii.String(*stack.StackName() + "-NLB"),
		Vpc:              vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		InternetFacing:   jsii.Bool(true),
		CrossZoneEnabled: jsii.Bool(true),
	})

	listener := nlb.AddListener(jsii.String("Listener"), &elbv2.BaseNetworkListenerProps{
		Port:     jsii.Number(7891),
		Protocol: elbv2.Protocol_TCP,
	})

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

	nodeSG := awsec2.NewSecurityGroup(stack, jsii.String("SG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("NEAR PoC communicate with external."),
	})
	nodeSG.Connections().AllowFrom(nodeSG, awsec2.Port_AllTraffic(),
		jsii.String("Allow all nodes communicate each other with the this SG."))
	nodeSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(80),
			ToPort:               jsii.Number(80),
			StringRepresentation: jsii.String("Receive HTTP requests."),
		}),
		jsii.String("Allow requests to HTTP server."),
		jsii.Bool(false))

	// Get key-pair pointer.
	var keyPair *string = nil
	if len(config.KeyPairName(stack)) > 0 {
		keyPair = jsii.String(config.KeyPairName(stack))
	}

	// Create cluster node role.
	clusterNodeRole := awsiam.NewRole(stack, jsii.String("ClusterNodeRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEKSWorkerNodePolicy")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEC2ContainerRegistryReadOnly")),
		},
		RoleName: jsii.String(*stack.StackName() + "-ClusterNodeRole"),
	})

	subnetType := awsec2.SubnetType_PUBLIC
	if config.DeploymentStage(stack) == config.DeploymentStage_PROD {
		subnetType = awsec2.SubnetType_PRIVATE_ISOLATED
	}

	asg := awsautoscaling.NewAutoScalingGroup(stack, jsii.String("ASG"), &awsautoscaling.AutoScalingGroupProps{
		//AllowAllOutbound:                 new(bool),
		//AssociatePublicIpAddress:         new(bool),
		AutoScalingGroupName: jsii.String(*stack.StackName() + "-ASG"),
		//BlockDevices:                     &[]*awsautoscaling.BlockDevice{},
		//Cooldown:                         nil,
		DesiredCapacity: jsii.Number(2),
		//GroupMetrics:                     &[]awsautoscaling.GroupMetrics{},
		HealthCheck: awsautoscaling.HealthCheck_Elb(&awsautoscaling.ElbHealthCheckOptions{
			Grace: awscdk.Duration_Seconds(jsii.Number(30)),
		}),
		//IgnoreUnmodifiedSizeProperties:   new(bool),
		//InstanceMonitoring:               "",
		KeyName:     keyPair,
		MaxCapacity: jsii.Number(3),
		//MaxInstanceLifetime:              nil,
		MinCapacity: jsii.Number(1),
		//NewInstancesProtectedFromScaleIn: new(bool),
		//Notifications:                    &[]*awsautoscaling.NotificationConfiguration{},
		//Signals:                          nil,
		//SpotPrice:                        new(string),
		//TerminationPolicies:              &[]awsautoscaling.TerminationPolicy{},
		UpdatePolicy: awsautoscaling.UpdatePolicy_RollingUpdate(&awsautoscaling.RollingUpdateOptions{
			MaxBatchSize:          jsii.Number(1),
			MinInstancesInService: jsii.Number(1),
		}),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: subnetType,
		},
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_COMPUTE5, awsec2.InstanceSize_LARGE),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		Vpc:           vpc,
		Init:          nil,
		RequireImdsv2: jsii.Bool(true),
		Role:          clusterNodeRole,
		SecurityGroup: nodeSG,
		UserData:      userData,
	})

	listener.AddTargets(jsii.String("Target"), &elbv2.AddNetworkTargetsProps{
		Port:                jsii.Number(80),
		DeregistrationDelay: nil,
		HealthCheck: &elbv2.HealthCheck{
			UnhealthyThresholdCount: jsii.Number(3),
			HealthyThresholdCount:   jsii.Number(3),
			Interval:                awscdk.Duration_Seconds(jsii.Number(30)),
		},
		PreserveClientIp: jsii.Bool(true),
		Protocol:         elbv2.Protocol_TCP,
		TargetGroupName:  jsii.String("MyTG"),
		Targets: &[]elbv2.INetworkLoadBalancerTarget{
			asg,
		},
	})

	asg.ScaleOnCpuUtilization(jsii.String("cpu-util-scaling"), &awsautoscaling.CpuUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(75),
	})

	awscdk.NewCfnOutput(stack, jsii.String("albDNS"), &awscdk.CfnOutputProps{
		Value: nlb.LoadBalancerDnsName(),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewNEARStack(app, config.StackName(app), &NEARStackProps{
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
