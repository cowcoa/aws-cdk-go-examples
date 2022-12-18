package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsopensearchservice"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/cowcoa/cdk/opensearch-cognito/config"
)

type DayOneAccountStackProps struct {
	awscdk.StackProps
}

func NewDayOneAccountStack(scope constructs.Construct, id string, props *DayOneAccountStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	awscdk.NewCfnOutput(stack, jsii.String("region"), &awscdk.CfnOutputProps{
		Value:       stack.Region(),
		Description: jsii.String("Region of this deployment."),
	})

	// Create role for lambda function.
	/*
		cognitoTriggerRole := awsiam.NewRole(stack, jsii.String("CognitoEventRole"), &awsiam.RoleProps{
			RoleName:  jsii.String("DayOne-CognitoEventRole-" + string(config.DeploymentStage(stack))),
			AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
			ManagedPolicies: &[]awsiam.IManagedPolicy{
				awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchFullAccess")),
				awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonCognitoPowerUser")),
			},
		})

		// Create Cognito's post confirmation Lambda trigger functions.
		signUpConfirmFunction := awslambda.NewFunction(stack, jsii.String("HandlePostConfirmEvent"), &awslambda.FunctionProps{
			FunctionName: jsii.String("DayOne-HandlePostConfirmEvent-" + string(config.DeploymentStage(stack))),
			Runtime:      awslambda.Runtime_GO_1_X(),
			MemorySize:   jsii.Number(128),
			Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
			Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/handle-post-confirm-event/."), nil),
			Handler:      jsii.String("handle-post-confirm-event"),
			Architecture: awslambda.Architecture_X86_64(),
			Role:         cognitoTriggerRole,
			LogRetention: awslogs.RetentionDays_ONE_WEEK,
				Environment: &map[string]*string{
					"DYNAMODB_TABLE": playersTable.TableName(),
				},
		})
	*/

	// Create Cognito player pool
	playerPool := awscognito.NewUserPool(stack, jsii.String("PlayerPool"), &awscognito.UserPoolProps{
		UserPoolName:    jsii.String("OpenSearch-ManagerPool-" + string(config.DeploymentStage(stack))),
		AccountRecovery: awscognito.AccountRecovery_EMAIL_ONLY,
		AutoVerify: &awscognito.AutoVerifiedAttrs{
			Email: jsii.Bool(true),
			Phone: jsii.Bool(false),
		},
		Email: awscognito.UserPoolEmail_WithCognito(jsii.String("zxaws@gmail.com")),
		Mfa:   awscognito.Mfa_OFF,
		PasswordPolicy: &awscognito.PasswordPolicy{
			MinLength:            jsii.Number(6),
			RequireDigits:        jsii.Bool(false),
			RequireLowercase:     jsii.Bool(false),
			RequireSymbols:       jsii.Bool(false),
			RequireUppercase:     jsii.Bool(false),
			TempPasswordValidity: awscdk.Duration_Days(jsii.Number(7)),
		},
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		SelfSignUpEnabled: jsii.Bool(true),
		SignInAliases: &awscognito.SignInAliases{
			Email:             jsii.Bool(true),
			Username:          jsii.Bool(false),
			Phone:             jsii.Bool(false),
			PreferredUsername: jsii.Bool(false),
		},
		SignInCaseSensitive: jsii.Bool(false),
		StandardAttributes: &awscognito.StandardAttributes{
			Email: &awscognito.StandardAttribute{
				Mutable:  jsii.Bool(false),
				Required: jsii.Bool(true),
			},
		},
		/*
			LambdaTriggers: &awscognito.UserPoolTriggers{
				PostConfirmation: signUpConfirmFunction,
			},
		*/
	})
	awscdk.NewCfnOutput(stack, jsii.String("poolId"), &awscdk.CfnOutputProps{
		Value:       playerPool.UserPoolId(),
		Description: jsii.String("Players' UserPool Id."),
	})

	// Add domain and hosted UI
	playerPool.AddDomain(jsii.String("PlayerPoolDomain"), &awscognito.UserPoolDomainOptions{
		CognitoDomain: &awscognito.CognitoDomainOptions{
			DomainPrefix: jsii.String("opensearch-cow"),
		},
	})

	// userpool group
	masterUserRole := awsiam.NewRole(stack, jsii.String("OpenSearchMasterRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{}, jsii.String("sts:AssumeRoleWithWebIdentity")),
		RoleName:  jsii.String("opensearch-master-role-123"),
	})
	limitedUserRole := awsiam.NewRole(stack, jsii.String("OpenSearchLimitedRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{}, jsii.String("sts:AssumeRoleWithWebIdentity")),
		RoleName:  jsii.String("opensearch-limited-role-123"),
	})
	awscognito.NewCfnUserPoolGroup(stack, jsii.String("mastergroup"), &awscognito.CfnUserPoolGroupProps{
		UserPoolId: playerPool.UserPoolId(),
		GroupName:  jsii.String("master-user-group"),
		Precedence: new(float64),
		RoleArn:    masterUserRole.RoleArn(),
	})
	awscognito.NewCfnUserPoolGroup(stack, jsii.String("limitedgroup"), &awscognito.CfnUserPoolGroupProps{
		UserPoolId: playerPool.UserPoolId(),
		GroupName:  jsii.String("limited-user-group"),
		Precedence: new(float64),
		RoleArn:    limitedUserRole.RoleArn(),
	})

	identityPool := awscognito.NewCfnIdentityPool(stack, jsii.String("IdentityPool"), &awscognito.CfnIdentityPoolProps{
		AllowUnauthenticatedIdentities: jsii.Bool(true),
		IdentityPoolName:               jsii.String("opensearch-test-pool"),
	})
	authRole := awsiam.NewRole(stack, jsii.String("IdentityAuthRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{
			"StringEquals": &map[string]string{
				"cognito-identity.amazonaws.com:aud": *identityPool.Ref(),
			},
			"ForAnyValue:StringLike": &map[string]string{
				"cognito-identity.amazonaws.com:amr": "authenticated",
			},
		}, jsii.String("sts:AssumeRoleWithWebIdentity")),
		RoleName: jsii.String("identity-auth-role-123"),
	})
	authRole.Node().AddDependency(identityPool)
	awscognito.NewCfnIdentityPoolRoleAttachment(stack, jsii.String("IdentityRoleAttach"), &awscognito.CfnIdentityPoolRoleAttachmentProps{
		IdentityPoolId: identityPool.Ref(),
		Roles: &map[string]interface{}{
			"authenticated": authRole.RoleArn(),
		},
	})

	// Create cluster node role.
	/*
		idPoolCondBytes := []byte(fmt.Sprintf(`"cognito-identity.amazonaws.com:aud": "%s"`, *identityPool.Ref()))
		var idPoolCondJsonMap map[string]interface{}
		json.Unmarshal([]byte(idPoolCondBytes), &idPoolCondJsonMap)
		authCondBytes := []byte(`"cognito-identity.amazonaws.com:amr": "authenticated"`)
		var authCondJsonMap map[string]interface{}
		json.Unmarshal([]byte(authCondBytes), &authCondJsonMap)
	*/
	/*&map[string]interface{}{
		"StringEquals":           idPoolCondJsonMap,
		"ForAnyValue:StringLike": authCondJsonMap,
	},*/

	// opensearch
	cognitoRole := awsiam.NewRole(stack, jsii.String("CognitoRole"), &awsiam.RoleProps{
		RoleName:  jsii.String("opensearch-CognitoRole-" + string(config.DeploymentStage(stack))),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("opensearchservice.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonOpenSearchServiceCognitoAccess")),
		},
	})
	domain := awsopensearchservice.NewDomain(stack, jsii.String("myopensearch"), &awsopensearchservice.DomainProps{
		Version: awsopensearchservice.EngineVersion_OpenSearch(jsii.String("2.3")),
		AccessPolicies: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewAnyPrincipal(),
				},
				Actions: &[]*string{
					jsii.String("es:*"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
		},
		Capacity: &awsopensearchservice.CapacityConfig{
			DataNodeInstanceType:   jsii.String("r6g.large.search"),
			DataNodes:              jsii.Number(2),
			MasterNodeInstanceType: jsii.String("c6g.large.search"),
			MasterNodes:            jsii.Number(3),
			WarmInstanceType:       jsii.String("ultrawarm1.medium.search"),
			WarmNodes:              jsii.Number(2),
		},
		CognitoDashboardsAuth: &awsopensearchservice.CognitoOptions{
			IdentityPoolId: identityPool.Ref(),
			UserPoolId:     playerPool.UserPoolId(),
			Role:           cognitoRole,
		},
		DomainName: jsii.String("my-test-env"),
		Ebs: &awsopensearchservice.EbsOptions{
			Enabled:    jsii.Bool(true),
			Iops:       jsii.Number(3000),
			VolumeSize: jsii.Number(10),
			VolumeType: awsec2.EbsDeviceVolumeType_GP3,
		},
		FineGrainedAccessControl: &awsopensearchservice.AdvancedSecurityOptions{
			MasterUserArn: masterUserRole.RoleArn(),
		},
		NodeToNodeEncryption: jsii.Bool(true),
		EnforceHttps:         jsii.Bool(true),
		EncryptionAtRest: &awsopensearchservice.EncryptionAtRestOptions{
			Enabled: jsii.Bool(true),
		},
		ZoneAwareness: &awsopensearchservice.ZoneAwarenessConfig{
			AvailabilityZoneCount: jsii.Number(2),
			Enabled:               jsii.Bool(true),
		},
	})
	domain.Node().AddDependency(identityPool)
	domain.Node().AddDependency(playerPool)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewDayOneAccountStack(app, config.StackName(app), &DayOneAccountStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

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
