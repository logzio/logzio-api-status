# API Status Auto-Deployment

Auto-deployment of Lambda function that collects API status metrics of user API and sends them to Logz.io.

* The Lambda function will be deployed with the layer LogzioLambdaExtensionLogs.
  For more information about the extension [click here](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs).

## Getting Started

To start just press the button and follow the instructions:

| Region           | Deployment                                                                                                                                                                                                                                                                                                                                                               |
|------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `us-east-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-us-east-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `us-east-2`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-2#/stacks/create/template?templateURL=https://logzio-aws-integrations-us-east-2.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `us-west-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-us-west-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `us-west-2`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/create/template?templateURL=https://logzio-aws-integrations-us-west-2.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `eu-central-1`   | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-central-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-eu-central-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)     | 
| `eu-north-1`     | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-north-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-eu-north-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)         | 
| `eu-west-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-eu-west-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `eu-west-2`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-2#/stacks/create/template?templateURL=https://logzio-aws-integrations-eu-west-2.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `eu-west-3`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-3#/stacks/create/template?templateURL=https://logzio-aws-integrations-eu-west-3.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `sa-east-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=sa-east-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-sa-east-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)           | 
| `ap-northeast-1` | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-ap-northeast-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment) | 
| `ap-northeast-2` | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-2#/stacks/create/template?templateURL=https://logzio-aws-integrations-ap-northeast-2.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment) | 
| `ap-northeast-3` | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-3#/stacks/create/template?templateURL=https://logzio-aws-integrations-ap-northeast-3.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment) | 
| `ap-south-1`     | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-south-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-ap-south-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)         | 
| `ap-southeast-1` | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-ap-southeast-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment) | 
| `ap-southeast-2` | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-2#/stacks/create/template?templateURL=https://logzio-aws-integrations-ap-southeast-2.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment) | 
| `ca-central-1`   | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ca-central-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-ca-central-1.s3.amazonaws.com/api-status-auto-deployment/1.1.0/sam-template.yaml&stackName=logzio-api-status-auto-deployment)     |

### Parameters

| Parameter | Description | Required/Optional | Default |
| --- | --- | --- | --- |
| ApiURL | Your API URL to collect status from (for example: https://example.api:1234). | Required | - |
| Method | Your API HTTP request method. Can be `GET` or `POST` | Required | `GET` |
| ApiResponseTimeout | Your API response timeout (seconds). | Required | `10 (seconds)` |
| ExpectedStatusCode | The expected HTTP response status code your API should return. | Required | `200` |
| ExpectedBody | The expected HTTP response body your API should return (leave empty if your API HTTP response body is empty). | Required | ` ` |
| LogzioListener | The Logz.io listener URL for your region. (For more details, see the regions page: https://docs.logz.io/user-guide/accounts/account-region.html) | Required | `https://listener.logz.io` |
| LogzioMetricsToken | Your Logz.io metrics token (Can be retrieved from the Manage Token page). | Required | - |
| LogzioLogsToken | Your Logz.io logs token (Can be retrieved from the Manage Token page). | Required | - |
| SchedulingInterval | The scheduling expression that determines when and how often the Lambda function runs. Rate below 6 minutes will cause the lambda to behave unexpectedly due to cold start and custom resource invocation. | Required | `rate(30 minutes)` |
| Headers | Your API headers separated by comma and each header's key and value are separated by `=` (`header_key_1=header_value_1,header_key_2=header_value_2`). | Optional | - |
| Body | Your API HTTP request body. | Optional | - |
| BearerToken | Your API bearer token. | Optional | - |
| Username | Your API username. | Optional | - |
| Password | Your API password. | Optional | - |

## Searching in Logz.io

All metrics that were sent from the Lambda function will have the prefix `api_status` in their name.


## Changlog

- **1.1.0**: Add geohash to metrics