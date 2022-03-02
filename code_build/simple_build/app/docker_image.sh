#!/bin/bash

# Habby PoC docker image generator.
# Copyright (c) 2022 AWS. All Rights Reserved.
# Created by Xin Zhang <zxaws@amazon.com>

# docker container run --name gin-test -p 8000:8080 168228779762.dkr.ecr.ap-northeast-2.amazonaws.com/ekscdkstack-repo:latest

# Get script location.
SHELL_PATH=$(cd "$(dirname "${0}")";pwd)

# Get stack name
stack_name=""
pushd ${SHELL_PATH}/../
stack_name="$(cdk ls)"
popd

# Declare variables
aws_account_id="$(aws sts get-caller-identity --output text --query 'Account')"
deployment_region="$(aws configure get region)"
habby_gin_poc_repo="$(aws cloudformation describe-stacks \
                    --stack-name ${stack_name} \
                    --query "Stacks[0].Outputs" \
                    --output json | jq -rc '.[] | select(.OutputKey=="EcrRepositoryName") | .OutputValue '
                    )"
habby_gin_poc_repo_uri=${aws_account_id}.dkr.ecr.${deployment_region}.amazonaws.com/${habby_gin_poc_repo}
docker_file="Dockerfile_arm64"

echo "Build docker image..."
docker build -t ${habby_gin_poc_repo} -f ${SHELL_PATH}/${docker_file} ${SHELL_PATH}

image_tag="$(echo $(date '+%Y.%m.%d.%H%M%S' -d '+8 hours'))"
#image_tag="$(printf '%(%Y.%m.%d)T\n' -1)"

echo "habby_gin_poc_repo: $habby_gin_poc_repo"
echo "habby_gin_poc_repo_uri: $habby_gin_poc_repo_uri"

echo "Upload docker image to ECR..."
eval "aws ecr get-login-password --region ${deployment_region} | docker login --username AWS --password-stdin ${aws_account_id}.dkr.ecr.${deployment_region}.amazonaws.com"
#DOCKER_LOGIN_CMD=$(aws ecr get-login --no-include-email --region $deployment_region)
#eval "${DOCKER_LOGIN_CMD}"
docker tag $habby_gin_poc_repo:latest $habby_gin_poc_repo_uri:$image_tag
docker push $habby_gin_poc_repo_uri:$image_tag
docker tag $habby_gin_poc_repo:latest $habby_gin_poc_repo_uri:latest
docker push $habby_gin_poc_repo_uri:latest

echo
echo "Done"
