package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsopensearchservice"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/cowcoa/cdk/opensearch-cognito/config"
)

type OpensearchCognitoStackProps struct {
	awscdk.StackProps
}

func NewOpensearchCognitoStack(scope constructs.Construct, id string, props *OpensearchCognitoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create Cognito user pool
	userPool := awscognito.NewUserPool(stack, jsii.String("UserPool"), &awscognito.UserPoolProps{
		UserPoolName:    jsii.String(*stack.StackName() + "-UserPool"),
		AccountRecovery: awscognito.AccountRecovery_EMAIL_ONLY,
		AutoVerify: &awscognito.AutoVerifiedAttrs{
			Email: jsii.Bool(true),
			Phone: jsii.Bool(false),
		},
		Email: awscognito.UserPoolEmail_WithCognito(jsii.String("zxaws@amazon.com")),
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
		SelfSignUpEnabled: jsii.Bool(false),
		SignInAliases: &awscognito.SignInAliases{
			Email:             jsii.Bool(true),
			Username:          jsii.Bool(true),
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
	})
	// Add Domain and Hosted UI
	userPool.AddDomain(jsii.String("UserPoolDomain"), &awscognito.UserPoolDomainOptions{
		CognitoDomain: &awscognito.CognitoDomainOptions{
			DomainPrefix: jsii.String("opensearch-signin"),
		},
	})

	// Create Cognito identity pool
	identityPool := awscognito.NewCfnIdentityPool(stack, jsii.String("IdentityPool"), &awscognito.CfnIdentityPoolProps{
		IdentityPoolName:               jsii.String(*stack.StackName() + "-IdentityPool"),
		AllowUnauthenticatedIdentities: jsii.Bool(true),
	})
	authRole := awsiam.NewRole(stack, jsii.String("IdentityAuthRole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-IdentityAuthRole"),
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{
			"StringEquals": &map[string]string{
				"cognito-identity.amazonaws.com:aud": *identityPool.Ref(),
			},
			"ForAnyValue:StringLike": &map[string]string{
				"cognito-identity.amazonaws.com:amr": "authenticated",
			},
		}, jsii.String("sts:AssumeRoleWithWebIdentity")),
	})
	authRole.Node().AddDependency(identityPool)
	unauthRole := awsiam.NewRole(stack, jsii.String("IdentityUnAuthRole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-IdentityUnAuthRole"),
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{
			"StringEquals": &map[string]string{
				"cognito-identity.amazonaws.com:aud": *identityPool.Ref(),
			},
			"ForAnyValue:StringLike": &map[string]string{
				"cognito-identity.amazonaws.com:amr": "authenticated",
			},
		}, jsii.String("sts:AssumeRoleWithWebIdentity")),
	})
	unauthRole.Node().AddDependency(identityPool)

	// Add group to Cognito user pool
	masterUserRole := awsiam.NewRole(stack, jsii.String("OpensearchMasterRole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-OpensearchMasterRole"),
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{
			"StringEquals": &map[string]string{
				"cognito-identity.amazonaws.com:aud": *identityPool.Ref(),
			},
			"ForAnyValue:StringLike": &map[string]string{
				"cognito-identity.amazonaws.com:amr": "authenticated",
			},
		}, jsii.String("sts:AssumeRoleWithWebIdentity")),
	})
	limitedUserRole := awsiam.NewRole(stack, jsii.String("OpensearchLimitedRole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-OpensearchLimitedRole"),
		AssumedBy: awsiam.NewFederatedPrincipal(jsii.String("cognito-identity.amazonaws.com"), &map[string]interface{}{
			"StringEquals": &map[string]string{
				"cognito-identity.amazonaws.com:aud": *identityPool.Ref(),
			},
			"ForAnyValue:StringLike": &map[string]string{
				"cognito-identity.amazonaws.com:amr": "authenticated",
			},
		}, jsii.String("sts:AssumeRoleWithWebIdentity")),
	})
	awscognito.NewCfnUserPoolGroup(stack, jsii.String("MasterGroup"), &awscognito.CfnUserPoolGroupProps{
		UserPoolId: userPool.UserPoolId(),
		GroupName:  jsii.String("master-user-group"),
		Precedence: new(float64),
		RoleArn:    masterUserRole.RoleArn(),
	})
	awscognito.NewCfnUserPoolGroup(stack, jsii.String("LimitedGroup"), &awscognito.CfnUserPoolGroupProps{
		UserPoolId: userPool.UserPoolId(),
		GroupName:  jsii.String("limited-user-group"),
		Precedence: new(float64),
		RoleArn:    limitedUserRole.RoleArn(),
	})

	// Create Opensearch domain
	cognitoRole := awsiam.NewRole(stack, jsii.String("CognitoRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(*stack.StackName() + "-CognitoRole"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("opensearchservice.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonOpenSearchServiceCognitoAccess")),
		},
	})
	domain := awsopensearchservice.NewDomain(stack, jsii.String("Opensearch"), &awsopensearchservice.DomainProps{
		DomainName:    jsii.String("opensearch-cognito-poc"),
		Version:       awsopensearchservice.EngineVersion_OpenSearch(jsii.String("2.3")),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		ZoneAwareness: &awsopensearchservice.ZoneAwarenessConfig{
			AvailabilityZoneCount: jsii.Number(2),
			Enabled:               jsii.Bool(true),
		},
		Capacity: &awsopensearchservice.CapacityConfig{
			DataNodeInstanceType:   jsii.String("r6g.large.search"),
			DataNodes:              jsii.Number(2),
			MasterNodeInstanceType: jsii.String("c6g.large.search"),
			MasterNodes:            jsii.Number(3),
			WarmInstanceType:       jsii.String("ultrawarm1.medium.search"),
			WarmNodes:              jsii.Number(2),
		},
		Ebs: &awsopensearchservice.EbsOptions{
			Enabled:    jsii.Bool(true),
			Iops:       jsii.Number(3000),
			VolumeSize: jsii.Number(10),
			VolumeType: awsec2.EbsDeviceVolumeType_GP3,
		},
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
		FineGrainedAccessControl: &awsopensearchservice.AdvancedSecurityOptions{
			MasterUserArn: masterUserRole.RoleArn(),
		},
		CognitoDashboardsAuth: &awsopensearchservice.CognitoOptions{
			IdentityPoolId: identityPool.Ref(),
			UserPoolId:     userPool.UserPoolId(),
			Role:           cognitoRole,
		},
		EnforceHttps:         jsii.Bool(true),
		NodeToNodeEncryption: jsii.Bool(true),
		EncryptionAtRest: &awsopensearchservice.EncryptionAtRestOptions{
			Enabled: jsii.Bool(true),
		},
	})
	domain.Node().AddDependency(identityPool)
	domain.Node().AddDependency(userPool)

	// Get Cognito user pool's client id that added by OpenSearch Domain automatically.
	userPoolClients := customresources.NewAwsCustomResource(stack, jsii.String("ClientIdResource"), &customresources.AwsCustomResourceProps{
		Policy: customresources.AwsCustomResourcePolicy_FromSdkCalls(&customresources.SdkCallsPolicyOptions{
			Resources: &[]*string{
				userPool.UserPoolArn(),
			},
		}),
		OnCreate: &customresources.AwsSdkCall{
			Service: jsii.String("CognitoIdentityServiceProvider"),
			Action:  jsii.String("listUserPoolClients"),
			Parameters: &map[string]interface{}{
				"UserPoolId": userPool.UserPoolId(),
			},
			PhysicalResourceId: customresources.PhysicalResourceId_Of(jsii.String(fmt.Sprintf("ClientId-%s", *domain.DomainName()))),
		},
	})
	userPoolClients.Node().AddDependency(domain)
	clientId := userPoolClients.GetResponseField(jsii.String("UserPoolClients.0.ClientId"))

	// Modify identity pool's role mapping
	providerName := fmt.Sprintf("cognito-idp.%s.amazonaws.com/%s:%s", *stack.Region(), *userPool.UserPoolId(), *clientId)
	awscognito.NewCfnIdentityPoolRoleAttachment(stack, jsii.String("IdentityRoleAttach"), &awscognito.CfnIdentityPoolRoleAttachmentProps{
		IdentityPoolId: identityPool.Ref(),
		Roles: &map[string]interface{}{
			"authenticated":   authRole.RoleArn(),
			"unauthenticated": unauthRole.RoleArn(),
		},
		RoleMappings: awscdk.NewCfnJson(stack, jsii.String("RoleMappings"), &awscdk.CfnJsonProps{
			Value: &map[string]interface{}{
				providerName: &map[string]string{
					"Type":                    "Token",
					"AmbiguousRoleResolution": "Deny",
				},
			},
		}),
	})
	awscdk.NewCfnOutput(stack, jsii.String("providerName"), &awscdk.CfnOutputProps{
		Value: jsii.String(providerName),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewOpensearchCognitoStack(app, config.StackName(app), &OpensearchCognitoStackProps{
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
