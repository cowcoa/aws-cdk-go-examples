package main

import (
	"os"
	"simple-build/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CodeBuildCdkStackProps struct {
	awscdk.StackProps
}

func NewCodeBuildCdkStack(scope constructs.Construct, id string, props *CodeBuildCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	/*
		gitHubSource := awscodebuild.Source_GitHub(&awscodebuild.GitHubSourceProps{
			Owner:       jsii.String("cowcoa"),
			Repo:        jsii.String("aws-cdk-go-examples"),
			BranchOrRef: jsii.String("master"),
		})
	*/

	/*
		ecrRepo := awsecr.Repository_FromRepositoryName(stack, jsii.String("EcrBuildRepo"), jsii.String("public.ecr.aws/d8k9t1f2/codebuild-linux-arm64"))
		awscdk.NewCfnOutput(stack, jsii.String("EcrBuildRepoName"), &awscdk.CfnOutputProps{
			Value: ecrRepo.RepositoryUri(),
		})
	*/

	/*
		awscodebuild.NewProject(stack, jsii.String("CodeBuildProject"), &awscodebuild.ProjectProps{
			BuildSpec:                           awscodebuild.BuildSpec_FromSourceFilename(jsii.String("code_build/simple_build/app/buildspec_aarch64.yml")),
			CheckSecretsInPlainTextEnvVariables: jsii.Bool(true),
			ConcurrentBuildLimit:                jsii.Number(1),
			Description:                         jsii.String("Test"),
			Environment: &awscodebuild.BuildEnvironment{
				// BuildImage:  awscodebuild.LinuxBuildImage_FromEcrRepository(ecrRepo, jsii.String("v1")),
				BuildImage:  awscodebuild.LinuxBuildImage_FromDockerRegistry(jsii.String("public.ecr.aws/d8k9t1f2/codebuild-linux-arm64"), &awscodebuild.DockerImageOptions{}),
				ComputeType: awscodebuild.ComputeType_SMALL,
			},
			ProjectName: jsii.String("HabbyTest"),
			Source:      gitHubSource,
		})
	*/

	////

	env := awscodebuild.CfnProject_EnvironmentProperty{
		ComputeType: jsii.String("BUILD_GENERAL1_SMALL"),
		Image:       jsii.String("public.ecr.aws/d8k9t1f2/codebuild-linux-arm64:v1"),
		Type:        jsii.String("ARM_CONTAINER"),
	}
	awscodebuild.NewCfnProject(stack, jsii.String("another"), &awscodebuild.CfnProjectProps{
		Artifacts: awscodebuild.CfnProject_ArtifactsProperty{
			Type: jsii.String("NO_ARTIFACTS"),
		},
		Environment: env,
		ServiceRole: jsii.String("arn:aws:iam::168228779762:role/service-role/codebuild-codebuild-demo-project-service-role"),
		Source: awscodebuild.CfnProject_SourceProperty{
			Type:     jsii.String("GITHUB"),
			Location: jsii.String("https://github.com/cowcoa/aws-cdk-go-examples"),
		},
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewCodeBuildCdkStack(app, config.StackName, &CodeBuildCdkStackProps{
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
