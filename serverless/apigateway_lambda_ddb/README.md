## AWS Serverless PoC (API Gateway + Lambda + DynamoDB)
We use the classic AWS serverless architecture to demonstrate how to build a simple chat room service.<br />

## Prerequisite
1. Install and configure AWS CLI environment:<br />
   [Installation] - Installing or updating the latest version of the AWS CLI.<br />
   [Configuration] - Configure basic settings that AWS CLI uses to interact with AWS.
2. Install AWS CDK Toolkit:
    ```sh
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash
    nvm install 16.3.0
    npm install -g aws-cdk
    ```
3. Install Golang:<br />
   [Download and Install] - Download and install Go quickly with the steps described here.
4. Install Docker:<br />
   [Install Docker Engine] - Find the corresponding platform and install Docker.
5. Make sure you also have GNU Make installed.

## Deployment
Run the following command to deploy AWS infra and code using CDK Toolkit:<br />
  ```sh
  cdk-cli-wrapper-dev.sh deploy
  ```
If all goes well, you will see the following output:<br />
  ```sh
  Outputs:
  CdkGolangExample-ApiGtwLambdaDdb.LambdaRestApiEndpointCCECE4C1 = https://b12gqp2av5.execute-api.ap-northeast-2.amazonaws.com/dev/
  
  ✨  Total time: 133.05s
  ```
You can also clean up the deployment by running command:<br />
  ```sh
  cdk-cli-wrapper-dev.sh destroy
  ```

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
