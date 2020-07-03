# planroom-mailer
This is an AWS Lambda function to separate emailing functionality from planroom-api and send the mail in an asynchronous manner.
It is designed to be triggered by adding an object containing the details of the emails to be sent to an S3 bucket.

## Required Environment Variables
- PE_SERVER   - Url of the smtp server to use
- PE_PORT     - Port to use for connecting to the smtp server
- PE_EMAIL    - Email addreess to use to login to smtp server
- PE_PASSWORD - Password to use for connecting to smtp server
- PE_ERROR    - Email address to notify when emails aren't successful

## Usage
This should be deployed as an AWS Lambda and triggered by adding an object to the S3 bucket containing json in the following format:
```
{
    "jobName": "some job name",
    "expiration": "human readable string indicating when the links expire",
    "message": "message to include in emails. may be blank",
    "recipients": [
        "to": "email address",
        "link": "link for this recipient"
    ] 
}
```

## Deployment
### KMS
- Create a key to use for encrypting environment variables.
### S3
- This should be a bucket for use only with this lambda
- Most options can be left off.  Public access isn't reccomended, and encryption can be turned on if desired.
### Lambda
- The runtime should be Go 1.x
- It should use a role with the AWSBasicLambdaExecutionRole policy and policies allowing GetObject and DeleteObject in S3, and decrypting with the KMS key.
- The S3 bucket should be configured as a trigger and does not need a prefix or suffix.
- Setup required environment variables with KMS encryption.
- 128 MB Memory
- 60 second timeout
- Handler - main
- No X-Ray
- No VPC
- Defaults for concurency, asynchronous invocation, filesystems and database proxies
