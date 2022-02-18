#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(aws configure get region)"

cdk bootstrap aws://${CDK_ACC}/${CDK_REGION}

# Deploy pre-process.
# Compile code.
pushd function
GOARCH=amd64 GOOS=linux go build main.go
popd

$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"

# Destroy post-process.
if [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/
fi
