package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	secretmgr "github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"nlb-ec2-aurora-ecredis/config"
)

type DBSecret struct {
	Username string `json:"username"`
}

type NEARStackProps struct {
	awscdk.StackProps
}

func NewVpc(stack awscdk.Stack) awsec2.Vpc {
	vpc := awsec2.NewVpc(stack, jsii.String("VPC"), &awsec2.VpcProps{
		VpcName:            jsii.String(*stack.StackName() + "/VPC"),
		Cidr:               jsii.String(config.VpcCidr),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		MaxAzs:             jsii.Number(float64(config.MaxAzs)),
		NatGateways:        jsii.Number(float64(0)),
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
		//DesiredCapacity: jsii.Number(2),
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

	asg.ScaleOnCpuUtilization(jsii.String("cpu-util-scaling"), &awsautoscaling.CpuUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(75),
	})

	return asg
}

func NewNlb(stack awscdk.Stack, vpc awsec2.Vpc, asg awsautoscaling.AutoScalingGroup) elbv2.NetworkLoadBalancer {
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

	return nlb
}

func NewAurora(stack awscdk.Stack, vpc awsec2.Vpc) awsrds.DatabaseCluster {
	arrSG := awsec2.NewSecurityGroup(stack, jsii.String("AuroraSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("Aurora MySQL communicate with external."),
	})
	arrSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(3306),
			ToPort:               jsii.Number(3306),
			StringRepresentation: jsii.String("Receive MySQL client requests."),
		}),
		jsii.String("Allow MySQL requests to DB server."),
		jsii.Bool(false),
	)

	dbsec := &DBSecret{
		Username: "cow",
	}
	dbsecjson, err := json.Marshal(dbsec)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	dbSecret := secretmgr.NewSecret(stack, jsii.String("DBSecret"), &secretmgr.SecretProps{
		SecretName: jsii.String(*stack.StackName() + "-Secret"),
		GenerateSecretString: &secretmgr.SecretStringGenerator{
			SecretStringTemplate: jsii.String(string(dbsecjson)),
			ExcludePunctuation:   jsii.Bool(true),
			IncludeSpace:         jsii.Bool(true),
			GenerateStringKey:    jsii.String("password"),
		},
	})

	dbEngine := awsrds.DatabaseClusterEngine_AuroraMysql(&awsrds.AuroraMysqlClusterEngineProps{
		Version: awsrds.AuroraMysqlEngineVersion_VER_2_09_2(),
	})

	pg := awsrds.NewParameterGroup(stack, jsii.String("ParameterGroup"), &awsrds.ParameterGroupProps{
		Engine:      dbEngine,
		Description: jsii.String("Aurora RDS Instance Parameter Group for database NEAR"),
		Parameters: &map[string]*string{
			"binlog_cache_size": jsii.String("62555"),
		},
	})

	subnetType := awsec2.SubnetType_PUBLIC
	if config.DeploymentStage(stack) == config.DeploymentStage_PROD {
		subnetType = awsec2.SubnetType_PRIVATE_ISOLATED
	}

	arr := awsrds.NewDatabaseCluster(stack, jsii.String("Aurora"), &awsrds.DatabaseClusterProps{
		Engine: dbEngine,
		InstanceProps: &awsrds.InstanceProps{
			InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_MEMORY5, awsec2.InstanceSize_LARGE),
			Vpc:          vpc,
			VpcSubnets: &awsec2.SubnetSelection{
				SubnetType: subnetType,
			},
			SecurityGroups: &[]awsec2.ISecurityGroup{
				arrSG,
			},
			ParameterGroup: pg,
		},
		Backup: &awsrds.BackupProps{
			Retention:       awscdk.Duration_Days(jsii.Number(7)),
			PreferredWindow: jsii.String("03:00-04:00"),
		},
		Credentials:             awsrds.Credentials_FromSecret(dbSecret, jsii.String(dbsec.Username)),
		Instances:               jsii.Number(2),
		CloudwatchLogsRetention: awslogs.RetentionDays_ONE_WEEK,
		DefaultDatabaseName:     jsii.String("cowdatabase"),
		IamAuthentication:       jsii.Bool(false),
		ClusterIdentifier:       jsii.String(*stack.StackName() + "-AuroraDB"),
		SubnetGroup: awsrds.NewSubnetGroup(stack, jsii.String("SubnetGroupNEAR"), &awsrds.SubnetGroupProps{
			Description:     jsii.String("Aurora RDS Subnet Group for database NEAR"),
			SubnetGroupName: jsii.String(*stack.StackName() + "-SubnetGroup"),
			Vpc:             vpc,
			RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
			VpcSubnets: &awsec2.SubnetSelection{
				SubnetType: subnetType,
			},
		}),
	})

	return arr
}

func NewECRedis(stack awscdk.Stack, vpc awsec2.Vpc) awselasticache.CfnReplicationGroup {
	subnets := vpc.SelectSubnets(&awsec2.SubnetSelection{
		SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
	})

	// Create Redis Cluster
	subnetGroup := awselasticache.NewCfnSubnetGroup(stack, jsii.String("SubnetGroup"), &awselasticache.CfnSubnetGroupProps{
		CacheSubnetGroupName: jsii.String(*stack.StackName() + "-private"),
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
		ReplicationGroupId:          jsii.String(*stack.StackName() + "-MyReplicaGroup"),
		SecurityGroupIds: &[]*string{
			vpc.VpcDefaultSecurityGroup(),
		},
	})
	replicaGroup.AddDependsOn(subnetGroup)

	return replicaGroup
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
	asg := NewAsg(stack, vpc)
	nlb := NewNlb(stack, vpc, asg)
	NewAurora(stack, vpc)
	NewECRedis(stack, vpc)

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
