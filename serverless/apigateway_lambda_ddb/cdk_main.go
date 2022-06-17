package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"apigtw-lambda-ddb/config"
)

type ApiGtwLambdaDdbStackProps struct {
	awscdk.StackProps
}

func NewApiGtwLambdaDdbStack(scope constructs.Construct, id string, props *ApiGtwLambdaDdbStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create role for lambda function.
	lambdaRole := awsiam.NewRole(stack, jsii.String("LambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(*stack.StackName() + "-LambdaRole"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBFullAccess")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchFullAccess")),
		},
	})

	// Create put-chat-records function.
	putFunction := awslambda.NewFunction(stack, jsii.String("PutFunction"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-PutChatRecords"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/put-chat-records/."), nil),
		Handler:      jsii.String("put-chat-records"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-" + config.DynamoDBTable),
		},
	})

	// Create get-chat-records function.
	getFunction := awslambda.NewFunction(stack, jsii.String("GetChatRecords"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-GetChatRecords"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/get-chat-records/."), nil),
		Handler:      jsii.String("get-chat-records"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-" + config.DynamoDBTable),
			"DYNAMODB_GSI":   jsii.String(config.DynamoDBGSI),
		},
		// ReservedConcurrentExecutions: jsii.Number(1),
	})

	// Create API Gateway rest api.
	restApi := awsapigateway.NewRestApi(stack, jsii.String("LambdaRestApi"), &awsapigateway.RestApiProps{
		RestApiName:        jsii.String(*stack.StackName() + "-LambdaRestApi"),
		RetainDeployments:  jsii.Bool(false),
		EndpointExportName: jsii.String("RestApiUrl"),
		Deploy:             jsii.Bool(true),
		DeployOptions: &awsapigateway.StageOptions{
			StageName:           jsii.String("dev"),
			CacheClusterEnabled: jsii.Bool(true),
			CacheClusterSize:    jsii.String("0.5"),
			CacheTtl:            awscdk.Duration_Minutes(jsii.Number(1)),
			// https://www.petefreitag.com/item/853.cfm
			// This can help you better understand what burst and rate limite are.
			ThrottlingBurstLimit: jsii.Number(100),
			ThrottlingRateLimit:  jsii.Number(1000),
		},
	})

	// Add path resources to rest api.
	// You MUST associate ApiKey with the methods for the UsagePlane to work.
	putRecordsRes := restApi.Root().AddResource(jsii.String("put-chat-records"), nil)
	putRecordsRes.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(putFunction, nil), &awsapigateway.MethodOptions{
		ApiKeyRequired: jsii.Bool(true),
	})
	getRecordsRes := restApi.Root().AddResource(jsii.String("get-chat-records"), nil)
	getMethod := getRecordsRes.AddMethod(jsii.String("GET"), awsapigateway.NewLambdaIntegration(getFunction, nil), &awsapigateway.MethodOptions{
		ApiKeyRequired: jsii.Bool(true),
	})

	// UsagePlane's throttle can override Stage's DefaultMethodThrottle,
	// while UsagePlanePerApiStage's throttle can override UsagePlane's throttle.
	usagePlane := restApi.AddUsagePlan(jsii.String("UsagePlane"), &awsapigateway.UsagePlanProps{
		Name: jsii.String(*stack.StackName() + "-UsagePlane"),
		Throttle: &awsapigateway.ThrottleSettings{
			BurstLimit: jsii.Number(10),
			RateLimit:  jsii.Number(100),
		},
		Quota: &awsapigateway.QuotaSettings{
			Limit:  jsii.Number(100),
			Offset: jsii.Number(0),
			Period: awsapigateway.Period_DAY,
		},
		ApiStages: &[]*awsapigateway.UsagePlanPerApiStage{
			{
				Api:   restApi,
				Stage: restApi.DeploymentStage(),
				Throttle: &[]*awsapigateway.ThrottlingPerMethod{
					{
						Method: getMethod,
						Throttle: &awsapigateway.ThrottleSettings{
							BurstLimit: jsii.Number(1),
							RateLimit:  jsii.Number(10),
						},
					},
				},
			},
		},
	})

	// Create ApiKey and associate it with UsagePlane.
	apiKey := restApi.AddApiKey(jsii.String("ApiKey"), &awsapigateway.ApiKeyOptions{})
	usagePlane.AddApiKey(apiKey, &awsapigateway.AddApiKeyOptions{})

	// Create DynamoDB Base table.
	// Data Modeling
	// name(PK), time(SK),                  comment, chat_room
	// string    string(micro sec unixtime)	string   string
	chatTable := awsdynamodb.NewTable(stack, jsii.String(config.DynamoDBTable), &awsdynamodb.TableProps{
		TableName:     jsii.String(*stack.StackName() + "-" + config.DynamoDBTable),
		BillingMode:   awsdynamodb.BillingMode_PROVISIONED,
		ReadCapacity:  jsii.Number(1),
		WriteCapacity: jsii.Number(1),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("name"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("time"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		PointInTimeRecovery: jsii.Bool(true),
	})

	// Create DynamoDB GSI table.
	// Data Modeling
	// chat_room(PK), time(SK),                  comment, name
	// string         string(micro sec unixtime) string   string
	chatTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String(config.DynamoDBGSI),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("chat_room"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("time"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
	})

	// Grant access to lambda functions.
	chatTable.GrantWriteData(putFunction)
	chatTable.GrantReadData(getFunction)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewApiGtwLambdaDdbStack(app, config.StackName(app), &ApiGtwLambdaDdbStackProps{
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
