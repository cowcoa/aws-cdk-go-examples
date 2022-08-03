#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(jq -r .context.deploymentRegion ./cdk.json)"
if [ -z "$CDK_REGION" ]; then
    CDK_REGION="$(aws configure get region)"
fi

# CDK command pre-process.
if [ "$CDK_CMD" == "deploy" ]; then
    # Remove PVRE hook auto-added policy before executing destroy.
    node_role_name="$(jq -r .context.stackName ./cdk.json)-$(jq -r .context.targetArch ./cdk.json)-ClusterNodeRole"
    aws iam detach-role-policy --role-name $node_role_name --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
fi

if [ "$CDK_CMD" == "destroy" ]; then
    echo ""
fi

# CDK command.
# Valid deploymentStage are: [DEV, PROD]
set -- "$@" "-c" "deploymentStage=DEV"
$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"
cdk_exec_result=$?

# CDK command post-process.
if [ "$CDK_CMD" == "deploy" ]; then
    echo ""
fi

if [ "$CDK_CMD" == "destroy" ]; then
    echo ""
fi
