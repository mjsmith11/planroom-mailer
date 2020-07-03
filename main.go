package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"encoding/base64"

	"gopkg.in/gomail.v2"
	"github.com/hashicorp/go-multierror"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/kms"
)

// struct containing data for all the emails being processed
type emailInfo struct {
	JobName    string          `json:"jobName"`
	Expiration string          `json:"expiration"`
	Message    string          `json:"message"`
	Recipients []recipientInfo `json:"recipients"`
}

// struct containing data specific to one individual email being processed
type recipientInfo struct {
	To   string `json:"to"`
	Link string `json:"link"`
}
var functionName string = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
var decryptedServer string
var decryptedPort string
var decryptedEmail string
var decryptedPassword string
var decryptedError string

func init() {
	decryptedServer = decrypt(os.Getenv("PE_SERVER"))
	decryptedPort = decrypt(os.Getenv("PE_PORT"))
	decryptedEmail = decrypt(os.Getenv("PE_EMAIL"))
	decryptedPassword= decrypt(os.Getenv("PE_PASSWORD"))
	decryptedError = decrypt(os.Getenv("PE_ERROR"))
}

func decrypt(encrypted string) string {
		kmsClient := kms.New(session.New())
		decodedBytes, err := base64.StdEncoding.DecodeString(encrypted)
		if err != nil {
			panic(err)
		}
		input := &kms.DecryptInput{
			CiphertextBlob: decodedBytes,
			EncryptionContext: aws.StringMap(map[string]string{
				"LambdaFunctionName": functionName,
			}),
		}
		response, err := kmsClient.Decrypt(input)
		if err != nil {
			panic(err)
		}
		// Plaintext is a byte array, so convert to string
		return string(response.Plaintext[:])
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(s3Event events.S3Event) error {
	var allErrors error
	for _, record := range s3Event.Records {
		var err error
		recordS3 := record.S3

		//Get Object from S3
		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String(record.AWSRegion)},
		)
	
		downloader := s3manager.NewDownloader(sess)
		
		buf := aws.NewWriteAtBuffer([]byte{})

		_, err = downloader.Download(buf,
			&s3.GetObjectInput{
				Bucket: aws.String(recordS3.Bucket.Name),
				Key:    aws.String(recordS3.Object.Key),
			})
		if err != nil {
			fmt.Printf("Error downloading S3 Object %v", err)
			allErrors = multierror.Append(allErrors,err)
			sendError(err)
			continue
		}

		S3String := string(buf.Bytes())

		//Get Object to a struct
		var request *emailInfo
		if request, err = unmarshalRequest(S3String); err != nil {
			fmt.Printf("Error Unmarshalling request %v", err)
			allErrors = multierror.Append(allErrors,err)
			sendError(err)
			continue
		}

		//Send the emails
		if err = sendMail(request); err != nil {
			fmt.Printf("Sending emails %v", err)
			allErrors = multierror.Append(allErrors,err)
			sendError(err)
			continue
		}

		//Delete the object from S3
		if err == nil {
			// if it's made it this far without erroring, it's safe to delete the S3 object
			svc := s3.New(sess)
			_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(recordS3.Bucket.Name), Key: aws.String(recordS3.Object.Key)})
			if err != nil {
				fmt.Printf("Error downloading S3 Object %v", err)
				allErrors = multierror.Append(allErrors,err)
				sendError(err)
				continue
			}
		}
	}
	return allErrors
}

func sendError(errorToSend error) {
	server := decryptedServer
	port, err := strconv.Atoi(decryptedPort)
	if err != nil {
		fmt.Printf("Failed to send error notification %v",err)
	}
	email := decryptedEmail
	password := decryptedPassword
	to := decryptedError

	d := gomail.NewDialer(server, port, email, password)
	s, err := d.Dial()
	if err != nil {
		fmt.Printf("Failed to send error notification %v",err)
	}
	defer s.Close()
	m := gomail.NewMessage()

	m.SetAddressHeader("From", email, "Benchmark Planroom")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Planroom Email Error")
	m.AddAlternative("text/plain", fmt.Sprintf("There was an error sending emails %v", errorToSend))
	m.AddAlternative("text/html", fmt.Sprintf("There was an error sending emails %v", errorToSend))

	if err := gomail.Send(s, m); err != nil {
		fmt.Printf("Failed to send error notification %v",err)
	}
	m.Reset()
	
}

