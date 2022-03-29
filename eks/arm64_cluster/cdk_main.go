package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseks"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"arm64-cluster/config"
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

	vpc := createVpc(stack)
	createEcrRepository(stack)
	createEksCluster(stack, vpc)

	return stack
}

func createEksCluster(stack awscdk.Stack, vpc *awsec2.Vpc) {
	// Create node group security group.
	nodeSG := awsec2.NewSecurityGroup(stack, jsii.String("EksNodeSG"), &awsec2.SecurityGroupProps{
		Vpc:              *vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("EKS worker nodes communicate with each other."),
	})
	nodeSG.Connections().AllowFrom(nodeSG, awsec2.Port_AllTraffic(),
		jsii.String("Allow all nodes communicate each other with the this SG."))
	nodeSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(8000),
			ToPort:               jsii.Number(9000),
			StringRepresentation: jsii.String("Receive HTTP(s) based request"),
		}),
		jsii.String("Allow testing requests"),
		jsii.Bool(false))

	// Create EKS cluster.
	subnetType := awsec2.SubnetType_PUBLIC
	if config.CurrentDeploymentStage == config.DeploymentStage_PROD {
		subnetType = awsec2.SubnetType_PRIVATE_WITH_NAT
	}
	cluster := awseks.NewCluster(stack, jsii.String("EksCluster"), &awseks.ClusterProps{
		ClusterName: jsii.String(*stack.StackName() + "-Cluster"),
		Version:     awseks.KubernetesVersion_V1_21(),
		Vpc:         *vpc,
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

	awscdk.NewCfnOutput(stack, jsii.String("EksClusterName"), &awscdk.CfnOutputProps{
		Value: cluster.ClusterName(),
	})

	// Add custom node group.
	nodeGroupLT := awsec2.NewLaunchTemplate(stack, jsii.String("EksNodeGroupLT"), &awsec2.LaunchTemplateProps{
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
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEKS_CNI_Policy")),
		},
		RoleName: jsii.String(*stack.StackName() + "-ClusterNodeRole"),
	})

	cluster.AddNodegroupCapacity(jsii.String("NewAdd"), &awseks.NodegroupOptions{
		AmiType:       awseks.NodegroupAmiType_AL2_X86_64,
		CapacityType:  awseks.CapacityType_ON_DEMAND,
		DesiredSize:   jsii.Number(3),
		InstanceTypes: &[]awsec2.InstanceType{awsec2.InstanceType_Of(awsec2.InstanceClass_COMPUTE5, awsec2.InstanceSize_LARGE)},
		Labels: &map[string]*string{
			"deployment-stage": jsii.String("dev"),
		},
		LaunchTemplateSpec: &awseks.LaunchTemplateSpec{Id: nodeGroupLT.LaunchTemplateId(), Version: nodeGroupLT.LatestVersionNumber()},
		MaxSize:            jsii.Number(5),
		MinSize:            jsii.Number(1),
		NodegroupName:      jsii.String(*stack.StackName() + "-CustomNodeGroup"),
		NodeRole:           clusterNodeRole,
		Subnets: &awsec2.SubnetSelection{
			SubnetType: subnetType,
		},
	})

	// Mapping IAM user to K8S group.
	for _, userName := range config.EksMasterUsers {
		masterUser := awsiam.User_FromUserName(stack, jsii.String("EksMasterUser-"+userName), jsii.String(userName))
		cluster.AwsAuth().AddUserMapping(masterUser, &awseks.AwsAuthMapping{
			Groups: &[]*string{
				jsii.String("system:masters"),
			},
		})
	}

	// Create IAM Policy for Cluster Autoscaler
	caStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("autoscaling:DescribeAutoScalingGroups"),
			jsii.String("autoscaling:DescribeAutoScalingInstances"),
			jsii.String("autoscaling:DescribeLaunchConfigurations"),
			jsii.String("autoscaling:DescribeTags"),
			jsii.String("autoscaling:SetDesiredCapacity"),
			jsii.String("autoscaling:TerminateInstanceInAutoScalingGroup"),
			jsii.String("ec2:DescribeLaunchTemplateVersions"),
		},
		Resources: &[]*string{
			jsii.String("*"),
		},
	})
	caPolicy := awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		AssignSids: jsii.Bool(true),
		Statements: &[]awsiam.PolicyStatement{
			caStatement,
		},
	})

	// Create IAM Role for EKS Cluster Autoscaler.
	caConditionVal := make(map[string]string)
	provider := *cluster.OpenIdConnectProvider().OpenIdConnectProviderIssuer() + ":sub"
	caConditionVal[provider] = "system:serviceaccount:kube-system:cluster-autoscaler"

	jsonCondition := awscdk.NewCfnJson(stack, jsii.String("JsonCondition"), &awscdk.CfnJsonProps{
		Value: caConditionVal,
	})

	caRole := awsiam.NewRole(stack, jsii.String("EksCARole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-EksCARole"),
		AssumedBy: awsiam.NewWebIdentityPrincipal(cluster.OpenIdConnectProvider().OpenIdConnectProviderArn(), &map[string]interface{}{
			"StringEquals": &jsonCondition,
		}),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"EksCAPolicy": caPolicy,
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("EksCARoleArn"), &awscdk.CfnOutputProps{
		Value: caRole.RoleArn(),
	})

	caSa := awseks.NewServiceAccount(stack, jsii.String("ClusterAutoscaler-SA"), &awseks.ServiceAccountProps{
		Name:      jsii.String("cluster-autoscaler-sa"),
		Cluster:   cluster,
		Namespace: jsii.String("kube-system"),
	})

	awsiam.NewPolicy(stack, jsii.String("CA_Policy"), &awsiam.PolicyProps{
		Document:   caPolicy,
		PolicyName: jsii.String(*stack.StackName() + "-CA_Policy"),
		Roles: &[]awsiam.IRole{
			caSa.Role(),
		},
	})

	awseks.NewHelmChart(stack, jsii.String("clusterautoscaler-deploy"), &awseks.HelmChartProps{
		Repository: jsii.String("https://kubernetes.github.io/autoscaler"),
		Release:    jsii.String("cluster-autoscaler"),
		Cluster:    cluster,
		Chart:      jsii.String("cluster-autoscaler"),
		Namespace:  jsii.String("kube-system"),
		Wait:       jsii.Bool(true),
		Version:    jsii.String("9.13.1"),
		Values: &map[string]interface{}{
			"cloudProvider": jsii.String("aws"),
			"awsRegion":     jsii.String(*stack.Region()),
			"autoDiscovery": map[string]string{
				"clusterName": *cluster.ClusterName(),
			},
			"rbac": map[string]map[string]interface{}{
				"serviceAccount": {
					"create": jsii.Bool(false),
					"name":   caSa.ServiceAccountName(),
				},
			},
			"extraArgs": map[string]interface{}{
				"skip-nodes-with-system-pods":   jsii.Bool(false),
				"skip-nodes-with-local-storage": jsii.Bool(false),
				"balance-similar-node-groups":   jsii.Bool(true),
				// How long a node should be unneeded before it is eligible for scale down
				"scale-down-unneeded-time": jsii.String("300s"),
				// How long after scale up that scale down evaluation resumes
				"scale-down-delay-after-add": jsii.String("300s"),
			},
		},
	})

	// Install EBS CSI driver
	// Create IAM Policy for EBS CSI driver
	ebsCsiPolicy := awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		AssignSids: jsii.Bool(true),
		Statements: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateSnapshot"),
					jsii.String("ec2:AttachVolume"),
					jsii.String("ec2:DetachVolume"),
					jsii.String("ec2:ModifyVolume"),
					jsii.String("ec2:DescribeAvailabilityZones"),
					jsii.String("ec2:DescribeInstances"),
					jsii.String("ec2:DescribeSnapshots"),
					jsii.String("ec2:DescribeTags"),
					jsii.String("ec2:DescribeVolumes"),
					jsii.String("ec2:DescribeVolumesModifications"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateTags"),
				},
				Resources: &[]*string{
					jsii.String("arn:aws:ec2:*:*:volume/*"),
					jsii.String("arn:aws:ec2:*:*:snapshot/*"),
				},
				Conditions: &map[string]interface{}{
					"StringEquals": awscdk.NewCfnJson(stack, jsii.String("JsonCondition2"), &awscdk.CfnJsonProps{
						Value: map[string][]string{
							"ec2:CreateAction": {
								"CreateVolume",
								"CreateSnapshot",
							},
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DeleteTags"),
				},
				Resources: &[]*string{
					jsii.String("arn:aws:ec2:*:*:volume/*"),
					jsii.String("arn:aws:ec2:*:*:snapshot/*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateVolume"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition3"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/ebs.csi.aws.com/cluster": "true",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateVolume"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition4"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/CSIVolumeName": "*",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateVolume"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition5"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/kubernetes.io/cluster/*": "owned",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DeleteVolume"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition6"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"ec2:ResourceTag/ebs.csi.aws.com/cluster": "true",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DeleteVolume"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition7"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"ec2:ResourceTag/CSIVolumeName": "*",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DeleteVolume"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition8"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"ec2:ResourceTag/kubernetes.io/cluster/*": "owned",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DeleteSnapshot"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition9"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"ec2:ResourceTag/CSIVolumeSnapshotName": "*",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DeleteSnapshot"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringLike": awscdk.NewCfnJson(stack, jsii.String("JsonCondition10"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"ec2:ResourceTag/ebs.csi.aws.com/cluster": "true",
						},
					}),
				},
			}),
		},
	})

	ebsCsiConditionVal := make(map[string]string)
	ebsCsiprovider := *cluster.OpenIdConnectProvider().OpenIdConnectProviderIssuer() + ":sub"
	ebsCsiConditionVal[ebsCsiprovider] = "system:serviceaccount:kube-system:ebs-csi-controller-sa"

	ebsCsiJsonCondition := awscdk.NewCfnJson(stack, jsii.String("JsonCondition11"), &awscdk.CfnJsonProps{
		Value: ebsCsiConditionVal,
	})

	ebsCsiRole := awsiam.NewRole(stack, jsii.String("EksEBSCSIRole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-EksEBSCSIRole"),
		AssumedBy: awsiam.NewWebIdentityPrincipal(cluster.OpenIdConnectProvider().OpenIdConnectProviderArn(), &map[string]interface{}{
			"StringEquals": &ebsCsiJsonCondition,
		}),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"EksEBSCSIPolicy": ebsCsiPolicy,
		},
	})

	awseks.NewCfnAddon(stack, jsii.String("csi-addon"), &awseks.CfnAddonProps{
		AddonName:             jsii.String("aws-ebs-csi-driver"),
		ClusterName:           cluster.ClusterName(),
		ServiceAccountRoleArn: ebsCsiRole.RoleArn(),
		AddonVersion:          jsii.String("v1.4.0-eksbuild.preview"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EksEBSCSIRoleArn"), &awscdk.CfnOutputProps{
		Value: ebsCsiRole.RoleArn(),
	})

	// Install CoreDNS add-on
	awseks.NewCfnAddon(stack, jsii.String("core-dns"), &awseks.CfnAddonProps{
		AddonName:        jsii.String("coredns"),
		ResolveConflicts: jsii.String("OVERWRITE"),
		ClusterName:      cluster.ClusterName(),
		AddonVersion:     jsii.String("v1.8.4-eksbuild.1"),
	})

	// Install kube-proxy add-on
	awseks.NewCfnAddon(stack, jsii.String("kube-proxy"), &awseks.CfnAddonProps{
		AddonName:        jsii.String("kube-proxy"),
		ResolveConflicts: jsii.String("OVERWRITE"),
		ClusterName:      cluster.ClusterName(),
		AddonVersion:     jsii.String("v1.21.2-eksbuild.2"),
	})

	// Install VPC CNI add-on
	cniRole := awsiam.NewRole(stack, jsii.String("EksCNIRole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-EksCNIRole"),
		AssumedBy: awsiam.NewWebIdentityPrincipal(cluster.OpenIdConnectProvider().OpenIdConnectProviderArn(), &map[string]interface{}{
			"StringEquals": awscdk.NewCfnJson(stack, jsii.String("JsonConditionCNI"), &awscdk.CfnJsonProps{
				Value: map[string]string{
					*cluster.OpenIdConnectProvider().OpenIdConnectProviderIssuer() + ":sub": "system:serviceaccount:kube-system:aws-node",
				},
			}),
		}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonEKS_CNI_Policy")),
		},
	})
	awseks.NewCfnAddon(stack, jsii.String("vpc-cni"), &awseks.CfnAddonProps{
		AddonName:             jsii.String("vpc-cni"),
		ResolveConflicts:      jsii.String("OVERWRITE"),
		ClusterName:           cluster.ClusterName(),
		AddonVersion:          jsii.String("v1.10.2-eksbuild.1"),
		ServiceAccountRoleArn: cniRole.RoleArn(),
	})
}

