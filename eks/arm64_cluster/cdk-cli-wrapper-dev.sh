#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(jq -r .context.deploymentRegion ./cdk.json)"

# Check execution env.
if [ -z $CODEBUILD_BUILD_ID ]
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

# CDK command.
# Valid deploymentStage are: [dev, prod]
# set -- "$@" "--context deploymentStage=dev"
$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"

# CDK command post-process.
init_state_file=$SHELL_PATH/cdk.out/init.state
if [ $? -eq 0 ] && [ "$CDK_CMD" == "deploy" ] && [ ! -f "$init_state_file" ]; then
    # Update kubeconfig
    stack_name="$(jq -r .context.stackName ./cdk.json)"
    eks_cluster_name="$(jq -r .context.clusterName ./cdk.json)"
    aws eks update-kubeconfig --region ${CDK_REGION} --name ${eks_cluster_name}

    # Add the following annotation to your service accounts to use the AWS Security Token Service AWS Regional endpoint, 
    # rather than the global endpoint.
    kubectl annotate serviceaccount -n kube-system aws-node eks.amazonaws.com/sts-regional-endpoints=true

    # Change init state.
    if [ $? -eq 0 ]; then
        echo "0" > $init_state_file
    fi
fi

# Destroy post process.
if [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/
fi
