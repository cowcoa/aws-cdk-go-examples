package lib

import (
	"simple-pipeline/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/pipelines"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewSimplePipelineStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// CodeStar connection policy for GitHub source.
	connStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("codestar-connections:UseConnection"),
		},
		Resources: &[]*string{
			jsii.String(config.CodeStarConnectionArn),
		},
	})

	// Construct CodePipeline
	pipeline := pipelines.NewCodePipeline(stack, jsii.String("SimplePipeline"), &pipelines.CodePipelineProps{
		PipelineName: jsii.String(*stack.StackName()),
		CodeBuildDefaults: &pipelines.CodeBuildOptions{
			RolePolicy: &[]awsiam.PolicyStatement{
				connStatement,
			},
			BuildEnvironment: &awscodebuild.BuildEnvironment{
				BuildImage:  awscodebuild.LinuxBuildImage_AMAZON_LINUX_2_3(),
				ComputeType: awscodebuild.ComputeType_MEDIUM,
			},
		},
		Synth: pipelines.NewShellStep(jsii.String("Synth"), &pipelines.ShellStepProps{
			InstallCommands: &[]*string{
				jsii.String("cd " + config.AppRootPath),
				jsii.String("source ./setup-codebuild-env.sh"),
				jsii.String("echo $PATH"),
				jsii.String("echo $GOROOT"),
				jsii.String("echo $GOPATH"),
				jsii.String("go version"),
				jsii.String("node --version"),
				jsii.String("uname -m"),
				jsii.String("go mod tidy"),
			},
			Commands: &[]*string{
				jsii.String("./cdk-cli-wrapper-dev.sh synth"),
			},
			Input: pipelines.CodePipelineSource_Connection(jsii.String(config.GitHubOwner+"/"+config.GitHubRepo), jsii.String(config.GitHubBranch), &pipelines.ConnectionSourceOptions{
				ConnectionArn:        jsii.String(config.CodeStarConnectionArn),
				CodeBuildCloneOutput: jsii.Bool(true),
				TriggerOnPush:        jsii.Bool(true),
			}),
			PrimaryOutputDirectory: jsii.String(config.AppRootPath + "/cdk.out"),
		}),
	})

	testingStage := pipeline.AddStage(NewSimplePipelineAppStage(stack, jsii.String("Test"), &awscdk.StageProps{
		Env: props.Env,
	}), &pipelines.AddStageOpts{})

	testingStage.AddPre(pipelines.NewManualApprovalStep(jsii.String("Approval"), &pipelines.ManualApprovalStepProps{}))

	return stack
}
