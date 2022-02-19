#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(aws configure get region)"
LAMBDA_PATH="function"
LAMBDA_HANDLER="main"

cdk bootstrap aws://${CDK_ACC}/${CDK_REGION}

# Deploy pre-process.
# Compile lambda function.
pushd ${LAMBDA_PATH}
if [ -f ${LAMBDA_HANDLER} ]; then
    rm -rf ${LAMBDA_HANDLER}
fi
GOARCH=amd64 GOOS=linux go build main.go
result=$?
if test $result -ne 0; then
    echo "Failed to build custom resource lambda function."
    exit $result
fi
popd

$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"

# Destroy post-process.
if [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/
fi
