package main

import (
	"os"
	"redis-cluster/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type RedisClusterStackProps struct {
	awscdk.StackProps
}

func NewRedisClusterStack(scope constructs.Construct, id string, props *RedisClusterStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	// Create VPC
	vpc := awsec2.NewVpc(stack, jsii.String("VPC"), &awsec2.VpcProps{
		VpcName:            jsii.String(*stack.StackName() + "-VPC"),
		Cidr:               jsii.String(config.VpcCidr),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		MaxAzs:             jsii.Number(float64(config.MaxAzs)),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:                jsii.String("Public"),
				MapPublicIpOnLaunch: jsii.Bool(true),
				SubnetType:          awsec2.SubnetType_PUBLIC,
				CidrMask:            jsii.Number(float64(config.SubnetMask)),
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(float64(config.SubnetMask)),
			},
		},
	})

	subnets := vpc.SelectSubnets(&awsec2.SubnetSelection{
		SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
	})

	// Create SecurityGroup
	sg := awsec2.NewSecurityGroup(stack, jsii.String("SG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(*stack.StackName() + "-SG"),
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("SecurityGroup for ElastiCache Redis cluster"),
	})
	sg.Connections().AllowFrom(sg, awsec2.Port_AllTraffic(),
		jsii.String("Allow all nodes in this SG to communicate with each other"))
	sg.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(config.Port(stack)),
			ToPort:               jsii.Number(config.Port(stack)),
			StringRepresentation: jsii.String("Allow incoming Redis requests"),
		}),
		jsii.String("Allow incoming Redis requests"),
		jsii.Bool(false),
	)

	// Create SubnetGroup
	subnetGroup := awselasticache.NewCfnSubnetGroup(stack, jsii.String("SubnetGroup"), &awselasticache.CfnSubnetGroupProps{
		CacheSubnetGroupName: jsii.String(*stack.StackName() + "-SubnetGroup"),
		SubnetIds:            subnets.SubnetIds,
		Description:          jsii.String(""),
	})

	// Create ParameterGroup
	paramGroup := awselasticache.NewCfnParameterGroup(stack, jsii.String("ParameterGroup"), &awselasticache.CfnParameterGroupProps{
		CacheParameterGroupFamily: jsii.String("redis6.x"),
		Description:               jsii.String(*stack.StackName() + " parameter group"),
		Properties: map[string]*string{
			"cluster-enabled": jsii.String("yes"),
		},
	})

	// Create LogGroup
	loggroup := awslogs.NewLogGroup(stack, jsii.String("LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String(*stack.StackName()),
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
		Retention:     awslogs.RetentionDays_FIVE_DAYS,
	})

	// Enable password or not
	var transitEnc bool = false
	var password *string = nil
	if len(config.Password(stack)) > 0 {
		transitEnc = true
		password = jsii.String(config.Password(stack))
	}

	// Create Redis Cluster
	replicaGroup := awselasticache.NewCfnReplicationGroup(stack, jsii.String("ReplicationGroup"), &awselasticache.CfnReplicationGroupProps{
		ReplicationGroupId:          jsii.String(*stack.StackName()),
		ReplicationGroupDescription: jsii.String("Demonstrate how to create a cluster mode enabled redis cluster."),
		Engine:                      jsii.String("redis"),
		EngineVersion:               jsii.String("6.2"),
		CacheNodeType:               jsii.String("cache.r6g.2xlarge"),
		SecurityGroupIds: &[]*string{
			sg.SecurityGroupId(),
		},
		CacheParameterGroupName:    paramGroup.Ref(),
		CacheSubnetGroupName:       subnetGroup.CacheSubnetGroupName(),
		MultiAzEnabled:             jsii.Bool(true),
		NumNodeGroups:              jsii.Number(2), // 2 shards in total
		ReplicasPerNodeGroup:       jsii.Number(1), // 1 replica per shard
		TransitEncryptionEnabled:   jsii.Bool(transitEnc),
		AuthToken:                  password,
		Port:                       jsii.Number(config.Port(stack)),
		AtRestEncryptionEnabled:    jsii.Bool(false),
		AutomaticFailoverEnabled:   jsii.Bool(true),
		AutoMinorVersionUpgrade:    jsii.Bool(false),
		PreferredMaintenanceWindow: jsii.String("wed:16:40-wed:17:40"),
		SnapshotRetentionLimit:     jsii.Number(7),
		SnapshotWindow:             jsii.String("05:00-09:00"),
		LogDeliveryConfigurations: &[]*awselasticache.CfnReplicationGroup_LogDeliveryConfigurationRequestProperty{
			{
				DestinationDetails: &awselasticache.CfnReplicationGroup_DestinationDetailsProperty{
					CloudWatchLogsDetails: &awselasticache.CfnReplicationGroup_CloudWatchLogsDestinationDetailsProperty{
						LogGroup: loggroup.LogGroupName(), // you MUST create CloudWatch LogGroup at first by yourself
					},
				},
				DestinationType: jsii.String("cloudwatch-logs"), // or kinesis-firehose
				LogFormat:       jsii.String("json"),            // or text
				LogType:         jsii.String("slow-log"),        // or engine-log
			},
		},
		// CacheSecurityGroupNames:  &[]*string{}, // specify SG by SecurityGroupIds
		// DataTieringEnabled:       nil,          // only supported for replication groups using the r6gd node type
		// NodeGroupConfiguration:   nil,          // parameters are repeated with the previous setting
		// NumCacheClusters:         new(float64), // is not used if there is more than one node group (shard)
		// PrimaryClusterId:         new(string),  // is not required if NumNodeGroups or ReplicasPerNodeGroup is specified
		// SnapshotArns:             &[]*string{}, // snapshot files are used to populate the new replication group
		// SnapshotName:             new(string),  // The name of a snapshot from which to restore data into the new replication group
		// SnapshottingClusterId:    new(string),  // cannot be set for Redis (cluster mode enabled) replication groups
		// PreferredCacheClusterAZs: &[]*string{},
		// NotificationTopicArn:     new(string),
		// GlobalReplicationGroupId: new(string),
		// KmsKeyId:                 new(string),
		// Tags:                     &[]*awscdk.CfnTag{},
		// UserGroupIds:             &[]*string{},
	})
	replicaGroup.AddDependsOn(subnetGroup)

	awscdk.NewCfnOutput(stack, jsii.String("RedisClusterName"), &awscdk.CfnOutputProps{
		Value: jsii.String(*replicaGroup.ReplicationGroupId()),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ConfigurationEndpointAddress"), &awscdk.CfnOutputProps{
		Value: jsii.String(*replicaGroup.AttrConfigurationEndPointAddress()),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ConfigurationEndpointPort"), &awscdk.CfnOutputProps{
		Value: jsii.String(*replicaGroup.AttrConfigurationEndPointPort()),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewRedisClusterStack(app, config.StackName(app), &RedisClusterStackProps{
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
