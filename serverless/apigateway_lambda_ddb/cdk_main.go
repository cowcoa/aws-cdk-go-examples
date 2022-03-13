package main

import (
	"apigateway-lambda-ddb/config"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ApigatewayLambdaDdbStackProps struct {
	awscdk.StackProps
}

func NewApigatewayLambdaDdbStack(scope constructs.Construct, id string, props *ApigatewayLambdaDdbStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	ddbLambdaRole := awsiam.NewRole(stack, jsii.String(config.RoleName), &awsiam.RoleProps{
		RoleName:  jsii.String(*stack.StackName() + "-" + config.RoleName),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBFullAccess")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchFullAccess")),
		},
	})

	// Create custom resource lambda function
	pubRecordFunc := awslambda.NewFunction(stack, jsii.String("PutChatRecords"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-PutChatRecords"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/put-chat-records/."), nil),
		Handler:      jsii.String("put-chat-records"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         ddbLambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	getRecordFunc := awslambda.NewFunction(stack, jsii.String("GetChatRecords"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-GetChatRecords"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/get-chat-records/."), nil),
		Handler:      jsii.String("get-chat-records"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         ddbLambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	restApi := awsapigateway.NewRestApi(stack, jsii.String("LambdaRestApi"), &awsapigateway.RestApiProps{
		Description: jsii.String("example rest api"),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String("dev"),
		},
	})

	putRecords := restApi.Root().AddResource(jsii.String("put-chat-records"), &awsapigateway.ResourceOptions{})
	putRecords.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(pubRecordFunc, &awsapigateway.LambdaIntegrationOptions{}), &awsapigateway.MethodOptions{})

	getRecords := restApi.Root().AddResource(jsii.String("get-chat-records"), &awsapigateway.ResourceOptions{})
	getRecords.AddMethod(jsii.String("GET"), awsapigateway.NewLambdaIntegration(getRecordFunc, &awsapigateway.LambdaIntegrationOptions{}), &awsapigateway.MethodOptions{})

	// Data Modeling
	// name(PK), time(SK),                  comment, chat_room
	// string    string(micro sec unixtime)	string   string
	chatTable := awsdynamodb.NewTable(stack, jsii.String("ChatTable"), &awsdynamodb.TableProps{
		TableName:     jsii.String("ChatTable"),
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

	// Data Modeling
	// chat_room(PK), time(SK),                  comment, name
	// string         string(micro sec unixtime) string   string
	chatTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("ChatTableGSI"),
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

	chatTable.GrantWriteData(pubRecordFunc)
	chatTable.GrantReadData(getRecordFunc)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewApigatewayLambdaDdbStack(app, config.StackName(app), &ApigatewayLambdaDdbStackProps{
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
