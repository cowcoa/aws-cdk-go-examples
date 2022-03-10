package main

import (
	"apigateway-lambda-ddb/config"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
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
	/*
		crLambdaRole := awsiam.NewRole(stack, jsii.String(config.RoleName), &awsiam.RoleProps{
			RoleName:  jsii.String(*stack.StackName() + "-" + config.RoleName),
			AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
			ManagedPolicies: &[]awsiam.IManagedPolicy{
				awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMFullAccess")),
			},
		})
	*/

	// Create custom resource lambda function
	awslambda.NewFunction(stack, jsii.String(config.FuncionName), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-" + config.FuncionName),
		Runtime:      awslambda.Runtime_FROM_IMAGE(),
		MemorySize:   jsii.Number(config.MemorySize),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(config.MaxDuration)),
		Code: awslambda.AssetImageCode_FromAssetImage(jsii.String("functions"), &awslambda.AssetImageCodeProps{
			File: jsii.String("Dockerfile"),
		}),
		//Handler: jsii.String("/var/task/put-chat-records"),
		Handler:               awslambda.Handler_FROM_IMAGE(),
		LogRetention:          awslogs.RetentionDays_ONE_DAY,
		Architecture:          awslambda.Architecture_ARM_64(),
		CurrentVersionOptions: &awslambda.VersionOptions{},
	})

	/*
		httpApi := awsapigv2.NewHttpApi(stack, jsii.String("HttpApi"), &awsapigv2.HttpApiProps{
			ApiName:     jsii.String("HttpApi"),
			Description: jsii.String("Just a api gateway test."),
		})
	*/

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewApigatewayLambdaDdbStack(app, "ApigatewayLambdaDdbStack", &ApigatewayLambdaDdbStackProps{
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