func createEcrRepository(stack awscdk.Stack) {
	repo := awsecr.NewRepository(stack, jsii.String("EcrRepository"), &awsecr.RepositoryProps{
		RepositoryName:     jsii.String(strings.ToLower(*stack.StackName()) + "-repo"),
		RemovalPolicy:      awscdk.RemovalPolicy_DESTROY,
		ImageTagMutability: awsecr.TagMutability_MUTABLE,
		ImageScanOnPush:    jsii.Bool(false),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EcrRepositoryName"), &awscdk.CfnOutputProps{
		Value: repo.RepositoryName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("EcrRepositoryUri"), &awscdk.CfnOutputProps{
		Value: repo.RepositoryUri(),
	})
}

func createVpc(stack awscdk.Stack) *awsec2.Vpc {
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
		awscdk.Tags_Of(subnet).Add(jsii.String("kubernetes.io/role/elb"), jsii.String("1"), &awscdk.TagProps{})
	}
	// Private subnets
	for index, subnet := range *vpc.PrivateSubnets() {
		subnetName := *stack.StackName() + "-PrivateSubnet0" + strconv.Itoa(index+1)
		awscdk.Tags_Of(subnet).Add(jsii.String("Name"), jsii.String(subnetName), &awscdk.TagProps{})
		awscdk.Tags_Of(subnet).Add(jsii.String("kubernetes.io/role/internal-elb"), jsii.String("1"), &awscdk.TagProps{})
	}

	return &vpc
}

func main() {
	app := awscdk.NewApp(nil)

	NewEksCdkStack(app, config.StackName, &EksCdkStackProps{
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
