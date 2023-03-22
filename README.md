# Birthday Message Automation

This project is a simple and efficient way to automate sending happy birthday messages using Golang, AWS Lambda, and AWS EventBridge. It supports sending birthday messages through Twilio (SMS) and Discord.

## Table of Contents

1. [Features](#features)
2. [Prerequisites](#prerequisites)
3. [Setup and Deployment](#setup-and-deployment)
4. [Usage](#usage)
5. [Contributing](#contributing)
6. [License](#license)
7. [Support](#support)
8. [Acknowledgments](#acknowledgments)

## Features

* Automatically sends happy birthday messages on the specified date.
* Supports sending messages through Twilio (SMS) and Discord.
* Securely stores API secrets in AWS Secrets Manager.
* Easily customizable by editing a CSV file with birthday information.

## Prerequisites

To use this project, you'll need:

* An AWS account with appropriate permissions to create Lambda functions, EventBridge rules, and Secrets Manager secrets.
* Twilio account with SMS capabilities and a Twilio phone number.
* Discord account with the ability to create webhooks (if using Discord for messaging).
* [Go](https://golang.org/dl/) installed on your local machine.

## Setup and Deployment

1. Clone this repository to your local machine.
```
git clone https://github.com/sahelanthropus/Birthday-Message-Automation.git
cd birthday-message-automation
```

2. Fill in the `birthdays.csv` file with relevant birthday information:

   * Format the CSV file with the following headers: `Date,Name,PhoneNumber,Message,Discord`
   * Each row should contain the contact's name, phone number (for Twilio), custom message if desired, and something in the Discord field (if using Discord).

3. Build the Lambda function package according to your system:
(refrence: https://github.com/aws/aws-lambda-go#building-your-function)
```PowerShell
$env:GOOS = "linux"
go build -o bootstrap main.go
~\Go\Bin\build-lambda-zip.exe -o lambda-handler.zip bootstrap
Compress-Archive -Path ".\birthdays.csv" -DestinationPath ".\lambda-handler.zip" -Update
```
4. Create a new Lambda function in the AWS Management Console, and upload the `lambda-handler.zip` file as the function package. Set the runtime to `Go 1.x` and the handler to `bootstrap`.

5. Add the necessary environment variables to the Lambda function:

   * `REGION`: The AWS region where the Lambda function is deployed (e.g., `us-east-1`).
   * `DISCORD_FRIENDZONE_CHANNEL`: The link to the Discord channel's webhook url for posting messages (e.g., `https://discord.com/api/v9/channels/123456789/messages`)
   * `DISCORD_SECRET_NAME`: The Secret name for your Discord bot authorization token.
   * `TWILIO_SECRET_NAME`: The Secret name for your Twilio account SID and authorization token.
   * `TWILIO_PHONE_NUMBER`: The from phone number used with the Twilio API. 


6. Create the new secrets in AWS Secrets Manager to store your Twilio and Discord API keys (Note that after the 30-day free trial for AWS Secrets Manager, AWS charges $0.40 per secret stored in AWS Secrets Manager):

   * Secret name (rename if desired): `prod\birthday\twilio`
   * Secret key-value pairs:
     * `AccountSID`: Your Twilio account SID.
     * `AuthToken`: Your Twilio auth token.
     
   * Secret name (rename if desired): `prod\birthday\discord`
   * Secret key-value pair:
     * `authorization`: Your Discord bot authorization token (if using Discord for messaging).

7. Update the IAM role associated with the Lambda function to include the necessary permissions to access Secrets Manager. Add the `secretsmanager:GetSecretValue` permission for the specific ARN of the two secrets:

   * In the AWS Management Console, navigate to the IAM service.
   * Find the role associated with your Lambda function and click on it.
   * Click "Add inline policy."
   * Select the "Secrets Manager" service.
   * Choose the "GetSecretValue" action.
   * In the "Resources" section, specify the ARNs for both secrets.
   * Review and create the policy.

8. Create a new CloudWatch EventBridge rule to trigger the Lambda function daily at a specific time:

   * Rule type: Schedule expression
   * Schedule expression: `cron(0 <minute> <hour> * * ? *)` (replace `<minute>` and `<hour>` with the desired time in UTC).
   * Target: Your Lambda function.

9. Test the Lambda function to ensure it's working correctly.

## Usage

To add or modify birthday messages, simply update the `birthdays.csv` file with the relevant contact information and re-zip and upload it to the Lambda function.

## Contributing

I welcome contributions to this project! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch with a descriptive name.
3. Make your changes and commit them to your branch.
4. Create a pull request targeting the main branch in the original repository.

When submitting a pull request, please provide a clear and concise description of your changes and any relevant issue numbers.

## License

This project is licensed under the [MIT License](LICENSE). By contributing to this project, you agree to abide by the terms of the license.

## Support

If you encounter any issues or have questions about this project, please open a new issue in the GitHub repository. I'll do my best to help you out.

## Acknowledgments

Thank you to [OpenAI](https://www.openai.com/) for providing guidance and suggestions on best practices and code improvements and to [Travis Media](https://travis.media/) for the project idea.
