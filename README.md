# API Status Auto-Deployment

Auto-deployment of Lambda function that collects API status metrics of user API and sends them to Logz.io.

* The Lambda function will be deployed with the layer LogzioLambdaExtensionLogs.
  For more information about the extension [click here](https://github.com/logzio/logzio-lambda-extensions/tree/main/logzio-lambda-extensions-logs).

## Getting Started

To start just press the button and follow the instructions:

[![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/template?templateURL=https://logzio-aws-integrations-us-east-1.s3.amazonaws.com/api-status-auto-deployment/auto-deployment.yaml&stackName=logzio-api-status-auto-deployment)

### Parameters

| Parameter | Description | Required/Optional | Default |
| --- | --- | --- | --- |
| ApiURL | Your API URL to collect status from. | Required | - |
| Method | Your API HTTP request method. Can be `GET` or `POST` | Required | `GET` |
| ApiResponseTimeout | Your API response timeout (seconds). | Required | `10 (seconds)` |
| ExpectedStatusCode | The expected HTTP response status code your API should return. | Required | `200` |
| ExpectedBody | The expected HTTP response body your API should return (leave empty if your API HTTP response body is empty). | Required | ` ` |
| LogzioListener | The Logz.io listener URL for your region. (For more details, see the regions page: https://docs.logz.io/user-guide/accounts/account-region.html) | Required | `https://listener.logz.io` |
| LogzioMetricsToken | Your Logz.io metrics token (Can be retrieved from the Manage Token page). | Required | - |
| LogzioLogsToken | Your Logz.io logs token (Can be retrieved from the Manage Token page). | Required | - |
| SchedulingInterval | The scheduling expression that determines when and how often the Lambda function runs. | Required | `rate(30 minutes)` |
| Headers | Your API headers separated by comma and each header's key and value are separated by `=`. | Optional | - |
| Body | Your API HTTP request body. | Optional | - |
| BearerToken | Your API bearer token. | Optional | - |
| Username | Your API username. | Optional | - |
| Password | Your API password. | Optional | - |

## Searching in Logz.io

All metrics that were sent from the Lambda function will have the prefix `api_status` in their name. 