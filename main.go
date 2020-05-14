package main

import (
	"encoding/json"
	"fmt"

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

func unmarshalRequest(data string) (*emailInfo, error) {
	req := emailInfo{}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return nil, err
	}
	return &req, nil
}
