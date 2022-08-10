# EKS Simple Cluster

Demonstrate how to create an EKS cluster and manage related addons.

This example will automatically install the following K8s addons:
- vpc-cni
- kube-proxy
- coredns
- ebs-csi-driver
- metrics-server
- cluster-autoscaler
- aws-load-balancer-controller
- external-dns
- node-termination-handler
- aws-xray
- cloudwatch-agent
- fluent-bit-for-aws

## Configuration

You can edit the cdk.json file to modify the deployment configuration.

| Key | Example Value | Description |
| ------ | ------ | ------ |
| stackName | CDKGoExample-EKSCluster | CloudFormation stack name. |
| deploymentRegion | ap-northeast-1 | CloudFormation stack deployment region. If the value is empty, the default is the same as the region where deploy is executed. |
| targetArch | amd64/arm64 | Node archtecture type of EKS Nodegroup. The default EC2 instance size is c5.large/m6.large. |
| clusterName | CDKGoExample-EKSCluster | EKS cluster name. |
| keyPairName | my-key-pair | EC2 instance keypair of EKS Nodegroup. If the value is non-empty, the keypair MUST exist. |
| masterUsers | [Cow, Admin] | Master users in K8s system:masters group. All users listed here must be existing IAM Users. If the value is empty, you have to manually configure the local kubeconfig environment. |
| externalDnsRole | arn:aws:iam::123456789012:role/AWSAccount-EKSExternalDNSRole | IAM role in different AWS account. Cross-account access for K8s External-DNS addon. Please reference to config.go->func ExternalDnsRole for more information. |

## Output

After the deployment is complete, the EKS cluster information will be written to cdk.out/cluster-info.json file:<br />
| Name | Example Value |
| ------ | ------ |
| clusterSecurityGroupId | sg-0cb7ee5b03a23bb74 |
| apiServerEndpoint | https:<span>//AB123D8E12345CD123AA92855957B4F8.gr7.ap-northeast-1.eks.amazonaws.com |
| vpcId | vpc-0445143cc39ee48f6 |
| clusterName | CDKGoExample-EKSCluster |
| certificateAuthorityData | LS0tLS1CRUdJTi...BDRVJUSU0tCg== |
| kubectlRoleArn | arn:aws:iam::123456789012:role/CDKGoExample-EKSCluster-EksClusterCreationRole75AA-1UKOP8JQ8R9DN |
| region | ap-northeast-1 |
| oidcIdpArn | arn:aws:iam::123456789012:oidc-provider/oidc.eks.ap-northeast-1.amazonaws.com/id/AB123D8E12345CD123AA92855957B4F8 |

You can call awseks.Cluster_FromClusterAttributes to import this cluster in other CDK8s projects.
