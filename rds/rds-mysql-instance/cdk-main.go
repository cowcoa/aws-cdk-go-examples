package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	secretmgr "github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"rds-mysql-cluster/config"
)

type RdsMySqlClusterStackProps struct {
	awscdk.StackProps
}

func NewRdsMySqlClusterStack(scope constructs.Construct, id string, props *RdsMySqlClusterStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Import default VPC.
	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("DefaultVPC"), &awsec2.VpcLookupOptions{
		IsDefault: jsii.Bool(true),
	})
	// Create MySQL 3306 inbound Security Group.
	sg := awsec2.NewSecurityGroup(stack, jsii.String("MySQLSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(*stack.StackName() + "-MySQLSG"),
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("RDS MySQL DB instances communication SG."),
	})
	sg.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(3306),
			ToPort:               jsii.Number(3306),
			StringRepresentation: jsii.String("Standard MySQL listen port."),
		}),
		jsii.String("Allow requests to MySQL DB instance."),
		jsii.Bool(false),
	)
	// Database engine version.
	var engine awsrds.IInstanceEngine
	if config.MySQLVersion(stack) < 7 {
		engine = awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
			Version: awsrds.MysqlEngineVersion_VER_5_7_34(),
		})
	} else {
		engine = awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
			Version: awsrds.MysqlEngineVersion_VER_8_0_26(),
		})
	}
	// Database subnet group.
	subnetGrp := awsrds.NewSubnetGroup(stack, jsii.String("SubnetGroup"), &awsrds.SubnetGroupProps{
		Vpc:             vpc,
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
		SubnetGroupName: jsii.String(*stack.StackName() + "-SubnetGroup"),
		VpcSubnets:      &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
		Description:     jsii.String("Custom SubnetGroup"),
	})
	// Database parameter group.
	// https://aws.amazon.com/blogs/database/best-practices-for-configuring-parameters-for-amazon-rds-for-mysql-part-1-parameters-related-to-performance/
	paramGrp := awsrds.NewParameterGroup(stack, jsii.String("ParameterGroup"), &awsrds.ParameterGroupProps{
		Engine:      engine,
		Description: jsii.String("Custom ParameterGroup"),
		Parameters: &map[string]*string{
			// "event_scheduler":        jsii.String("ON"),
			// "innodb_sync_array_size": jsii.String("16"),
			"lower_case_table_names": jsii.String("1"),
		},
	})

	// Database credential in SecretManager
	// The Secret must be a JSON string with a “username“ and “password“ field
	dbSecret := secretmgr.NewSecret(stack, jsii.String("DBSecret"), &secretmgr.SecretProps{
		SecretName: jsii.String(*stack.StackName() + "-Secret"),
		GenerateSecretString: &secretmgr.SecretStringGenerator{
			SecretStringTemplate: jsii.String(fmt.Sprintf(`{"username":"%s"}`, config.MasterUser(stack))),
			ExcludePunctuation:   jsii.Bool(true),
			IncludeSpace:         jsii.Bool(false),
			GenerateStringKey:    jsii.String("password"),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	// Create RDS MySQL DB primary instance.
	dbPrimInstance := awsrds.NewDatabaseInstance(stack, jsii.String("PrimaryDBInstance"), &awsrds.DatabaseInstanceProps{
		InstanceIdentifier: jsii.String(*stack.StackName() + "-PrimaryDBInstance"),
		Vpc:                vpc,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			sg,
		},
		InstanceType:               awsec2.InstanceType_Of(awsec2.InstanceClass_MEMORY5, awsec2.InstanceSize_LARGE),
		SubnetGroup:                subnetGrp,
		ParameterGroup:             paramGrp,
		StorageType:                awsrds.StorageType_IO1,
		Iops:                       jsii.Number(5000),
		AllocatedStorage:           jsii.Number(100),
		MaxAllocatedStorage:        jsii.Number(500),
		StorageEncrypted:           jsii.Bool(false),
		MultiAz:                    jsii.Bool(config.EnableMultiAz(stack)),
		DatabaseName:               jsii.String(config.InitDatabase(stack)),
		Engine:                     engine,
		Port:                       jsii.Number(3306),
		PubliclyAccessible:         jsii.Bool(true),
		Credentials:                awsrds.Credentials_FromSecret(dbSecret, jsii.String(config.MasterUser(stack))),
		IamAuthentication:          jsii.Bool(false),
		AllowMajorVersionUpgrade:   jsii.Bool(false),
		AutoMinorVersionUpgrade:    jsii.Bool(true),
		BackupRetention:            awscdk.Duration_Days(jsii.Number(7)),
		CopyTagsToSnapshot:         jsii.Bool(true),
		DeleteAutomatedBackups:     jsii.Bool(true),
		PreferredBackupWindow:      jsii.String("15:30-16:30"),
		PreferredMaintenanceWindow: jsii.String("wed:16:40-wed:17:40"),
		CloudwatchLogsExports: &[]*string{
			jsii.String("error"),
			jsii.String("general"),
			jsii.String("slowquery"),
		},
		CloudwatchLogsRetention:     awslogs.RetentionDays_FIVE_DAYS,
		EnablePerformanceInsights:   jsii.Bool(true),
		PerformanceInsightRetention: awsrds.PerformanceInsightRetention_DEFAULT,
		MonitoringInterval:          awscdk.Duration_Seconds(jsii.Number(60)),
		DeletionProtection:          jsii.Bool(false),
		RemovalPolicy:               awscdk.RemovalPolicy_DESTROY,
		// AvailabilityZone:                jsii.String(*stack.Region() + "a"), // Requesting a specific availability zone is not valid for Multi-AZ instances
		// CharacterSetName:                jsii.String("utf8mb4"),             // isn't supported when creating an instance using version 5.7 of mysql
		// CloudwatchLogsRetentionRole:     nil,                                // CDK utility role, automatically created
		// MonitoringRole:                  nil,                                // will be automatically created
		// OptionGroup:                     nil,                                // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.MySQL.Options.html
		// PerformanceInsightEncryptionKey: nil,                                // no need to specify if you don't encrypt your PI
		// ProcessorFeatures:               &awsrds.ProcessorFeatures{},        // no need to specify
		// S3ExportBuckets:                 &[]awss3.IBucket{},                 // for exporting snapshot to s3 bucket
		// S3ExportRole:                    nil,
		// VpcSubnets:                      &awsec2.SubnetSelection{},          // specified in subnet group
		// Parameters:                      &map[string]*string{},              // specified in parameter group
		// LicenseModel:                    "",                                 // for Microsoft SQL Server
		// Timezone:                        new(string),                        // only support by Microsoft SQL Server
		// S3ImportBuckets:                 &[]awss3.IBucket{},                 // only support by Microsoft SQL Server
		// S3ImportRole:                    nil,
		// Domain:                          new(string),                        // using MS AD for Microsoft SQL Server DB instance
		// DomainRole:                      nil,
		// StorageEncryptionKey:            nil,                                // no need to specify if you don't encrypt your database
	})

	// Create RDS MySQL DB replica instance.
	if config.EnableReplica(stack) {
		awsrds.NewDatabaseInstanceReadReplica(stack, jsii.String("ReplicaDBInstance"), &awsrds.DatabaseInstanceReadReplicaProps{
			InstanceIdentifier: jsii.String(*stack.StackName() + "-ReplicaDBInstance"),
			Vpc:                vpc,
			SecurityGroups: &[]awsec2.ISecurityGroup{
				sg,
			},
			InstanceType:           awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_LARGE),
			SubnetGroup:            subnetGrp,
			ParameterGroup:         paramGrp,
			StorageType:            awsrds.StorageType_GP2,
			SourceDatabaseInstance: dbPrimInstance,
		})
	}

	// Enable RDS Proxy
	if config.EnableProxy(stack) {
		dbProxy := awsrds.NewDatabaseProxy(stack, jsii.String("RDSProxy"), &awsrds.DatabaseProxyProps{
			DbProxyName: jsii.String(*stack.StackName() + "-RDSProxy"),
			Vpc:         vpc,
			VpcSubnets:  &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
			SecurityGroups: &[]awsec2.ISecurityGroup{
				sg,
			},
			ProxyTarget: awsrds.ProxyTarget_FromInstance(dbPrimInstance),
			Secrets: &[]secretmgr.ISecret{
				dbSecret,
			},
			IamAuth:                   jsii.Bool(false),
			RequireTLS:                jsii.Bool(false),
			BorrowTimeout:             awscdk.Duration_Seconds(jsii.Number(30)),
			IdleClientTimeout:         awscdk.Duration_Hours(jsii.Number(1)), // Minimum 1 minute. Maximum: 8 hours
			MaxConnectionsPercent:     jsii.Number(95),
			MaxIdleConnectionsPercent: jsii.Number(95), // between 0 and MaxConnectionsPercent
			DebugLogging:              jsii.Bool(true),
			// InitQuery:                 new(string),
			// Role:                      nil,
			// SessionPinningFilters: &[]awsrds.SessionPinningFilter{},
		})

		awscdk.NewCfnOutput(stack, jsii.String("RDSProxyEndpoint"), &awscdk.CfnOutputProps{
			Value: jsii.String(*dbProxy.Endpoint()),
		})
	}

	// Setup RDS 'failover' event subscription.
	if len(config.EventSubEmail(stack)) >= 5 {
		snsTopic := awssns.NewTopic(stack, jsii.String("Events"), &awssns.TopicProps{
			TopicName:   jsii.String(*stack.StackName() + "-Events"),
			DisplayName: jsii.String("RDS events subscription"),
			Fifo:        jsii.Bool(false),
		})
		emailSub := awssnssubscriptions.NewEmailSubscription(jsii.String(config.EventSubEmail(stack)), &awssnssubscriptions.EmailSubscriptionProps{})
		snsTopic.AddSubscription(emailSub)

		awsrds.NewCfnEventSubscription(stack, jsii.String("EventSubscription"), &awsrds.CfnEventSubscriptionProps{
			Enabled:     true,
			SnsTopicArn: snsTopic.TopicArn(),
			SourceIds: &[]*string{
				dbPrimInstance.InstanceIdentifier(),
			},
			SourceType: jsii.String("db-instance"),
			EventCategories: &[]*string{
				// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Events.Messages.html
				jsii.String("failover"),
			},
		})
	}

	// Output data source info.
	awscdk.NewCfnOutput(stack, jsii.String("RDSPrimaryEndpoint"), &awscdk.CfnOutputProps{
		Value: dbPrimInstance.InstanceEndpoint().Hostname(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("Database"), &awscdk.CfnOutputProps{
		Value: jsii.String(config.InitDatabase(stack)),
	})
	awscdk.NewCfnOutput(stack, jsii.String("User"), &awscdk.CfnOutputProps{
		Value: jsii.String(config.MasterUser(stack)),
	})

	// Output secet info.
	awscdk.NewCfnOutput(stack, jsii.String("SecretName"), &awscdk.CfnOutputProps{
		Value: dbSecret.SecretName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("SecretArn"), &awscdk.CfnOutputProps{
		Value: dbSecret.SecretArn(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("SecretFullArn"), &awscdk.CfnOutputProps{
		Value: dbSecret.SecretFullArn(),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewRdsMySqlClusterStack(app, config.StackName(app), &RdsMySqlClusterStackProps{
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
