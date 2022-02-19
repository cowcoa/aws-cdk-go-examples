package main

import (
	"os"
	"simple-resource/config"
	"time"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CustomResCdkStackProps struct {
	awscdk.StackProps
}

func NewCustomResCdkStack(scope constructs.Construct, id string, props *CustomResCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	// Create lambda role
	lambdaRole := awsiam.NewRole(stack, jsii.String(config.RoleName), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMFullAccess")),
		},
		RoleName: jsii.String(*stack.StackName() + "-" + config.RoleName),
	})

	// Create lambda function
	lambdaFunc := awslambda.NewFunction(stack, jsii.String(config.FuncionName), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-" + config.FuncionName),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(config.MemorySize),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(config.MaxDuration)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String(config.CodePath), nil),
		Handler:      jsii.String(config.Handler),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_DAY,
	})

	crProvider := customresources.NewProvider(stack, jsii.String("MyProvider"), &customresources.ProviderProps{
		OnEventHandler:       lambdaFunc,
		LogRetention:         awslogs.RetentionDays_ONE_WEEK,
		ProviderFunctionName: jsii.String(*stack.StackName() + "-" + "MyProvider"),
	})

	// After the first deploy. If you:
	// Update the properties other than PhysicalResourceId, your Lambda will receive Update event.
	// Update the PhysicalResourceId properties, your Lambda will receive Update & Delete events.
	// Update NewCustomResource Id, your Lambda will receive Create & Delete events.
	customRes := awscdk.NewCustomResource(stack, jsii.String("MyCustomRes2"), &awscdk.CustomResourceProps{
		ServiceToken: crProvider.ServiceToken(),
		ResourceType: jsii.String("Custom::CowCustomRes"),
		Properties: &map[string]interface{}{
			"PhysicalResourceId": "abcd1234",
			"SSMParamName":       "my-parameter",
			"SSMParamValue":      "AWS yyds!",
		},
	})

	ssmParamName := customRes.GetAtt(jsii.String("SSMParamName")).ToString()
	awscdk.NewCfnOutput(stack, jsii.String("SSMParamNameOutput"), &awscdk.CfnOutputProps{
		Value: ssmParamName,
	})
	ssmParamValue := customRes.GetAtt(jsii.String("SSMParamValue")).ToString()
	awscdk.NewCfnOutput(stack, jsii.String("SSMParamValueOutput"), &awscdk.CfnOutputProps{
		Value: ssmParamValue,
	})

	currentTime := time.Now()
	getParameter := customresources.NewAwsCustomResource(stack, jsii.String("GetParameter"), &customresources.AwsCustomResourceProps{
		OnUpdate: &customresources.AwsSdkCall{
			Service: jsii.String("SSM"),
			Action:  jsii.String("getParameter"),
			Parameters: &map[string]interface{}{
				"Name":           ssmParamName,
				"WithDecryption": true,
			},
			PhysicalResourceId: customresources.PhysicalResourceId_Of(jsii.String(currentTime.String())),
		},
		Policy: customresources.AwsCustomResourcePolicy_FromSdkCalls(&customresources.SdkCallsPolicyOptions{
			Resources: customresources.AwsCustomResourcePolicy_ANY_RESOURCE(),
		}),
	})

	param := getParameter.GetResponseField(jsii.String("Parameter.Value"))
	awscdk.NewCfnOutput(stack, jsii.String("GetParameterOutput"), &awscdk.CfnOutputProps{
		Value: param,
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewCustomResCdkStack(app, config.StackName, &CustomResCdkStackProps{
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
