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

# Destroy pre-process.
if [ "$CDK_CMD" == "destroy" ]; then
    # Remove PVRE hook auto-added policy before executing destroy.
    node_role_name="$(jq -r .context.stackName ./cdk.json)-ClusterNodeRole"
    aws iam detach-role-policy --role-name $node_role_name --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
fi

# CDK command.
# Valid deploymentStage are: [DEV, PROD]
set -- "$@" "-c" "deploymentStage=DEV"
$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"
cdk_exec_result=$?

# CDK command post-process.
init_state_file=$SHELL_PATH/cdk.out/init.state
if [ $cdk_exec_result -eq 0 ] && [ "$CDK_CMD" == "deploy" ] && [ ! -f "$init_state_file" ]; then
    # Update kubeconfig
    echo "Update kubeconfig..."
    stack_name="$(jq -r .context.stackName ./cdk.json)"
    eks_cluster_name="$(jq -r .context.clusterName ./cdk.json)"
    aws eks update-kubeconfig --region ${CDK_REGION} --name ${eks_cluster_name}

    # Add the following annotation to your service accounts to use the AWS Security Token Service AWS Regional endpoint, 
    # rather than the global endpoint.
    echo "Update service account annotate..."
    kubectl annotate serviceaccount -n kube-system aws-node eks.amazonaws.com/sts-regional-endpoints=true
    kubectl annotate serviceaccount -n kube-system aws-load-balancer-controller eks.amazonaws.com/sts-regional-endpoints=true
    kubectl patch deployment cluster-autoscaler-aws-cluster-autoscaler -n kube-system -p '{"spec":{"template":{"metadata":{"annotations":{"cluster-autoscaler.kubernetes.io/safe-to-evict": "false"}}}}}'

    # Change init state.
    if [ $? -eq 0 ]; then
        echo "Update init state..."
        echo $(date '+%Y.%m.%d.%H%M%S' -d '+8 hours') > $init_state_file
        echo "The first deployment is complete."
        echo ""
    fi
fi

# Destroy post-process.
if [ $cdk_exec_result -eq 0 ] && [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/
fi
