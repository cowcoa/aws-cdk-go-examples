package lib

import (
	"simple-pipeline/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

func NewSimplePipelineAppStage(scope constructs.Construct, id *string, props *awscdk.StageProps) awscdk.Stage {

	stage := awscdk.NewStage(scope, id, props)
	NewSimplePipelineLambdaStack(stage, config.StackName(scope)+"-LambdaStack", &awscdk.StackProps{
		Env: props.Env,
	})

	return stage
}
