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
