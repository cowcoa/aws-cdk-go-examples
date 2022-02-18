#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(aws configure get region)"

cdk bootstrap aws://${CDK_ACC}/${CDK_REGION}

$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"

# Deployment post process.
eks_ca_yaml=$SHELL_PATH/cdk.out/cluster-autoscaler-autodiscover.yaml
if [ "$CDK_CMD" == "deploy" ] && [ ! -f "$eks_ca_yaml" ]; then
    stack_name="$(cdk ls)"
    eks_cluster_name="$(aws cloudformation describe-stacks \
                      --stack-name ${stack_name} \
                      --query "Stacks[0].Outputs" \
                      --output json | jq -rc '.[] | select(.OutputKey=="EksClusterName") | .OutputValue '
                      )"
    eks_ca_role_arn="$(aws cloudformation describe-stacks \
                      --stack-name ${stack_name} \
                      --query "Stacks[0].Outputs" \
                      --output json | jq -rc '.[] | select(.OutputKey=="EksCARoleArn") | .OutputValue '
                      )"

    # Update kubeconfig
    aws eks update-kubeconfig --region ${CDK_REGION} --name ${eks_cluster_name}

    # Install Cluster Autoscaler
    cluster_autoscaler_version="1.21.2"
    curl -o ${eks_ca_yaml}.origin https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml
    sed -i '/<YOUR CLUSTER NAME>/a \            - --skip-nodes-with-system-pods=false' ${eks_ca_yaml}.origin
    sed -i '/<YOUR CLUSTER NAME>/a \            - --balance-similar-node-groups' ${eks_ca_yaml}.origin
    sed "s/<YOUR CLUSTER NAME>/${eks_cluster_name}/g" ${eks_ca_yaml}.origin > ${eks_ca_yaml}

    kubectl apply -f ${eks_ca_yaml}
    kubectl annotate --overwrite serviceaccount cluster-autoscaler -n kube-system eks.amazonaws.com/role-arn=${eks_ca_role_arn}
    kubectl patch deployment cluster-autoscaler -n kube-system -p '{"spec":{"template":{"metadata":{"annotations":{"cluster-autoscaler.kubernetes.io/safe-to-evict": "false"}}}}}'
    kubectl set image deployment cluster-autoscaler -n kube-system cluster-autoscaler=k8s.gcr.io/autoscaling/cluster-autoscaler:v${cluster_autoscaler_version}
fi

# Destroy post process.
if [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/
fi
