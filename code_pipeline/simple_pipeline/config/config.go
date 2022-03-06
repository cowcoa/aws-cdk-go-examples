package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	// Config your GitHub connection by CodeStart connection.
	// https://docs.aws.amazon.com/codepipeline/latest/userguide/connections-github.html
	CodeStarConnectionArn = "arn:aws:codestar-connections:ap-northeast-2:168228779762:connection/56f836ec-68cf-48be-a528-0f4e93544ceb"
	// GitHub info.
	GitHubOwner  = "cowcoa"
	GitHubRepo   = "aws-cdk-go-examples"
	GitHubBranch = "master"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "SimplePipeline"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}
