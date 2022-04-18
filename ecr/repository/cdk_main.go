package main

import (
	"eks-repository/config"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/constructs-go/constructs/v10"
)

type EcrRepositoryStackProps struct {
	awscdk.StackProps
}

func NewEcrRepositoryStack(scope constructs.Construct, id string, props *EcrRepositoryStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	repo := awsecr.NewRepository(stack, jsii.String("ECRRepository"), &awsecr.RepositoryProps{
		RepositoryName:     jsii.String(config.EcrRepoName(stack)),
		RemovalPolicy:      awscdk.RemovalPolicy_DESTROY,
		ImageTagMutability: awsecr.TagMutability_MUTABLE,
		ImageScanOnPush:    jsii.Bool(false),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ECRRepositoryName"), &awscdk.CfnOutputProps{
		Value: repo.RepositoryName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ECRRepositoryURI"), &awscdk.CfnOutputProps{
		Value: repo.RepositoryUri(),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewEcrRepositoryStack(app, config.StackName(app), &EcrRepositoryStackProps{
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
