package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseks"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"arm64-cluster/config"
	"arm64-cluster/constructs/addons"
	"arm64-cluster/constructs/ecr"
	"arm64-cluster/constructs/vpc"
)

type EksCdkStackProps struct {
	awscdk.StackProps
}

func NewEksCdkStack(scope constructs.Construct, id string, props *EksCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	// Create ECR repo
	ecr.NewEcrRepository(stack)

	// Create VPC
	vpc := vpc.NewEksVpc(stack)

	// Create EKS cluster
	cluster := createEksCluster(stack, vpc)
	addons.NewEksCoreDns(stack, cluster)
	addons.NewEksKubeProxy(stack, cluster)
	addons.NewEksVpcCni(stack, cluster)
	addons.NewEksEbsCsiDriver(stack, cluster)
	addons.NewEksClusterAutoscaler(stack, cluster)

	return stack
}

func createEksCluster(stack awscdk.Stack, vpc awsec2.Vpc) awseks.Cluster {
	// Create NodeGroup security group.
	nodeSG := awsec2.NewSecurityGroup(stack, jsii.String("NodeSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("EKS worker nodes communicate with external."),
	})
	nodeSG.Connections().AllowFrom(nodeSG, awsec2.Port_AllTraffic(),
		jsii.String("Allow all nodes communicate each other with the this SG."))
	nodeSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(30000),
			ToPort:               jsii.Number(32767),
			StringRepresentation: jsii.String("Receive K8s NodePort requests."),
		}),
		jsii.String("Allow requests to K8s NodePort range."),
		jsii.Bool(false))
	nodeSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(8000),
			ToPort:               jsii.Number(9000),
			StringRepresentation: jsii.String("Receive HTTP requests."),
		}),
		jsii.String("Allow requests to common app range."),
		jsii.Bool(false))

	// Create EKS cluster.
	subnetType := awsec2.SubnetType_PUBLIC
	if config.DeploymentStage(stack) == config.DeploymentStage_PROD {
		subnetType = awsec2.SubnetType_PRIVATE_WITH_NAT
	}
	cluster := awseks.NewCluster(stack, jsii.String("EksCluster"), &awseks.ClusterProps{
		ClusterName: jsii.String(config.ClusterName(stack)),
		Version:     awseks.KubernetesVersion_V1_21(),
		Vpc:         vpc,
		VpcSubnets: &[]*awsec2.SubnetSelection{
			{
				SubnetType: subnetType,
			},
		},
		DefaultCapacity: jsii.Number(0), // Disable creation of default node group.
		AlbController: &awseks.AlbControllerOptions{
			Version: awseks.AlbControllerVersion_V2_3_1(),
		},
		OutputConfigCommand: jsii.Bool(false),
		SecurityGroup:       nodeSG, // Set additional cluster security group.
	})
	awscdk.NewCfnOutput(stack, jsii.String("EKSClusterName"), &awscdk.CfnOutputProps{
		Value: cluster.ClusterName(),
	})

	// Add custom NodeGroup.
	nodeGroupLT := awsec2.NewLaunchTemplate(stack, jsii.String("NodeGroupLT"), &awsec2.LaunchTemplateProps{
		BlockDevices: &[]*awsec2.BlockDevice{
			{
				DeviceName: jsii.String("/dev/xvda"),
				Volume: awsec2.BlockDeviceVolume_Ebs(jsii.Number(100), &awsec2.EbsDeviceOptions{
					DeleteOnTermination: jsii.Bool(true),
					VolumeType:          awsec2.EbsDeviceVolumeType_GP2,
					Encrypted:           jsii.Bool(false),
				}),
			},
		},
		KeyName:            jsii.String(config.KeyPairName),
		LaunchTemplateName: jsii.String(*stack.StackName() + "-NodeLT"),
		SecurityGroup:      nodeSG,
	})

	clusterNodeRole := awsiam.NewRole(stack, jsii.String("ClusterNodeRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEKSWorkerNodePolicy")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEC2ContainerRegistryReadOnly")),
		},
		RoleName: jsii.String(*stack.StackName() + "-ClusterNodeRole"),
	})

	cluster.AddNodegroupCapacity(jsii.String("CustomNodeGroupCapacity"), &awseks.NodegroupOptions{
		AmiType:       awseks.NodegroupAmiType_AL2_X86_64,
		CapacityType:  awseks.CapacityType_ON_DEMAND,
		DesiredSize:   jsii.Number(3),
		InstanceTypes: &[]awsec2.InstanceType{awsec2.InstanceType_Of(awsec2.InstanceClass_COMPUTE5, awsec2.InstanceSize_LARGE)},
		Labels: &map[string]*string{
			"deployment-stage": jsii.String("dev"),
		},
		LaunchTemplateSpec: &awseks.LaunchTemplateSpec{Id: nodeGroupLT.LaunchTemplateId(), Version: nodeGroupLT.LatestVersionNumber()},
		MaxSize:            jsii.Number(5),
		MinSize:            jsii.Number(2),
		NodegroupName:      jsii.String(*stack.StackName() + "-CustomNodeGroup"),
		NodeRole:           clusterNodeRole,
		Subnets: &awsec2.SubnetSelection{
			SubnetType: subnetType,
		},
	})

	// Mapping IAM user to K8s group.
	for _, userName := range config.EksMasterUsers {
		masterUser := awsiam.User_FromUserName(stack, jsii.String("ClusterMasterUser-"+userName), jsii.String(userName))
		cluster.AwsAuth().AddUserMapping(masterUser, &awseks.AwsAuthMapping{
			Groups: &[]*string{
				jsii.String("system:masters"),
			},
		})
	}

	return cluster
}

func main() {
	app := awscdk.NewApp(nil)

	NewEksCdkStack(app, config.StackName(app), &EksCdkStackProps{
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
