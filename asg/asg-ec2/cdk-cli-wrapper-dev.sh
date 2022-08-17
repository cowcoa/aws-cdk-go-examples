#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(jq -r .context.deploymentRegion ./cdk.json)"

# Check execution env.
if [ -z "$CODEBUILD_BUILD_ID" ]
then
    if [ -z "$CDK_REGION" ]; then
        CDK_REGION="$(aws configure get region)"
    fi

    echo "Run bootstrap..."
    export CDK_NEW_BOOTSTRAP=1 
    npx cdk bootstrap aws://${CDK_ACC}/${CDK_REGION} --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
else
    CDK_REGION=$AWS_DEFAULT_REGION
fi

# CDK command pre-process.

# Destroy pre-process.
if [ "$CDK_CMD" == "destroy" ]; then
    # Remove PVRE hook auto-added policy before executing destroy.
    # Feel free to remove this code block if you are not an AWS employee.
    role_name="$(jq -r .context.stackName ./cdk.json)-${CDK_REGION}-EC2Role"
    policy_arn="$(aws iam list-attached-role-policies --role-name ${role_name} --query 'AttachedPolicies[?PolicyName==`AmazonSSMManagedInstanceCore`].PolicyArn' --output text)"

    if [ ! -z "$policy_arn" ]; then
        aws iam detach-role-policy --role-name $role_name --policy-arn $policy_arn
    fi
fi

# CDK command.
# Valid deploymentStage are: [DEV, PROD]
set -- "$@" "-c" "deploymentStage=DEV"
$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"
cdk_exec_result=$?

# CDK command post-process.
