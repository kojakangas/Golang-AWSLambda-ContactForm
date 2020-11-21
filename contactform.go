package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	//go get -u github.com/aws/aws-sdk-go

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "kojakangas@drury.edu"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	Recipient = "kojakangas@drury.edu"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	Subject = "Portfolio Inquiry"

	// The character encoding for the email.
	CharSet = "UTF-8"
)

// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	// Double check it's a post request being made
	// if request.HTTPMethod != http.MethodPost {
	// 	return events.APIGatewayProxyResponse{
	// 		Body:       "Invalid HTTP Method",
	// 		StatusCode: 502,
	// 	}, nil
	// }

	r := http.Request{}
	r.Header = make(map[string][]string)
	for k, v := range request.Headers {
		if k == "content-type" || k == "Content-Type" {
			r.Header.Set(k, v)
		}
	}
	// NOTE: API Gateway is set up with */* as binary media type, so all APIGatewayProxyRequests will be base64 encoded
	body, err := base64.StdEncoding.DecodeString(request.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Could not read response body",
			StatusCode: 502,
		}, nil
	}

	// err = r.ParseMultipartForm(maxRequestBodySize)

	var email = r.FormValue("email")
	var name = r.FormValue("name")
	var content = r.FormValue("content")

	// The HTML body for the email.
	var HtmlBody = "<h1>Inquiry from: " + name + "</h1><p>Email: " + email +
		"</p><p>" + content + "</p>"

	//The email body for recipients with non-HTML email clients.
	var TextBody = "This is an inquiry from " + name + " Email at: " + email +
		" " + content

	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return events.APIGatewayProxyResponse{
			Body:       "Unknown Failure",
			StatusCode: 502,
		}, err
	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)

	return events.APIGatewayProxyResponse{
		Body:       "Hello " + r.FormValue("name"),
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
