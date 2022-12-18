## AWS Serverless PoC (API Gateway + Lambda + DynamoDB)
We use the classic AWS serverless architecture to demonstrate how to build a simple chat room service.<br />

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
  Outputs:
  CdkGolangExample-ApiGtwLambdaDdb.LambdaRestApiEndpointCCECE4C1 = https://b12gqp2av5.execute-api.ap-northeast-2.amazonaws.com/dev/
  Stack ARN:
  arn:aws:cloudformation:ap-northeast-2:123456789012:stack/CdkGolangExample-ApiGtwLambdaDdb/225b9050-a414-11ec-b5c2-0ab842e4df54
  
  ✨  Total time: 133.05s
  ```
You can also clean up the deployment by running command:<br />
  ```sh
  cdk-cli-wrapper-dev.sh destroy
  ```

## Testing
As you see in the output of the Deployment step, the URL is your API Gateway endpoint:<br />
  ```sh
  https://b12gqp2av5.execute-api.ap-northeast-2.amazonaws.com/dev/
  ```
We have integrated two Lambda functions with the following resource paths:
  ```sh
  put-chat-records
  get-chat-records
  ```
You can POST user comment by following API:
  ```sh
  POST https://b12gqp2av5.execute-api.ap-northeast-2.amazonaws.com/dev/put-chat-records
  Content-Type: application/json
  x-api-key: dI65dhFd3742OmUhbdxYo4CT2eOwfoUT1FCtm8ml
  Body:
  {
    "name"    : string,
    "comment" : string,
    "chatRoom": string
  }
  Status Code: 201 Created
  ```
Or you can QUERY user comments by following API:
  ```sh
  GET https://b12gqp2av5.execute-api.ap-northeast-2.amazonaws.com/dev/get-chat-records?chatroom=abc123
  x-api-key: dI65dhFd3742OmUhbdxYo4CT2eOwfoUT1FCtm8ml
  Status Code: 200 OK
  [
    {
      "name"   :string,
      "comment":string,
      "time"   :string
    },
    ...
  ]
  ```

## Development
In your day-to-day development work, running Lambda functions locally can improve productivity.<br />
All scripting tools related to Lambda functions are in the 'functions' directory.<br />
Run the following command to build a Docker image and run the container:<br />
  ```sh
  ./run_local_test.sh
  Script for running RIE in local.
  
  Usage:
  
        run_local_test.sh <lambda_function_name>
        
  Examples:
  
        run_local_test.sh put-chat-records
  
  ```
Keep run_local_test.sh running and open another terminal, run the test script:<br />
  ```sh
  ./do_local_test.sh
  Script for testing lambda function in local.
  
  Usage:
  
        do_local_test.sh <json body>
        
  Examples:
  
        do_local_test.sh '{"body":"{\"name\":\"Cow\",\"comment\":\"Sample comment!\",\"chatRoom\":\"101\"}"}'
        
        do_local_test.sh @put-chat-records/sample_body.json
        
        do_local_test.sh '{"queryStringParameters":{"chatroom":"101"}}'
        
        do_local_test.sh @get-chat-records/sample_query_string.json
  
  ```
The first two examples are for put-chat-records function, and the last two examples are for get-chat-records function.<br />
When you are done modifying the Lambda function code, you can run the following command again:<br />
  ```sh
  cdk-cli-wrapper-dev.sh deploy
  ```

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Install NVM]: <https://github.com/nvm-sh/nvm#install--update-script>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
