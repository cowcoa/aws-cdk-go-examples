package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewSimplePipelineLambdaStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	awslambda.NewFunction(stack, jsii.String("SimpleFunction"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "SimpleFunction"),
		Runtime:      awslambda.Runtime_NODEJS_14_X(),
		Handler:      jsii.String("index.handler"),
		Code:         awslambda.AssetCode_FromInline(jsii.String(`exports.handler = _ => "Hello, CDK";`)),
	})

	return stack
}
