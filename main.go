package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(s3Event events.S3Event) error {

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
			return err
		}

		S3String := string(buf.Bytes())

		//Get Object to a struct
		var request *emailInfo
		if request, err = unmarshalRequest(S3String); err != nil {
			return err
		}

		//Send the emails
		if err = sendMail(request); err != nil {
			return err
		}

		//Log errors and try to email someone on error. maybe use multiple error plugin thing

		//Delete the object from S3
	}
	return nil
}

func sendMail(data *emailInfo) error {
	server := os.Getenv("PE_SERVER")
	port, err := strconv.Atoi(os.Getenv("PE_PORT"))
	if err != nil {
		return err
	}
	email := os.Getenv("PE_EMAIL")
	password := os.Getenv("PE_PASSWORD")

	d := gomail.NewDialer(server, port, email, password)
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()
	m := gomail.NewMessage()
	for _, e := range data.Recipients {
		m.SetAddressHeader("From", email, "Benchmark Planroom")
		m.SetHeader("To", e.To)
		m.SetHeader("Subject", buildSubject(data.JobName))
		m.AddAlternative("text/plain", buildAltMessage(data.JobName, data.Expiration, data.Message, e.Link))
		m.AddAlternative("text/html", buildMessage(data.JobName, data.Expiration, data.Message, e.Link))

		if err := gomail.Send(s, m); err != nil {
			fmt.Printf("Could not send email to %q: %v", e.To, err)
		}
		m.Reset()
	}
	return nil
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
