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
| masterUsers | [Cow, CowAdmin] | Master users in K8s system:masters group. All users listed here must be existing IAM Users. If the value is empty, you have to manually configure the local kubeconfig environment. |
| externalDnsRole | arn:aws:iam::123456789012:role/AWSIsengardAccount-EKSExternalDNSRole | IAM role in different AWS account. Cross-account access for K8s External-DNS addon. Please reference to config.go->func ExternalDnsRole for more information. |
