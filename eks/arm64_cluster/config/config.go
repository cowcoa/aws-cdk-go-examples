package config

import (
	"reflect"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseks"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "MyEKSClusterStack"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, change EKS cluster name by 'cdk.json/context/clusterName'.
func ClusterName(scope constructs.Construct) string {
	clusterName := "MyEKSCluster"

	ctxValue := scope.Node().TryGetContext(jsii.String("clusterName"))
	if v, ok := ctxValue.(string); ok {
		clusterName = v
	}

	return clusterName
}

// DO NOT modify this function, change EC2 key pair name by 'cdk.json/context/keyPairName'.
func KeyPairName(scope constructs.Construct) string {
	keyPairName := "MyKeyPair"

	ctxValue := scope.Node().TryGetContext(jsii.String("keyPairName"))
	if v, ok := ctxValue.(string); ok {
		keyPairName = v
	}

	return keyPairName
}

// Master users in k8s system:masters. All users must be existing IAM Users
// DO NOT modify this function, change master IAM users by 'cdk.json/context/masterUsers'.
func MasterUsers(scope constructs.Construct) []string {
	var masterUsers []string

	ctxValue := scope.Node().TryGetContext(jsii.String("masterUsers"))
	iamUsers := reflect.ValueOf(ctxValue)
	if iamUsers.Kind() != reflect.Slice {
		return masterUsers
	}

	for i := 0; i < iamUsers.Len(); i++ {
		user := iamUsers.Index(i).Interface().(string)
		masterUsers = append(masterUsers, user)
	}

	return masterUsers
}

// Deployment stage config
type DeploymentStageType string

const (
	DeploymentStage_DEV  DeploymentStageType = "DEV"
	DeploymentStage_PROD DeploymentStageType = "PROD"
)

// DO NOT modify this function, change EKS cluster name by 'cdk-cli-wrapper-dev.sh/--context deploymentStage='.
func DeploymentStage(scope constructs.Construct) DeploymentStageType {
	deploymentStage := DeploymentStage_PROD

	ctxValue := scope.Node().TryGetContext(jsii.String("deploymentStage"))
	if v, ok := ctxValue.(string); ok {
		deploymentStage = DeploymentStageType(v)
	}

	return deploymentStage
}

// VPC config
const vpcMask = 16
const vpcIpv4 = "192.168.0.0"

var VpcCidr = vpcIpv4 + "/" + strconv.Itoa(vpcMask)

const MaxAzs = 3
const SubnetMask = vpcMask + MaxAzs

// DO NOT modify this function, change ExternalDNS role ARN by 'cdk.json/context/externalDnsRole'.
func ExternalDnsRole(scope constructs.Construct) string {
	// The 'cdk.json/context/externalDnsRole' is a role defined in target account with permission policy:
	/*
		{
			"Version": "2012-10-17",
			"Statement": [
				{
				"Effect": "Allow",
				"Action": [
					"route53:ChangeResourceRecordSets"
				],
				"Resource": [
					"arn:aws:route53:::hostedzone/*"
				]
				},
				{
				"Effect": "Allow",
				"Action": [
					"route53:ListHostedZones",
					"route53:ListResourceRecordSets"
				],
				"Resource": [
					"*"
				]
				}
			]
		}
	*/
	// and trust policy:
	/*
		{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": "Role ARN in cluster account"
					},
					"Action": "sts:AssumeRole",
					"Condition": {}
				}
			]
		}
	*/
	externalDnsRole := ""

	ctxValue := scope.Node().TryGetContext(jsii.String("externalDnsRole"))
	if v, ok := ctxValue.(string); ok {
		externalDnsRole = v
	}

	return externalDnsRole
}

// Init K8s Ingress/Service external DNS resources
func InitExternalDns(stack awscdk.Stack, cluster awseks.Cluster) {
	// The 'cdk.json/context/externalDnsRole' is a role defined in target account with permission policy:
	/*
		{
			"Version": "2012-10-17",
			"Statement": [
				{
				"Effect": "Allow",
				"Action": [
					"route53:ChangeResourceRecordSets"
				],
				"Resource": [
					"arn:aws:route53:::hostedzone/*"
				]
				},
				{
				"Effect": "Allow",
				"Action": [
					"route53:ListHostedZones",
					"route53:ListResourceRecordSets"
				],
				"Resource": [
					"*"
				]
				}
			]
		}
	*/
	// and trust policy:
	/*
		{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": "Role ARN in cluster account"
					},
					"Action": "sts:AssumeRole",
					"Condition": {}
				}
			]
		}
	*/
	externalDnsRole := ""
	ctxValue := stack.Node().TryGetContext(jsii.String("externalDnsRole"))
	if v, ok := ctxValue.(string); ok {
		externalDnsRole = v
	}

	var externalDnsPolicy awsiam.PolicyDocument
	// If the 'cdk.json/context/externalDnsRole' is not empty, we need to define a policy to assume that target role.
	if len(externalDnsRole) > 0 {
		externalDnsPolicy = awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
			AssignSids: jsii.Bool(true),
			Statements: &[]awsiam.PolicyStatement{
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Actions: &[]*string{
						jsii.String("sts:AssumeRole"),
					},
					Resources: &[]*string{
						jsii.String(externalDnsRole),
					},
				}),
			},
		})
	} else { // Otherwise, we define a policy with the corresponding permissions.
		externalDnsPolicy = awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
			AssignSids: jsii.Bool(true),
			Statements: &[]awsiam.PolicyStatement{
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Actions: &[]*string{
						jsii.String("route53:ChangeResourceRecordSets"),
					},
					Resources: &[]*string{
						jsii.String("arn:aws:route53:::hostedzone/*"),
					},
				}),
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Actions: &[]*string{
						jsii.String("route53:ListHostedZones"),
						jsii.String("route53:ListResourceRecordSets"),
					},
					Resources: &[]*string{
						jsii.String("*"),
					},
				}),
			},
		})
	}

	// Create K8s service account in default namespace.
	externalDnsSa := awseks.NewServiceAccount(stack, jsii.String("ExternalDNSSA"), &awseks.ServiceAccountProps{
		Name:      jsii.String("external-dns"),
		Cluster:   cluster,
		Namespace: jsii.String("default"),
	})

	// Associate the policy with that service account.
	awsiam.NewPolicy(stack, jsii.String("ExternalDNSPolicy"), &awsiam.PolicyProps{
		Document:   externalDnsPolicy,
		PolicyName: jsii.String(*stack.StackName() + "-ExternalDNSPolicy"),
		Roles: &[]awsiam.IRole{
			externalDnsSa.Role(),
		},
	})

	// You need to update the target role's Principal of trust policy with this ARN
	awscdk.NewCfnOutput(stack, jsii.String("ExternalDNSRoleArn"), &awscdk.CfnOutputProps{
		Value: externalDnsSa.Role().RoleArn(),
	})
}