func sendMail(data *emailInfo) error {
	server := decryptedServer
	port, err := strconv.Atoi(decryptedPort)
	if err != nil {
		return err
	}
	email := decryptedEmail
	password := decryptedPassword

	d := gomail.NewDialer(server, port, email, password)
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()
	m := gomail.NewMessage()
	var messageErrors error
	for _, e := range data.Recipients {
		m.SetAddressHeader("From", email, "Benchmark Planroom")
		m.SetHeader("To", e.To)
		m.SetHeader("Subject", buildSubject(data.JobName))
		m.AddAlternative("text/plain", buildAltMessage(data.JobName, data.Expiration, data.Message, e.Link))
		m.AddAlternative("text/html", buildMessage(data.JobName, data.Expiration, data.Message, e.Link))

		if err := gomail.Send(s, m); err != nil {
			messageErrors = multierror.Append(messageErrors, err)
		} else {
			fmt.Printf("%s email sent to %s",data.JobName,e.To)
		}
		m.Reset()
	}
	return messageErrors
}
func buildSubject(jobName string) string {
	return "Invitation to Bid: " + jobName
}

func buildMessage(name, expiration, message, link string) string {
	var body string
	if strings.TrimSpace(message) == "" {
		body = "<center>"
		body += "	<img src=\"https://benchmarkmechanical.com/Images/logo1.jpg\" />"
		body += "	<br><br><br>"
		body += "	<div style=\"width:60%;border:1px solid lightgrey\">"
		body += "		<h1>Invitation to Bid</h1>"
		body += fmt.Sprintf("		<h2>%s</h2>",name)
		body += fmt.Sprintf("		<a href=\"%s\">Click Here</a> to access bidding documents and project details.<br>This link will expire %s.",link,expiration)
		body += "		<br><br><br>"
		body += "		<span style=\"color:grey;font-size:10pt\"><em>Please do not reply to this email. The mailbox is not monitored.</em></span>"
		body += "	</div>"
		body += "</center>"
	} else {
		body = "<center>"
		body += "	<img src=\"https://benchmarkmechanical.com/Images/logo1.jpg\" />"
		body += "	<br><br><br>"
		body += "	<div style=\"width:60%;border:1px solid lightgrey\">"
		body += "		<h1>Invitation to Bid</h1>"
		body += fmt.Sprintf("		<h2>%s</h2>",name)
		body += "		<div style=\"width:70%\">"
		body += message
		body += "		</div><br>"
		body += fmt.Sprintf("		<a href=\"%s\">Click Here</a> to access bidding documents and project details.<br>This link will expire %s.",link,expiration)
		body += "		<br><br><br>"
		body += "		<span style=\"color:grey;font-size:10pt\"><em>Please do not reply to this email. The mailbox is not monitored.</em></span>"
		body += "	</div>"
		body += "</center>"
	}
	return body
}

func buildAltMessage(name, expiration, message, link string) string {
	body := fmt.Sprintf("This is an invitation from Benchmark Mechanical to bid on the %s project. Bidding documents ", name)
	body += fmt.Sprintf("and project details are available at the link below. The link will expire %s", expiration)
	if strings.TrimSpace(message) != "" {
		body += "\n\n"
		body += message
		body += "\n\n"
	}
	body += "\n\n"
	body += link
	body += "\n\n"
	body += "Please do not reply to this email. The mailbox is not monitored"
	return body
}

func unmarshalRequest(data string) (*emailInfo, error) {
	req := emailInfo{}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return nil, err
	}
	return &req, nil
}


