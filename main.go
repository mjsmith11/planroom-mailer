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
	var request *emailInfo
	var err error

	//Get Object from S3
	S3String := "fake Data"
	//Get Object to a struct
	if request, err = unmarshalRequest(S3String); err != nil {
		return err
	}
	fmt.Println(request) // temporarily print request so go will compile
	//Send the emails

	//Log errors and try to email someone on error

	//Delete the object from S3

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
	return nil
}
func buildMessage(name, expiration, message, link string) string {
	/*		if (($message == '') || (ctype_space($message))) {
			$body = '<center>
	<img src="https://benchmarkmechanical.com/Images/logo1.jpg" />
	<br><br><br>
	<div style="width:60%;border:1px solid lightgrey">
		<h1>Invitation to Bid</h1>
		<h2>' . $job['name'] . '</h2>
		<a href="' . $link . '">Click Here</a> to access bidding documents and project details.<br>This link will expire ' . $expStr . '.
		<br><br><br>
		<span style="color:grey;font-size:10pt"><em>Please do not reply to this email. The mailbox is not monitored.</em></span>
	</div>
</center>';
		} else {
			$body = '<center>
	<img src="https://benchmarkmechanical.com/Images/logo1.jpg" />
	<br><br><br>
	<div style="width:60%;border:1px solid lightgrey">
		<h1>Invitation to Bid</h1>
		<h2>' . $job['name'] . '</h2>
		<div style="width:70%">'
            . $message .
        '</div>
        <br>
		<a href="' . $link . '">Click Here</a> to access bidding documents and project details.<br>This link will expire ' . $expStr . '.
		<br><br><br>
		<span style="color:grey;font-size:10pt"><em>Please do not reply to this email. The mailbox is not monitored.</em></span>
	</div>
</center>';
		}
		return $body;*/
		return ""
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
