## AWS OpenSearch with Cognito PoC
Integrate OpenSearch and Cognito, and use Cognito to login to OpenSearch.<br />

## Prerequisites
1. Install and configure AWS CLI environment:<br />
   [Installation] - Installing or updating the latest version of the AWS CLI.<br />
   [Configuration] - Configure basic settings that AWS CLI uses to interact with AWS.<br />
   NOTE: Make sure your IAM User/Role has sufficient permissions.
2. Install Node Version Manager:<br />
   [Install NVM] - Install NVM and configure your environment according to this document.
3. Install Node.js:<br />
    ```sh
    nvm install 16.3.0
    ```
4. Install AWS CDK Toolkit:
    ```sh
    npm install -g aws-cdk
    ```
5. Install Golang:<br />
   [Download and Install] - Download and install Go quickly with the steps described here.
6. Install Docker:<br />
   [Install Docker Engine] - The installation section shows you how to install Docker on a variety of platforms.
7. Make sure you also have GNU Make, jq installed.

## Deployment
Run the following command to deploy AWS infra and code by CDK Toolkit:<br />
  ```sh
  cdk-cli-wrapper-dev.sh deploy
  ```
If all goes well, you will see the following output:<br />
  ```sh
  Stack ARN:
  arn:aws:cloudformation:ap-northeast-1:123456789012:stack/CDKGoExample-OpensearchCognito/28188b10-7ede-11ed-b67d-0ec11c169527
  
  ✨  Total time: 133.05s
  ```
You can also clean up the deployment by running command:<br />
  ```sh
  cdk-cli-wrapper-dev.sh destroy
  ```

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Install NVM]: <https://github.com/nvm-sh/nvm#install--update-script>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
