#!/bin/bash

# Get execution info
script_name=$(basename $0)
arg_count=$#
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

lambda_handler="/var/task/"
ecr_repo="apig-lambda-ddb"

if test $arg_count -eq 1; then
    lambda_handler+="$1"
else
    echo "Script for running RIE in local."
    echo ""
    echo "Usage:"
    echo ""
    echo "      $script_name <lambda_function_name>"
    echo ""
    echo "Examples:"
    echo ""
    echo "      $script_name put-chat-records"
    echo ""
    exit 0
fi

echo "Build docker image..."
docker build -t $ecr_repo $SHELL_PATH

AWS_ACCESS_KEY_ID=$(aws --profile default configure get aws_access_key_id)
AWS_SECRET_ACCESS_KEY=$(aws --profile default configure get aws_secret_access_key)
AWS_REGION="$(aws configure get region)"
DYNAMODB_TABLE="CdkGolangExample-ApiGtwLambdaDdb-ChatTable"
DYNAMODB_GSI="ChatTableGSI"

echo "Lambda runtime emulator is listening port 9000..."
docker run \
        -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
        -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
        -e AWS_REGION=$AWS_REGION \
        -e DYNAMODB_TABLE=$DYNAMODB_TABLE \
        -e DYNAMODB_GSI=$DYNAMODB_GSI \
        -p 9000:8080 ${ecr_repo}:latest \
        /var/task/"$@"
